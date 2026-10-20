# Leviosa

Leviosa is booking and activity management software for service businesses — massage therapy practices, studios, and similar appointment-based businesses that need to manage rooms, staff schedules, and clients in one place.

## Features

- **Room scheduling & availability** — booking calendar, gap detection between bookings to surface unused room time, room utilization metrics (efficiency score, fragmentation, idle time), and suggested availability block durations based on the partner's product catalog
- **Product catalog & pricing** — categories and products (treatments, classes, events) with capacity tracking and Stripe-backed pricing/checkout
- **Payments** — Stripe integration for checkout and customer billing
- **Client messaging** — direct conversations between businesses and their clients
- **Notifications** — booking confirmations and reminders over email (Gmail SMTP) and SMS (Twilio)
- **Auth & accounts** — email/OTP sign-in, sessions, and Google/Apple OAuth
- **Settings** — per-business configuration (OTP policy, notification preferences, etc.)

## Architecture

```
├── frontend/            # SvelteKit 5 application
├── backend/             # Go modular monolith
├── config/              # Infrastructure configuration (Loki, Vault, etc.)
├── infra/               # Terraform and Ansible deployment scripts
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
- **Notifications**: Gmail SMTP, Twilio SMS.

## Local Development

Two ways to run the app locally, depending on what you need:

```bash
make local-dev     # fast loop: frontend natively (pnpm), backend + deps in Docker (Postgres,
                    # Redis, RabbitMQ, S3 and KMS via LocalStack), live reload via air
make local-start   # whole app in Docker (frontend, backend, + the same deps), built from
                    # the real production Dockerfiles — no hot reload, mirrors what Ansible
                    # actually deploys to staging/production (infra/ansible/roles/app)
make local-stop    # stop the local-start stack
make local-restart # restart it
```

`make local-start` copies `.env.example` to `.env` on first run, and `make local-dev` copies `backend/example.env` to `backend/development.env` — both have working local defaults, you only need to fill in real Stripe/Twilio/SMTP credentials if you want to exercise those integrations. Both stacks are intentionally ephemeral: stopping either clears its Postgres and LocalStack (S3 bucket and the KMS key holding the encryption KEK) together, so a restart never leaves data encrypted under a KMS key that no longer exists. Re-seed with `go run ./cmd/seed` (from `/backend`) if you need data to work with.

See `make help` at the repo root for the full list of commands (local dev, Docker image builds, VPS deploy, Terraform, Ansible). Frontend- and backend-specific commands are documented in `frontend/CLAUDE.md` and `backend/CLAUDE.md`.
