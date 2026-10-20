package bootstrap_test

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
	userService "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/user"
	userRepository "github.com/Leviosa-care/leviosa/backend/internal/authuser/infrastructure/postgres/user"
	"github.com/Leviosa-care/leviosa/backend/internal/authuser/ports"

	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

var (
	pgContainer *tu.PostgresContainer
	testPool    *pgxpool.Pool
	crypto      encx.CryptoService
	repo        ports.UserRepository
	svc         *bootstrap.Service
	users       ports.UserService
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Postgres container + pool + migrations (the bootstrap-admin mode itself
	// never migrates; tests apply them like every other integration suite).
	pgContainer, err := tu.SetupPostgres(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer tu.TeardownPostgres(ctx, nil, pgContainer)

	poolCtx, poolCancel := context.WithTimeout(ctx, 30*time.Second)
	defer poolCancel()

	pgCfg, err := pgxpool.ParseConfig(pgContainer.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to parse pgxpool config: %v", err)
	}
	pgCfg.MaxConns = 5
	pgCfg.MinConns = 1

	testPool, err = pgxpool.NewWithConfig(poolCtx, pgCfg)
	if err != nil {
		log.Fatalf("Failed to open test database pool: %v", err)
	}
	if err = testPool.Ping(poolCtx); err != nil {
		log.Fatalf("Failed to ping database pool: %v", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err = goose.SetDialect("pgx"); err != nil {
		log.Fatalf("Setting dialect for migrations: %v", err)
	}
	gooseDB, err := sql.Open("pgx", testPool.Config().ConnString())
	if err != nil {
		log.Fatalf("Failed to open temp *sql.DB for goose migrations: %v", err)
	}
	defer gooseDB.Close()
	if err = goose.UpContext(ctx, gooseDB, "."); err != nil {
		log.Fatalf("Running all migrations: %v", err)
	}

	// encx as the app builds it: a KEK in (LocalStack) KMS and a generated pepper
	cryptoSetup, err := tu.SetupCrypto(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup crypto (LocalStack KMS): %v", err)
	}
	defer tu.TeardownCrypto(ctx, nil, cryptoSetup)
	crypto = cryptoSetup.Crypto

	repo = userRepository.New(ctx, testPool)
	svc = bootstrap.New(repo, crypto)

	// The login code path used to prove "can log in": VerifyUserPassword is
	// what SignIn calls after GetUserByEmail.
	users = userService.New(repo, crypto, nil, nil)

	code := m.Run()
	os.Exit(code)
}

// newTestLogger returns a context carrying a logger that writes to the
// returned buffer, mirroring how the app attaches its logger.
func newTestLogger(ctx context.Context) (context.Context, *loggedBuffer) {
	buf := &loggedBuffer{}
	testLogger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return context.WithValue(ctx, ctxutil.LoggerKey, testLogger), buf
}

type loggedBuffer struct {
	data []byte
}

func (b *loggedBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *loggedBuffer) String() string {
	return string(b.data)
}
