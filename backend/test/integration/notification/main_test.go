package notification

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/auth/session"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
	"github.com/Leviosa-care/leviosa/backend/internal/common/envmode"
	"github.com/Leviosa-care/leviosa/backend/internal/common/logger"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware"
	"github.com/Leviosa-care/leviosa/backend/internal/common/middleware/auth"
	"github.com/Leviosa-care/leviosa/backend/internal/common/migrations"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	notificationApp "github.com/Leviosa-care/leviosa/backend/internal/notification/application"
	notificationDomain "github.com/Leviosa-care/leviosa/backend/internal/notification/domain"
	notificationPostgres "github.com/Leviosa-care/leviosa/backend/internal/notification/infrastructure/postgres"
	notificationHandler "github.com/Leviosa-care/leviosa/backend/internal/notification/interface/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
)

// The notification HTTP handler behind the real session auth middleware:
// Postgres (users) and Redis (sessions) in testcontainers, encx with a KEK
// in LocalStack KMS. The notification service and the Web Push sender are
// recording fakes, so nothing is ever sent; push subscriptions are stored
// for real.
var (
	authCtx       *tu.AuthTestContext
	notifications *recordingNotificationService
	pushSender    *recordingPushSender
	pushService   *notificationApp.PushService
	testServerURL string
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loggerHandler, err := logger.SetHandler("debug", "dev")
	if err != nil {
		log.Fatalf("Failed to create logger handler: %v", err)
	}
	testLogger := slog.New(loggerHandler)
	slog.SetDefault(testLogger)
	ctx = context.WithValue(ctx, ctxutil.LoggerKey, testLogger)

	pgContainer, err := tu.SetupPostgres(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer tu.TeardownPostgres(ctx, nil, pgContainer)

	pool, err := pgxpool.New(ctx, pgContainer.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to open test database pool: %v", err)
	}
	defer pool.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("pgx"); err != nil {
		log.Fatalf("Setting dialect for migrations: %v", err)
	}
	gooseDB, err := sql.Open("pgx", pool.Config().ConnString())
	if err != nil {
		log.Fatalf("Failed to open *sql.DB for migrations: %v", err)
	}
	defer gooseDB.Close()
	if err := goose.UpContext(ctx, gooseDB, "."); err != nil {
		log.Fatalf("Running migrations: %v", err)
	}

	redisContainer, err := tu.SetupRedis(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup Redis container: %v", err)
	}
	defer tu.TeardownRedis(ctx, nil, redisContainer)
	redisClient := redisContainer.NewClient()

	cryptoSetup, err := tu.SetupCrypto(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to setup crypto (LocalStack KMS): %v", err)
	}
	defer tu.TeardownCrypto(ctx, nil, cryptoSetup)

	authCtx = &tu.AuthTestContext{Pool: pool, Redis: redisClient, Crypto: cryptoSetup.Crypto}

	sessionRepo := session.NewRedisSessionRepository(redisClient)
	authmw := auth.NewSessionAuthMiddleware(sessionRepo, cryptoSetup.Crypto, tu.NewServiceKeyStore(cryptoSetup.Crypto))

	notifications = &recordingNotificationService{}
	router := http.NewServeMux()
	notificationHandler.New(notifications, authmw).RegisterRoutes(router)

	pushSender = &recordingPushSender{gone: map[string]bool{}}
	pushService = notificationApp.NewPushService(
		notificationPostgres.NewPushSubscriptionRepository(pool),
		pushSender,
		cryptoSetup.Crypto,
		testVAPIDPublicKey,
	)
	notificationHandler.NewPushHandler(pushService, authmw).RegisterRoutes(router)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to listen for test server: %v", err)
	}
	testServerURL = "http://" + listener.Addr().String()
	server := &http.Server{Handler: middleware.AttachLogger(envmode.Dev, testLogger)(router)}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Test server failed to serve: %v", err)
		}
	}()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return m.Run()
}

// recordingNotificationService records contact-form submissions and sends
// nothing.
type recordingNotificationService struct {
	mu       sync.Mutex
	contacts []string
}

func (s *recordingNotificationService) SendContactFormNotification(_ context.Context, name, _, _, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contacts = append(s.contacts, name)
	return nil
}

func (s *recordingNotificationService) contactNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.contacts...)
}

func (s *recordingNotificationService) SendOTPEmail(context.Context, string, string, time.Duration) error {
	return nil
}
func (s *recordingNotificationService) SendWelcomeEmail(context.Context, string, string, string) error {
	return nil
}
func (s *recordingNotificationService) SendVerifyEmailEmail(context.Context, string, string, string) error {
	return nil
}
func (s *recordingNotificationService) SendEventNotificationEmail(context.Context, string, string, string, string, string) error {
	return nil
}
func (s *recordingNotificationService) SendPaymentNotificationEmail(context.Context, string, string, string, string, string, string) error {
	return nil
}
func (s *recordingNotificationService) SendOTPBySMS(context.Context, string, string) error {
	return nil
}
func (s *recordingNotificationService) SendGenericSMS(context.Context, string, string) error {
	return nil
}
func (s *recordingNotificationService) InvalidateSettingsCache() {}

// testVAPIDPublicKey is a placeholder: the recording sender never uses it.
const testVAPIDPublicKey = "BTestVapidPublicKey"

// recordingPushSender records the endpoints it is asked to push to, and
// answers "gone" for the endpoints in gone, as a push service does.
type recordingPushSender struct {
	mu        sync.Mutex
	endpoints []string
	gone      map[string]bool
}

func (s *recordingPushSender) Send(_ context.Context, sub notificationDomain.PushSubscription, _ []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.gone[sub.Endpoint] {
		return notificationDomain.ErrPushSubscriptionGone
	}
	s.endpoints = append(s.endpoints, sub.Endpoint)
	return nil
}

func (s *recordingPushSender) reset(gone ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.endpoints = nil
	s.gone = map[string]bool{}
	for _, endpoint := range gone {
		s.gone[endpoint] = true
	}
}

func (s *recordingPushSender) sentTo() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.endpoints...)
}
