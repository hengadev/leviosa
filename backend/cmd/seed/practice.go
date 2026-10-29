package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hengadev/encx"
	"github.com/jackc/pgx/v5/pgxpool"

	authDomain "github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	bookingDomain "github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	catalogDomain "github.com/Leviosa-care/leviosa/backend/internal/catalog/domain"
)

// PracticeSeed describes the single practice the seed creates: one partner
// (user + profile), one building, one room, one dedicated allocation, one
// product (60 min, buffer 0) with its price, and the practice page settings
// (docs/issues/009-practice-page.md).
//
// Every client-facing value here is a PLACEHOLDER: the real name, bio, title,
// address and price come from the practitioner and must be replaced (issue
// 021) before anything ships. The practice page video setting is deliberately
// left unset — the page shows its placeholder frame until the practice's own
// video is configured.
type PracticeSeed struct {
	Partner PracticePartnerSeed `json:"partner"`
	// The product needs a category to exist; one is created for the practice.
	Category PracticeCategorySeed `json:"category"`
	Product  PracticeProductSeed  `json:"product"`
	Building PracticeBuildingSeed `json:"building"`
	Room     PracticeRoomSeed     `json:"room"`
}

type PracticePartnerSeed struct {
	// Account fields, shared with the admin seed format
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Telephone  string `json:"telephone"`
	PostalCode string `json:"postal_code"`
	City       string `json:"city"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`

	// Public profile shown on the practice page (PLACEHOLDER values)
	Occupation string `json:"occupation"` // rendered as the partner's title
	Bio        string `json:"bio"`
}

type PracticeCategorySeed struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PracticeProductSeed struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	DurationMinutes   int    `json:"duration_minutes"`
	BufferMinutes     int    `json:"buffer_minutes"`
	CancellationHours int    `json:"cancellation_hours"`
	PriceCents        int    `json:"price_cents"`
	Currency          string `json:"currency"`
	// Deposit is a PLACEHOLDER (issue 013): the practitioner's real deposit
	// is unknown and is set in the admin product form before launch.
	Deposit *catalogDomain.Deposit `json:"deposit,omitempty"`
}

type PracticeBuildingSeed struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type PracticeRoomSeed struct {
	Name string `json:"name"`
}

// seed-local stand-ins for the Stripe identifiers, which are NOT NULL in the
// schema but do not exist until the product is created through the admin
// catalog against a real Stripe account. Clearly not Stripe IDs so they can
// never be mistaken for live objects.
const (
	seedStripeProductID = "seed-local-product"
	seedStripePriceID   = "seed-local-price"
)

func seedPractice(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, userSchema string, practice *PracticeSeed) error {
	// 1. The partner's user account (role: partner)
	userID, err := seedUser(ctx, db, crypto, userSchema, AccountSeed{
		Email:      practice.Partner.Email,
		Password:   practice.Partner.Password,
		FirstName:  practice.Partner.FirstName,
		LastName:   practice.Partner.LastName,
		Role:       "partner",
		Telephone:  practice.Partner.Telephone,
		PostalCode: practice.Partner.PostalCode,
		City:       practice.Partner.City,
		Address1:   practice.Partner.Address1,
		Address2:   practice.Partner.Address2,
	})
	if err != nil {
		return fmt.Errorf("seed practice partner user: %w", err)
	}

	// 2. The practice's category and product (60 min, buffer 0), published so
	// the public endpoints return them.
	categoryID, err := seedCategory(ctx, db, practice.Category)
	if err != nil {
		return fmt.Errorf("seed practice category: %w", err)
	}
	productID, err := seedProduct(ctx, db, categoryID, practice.Product)
	if err != nil {
		return fmt.Errorf("seed practice product: %w", err)
	}
	if err := seedPrice(ctx, db, productID, practice.Product); err != nil {
		return fmt.Errorf("seed practice price: %w", err)
	}

	// 3. The partner's public profile, linked to the product and category
	partnerID, err := seedPartnerProfile(ctx, db, crypto, userID, categoryID, productID, practice.Partner)
	if err != nil {
		return fmt.Errorf("seed practice partner profile: %w", err)
	}

	// 4. One building, one room, one dedicated allocation to the partner
	roomID, err := seedBuildingAndRoom(ctx, db, crypto, practice.Building, practice.Room)
	if err != nil {
		return fmt.Errorf("seed practice building and room: %w", err)
	}
	if err := seedDedicatedAllocation(ctx, db, crypto, roomID, userID); err != nil {
		return fmt.Errorf("seed practice allocation: %w", err)
	}
	if err := seedRoomOpeningHours(ctx, db, roomID); err != nil {
		return fmt.Errorf("seed practice room opening hours: %w", err)
	}

	// 5. The practice page settings: which partner and product the public
	// page shows, and the practice address it displays. No release needed to
	// change them — the admin settings endpoints rewrite these rows.
	settings := map[string]string{
		"practice_page_partner_id": partnerID.String(),
		"practice_page_product_id": productID.String(),
		"headquarters":             practice.Building.Address + ", " + practice.Building.PostalCode + " " + practice.Building.City,
	}
	for key, value := range settings {
		if _, err := db.Exec(ctx, `
			INSERT INTO settings.plain (key, value)
			VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		`, key, value); err != nil {
			return fmt.Errorf("seed practice setting %q: %w", key, err)
		}
	}

	log.Printf("Seeded practice: partner %s (%s %s), product %q (%d min, %d cents), room %q",
		partnerID, practice.Partner.FirstName, practice.Partner.LastName,
		practice.Product.Name, practice.Product.DurationMinutes, practice.Product.PriceCents,
		practice.Room.Name)
	log.Printf("NOTE: practice values (name, bio, title, address, price, deposit) are seed placeholders — replace with the practitioner's real content before launch")
	return nil
}

