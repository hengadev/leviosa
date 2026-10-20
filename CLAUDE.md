# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Root Level Commands
- `make help` - Show all available commands (local dev, local deploy harnesses, server setup, Terraform)
- `make local-dev` - Frontend natively (`pnpm run dev`) + backend and its dependencies via Docker (`backend/compose.yaml`: Postgres, Redis, RabbitMQ, and LocalStack for S3 and KMS), live reload via `air` (see `backend/local.Dockerfile`). Fastest loop for day-to-day work — edits under `backend/` rebuild and restart the server automatically via a bind mount.
- `make local-start` - Run the entire app (frontend, backend, Postgres, Redis, RabbitMQ, and LocalStack for S3 and KMS) in containers, mirroring `infra/compose/docker-compose.prod.yml` (ADR-0016): same service shape minus Caddy, both app images built from the real production Dockerfiles with `COMMIT_SHA=local`, and the `migrate` and `admin` (`bootstrap-admin`) one-offs running to completion before the backend starts, as a deploy runs them on the server. No hot reload — closest simulation of what actually ships to staging/production. Copies `.env.example` to `.env` on first run; the per-consumer env split of the server stacks is preserved (the frontend container receives no backend secret, `ADMIN_*` reaches only the admin one-off).
- `make local-stop` / `make local-restart` - Stop/restart the `local-start` stack
- `make env` - Render `backend/development.env` and `frontend/.env` from Infisical `dev` with the server's Agent templates (overwrites both; `infisical login` first). Without Infisical, the committed example files are the path (integrations off)
- `make env-check-<dev|staging|prod>` - Read-only: render that environment's files from Infisical and list key names and problems, never values
- `make stripe-listen` - Forward Stripe test-mode webhooks to the local backend

