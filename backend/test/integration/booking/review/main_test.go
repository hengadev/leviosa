package review

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

	bookingService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/booking"
	reviewService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/review"
	availabilityPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/availability"
	bookingPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/booking"
	reviewPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/review"
	bookingHandler "github.com/Leviosa-care/leviosa/backend/internal/booking/interface/booking"
	reviewHandler "github.com/Leviosa-care/leviosa/backend/internal/booking/interface/review"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"

	priceService "github.com/Leviosa-care/leviosa/backend/internal/catalog/application/price"
	productService "github.com/Leviosa-care/leviosa/backend/internal/catalog/application/product"
	pricePostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/price"
	productPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/product"
	sharedPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/shared"
	pricePayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/price"
	productPayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/product"

	authsession "github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/envmode"
	"github.com/Leviosa-care/leviosa/backend/internal/common/logger"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"

	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

var testTokenSecret = []byte("test_booking_token_secret_32bytes")

var (
	pgContainer      *tu.PostgresContainer
	testPool         *pgxpool.Pool
	redisContainer   *tu.RedisContainer
	redisClient      *redis.Client
	crypto           encx.CryptoService
	bookingRepo      ports.BookingRepository
	availabilityRepo ports.AvailabilityRepository
	reviewRepo       ports.ReviewRepository
	bookingSvc       ports.BookingService
	reviewSvc        ports.ReviewService
	authCtx          *tu.AuthTestContext
	testServerURL    string
	testServer       *http.Server
	authSessionRepo  authsession.SessionRepository
)

// TestMain wires a real Postgres + Redis + KMS-backed booking service, plus the
// review sub-module under test, onto a single HTTP server — mirroring production
// wiring where both handlers share one mux. Tests use the booking endpoints to set
// up (create + complete) bookings via HTTP, then exercise the review endpoints.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loggerHandler, err := logger.SetHandler("debug", "dev")
	if err != nil {
		log.Fatalf("Failed to create logger handler: %v", err)
	}
	testLogger := slog.New(loggerHandler)
	slog.SetDefault(testLogger)

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

	log.Println("Applying database migrations...")
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
	log.Println("Migrations applied.")

	redisContainer, err = tu.SetupRedis(ctx, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup redis container: %v", err))
	}
	defer tu.TeardownRedis(ctx, nil, redisContainer)

	redisClient = redisContainer.NewClient()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		tu.TeardownRedis(ctx, nil, redisContainer)
		panic(fmt.Sprintf("Failed to ping Redis: %v", err))
	}

	// encx as the app builds it: a KEK in (LocalStack) KMS and a generated pepper
	cryptoSetup, err := tu.SetupCrypto(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup crypto (LocalStack KMS): %v", err)
	}
	defer tu.TeardownCrypto(ctx, nil, cryptoSetup)
	crypto = cryptoSetup.Crypto

	authCtx = &tu.AuthTestContext{
		Pool:   testPool,
		Redis:  redisClient,
		Crypto: crypto,
	}

	ctx = context.WithValue(ctx, ctxutil.LoggerKey, testLogger)

	// Booking dependencies
	bookingRepo = bookingPostgres.New(ctx, testPool)
	availabilityRepo = availabilityPostgres.New(ctx, testPool)
	reviewRepo = reviewPostgres.New(ctx, testPool)

	paymentSvc := NewMockPaymentService()

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

	catalogPriceRepo := pricePostgres.New(ctx, testPool)
	catalogPriceService := priceService.New(catalogPriceRepo, sharedRepo, priceStripe)

	notificationSpy := NewSpyNotificationService()
	bookingSvc = bookingService.New(bookingRepo, availabilityRepo, paymentSvc, catalogProductService, catalogPriceService, notificationSpy, crypto,
		bookingService.WithTokenSecret(testTokenSecret))

	// Review sub-module under test
	reviewSvc = reviewService.New(reviewRepo, bookingSvc)

	authSessionRepo = authsession.NewRedisSessionRepository(redisClient)
	authmw := auth.NewSessionAuthMiddleware(authSessionRepo, crypto, nil)

	bookingH := bookingHandler.New(bookingSvc, paymentSvc, authmw)
	reviewH := reviewHandler.New(reviewSvc, authmw)

	os.Setenv("CLIENT_IP_HEADER", "X-Forwarded-For")
	os.Setenv("LOGGING_SALT", "test_logging_salt_12345")

	router := http.NewServeMux()
	bookingH.RegisterRoutes(router)
	reviewH.RegisterRoutes(router)

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
	log.Printf("Test HTTP server started at %s", testServerURL)

	time.Sleep(100 * time.Millisecond)

	code := m.Run()

	log.Println("Shutting down test HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := testServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Test server shutdown failed: %v", err)
	}

	os.Exit(code)
}