func seedCategory(ctx context.Context, db *pgxpool.Pool, category PracticeCategorySeed) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO catalog.categories (id, name, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'published', NOW(), NOW())
		ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description, status = 'published', updated_at = NOW()
		RETURNING id
	`, uuid.New(), category.Name, category.Description).Scan(&id)
	return id, err
}

func seedProduct(ctx context.Context, db *pgxpool.Pool, categoryID uuid.UUID, product PracticeProductSeed) (uuid.UUID, error) {
	domainProduct := catalogDomain.Product{
		ID:                uuid.New(),
		Name:              product.Name,
		Description:       product.Description,
		CategoryID:        categoryID,
		Duration:          product.DurationMinutes,
		Status:            catalogDomain.Published,
		Availability:      catalogDomain.InPerson,
		BufferTime:        product.BufferMinutes,
		CancellationHours: product.CancellationHours,
		StripeProductID:   seedStripeProductID,
		Deposit:           product.Deposit,
	}
	if err := domainProduct.Valid(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("product invalid: %w", err)
	}
	if err := domainProduct.Deposit.ValidateAgainstPrice(product.PriceCents); err != nil {
		return uuid.Nil, fmt.Errorf("product deposit invalid: %w", err)
	}
	var depositJSON []byte
	if domainProduct.Deposit != nil {
		var err error
		if depositJSON, err = json.Marshal(domainProduct.Deposit); err != nil {
			return uuid.Nil, fmt.Errorf("encode product deposit: %w", err)
		}
	}

	var id uuid.UUID
	err := db.QueryRow(ctx, `
		INSERT INTO catalog.products (
			id, name, description, category_id, duration, created_at, updated_at,
			status, availability, buffer_time, cancellation_hours, stripe_product_id,
			deposit
		) VALUES (
			$1, $2, $3, $4, $5, NOW(), NOW(),
			$6, $7, $8, $9, $10, $11
		)
		ON CONFLICT (name) DO UPDATE SET
			description = EXCLUDED.description,
			category_id = EXCLUDED.category_id,
			duration = EXCLUDED.duration,
			status = EXCLUDED.status,
			availability = EXCLUDED.availability,
			buffer_time = EXCLUDED.buffer_time,
			cancellation_hours = EXCLUDED.cancellation_hours,
			deposit = EXCLUDED.deposit,
			updated_at = NOW()
		RETURNING id
	`, domainProduct.ID, domainProduct.Name, domainProduct.Description, domainProduct.CategoryID,
		domainProduct.Duration, domainProduct.Status, domainProduct.Availability,
		domainProduct.BufferTime, domainProduct.CancellationHours, domainProduct.StripeProductID,
		depositJSON,
	).Scan(&id)
	return id, err
}

func seedPrice(ctx context.Context, db *pgxpool.Pool, productID uuid.UUID, product PracticeProductSeed) error {
	price := catalogDomain.Price{
		ID:            uuid.New(),
		ProductID:     productID,
		StripePriceID: seedStripePriceID,
		Amount:        product.PriceCents,
		Currency:      product.Currency,
		Interval:      catalogDomain.OneTime,
		IsActive:      true,
	}
	if err := price.Valid(ctx); err != nil {
		return fmt.Errorf("price invalid: %w", err)
	}

	_, err := db.Exec(ctx, `
		INSERT INTO catalog.prices (id, product_id, stripe_price_id, amount, currency, interval, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW(), NOW())
		ON CONFLICT (stripe_price_id) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			is_active = TRUE,
			updated_at = NOW()
	`, price.ID, price.ProductID, price.StripePriceID, price.Amount, price.Currency, price.Interval)
	return err
}

func seedPartnerProfile(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, userID, categoryID, productID uuid.UUID, partner PracticePartnerSeed) (uuid.UUID, error) {
	now := time.Now()
	domainPartner := &authDomain.Partner{
		ID:                  uuid.New(),
		UserID:              userID,
		Bio:                 partner.Bio,
		Occupation:          partner.Occupation,
		Tags:                []string{},
		CategoryIDs:         []uuid.UUID{categoryID},
		ProductIDs:          []uuid.UUID{productID},
		StripeAccountStatus: authDomain.StripeAccountStatusPending,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	partnerEncx, err := authDomain.ProcessPartnerEncx(ctx, crypto, domainPartner)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt partner profile: %w", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO auth.partners (
			id, user_id, bio, experience, occupation, quote, tags,
			category_ids, product_ids,
			stripe_connected_account_id_encrypted, stripe_account_status, stripe_onboarding_complete,
			dek_encrypted, key_version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, COALESCE($7, ARRAY[]::text[]), $8, $9, $10, $11, FALSE, $12, $13, $14, $15
		)
		ON CONFLICT (user_id) DO UPDATE SET
			bio = EXCLUDED.bio,
			occupation = EXCLUDED.occupation,
			category_ids = EXCLUDED.category_ids,
			product_ids = EXCLUDED.product_ids,
			updated_at = NOW()
	`, partnerEncx.ID, partnerEncx.UserID, partnerEncx.Bio, partnerEncx.Experience, partnerEncx.Occupation,
		partnerEncx.Quote, partnerEncx.Tags, partnerEncx.CategoryIDs, partnerEncx.ProductIDs,
		partnerEncx.StripeConnectedAccountIDEncrypted, partnerEncx.StripeAccountStatus,
		partnerEncx.DEKEncrypted, partnerEncx.KeyVersion, partnerEncx.CreatedAt, partnerEncx.UpdatedAt,
	)
	if err != nil {
		return uuid.Nil, err
	}

	// The conflict keeps the pre-existing partner ID stable (settings already
	// reference it); read back whichever row won.
	var id uuid.UUID
	err = db.QueryRow(ctx, `SELECT id FROM auth.partners WHERE user_id = $1`, userID).Scan(&id)
	return id, err
}

