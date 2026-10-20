package waitlist

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	waitlistService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/waitlist"
	waitlistPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/waitlist"
	waitlistHandler "github.com/Leviosa-care/leviosa/backend/internal/booking/interface/waitlist"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"

	productService "github.com/Leviosa-care/leviosa/backend/internal/catalog/application/product"
	productPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/product"
	sharedPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/shared"
	pricePayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/price"
	productPayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/product"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/envmode"
	"github.com/Leviosa-care/leviosa/backend/internal/common/logger"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

var (
	pgContainer   *tu.PostgresContainer
	testPool      *pgxpool.Pool
	crypto        encx.CryptoService
	waitlistRepo  ports.WaitlistRepository
	service       ports.WaitlistService
	handler       waitlistHandler.Handler
	testServerURL string
	testServer    *http.Server
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loggerHandler, err := logger.SetHandler("debug", "dev")
	if err != nil {
		log.Fatalf("Failed to create logger handler: %v", err)
	}
	testLogger := slog.New(loggerHandler)
	slog.SetDefault(testLogger)

	// Postgres container
	pgContainer, err = tu.SetupPostgres(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer tu.TeardownPostgres(ctx, nil, pgContainer)

	poolCtx, poolCancel := context.WithTimeout(ctx, 10*time.Second)
	defer poolCancel()

	pgCfg, err := pgxpool.ParseConfig(pgContainer.ConnectionString)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse pgxpool config: %v", err))
	}
	pgCfg.MaxConns = 5
	pgCfg.MinConns = 1
	pgCfg.MaxConnLifetime = 30 * time.Minute
	pgCfg.MaxConnIdleTime = 5 * time.Minute

	testPool, err = pgxpool.NewWithConfig(poolCtx, pgCfg)
	if err != nil {
		tu.TeardownPostgres(ctx, nil, pgContainer)
		panic(fmt.Sprintf("Failed to open test database pool: %v", err))
	}

	if err = testPool.Ping(poolCtx); err != nil {
		panic(fmt.Sprintf("Failed to ping database pool: %v", err))
	}

	// Database migrations
	goose.SetBaseFS(migrations.FS)
	if err = goose.SetDialect("pgx"); err != nil {
		log.Fatalf("Setting dialect for migrations: %s\n", err)
	}

	gooseDB, err := sql.Open("pgx", testPool.Config().ConnString())
	if err != nil {
		panic(fmt.Sprintf("Failed to open temp *sql.DB for goose migrations: %v", err))
	}
	defer gooseDB.Close()

	if err = goose.UpContext(ctx, gooseDB, "."); err != nil {
		panic(fmt.Sprintf("running all migrations: %s\n", err))
	}

	// encx as the app builds it: a KEK in (LocalStack) KMS and a generated pepper
	cryptoSetup, err := tu.SetupCrypto(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup crypto (LocalStack KMS): %v", err)
	}
	defer tu.TeardownCrypto(ctx, nil, cryptoSetup)
	crypto = cryptoSetup.Crypto

	ctx = context.WithValue(ctx, ctxutil.LoggerKey, testLogger)

	// Waitlist repository
	waitlistRepo = waitlistPostgres.New(ctx, testPool)

	// Real product service (catalog) so JoinWaitlist's product-existence check
	// exercises the same path as production.
	productRepo := productPostgres.New(ctx, testPool)
	sharedRepo := sharedPostgres.New(ctx, testPool)

	stripeTestKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeTestKey == "" {
		stripeTestKey = "sk_test_dummy_key_for_testing"
	}
	stripeBaseURL := os.Getenv("STRIPE_API_BASE_URL")

	productStripe := productPayment.NewProduct(stripeTestKey, stripeBaseURL)
	priceStripe := pricePayment.NewPrice(stripeTestKey, stripeBaseURL)

	catalogProductService := productService.New(productRepo, sharedRepo, productStripe, priceStripe, nil)

	service = waitlistService.New(waitlistRepo, catalogProductService, crypto)
	handler = waitlistHandler.New(service)

	os.Setenv("CLIENT_IP_HEADER", "X-Forwarded-For")
	os.Setenv("LOGGING_SALT", "test_logging_salt_12345")

	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	loggerMiddleware := middleware.AttachLogger(envmode.Dev, testLogger)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen for test server: %v", err)
	}
	testServerURL = "http://" + listener.Addr().String()
	testServer = &http.Server{Handler: loggerMiddleware(router)}

	go func() {
		if err := testServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Test server failed to serve: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := testServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Test server shutdown failed: %v", err)
	}

	os.Exit(code)
}