Both the `local-dev` backend stack and the `local-start` stack are intentionally ephemeral (no named volumes): stopping either wipes its Postgres and LocalStack (the S3 bucket and the KMS key holding encx's KEK) together, so a restart never has Postgres pointing at data encrypted under a KMS key that no longer exists. Each stack's LocalStack init container creates the bucket and the KEK alias (`alias/leviosa-dev`); the dev pepper is a public value in `backend/example.env` / `.env.example`. The `local-start` stack bootstraps a superadmin from the `ADMIN_*` values in `.env` (log in at `/auth` with them); re-seed with `go run ./cmd/seed` (from `/backend`) if you need more data.

Server setup (`make setup`: Ansible, see `infra/ansible/README.md`; run it twice, the second run must report `changed=0`) and infrastructure (`make infra-init|plan|apply|output`: Terraform through `infra/terraform/tf.sh`, credentials from Infisical `leviosa-infra`) are also driven from this Makefile. The Infisical Agent alone (install, or the 3-monthly credential rotation) is `make infisical-agent` from `infra/ansible`. Local tests of the deployment pieces, Docker/shell only, nothing remote: `make deploy-harness` (the deploy script against the real compose files), `make deploy-gate-test`, `make agent-templates-test`, `make infisical-agent-test`, `make backup-test`, `make promote-preflight-test`, and `docs/deployment/check-environment.sh` (every template key is in the environment reference). Releases are not make targets: CI runs `deploy <stack> <sha>` through the deploy gate (`infra/compose/README.md`). The deployment model is in `docs/deployment/README.md`.

### Frontend Commands (from /frontend)
- `pnpm run dev` - Start SvelteKit development server
- `pnpm run build` - Build for production
- `pnpm run preview` - Preview production build
- `pnpm run check` - Type checking
- `pnpm run check:watch` - Type checking in watch mode

### Backend Commands (from /backend)
- `docker compose up` - Start the backend plus its dependencies (PostgreSQL, Redis, RabbitMQ, LocalStack for S3 and KMS)
- `go run ./cmd/app [mode]` - Run the backend binary directly against already-running dependencies. Modes: `serve` (default; never migrates, refuses an older schema), `migrate`, `bootstrap-admin` (the first superadmin from `ADMIN_*`), `verify` (read-only decrypt + pepper canary check), `check-sms <to>` / `check-email <to>` (one test message through the real provider settings). See `backend/CLAUDE.md`
- `go run ./cmd/seed` - Seed data (configured via `cmd/seed/seed_data.example.json`)

### Testing Commands
Run from `/backend` (see `make test-help` for the full list):
- `make test-unit` / `make test-unit-<domain>` - Unit tests, all domains or one (`authuser`, `catalog`, `settings`, `booking`, `notification`)
- `make test-integration` / `make test-integration-<domain>` - Integration tests (spin up dependencies via testcontainers)
- `make test-coverage` - HTML coverage report across all domains

## Project Architecture

### Full-Stack Application Structure
```
├── frontend/            # SvelteKit 5 application
├── backend/             # Go modular monolith
├── infra/
│   ├── compose/         # Prod and staging compose files, deploy script, Caddyfile, local harness
│   ├── infisical/       # Infisical Agent env templates, make env, envcheck
│   ├── ansible/         # Server setup only (roles: deploy gate, Infisical Agent, shared services, backup, ...)
│   └── terraform/       # Hetzner, Cloudflare, S3, KMS, IAM; run through tf.sh
├── docs/                # ADRs (docs/adr), deployment docs (docs/deployment), runbooks
└── local.compose.yaml   # Full local stack mirroring the production compose file (make local-start)
```

### Frontend Architecture (SvelteKit 5)
- **Framework**: SvelteKit 5 with TypeScript and Tailwind CSS v4
- **Authentication**: Session-based with cookies; mock mode (`USE_MOCK_DATA=true`) is local-only, never set in a rendered env file
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
- External integrations: Stripe, SMTP (Zoho Mail), Twilio SMS (IE1 region, API key)

**Infrastructure**:
- One Hetzner VPS, two Docker Compose stacks (production owns Postgres, RabbitMQ and Caddy; staging runs its own backend, frontend and Redis) (ADR-0016)
- Caddy reverse proxy (TLS through Cloudflare DNS; `/webhooks/*` to the backend, everything else to the frontend)
- Infisical as the only source of configuration, rendered on the server by the Infisical Agent (ADR-0017); encx KEK in AWS KMS, pepper in Infisical (ADR-0015)
- Ansible (server setup only) and Terraform (infrastructure only)

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
- **Development**: committed example env files (no Infisical, integrations off), or `make env` from Infisical `dev`
- **Staging / Production**: env files rendered from Infisical by the Infisical Agent, one per consumer, root-only on the server (`docs/deployment/environment.md`)
- **Required Services**: PostgreSQL, Redis, RabbitMQ, AWS S3, AWS KMS
- **Docker**: Full containerized setup available via `local.compose.yaml` (`make local-start`)

### Testing Approach
- **Backend**: Unit tests alongside source files, integration tests in `test/integration/`
- **Frontend**: No testing framework currently configured
- Uses real adapters for black-box integration testing
- Test data helpers available in `test/testdata/`

### CI/CD

**CI** (`.github/workflows/ci.yaml`) runs on every PR to `main` and every push to `main`:
- Frontend: install, build (SvelteKit)
- Backend: build, unit tests (`go test ./...`)
- Backend integration tests: `make test-integration` (from `/backend`; dependencies are spun up per-test via testcontainers)
- Security: Go vulnerability scanning (`govulncheck`)
- Deploy scripts: deploy gate, Promote preflight and Agent template tests

On a push to `main` that passes every check, CI builds and pushes `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>` and deploys them to staging (`deploy staging <sha>`, `docs/deployment/ci.md`). Production is released only by the manual **Promote** workflow (`.github/workflows/promote.yaml`), which ships the SHA staging runs; a rollback is a Promote of an older SHA (`docs/deployment/promote.md`).

Releases never go through Ansible or an operator's computer: the single release entry point is `deploy <stack> <sha>`, which the CI key can run only through the deploy gate (`infra/compose/README.md`, `infra/ansible/roles/deploy_gate/README.md`). Ansible (`make setup`) only sets the server up; Terraform (`make infra-*`) only creates infrastructure. See `make help` at the repo root for the full command list.

**Merge Strategy:** Rebase and merge only (preserves full commit history)