func seedBuildingAndRoom(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, buildingSeed PracticeBuildingSeed, roomSeed PracticeRoomSeed) (uuid.UUID, error) {
	now := time.Now()
	building := &bookingDomain.Building{
		ID:         uuid.New(),
		Name:       buildingSeed.Name,
		Address:    buildingSeed.Address,
		City:       buildingSeed.City,
		PostalCode: buildingSeed.PostalCode,
		Country:    buildingSeed.Country,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	buildingEncx, err := bookingDomain.ProcessBuildingEncx(ctx, crypto, building)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt building: %w", err)
	}

	// Buildings are encrypted, so idempotency keys on the searchable hash.
	var buildingID uuid.UUID
	err = db.QueryRow(ctx, `SELECT id FROM booking.buildings WHERE name_hash = $1`, buildingEncx.NameHash).Scan(&buildingID)
	if err != nil {
		if _, err := db.Exec(ctx, `
			INSERT INTO booking.buildings (
				id, name_encrypted, name_hash, address_encrypted, address_hash,
				city_encrypted, city_hash, postal_code_encrypted,
				country_encrypted, country_hash, dek_encrypted, key_version,
				is_active, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, TRUE, $13, $14
			)
		`, buildingEncx.ID, buildingEncx.NameEncrypted, buildingEncx.NameHash,
			buildingEncx.AddressEncrypted, buildingEncx.AddressHash,
			buildingEncx.CityEncrypted, buildingEncx.CityHash, buildingEncx.PostalCodeEncrypted,
			buildingEncx.CountryEncrypted, buildingEncx.CountryHash,
			buildingEncx.DEKEncrypted, buildingEncx.KeyVersion,
			buildingEncx.CreatedAt, buildingEncx.UpdatedAt,
		); err != nil {
			return uuid.Nil, err
		}
		buildingID = buildingEncx.ID
	}

	room := &bookingDomain.Room{
		ID:         uuid.New(),
		BuildingID: buildingID,
		Name:       roomSeed.Name,
		Capacity:   1,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	roomEncx, err := bookingDomain.ProcessRoomEncx(ctx, crypto, room)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encrypt room: %w", err)
	}

	var roomID uuid.UUID
	err = db.QueryRow(ctx, `SELECT id FROM booking.rooms WHERE name_hash = $1 AND building_id = $2`, roomEncx.NameHash, buildingID).Scan(&roomID)
	if err != nil {
		if _, err := db.Exec(ctx, `
			INSERT INTO booking.rooms (
				id, building_id, name_encrypted, name_hash, description_encrypted,
				room_number_encrypted, room_number_hash, capacity, equipment_encrypted,
				is_active, created_at, updated_at,
				dek_encrypted, key_version, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, $10, $11, $12, $13, $14
			)
		`, roomEncx.ID, roomEncx.BuildingID, roomEncx.NameEncrypted, roomEncx.NameHash,
			roomEncx.DescriptionEncrypted, roomEncx.RoomNumberEncrypted, roomEncx.RoomNumberHash,
			roomEncx.Capacity, roomEncx.EquipmentEncrypted,
			roomEncx.CreatedAt, roomEncx.UpdatedAt,
			roomEncx.DEKEncrypted, roomEncx.KeyVersion, roomEncx.Metadata,
		); err != nil {
			return uuid.Nil, err
		}
		roomID = roomEncx.ID
	}
	return roomID, nil
}

