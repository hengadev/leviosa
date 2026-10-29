package booking

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

	bookingAdapters "github.com/Leviosa-care/leviosa/backend/internal/booking/adapters/postgres"
	availabilityAppService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/availability"
	bookingService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/booking"
	buildingAppService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/building"
	exceptionAppService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/exception"
	roomAppService "github.com/Leviosa-care/leviosa/backend/internal/booking/application/room"
	bookingNotification "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/notification"
	allocationPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/allocation"
	availabilityPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/availability"
	bookingPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/booking"
	buildingPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/building"
	exceptionPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/exception"
	roomPostgres "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/postgres/room"
	bookingHandler "github.com/Leviosa-care/leviosa/backend/internal/booking/interface/booking"
	exceptionHandler "github.com/Leviosa-care/leviosa/backend/internal/booking/interface/exception"
	"github.com/Leviosa-care/leviosa/backend/internal/booking/ports"

	authuserPartnerService "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/partner"
	authuserUserService "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/user"
	authuserPartnerPostgres "github.com/Leviosa-care/leviosa/backend/internal/authuser/infrastructure/postgres/partner"
	authuserUserPostgres "github.com/Leviosa-care/leviosa/backend/internal/authuser/infrastructure/postgres/user"
	authuserClient "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/authuser"
	bookingSettings "github.com/Leviosa-care/leviosa/backend/internal/booking/infrastructure/settings"

	settingsContracts "github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	settingsService "github.com/Leviosa-care/leviosa/backend/internal/settings/application"
	settingsDomain "github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	settingsNoop "github.com/Leviosa-care/leviosa/backend/internal/settings/infrastructure/noop"
	settingsPostgres "github.com/Leviosa-care/leviosa/backend/internal/settings/infrastructure/postgres"
	settingsPorts "github.com/Leviosa-care/leviosa/backend/internal/settings/ports"

	priceService "github.com/Leviosa-care/leviosa/backend/internal/catalog/application/price"
	productService "github.com/Leviosa-care/leviosa/backend/internal/catalog/application/product"
	pricePostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/price"
	productPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/product"
	sharedPostgres "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/postgres/shared"
	pricePayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/price"
	productPayment "github.com/Leviosa-care/leviosa/backend/internal/catalog/infrastructure/stripe/product"

	"github.com/Leviosa-care/leviosa/backend/internal/common/audit"
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

// testClaimServiceKey is the plaintext key injected into the test secret reader.
// All requests to POST /bookings/claim must supply this as X-Service-Key.
const testClaimServiceKey = "test-authuser-claim-service-key"

// testSecretReader is a static SecretReader that validates against a pre-hashed key.
// It allows the integration tests to exercise RequireServiceAuth without a service key store.
type testSecretReader struct {
	keyHash string // HMAC of testClaimServiceKey, computed once during TestMain
}

func (r *testSecretReader) Read(_ context.Context, _ string) (*auth.SecretData, error) {
	return &auth.SecretData{
		Data: map[string]interface{}{
			"data": map[string]interface{}{
				"key_hash": r.keyHash,
			},
		},
	}, nil
}

var testTokenSecret = []byte("test_booking_token_secret_32bytes")

var (
	pgContainer      *tu.PostgresContainer
	testPool         *pgxpool.Pool
	redisContainer   *tu.RedisContainer
	redisClient      *redis.Client
	crypto           encx.CryptoService
	bookingRepo      ports.BookingRepository
	availabilityRepo ports.AvailabilityRepository
	paymentService   ports.PaymentService
	service          ports.BookingService
	authCtx          *tu.AuthTestContext
	handler          bookingHandler.Handler
	testServerURL    string
	testServer       *http.Server
	authSessionRepo  authsession.SessionRepository
	notificationSpy  *SpyNotificationService

	// Real notification adapter (backed by real user/room/building/product
	// fetchers against the test database) plus SMS/email spies, used by tests
	// that need to verify actual SMS-send-attempt behavior end-to-end, not
	// just that the ports.BookingNotificationService interface was invoked.
	realNotificationAdapter *bookingNotification.BookingNotificationAdapter
	notificationSMSSpy      *IntegrationSMSSpy
	notificationEmailSpy    *IntegrationEmailSpy

	auditLogger *audit.PostgresLogger

	// Real settings repository behind the booking policy: tests change the
	// policy through it, as the admin settings endpoint does.
	settingsRepo settingsPorts.SettingsRepository

	// Real availability service: the partner area creates, edits and deletes
	// availabilities through it (recurring availability tests).
	availabilitySvc ports.AvailabilityService

	// Real availability exception service, also served over HTTP: the
	// partner area closes dates through it (issue 012).
	exceptionRepo ports.AvailabilityExceptionRepository
	exceptionSvc  ports.AvailabilityExceptionService
)

