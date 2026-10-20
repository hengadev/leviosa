# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Root Level Commands
- `make help` - Show all available commands (local dev, Docker image builds, VPS deploy, Terraform, Ansible)
- `make local-dev` - Frontend natively (`pnpm run dev`) + backend and its dependencies via Docker (`backend/compose.yaml`: Postgres, Redis, RabbitMQ, and LocalStack for S3 and KMS), live reload via `air` (see `backend/local.Dockerfile`). Fastest loop for day-to-day work — edits under `backend/` rebuild and restart the server automatically via a bind mount.
- `make local-start` - Run the entire app (frontend, backend, Postgres, Redis, RabbitMQ, and LocalStack for S3 and KMS) in containers, built from the real production Dockerfiles and mirroring `infra/ansible/roles/app/templates/docker-compose.yml.j2` + `env.j2` (same image pins, same service shape minus Caddy) (`local.compose.yaml`). No hot reload — closest simulation of what actually ships to staging/production. Copies `.env.example` to `.env` on first run.
- `make local-stop` / `make local-restart` - Stop/restart the `local-start` stack

Both the `local-dev` backend stack and the `local-start` stack are intentionally ephemeral (no named volumes): stopping either wipes its Postgres and LocalStack (the S3 bucket and the KMS key holding encx's KEK) together, so a restart never has Postgres pointing at data encrypted under a KMS key that no longer exists. Each stack's LocalStack init container creates the bucket and the KEK alias (`alias/leviosa-dev`); the dev pepper is a public value in `backend/example.env` / `.env.example`. Re-seed with `go run ./cmd/seed` (from `/backend`) if you need data.

Deployment to the VPS (image build/push, Ansible, Terraform) is also driven from this Makefile — see `make help` for the full list (`image-*`, `deploy*`, `prod-*`, `staging-*`, `infra-*`, `ansible-*`).

### Frontend Commands (from /frontend)
- `pnpm run dev` - Start SvelteKit development server
- `pnpm run build` - Build for production
- `pnpm run preview` - Preview production build
- `pnpm run check` - Type checking
- `pnpm run check:watch` - Type checking in watch mode

### Backend Commands (from /backend)
- `docker compose up` - Start the backend plus its dependencies (PostgreSQL, Redis, RabbitMQ)
- `go run ./cmd/app` - Run the backend binary directly against already-running dependencies
- `go run ./cmd/seed` - Seed data (configured via `cmd/seed/seed_data.example.json`)

### Testing Commands
Run from `/backend` (see `make test-help` for the full list):
- `make test-unit` / `make test-unit-<domain>` - Unit tests, all domains or one (`authuser`, `catalog`, `settings`, `booking`, `notification`)
- `make test-integration` / `make test-integration-<domain>` - Integration tests (spin up dependencies via testcontainers)
- `make test-coverage` - HTML coverage report across all domains

## Project Architecture

### Full-Stack Application Structure
```
├── frontend/           # SvelteKit 5 application
├── backend/           # Go modular monolith
├── config/            # Infrastructure configuration (Caddy, Loki, Vault, etc.)
└── infra/             # Terraform and Ansible deployment scripts
```

### Frontend Architecture (SvelteKit 5)
- **Framework**: SvelteKit 5 with TypeScript and Tailwind CSS v4
- **Authentication**: Session-based with cookies, environment-dependent (mock in dev/staging)
- **Route Structure**:
  - `(app)/` - Protected application routes
  - `auth/` - Authentication routes
  - `legal/` - Public legal pages
- **API Integration**: Automatic Bearer token injection via hooks.server.ts
- **Key Libraries**: bits-ui, sveltekit-superforms, arktype, @internationalized/date

### Backend Architecture (Go Modular Monolith)
A single Go binary (`backend/cmd/app`) composed of internal domain packages, each following hexagonal architecture:

```
backend/
├── cmd/app/            # Main application entry point (single binary)
├── cmd/seed/            # Data seeding tool
└── internal/
    ├── common/          # Shared contracts, utilities, error definitions
    ├── authuser/        # Authentication and user management
    ├── catalog/         # Product catalog and pricing (Stripe integration)
    ├── settings/        # System configuration
    ├── booking/         # Room scheduling, availability, utilization analytics
    ├── notification/    # Email and SMS notifications
    └── messaging/       # User-to-user conversations
```

**Hexagonal Architecture Pattern** within each domain package:
- `domain/` - Business entities (no external dependencies)
- `ports/` - Repository and service interfaces
- `application/` - Use cases and workflows, often split by sub-feature (e.g. `booking/application/availability/`)
- `interface/` - Inbound adapters (mostly REST API handlers), often split by sub-feature (e.g. `catalog/interface/product/`)
- `infrastructure/` (or `adapters/` in `booking`) - Outbound adapters to external systems:
  - `postgres/` - Database persistence
  - `stripe/` - Payment provider integration
  - `rabbitmq/` - Message queue integration
  - `redis/` - Caching layer
  - `s3/` - Object storage
  - `smtp/` / `twilio/` - Email/SMS providers
  - `sse/` - Server-sent events
  - Cross-domain adapters that call other domains directly (e.g. `booking/infrastructure/authuser/`)

See `backend/CLAUDE.md` for the full breakdown, including booking-domain features (gap detection, utilization metrics, availability suggestions).

### Technology Stack

**Frontend**:
- SvelteKit 5, TypeScript, Tailwind CSS v4
- Node.js adapter for SSR

**Backend**:
- Go 1.24.2 (modular monolith, single binary)
- PostgreSQL 17.5, Redis (alpine), RabbitMQ 3 (management)
- AWS S3, AWS KMS (encx KEK; the pepper comes from configuration — ADR-0015)
- External integrations: Stripe, Gmail SMTP, Twilio SMS

**Infrastructure**:
- Docker with docker-compose
- Caddy reverse proxy
- Grafana + Loki for logging
- Prometheus for monitoring

### Key Business Domains
- **User Management**: Authentication, OTP verification, sessions
- **Product Catalog**: Categories, products, pricing with Stripe
- **Settings**: Company configuration, OTP policies, notifications
- **Events**: Event creation with capacity tracking
- **Messaging**: User-to-user conversations
- **Payments**: Stripe integration for checkout

### Database Migrations
- Convention: `{timestamp}_{service}_{action}_{entity_or_scope}.sql`
- Managed through Go embed system
- Applied only by the backend's `migrate` run mode, never on `serve` startup; every migration must work with the previous release (see the migration rule in `backend/CLAUDE.md`)

### Error Handling
- Centralized error definitions in `internal/common/errs` package
- Domain-specific error constructors
- PostgreSQL and Redis error classification utilities
- Consistent error wrapping with context

### Development Workflow
1. Start backend + dependencies: `docker compose up` (from backend directory)
2. Start frontend: `pnpm run dev` (from frontend directory)
3. Access application: Frontend typically on http://localhost:5173, backend on http://localhost:3500
4. Run tests from the backend directory using `make test-unit` / `make test-integration`

### Environment Configuration
- **Development**: Uses `.env` files and mock data
- **Production**: Environment variables with real services
- **Required Services**: PostgreSQL, Redis, RabbitMQ, AWS S3, AWS KMS
- **Docker**: Full containerized setup available via `local.compose.yaml` (`make local-start`)

### Testing Approach
- **Backend**: Unit tests alongside source files, integration tests in `test/integration/`
- **Frontend**: No testing framework currently configured
- Uses real adapters for black-box integration testing
- Test data helpers available in `test/testdata/`

### CI/CD

**CI** (`.github/workflows/ci.yaml`) runs on every PR to `main` and validates the change before merge:
- Frontend: install, build (SvelteKit)
- Backend: build, unit tests (`go test ./...`)
- Backend integration tests: `make test-integration` (from `/backend`; dependencies are spun up per-test via testcontainers)
- Security: Go vulnerability scanning (`govulncheck`)

There is no automated deployment workflow. Deployment to staging/production VPS environments is manual, driven by the root `Makefile` (Docker image build/push to Docker Hub, then Ansible playbooks or quick SSH-based `make deploy`/`make deploy-staging`) and Terraform for infrastructure provisioning. See `make help` at the repo root for the full command list.

**Merge Strategy:** Rebase and merge only (preserves full commit history)