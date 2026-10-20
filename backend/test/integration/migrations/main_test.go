// Black-box tests for the backend's migration run mode and the startup schema
// check (issue 012): the `migrate` mode applies the embedded goose migrations
// as a one-off, and `serve` refuses to start against a schema older than the
// embedded set while accepting an equal or newer one (a rollback).
//
// The app binary is built once and driven as a real process against
// testcontainers Postgres/Redis/LocalStack KMS (ADR-0005: real adapters, no mocking) —
// the same dependency shape `make local-dev` boots.
package migrations_test

import (
	"context"
	"database/sql"
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

	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
	"github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver
)

var (
	appBinary   string
	pgContainer *testutils.PostgresContainer
	redisCont   *testutils.RedisContainer
	kmsCont     *testutils.KMSContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Build the real binary once; the tests below run it in its modes.
	binDir, err := os.MkdirTemp("", "leviosa-runmode-test-")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(binDir)

	appBinary = filepath.Join(binDir, "leviosa")
	build := exec.Command("go", "build", "-o", appBinary, "./cmd/app")
	// The test binary runs from backend/test/integration/migrations.
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
	if _, err := kmsCont.CreateKey(ctx, testKEKAlias); err != nil {
		log.Fatalf("Failed to create the KEK: %v", err)
	}

	code := m.Run()

	_ = pgContainer.Terminate(ctx)
	_ = redisCont.Terminate(ctx)
	testutils.TeardownKMS(ctx, nil, kmsCont)
	os.Exit(code)
}

// createTestDB creates a fresh database on the shared Postgres container and
// returns its DSN, so every test starts from an empty schema.
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

// simulateNewerSchema records a migration version above the newest embedded
// one, as if a newer release had migrated the database, and returns it.
func simulateNewerSchema(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	newer := migrations.NewestVersion() + 1
	if _, err := db.Exec(
		"INSERT INTO goose_db_version (version_id, is_applied) VALUES ($1, true)", newer,
	); err != nil {
		t.Fatalf("simulate a newer schema version: %v", err)
	}
	return newer
}

// baseEnv is the minimal environment the binary needs: everything else keeps
// its defaults (development mode, no RabbitMQ, integrations off).
func baseEnv(dsn string, extra map[string]string) []string {
	env := append(os.Environ(),
		"ENVIRONMENT=development",
		"DATABASE_URL="+dsn,
	)
	for k, v := range extra {
		env = append(env, k+"="+v)
	}
	return env
}

// testKEKAlias and testPepper are the encx settings serve runs with: a KEK
// created on the LocalStack KMS container and a fixed base64 32-byte pepper.
const (
	testKEKAlias = "alias/leviosa-migrations-test"
	testPepper   = "bGV2aW9zYS1taWdyYXRpb25zLXRlc3QtcGVwcGVyMzI="
)

// serveEnv is what serve needs to get past infrastructure setup: Postgres,
// Redis and the encx KEK on LocalStack KMS with a pepper.
func serveEnv(dsn string, port int) []string {
	enc := kmsCont.EncryptionConfig(testKEKAlias, testPepper)
	return baseEnv(dsn, map[string]string{
		"SERVER_PORT":           strconv.Itoa(port),
		"REDIS_ADDR":            redisCont.Host + ":" + redisCont.Port,
		"ENCX_KMS_KEY_ALIAS":    enc.KMSKeyAlias,
		"ENCX_PEPPER":           enc.Pepper,
		"AWS_REGION":            enc.AWSRegion,
		"AWS_ACCESS_KEY_ID":     enc.AWSAccessKeyID,
		"AWS_SECRET_ACCESS_KEY": enc.AWSSecretAccessKey,
		"AWS_ENDPOINT_URL":      enc.AWSEndpointURL,
	})
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

// serveProc is a serve process that has become healthy.
type serveProc struct {
	cmd    *exec.Cmd
	done   chan error // receives cmd.Wait's result exactly once
	output *strings.Builder
}

// startServe boots the serve mode and returns once /health answers 200.
func startServe(t *testing.T, dsn string, port int) *serveProc {
	t.Helper()

	proc := &serveProc{
		cmd:    exec.Command(appBinary),
		done:   make(chan error, 1),
		output: &strings.Builder{},
	}
	proc.cmd.Env = serveEnv(dsn, port)
	proc.cmd.Stdout = proc.output
	proc.cmd.Stderr = proc.output

	if err := proc.cmd.Start(); err != nil {
		t.Fatalf("Failed to start serve: %v", err)
	}
	go func() { proc.done <- proc.cmd.Wait() }()

	healthURL := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	deadline := time.After(120 * time.Second)
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-deadline:
			_ = proc.cmd.Process.Kill()
			t.Fatalf("serve did not become healthy in time; output:\n%s", proc.output.String())
		case err := <-proc.done:
			t.Fatalf("serve exited before becoming healthy (err=%v); output:\n%s", err, proc.output.String())
		case <-tick.C:
			resp, err := http.Get(healthURL)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return proc
			}
		}
	}
}

// stop shuts serve down with SIGTERM and reports once it has exited.
func (p *serveProc) stop(t *testing.T) {
	t.Helper()

	if err := p.cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("Failed to signal serve: %v", err)
	}
	select {
	case <-p.done:
	case <-time.After(60 * time.Second):
		_ = p.cmd.Process.Kill()
		t.Fatalf("serve did not shut down in time; output:\n%s", p.output.String())
	}
}
