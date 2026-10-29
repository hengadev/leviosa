package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/encryption"
)

type SeedConfig struct {
	DatabaseURL    string
	DatabaseSchema string
	Encryption     encryption.Config
	SeedDataPath   string
}

func loadConfig() (*SeedConfig, error) {
	if err := godotenv.Load("development.env"); err != nil {
		log.Printf("Warning: Could not load development.env: %v", err)
	}

	cfg := &SeedConfig{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DatabaseSchema: getEnv("DATABASE_SCHEMA", "auth"),
		Encryption:     encryption.ConfigFromEnv(os.Getenv),
		SeedDataPath:   getEnv("SEED_DATA_PATH", "cmd/seed/seed_data.json"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// AccountSeed is one user account to upsert (an administrator, or the
// practice's partner when the practice seed references one).
type AccountSeed struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	Telephone  string `json:"telephone"`
	PostalCode string `json:"postal_code"`
	City       string `json:"city"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
}

// SeedData is the seed data file: the admin accounts, test client accounts,
// plus the single practice created when a "practice" section is present (see
// PracticeSeed).
type SeedData struct {
	Admins   []AccountSeed `json:"admins"`
	Clients  []AccountSeed `json:"clients"`
	Practice *PracticeSeed `json:"practice,omitempty"`
}

func loadSeedData(path string) (*SeedData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read seed data file %q: %w", path, err)
	}
	var seedData SeedData
	if err := json.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("parse seed data file: %w", err)
	}
	return &seedData, nil
}

// seedUser upserts a user (by email) and returns the account's stable ID.
func seedUser(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, schema string, account AccountSeed) (uuid.UUID, error) {
	now := time.Now()

	user := &domain.User{
		ID:         uuid.New(),
		State:      domain.Active,
		Email:      account.Email,
		Password:   account.Password,
		FirstName:  account.FirstName,
		LastName:   account.LastName,
		Role:       account.Role,
		Telephone:  account.Telephone,
		PostalCode: account.PostalCode,
		City:       account.City,
		Address1:   account.Address1,
		Address2:   account.Address2,
		CreatedAt:  now,
		LoggedInAt: now,
	}

	userEncx, err := domain.ProcessUserEncx(ctx, crypto, user)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt user data: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.users (
			id, state,
			email_encrypted, email_hash,
			password_hash_secure,
			first_name_encrypted,
			last_name_encrypted,
			role_encrypted,
			telephone_encrypted, telephone_hash,
			postal_code_encrypted,
			city_encrypted,
			address1_encrypted,
			address2_encrypted,
			created_at_encrypted,
			logged_in_at_encrypted,
			picture_encrypted,
			birth_date_encrypted,
			gender_encrypted,
			google_id_encrypted,
			apple_id_encrypted,
			stripe_customer_id_encrypted,
			dek_encrypted, key_version
		) VALUES (
			$1, $2,
			$3, $4,
			$5,
			$6,
			$7,
			$8,
			$9, $10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19,
			$20,
			$21,
			$22,
			$23, $24
		)
		ON CONFLICT (email_hash) DO UPDATE SET
			email_encrypted = EXCLUDED.email_encrypted,
			password_hash_secure = EXCLUDED.password_hash_secure,
			first_name_encrypted = EXCLUDED.first_name_encrypted,
			last_name_encrypted = EXCLUDED.last_name_encrypted,
			role_encrypted = EXCLUDED.role_encrypted,
			telephone_encrypted = EXCLUDED.telephone_encrypted,
			telephone_hash = EXCLUDED.telephone_hash,
			postal_code_encrypted = EXCLUDED.postal_code_encrypted,
			city_encrypted = EXCLUDED.city_encrypted,
			address1_encrypted = EXCLUDED.address1_encrypted,
			address2_encrypted = EXCLUDED.address2_encrypted,
			created_at_encrypted = EXCLUDED.created_at_encrypted,
			logged_in_at_encrypted = EXCLUDED.logged_in_at_encrypted,
			dek_encrypted = EXCLUDED.dek_encrypted,
			key_version = EXCLUDED.key_version
	`, schema)

	_, err = db.Exec(ctx, query,
		userEncx.ID, userEncx.State,
		userEncx.EmailEncrypted, userEncx.EmailHash,
		userEncx.PasswordHashSecure,
		userEncx.FirstNameEncrypted,
		userEncx.LastNameEncrypted,
		userEncx.RoleEncrypted,
		userEncx.TelephoneEncrypted, userEncx.TelephoneHash,
		userEncx.PostalCodeEncrypted,
		userEncx.CityEncrypted,
		userEncx.Address1Encrypted,
		userEncx.Address2Encrypted,
		userEncx.CreatedAtEncrypted,
		userEncx.LoggedInAtEncrypted,
		userEncx.PictureEncrypted,
		userEncx.BirthDateEncrypted,
		userEncx.GenderEncrypted,
		userEncx.GoogleIDEncrypted,
		userEncx.AppleIDEncrypted,
		userEncx.StripeCustomerIDEncrypted,
		userEncx.DEKEncrypted, userEncx.KeyVersion,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert user into database: %w", err)
	}

	// The upsert keeps a pre-existing account's ID stable; read back the row
	// that won so callers (e.g. the practice seed) can reference it.
	var id uuid.UUID
	err = db.QueryRow(ctx, fmt.Sprintf(`SELECT id FROM %s.users WHERE email_hash = $1`, schema), userEncx.EmailHash).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("look up seeded user: %w", err)
	}

	log.Printf("Seeded user: %s (%s %s, role %s)", account.Email, account.FirstName, account.LastName, account.Role)
	return id, nil
}

func main() {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	crypto, err := encryption.NewCrypto(ctx, cfg.Encryption)
	if err != nil {
		log.Fatalf("Failed to setup crypto: %v", err)
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	admins, err := loadSeedData(cfg.SeedDataPath)
	if err != nil {
		log.Fatalf("Failed to load seed data: %v", err)
	}

	log.Println("Starting database seeding...")
	log.Println("====================================")

	for _, account := range admins.Admins {
		if _, err := seedUser(ctx, db, crypto, cfg.DatabaseSchema, account); err != nil {
			log.Fatalf("Failed to seed user %s: %v", account.Email, err)
		}
	}

	for _, account := range admins.Clients {
		if _, err := seedUser(ctx, db, crypto, cfg.DatabaseSchema, account); err != nil {
			log.Fatalf("Failed to seed client %s: %v", account.Email, err)
		}
	}

	if admins.Practice != nil {
		if err := seedPractice(ctx, db, crypto, cfg.DatabaseSchema, admins.Practice); err != nil {
			log.Fatalf("Failed to seed the practice: %v", err)
		}
	}

	log.Println("====================================")
	log.Println("Database seeding completed successfully!")
}