// suiteMinimumNoticeHours is the minimum notice the suite runs with: most
// tests book a few hours ahead, so the 12 h default would refuse them. The
// slot tests set their own policy and restore this one.
const suiteMinimumNoticeHours = 0

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and configure logger for tests
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

	// DB Pool
	log.Println("Creating pgxpool...")
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
	log.Println("pgxpool created.")

	// Ping database
	if err = testPool.Ping(poolCtx); err != nil {
		panic(fmt.Sprintf("Failed to ping database pool: %v", err))
	}
	log.Println("Database pool ping successful.")

	// Database migrations
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

	// Redis container
	redisContainer, err = tu.SetupRedis(ctx, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup redis container: %v", err))
	}
	defer tu.TeardownRedis(ctx, nil, redisContainer)

	// Redis client
	log.Println("Creating Redis client...")
	redisClient = redisContainer.NewClient()

	if err = redisClient.Ping(ctx).Err(); err != nil {
		tu.TeardownRedis(ctx, nil, redisContainer)
		panic(fmt.Sprintf("Failed to ping Redis: %v", err))
	}
	log.Println("Redis client connected successfully.")

	// encx as the app builds it: a KEK in (LocalStack) KMS and a generated pepper
	cryptoSetup, err := tu.SetupCrypto(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup crypto (LocalStack KMS): %v", err)
	}
	defer tu.TeardownCrypto(ctx, nil, cryptoSetup)
	crypto = cryptoSetup.Crypto

	// Initialize AuthTestContext for user/session testing
	authCtx = &tu.AuthTestContext{
		Pool:   testPool,
		Redis:  redisClient,
		Crypto: crypto,
	}

	ctx = context.WithValue(ctx, ctxutil.LoggerKey, testLogger)

	// Initialize application layers
	bookingRepo = bookingPostgres.New(ctx, testPool)
	availabilityRepo = availabilityPostgres.New(ctx, testPool)
	exceptionRepo = exceptionPostgres.New(ctx, testPool)
	exceptionSvc = exceptionAppService.New(exceptionRepo, availabilityRepo, bookingRepo, crypto)

	// Initialize mock payment service (Stripe integration not needed for booking tests)
	paymentService = NewMockPaymentService()

	// Initialize real product service with all dependencies
	productRepo := productPostgres.New(ctx, testPool)
	sharedRepo := sharedPostgres.New(ctx, testPool)

	// Use Stripe test key for product service
	stripeTestKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeTestKey == "" {
		stripeTestKey = "sk_test_dummy_key_for_testing" // Fallback for test environment
	}
	stripeBaseURL := os.Getenv("STRIPE_API_BASE_URL") // Empty for production Stripe API

	productStripe := productPayment.NewProduct(stripeTestKey, stripeBaseURL)
	priceStripe := pricePayment.NewPrice(stripeTestKey, stripeBaseURL)

	catalogProductService := productService.New(productRepo, sharedRepo, productStripe, priceStripe, nil)

	catalogPriceRepo := pricePostgres.New(ctx, testPool)
	catalogPriceService := priceService.New(catalogPriceRepo, sharedRepo, priceStripe)

	notificationSpy = NewSpyNotificationService()
	auditLogger = audit.NewPostgresLogger(testPool)

	// Real authuser user service: shared by the in-process AuthUserClient (so
	// the partner agenda can resolve registered clients' contact details) and
	// the notification adapter below.
	authUserRepo := authuserUserPostgres.New(ctx, testPool)
	realAuthUserService := authuserUserService.New(authUserRepo, crypto, nil, nil)

	// Real partner service: the public slot endpoint resolves a partner ID
	// (as the practice page settings hold it) to the partner's user ID.
	realPartnerService, err := authuserPartnerService.New(ctx, authuserPartnerPostgres.New(ctx, testPool), authUserRepo, catalogProductService, nil, crypto, nil)
	if err != nil {
		log.Fatalf("Failed to create partner service: %v", err)
	}

	// Booking policy from the real settings service, as the app wires it.
	settingsRepo = settingsPostgres.New(ctx, testPool)
	realSettingsService := settingsService.New(settingsRepo, nil, crypto, settingsNoop.NewPublisher(), nil)
	if err := settingsRepo.SetInt(ctx, &settingsDomain.Setting[int]{Key: settingsContracts.BookingMinimumNoticeHours, Value: suiteMinimumNoticeHours}); err != nil {
		log.Fatalf("Failed to set the suite's booking policy: %v", err)
	}

	service = bookingService.New(bookingRepo, availabilityRepo, paymentService, catalogProductService, catalogPriceService, notificationSpy, crypto,
		bookingService.WithTokenSecret(testTokenSecret),
		bookingService.WithAuditLogger(auditLogger),
		bookingService.WithLatePaymentNotifier(notificationSpy),
		bookingService.WithAuthUserClient(authuserClient.NewInProcessClient(realPartnerService, realAuthUserService)),
		bookingService.WithBookingPolicyProvider(bookingSettings.NewInProcessBookingPolicyProvider(realSettingsService)),
		bookingService.WithAvailabilityExceptionRepository(exceptionRepo))

	allocationRepo, err := allocationPostgres.New(ctx, testPool)
	if err != nil {
		log.Fatalf("Failed to create allocation repository: %v", err)
	}
	availabilitySvc = availabilityAppService.New(availabilityRepo, allocationRepo, roomPostgres.New(ctx, testPool),
		bookingAdapters.NewRoomScheduleRepository(testPool), catalogProductService, crypto)

	// Wire a real BookingNotificationAdapter (real user/room/building/product
	// fetchers against the test DB) backed by SMS/email spies, so tests can
	// assert that an SMS send is actually attempted end-to-end, not just that
	// the notification port was invoked.

	realBuildingRepo := buildingPostgres.New(ctx, testPool)
	realBuildingService := buildingAppService.New(realBuildingRepo, crypto)

	realRoomRepo := roomPostgres.New(ctx, testPool)
	realRoomService := roomAppService.New(realRoomRepo, realBuildingRepo, crypto)

	notificationSMSSpy = &IntegrationSMSSpy{}
	notificationEmailSpy = &IntegrationEmailSpy{}
	realNotificationAdapter = bookingNotification.NewBookingNotificationAdapter(
		notificationEmailSpy,
		notificationSMSSpy,
		"https://app.leviosa.test",
		bookingNotification.NewInProcessUserFetcher(realAuthUserService),
		bookingNotification.NewInProcessRoomFetcher(realRoomService),
		bookingNotification.NewInProcessBuildingFetcher(realBuildingService),
		catalogProductService,
		&stubSettingsProvider{},
	)

	// Pre-compute the hash of the test service key so the mock secret reader can
	// return it for RequireServiceAuth validation on POST /bookings/claim.
	claimKeyBytes, err := encx.SerializeValue(testClaimServiceKey)
	if err != nil {
		log.Fatalf("Failed to serialize test claim service key: %v", err)
	}
	claimKeyHash := crypto.HashBasic(ctx, claimKeyBytes)
	mockReader := &testSecretReader{keyHash: claimKeyHash}

	authSessionRepo = authsession.NewRedisSessionRepository(redisClient)
	authmw := auth.NewSessionAuthMiddleware(authSessionRepo, crypto, nil,
		auth.WithSecretReader(mockReader),
	)

	handler = bookingHandler.New(service, paymentService, authmw)

	// Set required environment variables for logger middleware
	os.Setenv("CLIENT_IP_HEADER", "X-Forwarded-For")
	os.Setenv("LOGGING_SALT", "test_logging_salt_12345")

	// HTTP server setup with logger middleware
	router := http.NewServeMux()
	handler.RegisterRoutes(router)
	exceptionHandler.New(exceptionSvc, authmw).RegisterRoutes(router)

	loggerMiddleware := middleware.AttachLogger(envmode.Dev, testLogger)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen for test server: %v", err)
	}
	testServerURL = "http://" + listener.Addr().String()
	testServer = &http.Server{Handler: loggerMiddleware(router)}

	// Start server in goroutine
	go func() {
		if err := testServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Test server failed to serve: %v", err)
		}
	}()
	log.Printf("Test HTTP server started at %s", testServerURL)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Run tests
	code := m.Run()

	// Graceful shutdown
	log.Println("Shutting down test HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := testServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Test server shutdown failed: %v", err)
	}
	log.Println("Test HTTP server shut down.")

	// Exit with test result code
	os.Exit(code)
}
