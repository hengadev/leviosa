# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
- `docker compose up --build` - Start the backend plus its dependencies (PostgreSQL, Redis, RabbitMQ, and LocalStack for S3 and KMS — its init container creates the bucket and the encx KEK alias); builds from `local.Dockerfile`, which runs `air` for live reload — the source tree is bind-mounted, so edits rebuild and restart the server automatically. Needs `development.env` (copy from `example.env`; `make local-dev` from the repo root does this for you).
- `go run ./cmd/app migrate` - Apply pending database migrations (`serve` never migrates; see Database Migrations below)
- `go run ./cmd/app` - Run the backend binary (`serve` mode) directly against already-running, migrated dependencies
- `go run ./cmd/seed` - Seed data (configured via `cmd/seed/seed_data.example.json`)

### Testing
Run from this directory (see `make test-help` for the full list):
- `make test-unit` / `make test-unit-<domain>` - Unit tests, all domains or one (`authuser`, `catalog`, `settings`, `booking`, `notification`)
- `make test-integration` / `make test-integration-<domain>` - Integration tests; dependencies are spun up per-test via testcontainers, no manual container management needed
- `make test-coverage` - HTML coverage report across all domains
- `make test-race` / `make test-benchmark` / `make test-smoke` - Other test modes
- `make test-file TEST_PATH=<path>` / `make test-func TEST_NAME=<name> TEST_PATH=<path>` - Run a single file or test function

## Project Architecture

### Modular Monolith Structure
A single Go binary (`cmd/app`), with business domains organized as internal packages following hexagonal architecture:

```
├── cmd/app/         # Main application entry point (single binary)
├── cmd/seed/        # Data seeding tool
└── internal/
    ├── common/      # Shared contracts, utilities, error definitions
    ├── authuser/    # Authentication and user management
    ├── catalog/     # Product catalog and pricing service
    ├── settings/    # System configuration service
    ├── booking/     # Room scheduling, availability, utilization analytics
    ├── notification/ # Email and SMS notification service
    └── messaging/   # User-to-user conversations
```

