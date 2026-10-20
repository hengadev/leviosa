// Black-box tests for the backend's `verify` run mode (issue 015): inside a
// read-only transaction it decrypts one auth.users record with the configured
// KMS key and checks the pepper canary (encryption.canary) — written by serve
// at startup — against the configured pepper, so a restored dump is proven
// readable with the environment's KMS key and pepper.
//
// The app binary is built once and driven as a real process (migrate,
// bootstrap-admin, serve, verify) against testcontainers Postgres, Redis and
// LocalStack KMS (ADR-0005: real adapters, no mocking).
package verify_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hengadev/encx"

	"github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver
)

// kekAlias is the KMS key the databases below are encrypted with; otherKEKAlias
// is a second key on the same LocalStack, standing in for another
// environment's key.
const (
	kekAlias      = "alias/leviosa-verify-test"
	otherKEKAlias = "alias/leviosa-verify-test-other"
)

var (
	appBinary   string
	pgContainer *testutils.PostgresContainer
	redisCont   *testutils.RedisContainer
	kmsCont     *testutils.KMSContainer

	// pepper is the pepper the databases below are written with.
	pepper string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	binDir, err := os.MkdirTemp("", "leviosa-verify-test-")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(binDir)

	appBinary = filepath.Join(binDir, "leviosa")
	build := exec.Command("go", "build", "-o", appBinary, "./cmd/app")
	// The test binary runs from backend/test/integration/verify.
	build.Dir = filepath.Join("..", "..", "..")
	if out, err := build.CombinedOutput(); err != nil {
		log.Fatalf("Failed to build cmd/app: %v\n%s", err, out)
	}

	pgContainer, err = testutils.SetupPostgres(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup postgres container: %v", err)
	}

	redisCont, err = testutils.SetupRedis(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup redis container: %v", err)
	}

	kmsCont, err = testutils.SetupKMS(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup LocalStack KMS container: %v", err)
	}
	for _, alias := range []string{kekAlias, otherKEKAlias} {
		if _, err := kmsCont.CreateKey(ctx, alias); err != nil {
			log.Fatalf("Failed to create KMS key %s: %v", alias, err)
		}
	}

	pepper = newPepper()

	code := m.Run()

	_ = pgContainer.Terminate(ctx)
	_ = redisCont.Terminate(ctx)
	testutils.TeardownKMS(ctx, nil, kmsCont)
	os.Exit(code)
}

// newPepper returns a fresh base64 32-byte pepper, as an operator would put in Infisical.
func newPepper() string {
	b := make([]byte, encx.PepperLength)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// createTestDB creates a fresh, empty database on the shared Postgres container
// and returns its DSN.
func createTestDB(t *testing.T, name string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := sql.Open("pgx", pgContainer.ConnectionString)
	if err != nil {
		t.Fatalf("Failed to open admin connection: %v", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", name)); err != nil {
		t.Fatalf("Failed to drop database %s: %v", name, err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name)); err != nil {
		t.Fatalf("Failed to create database %s: %v", name, err)
	}

	dsn := strings.Replace(pgContainer.ConnectionString, "dbname=testdb", "dbname="+name, 1)
	if dsn == pgContainer.ConnectionString {
		t.Fatalf("DSN did not contain dbname=testdb: %s", pgContainer.ConnectionString)
	}
	return dsn
}

// openDB opens a pinged database/sql connection to dsn, closed on test end.
func openDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// appEnv is the environment every mode below runs with: development mode, the
// database, and encx on alias with pepperValue. extra adds or overrides values.
func appEnv(dsn, alias, pepperValue string, extra map[string]string) []string {
	enc := kmsCont.EncryptionConfig(alias, pepperValue)
	env := append(os.Environ(),
		"ENVIRONMENT=development",
		"DATABASE_URL="+dsn,
		"REDIS_ADDR="+redisCont.Host+":"+redisCont.Port,
		"ENCX_KMS_KEY_ALIAS="+enc.KMSKeyAlias,
		"ENCX_PEPPER="+enc.Pepper,
		"AWS_REGION="+enc.AWSRegion,
		"AWS_ACCESS_KEY_ID="+enc.AWSAccessKeyID,
		"AWS_SECRET_ACCESS_KEY="+enc.AWSSecretAccessKey,
		"AWS_ENDPOINT_URL="+enc.AWSEndpointURL,
	)
	for k, v := range extra {
		env = append(env, k+"="+v) // os/exec keeps the last value of a duplicate key
	}
	return env
}

// envFor is appEnv with the suite's key and pepper.
func envFor(dsn string) []string {
	return appEnv(dsn, kekAlias, pepper, nil)
}

// appRun is the result of running the binary until it exits.
type appRun struct {
	exitCode int
	output   string
}

// runApp runs the binary in the given mode and waits for it to exit.
func runApp(t *testing.T, mode string, env []string) appRun {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, appBinary, mode)
	cmd.Env = env
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("%s mode timed out; output so far:\n%s", mode, out.String())
	}

	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("Failed to run %s mode: %v", mode, err)
		}
		code = exitErr.ExitCode()
	}
	return appRun{exitCode: code, output: out.String()}
}

// mustRun runs mode and fails the test unless it exits 0.
func mustRun(t *testing.T, mode string, env []string) {
	t.Helper()

	if run := runApp(t, mode, env); run.exitCode != 0 {
		t.Fatalf("%s exited %d; output:\n%s", mode, run.exitCode, run.output)
	}
}

// freePort reserves an unused loopback port for a serve process to listen on.
func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to reserve a port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// serveOnce boots serve with env until /health answers 200, then stops it with
// SIGINT and waits for a clean exit — the way a deploy's first start writes the
// pepper canary.
func serveOnce(t *testing.T, env []string) {
	t.Helper()

	port := freePort(t)
	cmd := exec.Command(appBinary)
	cmd.Env = append(env, "SERVER_PORT="+strconv.Itoa(port))
	output := &strings.Builder{}
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start serve: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	deadline := time.After(120 * time.Second)
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	healthy := false
	for !healthy {
		select {
		case <-deadline:
			_ = cmd.Process.Kill()
			t.Fatalf("serve did not become healthy in time; output:\n%s", output.String())
		case err := <-done:
			t.Fatalf("serve exited before becoming healthy (err=%v); output:\n%s", err, output.String())
		case <-tick.C:
			resp, err := http.Get(healthURL)
			if err != nil {
				continue
			}
			resp.Body.Close()
			healthy = resp.StatusCode == http.StatusOK
		}
	}

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to signal serve: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve did not exit cleanly (err=%v); output:\n%s", err, output.String())
		}
	case <-time.After(60 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatalf("serve did not shut down in time; output:\n%s", output.String())
	}
}
