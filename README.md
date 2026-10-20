# Leviosa

Leviosa is booking and activity management software for service businesses — massage therapy practices, studios, and similar appointment-based businesses that need to manage rooms, staff schedules, and clients in one place.

## Features

- **Room scheduling & availability** — booking calendar, gap detection between bookings to surface unused room time, room utilization metrics (efficiency score, fragmentation, idle time), and suggested availability block durations based on the partner's product catalog
- **Product catalog & pricing** — categories and products (treatments, classes, events) with capacity tracking and Stripe-backed pricing/checkout
- **Payments** — Stripe integration for checkout and customer billing
- **Client messaging** — direct conversations between businesses and their clients
- **Notifications** — booking confirmations and reminders over email (SMTP, Zoho Mail) and SMS (Twilio)
- **Auth & accounts** — email/OTP sign-in, sessions, and Google/Apple OAuth
- **Settings** — per-business configuration (OTP policy, notification preferences, etc.)

## Architecture

```
├── frontend/            # SvelteKit 5 application
├── backend/             # Go modular monolith
├── infra/
│   ├── compose/         # Production and staging compose files, the deploy script, Caddy
│   ├── infisical/       # Infisical Agent env templates, make env
│   ├── ansible/         # Server setup (never releases the app)
│   └── terraform/       # Server, DNS, S3, KMS, IAM (through tf.sh)
├── docs/                # ADRs, deployment docs, runbooks
└── local.compose.yaml   # Full local stack simulation (see Local Development below)
```

- **Frontend**: [SvelteKit 5](https://kit.svelte.dev/) + TypeScript + Tailwind CSS v4.
- **Backend**: [Go](https://golang.org/) modular monolith — a single binary built from `backend/cmd/app`, with business domains (`authuser`, `booking`, `catalog`, `settings`, `notification`, `messaging`) organized as internal packages under `backend/internal/`, each following a hexagonal (ports & adapters) structure. See `backend/CLAUDE.md` for details.
- **Database**: PostgreSQL.
- **Cache**: Redis.
- **Message queue**: RabbitMQ.
- **File storage**: AWS S3.
- **Encryption keys**: AWS KMS for the encx KEK, with the pepper supplied as configuration (ADR-0015).
- **Payments**: Stripe.
- **Configuration**: Infisical, the only source of every secret and environment value (ADR-0017).
- **Notifications**: SMTP (Zoho Mail), Twilio SMS (EU region IE1).

## Local Development

Two ways to run the app locally, depending on what you need:

```bash
make local-dev     # fast loop: frontend natively (pnpm), backend + deps in Docker (Postgres,
                    # Redis, RabbitMQ, S3 and KMS via LocalStack), live reload via air
make local-start   # whole app in Docker (frontend, backend, + the same deps), built from
                    # the real production Dockerfiles — no hot reload, mirrors
                    # infra/compose/docker-compose.prod.yml (migrate + admin one-offs run
                    # before the apps start, as a deploy runs them)
make local-stop    # stop the local-start stack
make local-restart # restart it
```

`make local-start` copies `.env.example` to `.env` on first run, and `make local-dev` copies `backend/example.env` to `backend/development.env` — both have working local defaults and leave the integrations off. With an Infisical login, `make env` renders `backend/development.env` and `frontend/.env` from Infisical `dev` with the server's own templates instead (it overwrites both). Both stacks are intentionally ephemeral: stopping either clears its Postgres and LocalStack (S3 bucket and the KMS key holding the encryption KEK) together, so a restart never leaves data encrypted under a KMS key that no longer exists. Re-seed with `go run ./cmd/seed` (from `/backend`) if you need data to work with.

See `make help` at the repo root for the full list of commands (local dev, `make env`, the local deploy harnesses and template tests, server setup with `make setup`, Terraform with `make infra-*`; releases go through CI, not make). Frontend- and backend-specific commands are documented in `frontend/CLAUDE.md` and `backend/CLAUDE.md`.

## Deployment

One Hetzner server runs two Docker Compose stacks, production and staging. Every configuration value comes from Infisical through the Infisical Agent, a push to `main` deploys staging through CI, and the manual **Promote** workflow ships the release staging runs to production (a rollback is a Promote of an older SHA).

- [docs/deployment/README.md](docs/deployment/README.md): the deployment model, releases and rollback, the migration rule, `COMING_SOON`, Infisical outages, backups and recovery, credential rotation, the public repository
- [docs/deployment/environment.md](docs/deployment/environment.md): every configuration value, its Infisical folder, its consumer and who writes it
- [docs/deployment/setup.md](docs/deployment/setup.md): setting up a new operator computer
- [docs/deployment/ci.md](docs/deployment/ci.md) and [docs/deployment/promote.md](docs/deployment/promote.md): CI, the staging deploy, Promote and rollback