// seedRoomOpeningHours opens the practice room every day from 07:00 to 22:00.
// One-off availabilities are refused on a day the room has no operating hours,
// and v1 hides buildings and rooms from the admin, so without these rows the
// partner could not open time at all. PLACEHOLDER hours: wide enough that the
// partner's own availabilities are the real constraint. Idempotent.
func seedRoomOpeningHours(ctx context.Context, db *pgxpool.Pool, roomID uuid.UUID) error {
	var existing int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM booking.room_availability_schedules WHERE room_id = $1`, roomID).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	now := time.Now()
	for day := 0; day < 7; day++ {
		if _, err := db.Exec(ctx, `
			INSERT INTO booking.room_availability_schedules (
				id, room_id, day_of_week, specific_date, open_time, close_time,
				priority, is_active, created_at, updated_at
			) VALUES ($1, $2, $3, NULL, '07:00', '22:00', 0, TRUE, $4, $4)
		`, uuid.New(), roomID, day, now); err != nil {
			return err
		}
	}
	return nil
}

func seedDedicatedAllocation(ctx context.Context, db *pgxpool.Pool, crypto encx.CryptoService, roomID, userID uuid.UUID) error {
	startDate := time.Now().UTC()
	allocation, err := bookingDomain.NewDedicatedAllocation(roomID, userID, &startDate, nil)
	if err != nil {
		return fmt.Errorf("build dedicated allocation: %w", err)
	}
	allocationEncx, err := bookingDomain.ProcessRoomAllocationEncx(ctx, crypto, allocation)
	if err != nil {
		return fmt.Errorf("encrypt allocation: %w", err)
	}

	// Re-use the partner's active allocation if the seed already ran, so a
	// re-seed never accumulates allocations.
	var allocationID uuid.UUID
	err = db.QueryRow(ctx, `
		SELECT id FROM booking.room_allocations
		WHERE user_id_hash = $1 AND is_active = TRUE
		ORDER BY created_at DESC LIMIT 1
	`, allocationEncx.UserIDHash).Scan(&allocationID)
	if err == nil {
		_, err = db.Exec(ctx, `
			UPDATE booking.room_allocations
			SET room_id = $1, allocation_type = 'dedicated', start_date = $2, end_date = NULL, updated_at = NOW()
			WHERE id = $3
		`, roomID, allocationEncx.StartDate, allocationID)
		return err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO booking.room_allocations (
			id, room_id, user_id_encrypted, user_id_hash, allocation_type,
			start_date, end_date, dek_encrypted, key_version,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NULL, $7, $8, TRUE, $9, $10
		)
	`, allocationEncx.ID, allocationEncx.RoomID, allocationEncx.UserIDEncrypted, allocationEncx.UserIDHash,
		allocationEncx.AllocationType, allocationEncx.StartDate,
		allocationEncx.DEKEncrypted, allocationEncx.KeyVersion,
		allocationEncx.CreatedAt, allocationEncx.UpdatedAt,
	)
	return err
}