### Hexagonal Architecture Pattern
Each domain package follows the ports & adapters pattern:
- **domain/** - Business entities and value objects (no external dependencies)
- **ports/** - Interfaces for repositories and services
- **application/** - Use cases and business workflows
- **infrastructure/** - Infrastructure implementations:
  - `http/` - REST API handlers and routes
  - `postgres/` - Database persistence
  - `rabbitmq/` - Message queue integration
  - `redis/` - Caching layer
  - `s3/` - Object storage

### Common Package
`internal/common/` contains shared components:
- **errs/** - Centralized error definitions and constructors
- **contracts/** - RabbitMQ message contracts and routing keys
- **messaging/** - RabbitMQ utilities for exchanges, queues, and payloads
- **ctxutil/** - Context utilities for logger and role validation
- **encryption/** - `NewCrypto`: the one way every entry point builds encx (KEK in AWS KMS by alias, pepper from `ENCX_PEPPER`; refuses to create keys or store a pepper)
- **httpx/** - HTTP utilities for CORS, JSON responses, error handling
- **logger/** - Structured logging configuration
- **middleware/** - Authentication middleware
- **validation/** - Email and phone validation utilities
- **testutils/** - Testcontainers helpers for Postgres, Redis, RabbitMQ, S3, KMS (LocalStack, `SetupKMS`), Stripe, and the legacy Vault helper still used by most integration suites

### Database Migrations
- Located in `internal/common/migrations/` (goose SQL files, embedded into the binary)
- Convention: `{timestamp}_{service}_{action}_{entity_or_scope}.sql`
- Example: `20250714103022_catalog_add_column_products_buffer_time.sql`
- Applied only by the `migrate` run mode (`go run ./cmd/app migrate`, or `leviosa migrate` in the image), which exits non-zero and names the migration that failed. Deploys run it as a one-off before the new backend container starts; the dev stack runs it before every air restart (`.air.toml`).
- `serve` never migrates. At startup it refuses to run if the database's migration version is **lower** than the newest embedded migration (the migrate step was skipped) and accepts a **higher** one (a rollback to an older release).

**Migration rule: every migration must work with the previous release — add first, remove later.** A deploy migrates while the previous release is still serving, and a rollback redeploys an older release without reverting the schema (down migrations are never run), so the previous release's code must keep working against the new schema:
- Add tables, and add columns as nullable or with a default. Don't add a constraint (`NOT NULL`, `CHECK`, unique) that the previous release's writes would violate.
- Rename or drop in steps: add the new column/table, ship the code that uses it and stops using the old one, and only drop the old one in a later release, once no release you might roll back to reads it.
- Never edit a migration that has already been applied anywhere; write a new one.

### Testing Structure
- **Unit tests**: Alongside source files (`*_test.go`)
- **Integration tests**: `test/integration/` directories per domain
- **Test data**: `test/testdata/` and `internal/common/testutils/` with helpers for database, RabbitMQ, HTTP setup
- Uses real adapters (via testcontainers) for black-box testing

### Technology Stack
- **Language**: Go 1.24.2
- **Database**: PostgreSQL 17.5
- **Cache**: Redis (alpine)
- **Message Queue**: RabbitMQ 3 (management)
- **Object Storage**: AWS S3
- **Configuration**: Viper + environment variables

### Environment Configuration
- Development uses `development.env` file
- Production uses environment variables
- Required services: PostgreSQL, Redis, RabbitMQ, AWS S3, AWS KMS
- Encryption (encx, ADR-0015): `ENCX_KMS_KEY_ALIAS` (an alias name, `alias/<name>`, never a key ID or ARN), `ENCX_PEPPER` (base64 of 32 bytes), `AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_ENDPOINT_URL` for LocalStack only. A missing or invalid pepper, or an alias that does not resolve, stops startup
- External integrations: Stripe, Gmail SMTP, Twilio SMS

### Error Handling
- Sentinel errors defined in `internal/common/errs/` package
- Domain-specific error constructors (e.g., `NewNotFoundErr`, `NewConflictErr`)
- PostgreSQL and Redis error classification utilities
- Consistent error wrapping with context

### Key Business Domains
- **User Management**: Authentication, OTP verification, session management
- **Product Catalog**: Categories, products, pricing with Stripe integration
- **Settings**: Company configuration, OTP policies, notification preferences
- **Events**: Event creation and management with capacity tracking
- **Messaging**: User-to-user conversations and notifications
- **Payments**: Stripe integration for pricing and checkout
- **Booking**: Room scheduling, availability management, utilization analytics

### Booking Service Features

The booking service provides advanced scheduling optimization for massage therapy practices:

#### 1. Gap Detection API
**Purpose**: Identifies unused time slots between bookings to maximize room utilization

**Key Features**:
- Analyzes room schedules for a specific date
- Finds gaps before first booking, between bookings, and after last booking
- Suggests products that fit within each gap (Duration + BufferTime <= gap)
- Sorts suggestions by duration (shortest first for flexibility)
- Respects room operating hours

**Endpoints**:
- `GET /availabilities/rooms/{room_id}/gaps?date=YYYY-MM-DD`

**Algorithm**: See `internal/booking/application/availability/get_room_gaps.go`

#### 2. Utilization Metrics
**Purpose**: Track room efficiency and identify scheduling improvements

**Key Features**:
- Daily metrics stored in materialized view (`booking.room_daily_metrics`)
- Tracks utilization percentage, fragmentation count, idle minutes
- Calculates efficiency score (utilization - fragmentation penalty - idle penalty)
- Supports room-level and partner-level aggregation
- Historical analysis with date ranges

**Endpoints**:
- `GET /rooms/{room_id}/metrics?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`
- `GET /partners/{partner_id}/metrics?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`

**Database**: Materialized view uses CTEs for complex calculations
**Migration**: `20251124110535_booking_add_metrics_materialized_view.sql`

#### 3. Availability Suggestions
**Purpose**: Recommend optimal availability block durations based on products

**Key Features**:
- Analyzes partner's product catalog (Duration + BufferTime)
- Generates single and multi-session block suggestions (1x, 2x, 3x, 4x)
- For shared rooms, suggests standard durations (60, 90, 120, 180, 240 min)
- Priority ranking:
  * Priority 1: Standard durations (60, 90, 120, 180 min)
  * Priority 2: 30-minute increments
  * Priority 3: Non-standard durations
- Consolidates multiple products suggesting same duration

**Endpoints**:
- `GET /partners/{partner_id}/rooms/{room_id}/suggest-blocks`

**Algorithm**: See `internal/booking/application/availability/suggestion_algorithm.go`

#### Error Handling Pattern
All booking handlers use `httpx.RespondWithServiceError()` for automatic error classification:
- Repository layer: `errs.ClassifyPgError()` converts DB errors to sentinel errors
- Service layer: Wraps errors with `fmt.Errorf` or returns sentinel errors
- Handler layer: `RespondWithServiceError()` maps to HTTP status codes (400/404/503/500)

### Development Workflow
1. Start the backend and its dependencies: `docker compose up`
2. Run tests: `make test-unit` / `make test-integration` (or the per-domain variants)
3. Database migrations are embedded goose files applied by the `migrate` run mode; follow the migration rule under Database Migrations
