package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hengadev/encx"
	hashicorpkeys "github.com/hengadev/encx/providers/keys/hashicorp"
	hashicorpsecrets "github.com/hengadev/encx/providers/secrets/hashicorp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	bootstrap "github.com/Leviosa-care/leviosa/backend/internal/authuser/application/bootstrap"
	userRepository "github.com/Leviosa-care/leviosa/backend/internal/authuser/infrastructure/postgres/user"
	"github.com/Leviosa-care/leviosa/backend/internal/common/ctxutil"
)

// bootstrapAdmin creates the first superadmin from the ADMIN_* values and
// exits. It is run as a one-off after migrations on every deploy: the only
// container that receives the admin env file, so the admin password never
// reaches the running app. It creates the account when no user with
// ADMIN_EMAIL exists, never modifies an existing one, and never migrates.
// Outside development a missing value or a weak ADMIN_PASSWORD makes it fail
// (non-zero exit) before anything is written.
func bootstrapAdmin(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("bootstrap-admin takes no arguments, got %q", args)
	}

	logger := setupLogger()

	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "development"
	}
	if environment == "development" {
		// Same development convenience as the serve mode; never overrides
		// variables already set in the environment.
		if err := godotenv.Load("development.env"); err != nil {
			fmt.Printf("Warning: Could not load development.env: %v\n", err)
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	cfg, err := bootstrap.LoadConfig(os.Getenv)
	if err != nil {
		return fmt.Errorf("load admin config: %w", err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	crypto, err := newBootstrapCrypto(ctx)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, ctxutil.LoggerKey, logger)
	repo := userRepository.New(ctx, pool)

	created, err := bootstrap.New(repo, crypto).Run(ctx, cfg)
	if err != nil {
		return fmt.Errorf("bootstrap superadmin: %w", err)
	}
	if created {
		logger.InfoContext(ctx, "bootstrap-admin: superadmin created", "email", cfg.Email)
	} else {
		logger.InfoContext(ctx, "bootstrap-admin: account already exists, nothing changed", "email", cfg.Email)
	}
	return nil
}

// newBootstrapCrypto builds the encx crypto service exactly the way the app
// container does today — the hashicorp key and secret providers, configured
// from the ENCX_* variables and authenticated by VAULT_ADDR/VAULT_TOKEN.
func newBootstrapCrypto(ctx context.Context) (encx.CryptoService, error) {
	keyProvider, err := hashicorpkeys.NewTransitService()
	if err != nil {
		return nil, fmt.Errorf("create hashicorp key provider: %w", err)
	}

	secretProvider, err := hashicorpsecrets.NewKVStore()
	if err != nil {
		return nil, fmt.Errorf("create hashicorp secret provider: %w", err)
	}

	kekAlias := os.Getenv("ENCX_KEK_ALIAS")
	if kekAlias == "" {
		kekAlias = "leviosa-kek"
	}
	pepperAlias := os.Getenv("ENCX_PEPPER_ALIAS")
	if pepperAlias == "" {
		pepperAlias = "leviosa"
	}

	crypto, err := encx.NewCrypto(ctx, keyProvider, secretProvider, encx.Config{
		KEKAlias:    kekAlias,
		PepperAlias: pepperAlias,
	})
	if err != nil {
		return nil, fmt.Errorf("create crypto service: %w", err)
	}
	return crypto, nil
}
