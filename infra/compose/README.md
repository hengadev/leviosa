# Compose deployment

Everything that releases Leviosa to the VPS, as plain files in the repository
(ADR-0016, ADR-0017; PRD `docs/prd/deployment-infisical-compose.md`): no
Ansible templating, no `${VAR}` substitution, no top-level `.env` on the server.

| File | On the server |
| --- | --- |
| `docker-compose.prod.yml` | installed by each production deploy as `/opt/leviosa/docker-compose.yml`, from the deployed commit (root runner, issue 022) |
| `docker-compose.staging.yml` | installed by each staging deploy as `/opt/leviosa-staging/docker-compose.yml`, the same way |
| `deploy` | never copied: the root runner runs it straight from the deployed commit |
| `test/` | stays in the repo: the local harness below |

The server reads both files from the repository at the SHA it deploys, so a
rollback rolls them back too, and CI uploads nothing.

## Production stack root

```
/opt/leviosa/                     compose project `leviosa`
├── docker-compose.yml            docker-compose.prod.yml from the deployed commit
├── env/                          Infisical Agent renderings, root 0600 (issues 024/027)
│   ├── backend.env               backend + migrate + admin one-offs
│   ├── frontend.env              frontend only
│   ├── admin.env                 ADMIN_* only — the admin one-off alone
│   ├── postgres.env              POSTGRES_USER / POSTGRES_PASSWORD / POSTGRES_DB
│   ├── rabbitmq.env              RABBITMQ_DEFAULT_USER / _PASS / _VHOST
│   ├── redis.env                 REDIS_PASSWORD
│   └── caddy.env                 CLOUDFLARE_API_TOKEN (+ staging basic auth, issue 020)
├── data/postgres/                Postgres data dir      (survives `down`)
├── data/rabbitmq/                RabbitMQ node data     (survives `down`)
├── data/redis/                   Redis AOF + RDB        (survives `down`)
└── caddy/{etc,data,config,logs}/ Caddyfile (Ansible), certificates, autosave, logs
```

All paths in the compose file are relative to the stack root, so the same file
works wherever the root is (the harness uses a temp dir).

### Services

| Service | Image | Env files | Notes |
| --- | --- | --- | --- |
| `backend` | `henga/leviosa-backend:prod` | `backend.env` | `serve`; healthcheck `/health` (busybox wget) on 3500 |
| `frontend` | `henga/leviosa-frontend:prod` | `frontend.env` | distroless: node healthcheck on `/healthz` (3000) |
| `migrate` | `henga/leviosa-backend:prod` | `backend.env` | one-off (`oneoff` profile): `./leviosa migrate` |
| `admin` | `henga/leviosa-backend:prod` | `backend.env`, `admin.env` | one-off (`oneoff` profile): `./leviosa bootstrap-admin` |
| `postgres` | `postgres:17-alpine` | `postgres.env` | shared with staging (own role and database each) |
| `rabbitmq` | `rabbitmq:3-management-alpine` | `rabbitmq.env` | shared with staging (own vhost and user each); fixed `hostname` so its data survives recreation |
| `redis` | `redis:7-alpine` | `redis.env` | production's own; password via an in-container config file, not argv |
| `caddy` | `serfriz/caddy-cloudflare:2.11` | `caddy.env` | the only service publishing host ports (80, 443, 443/udp) |

What each env file holds:

- **`backend.env`**: `ENVIRONMENT=production` (always explicit: `bootstrap-admin`
  falls back to development defaults when it is unset), `SERVER_PORT=3500`,
  `DATABASE_URL` (`postgres://…@postgres:5432/<db>`), `REDIS_ADDR=redis:6379`,
  `REDIS_PASSWORD`, `REDIS_DB`, `RABBITMQ_URL` (`amqp://…@rabbitmq:5672/<vhost>`),
  `ENCX_KMS_KEY_ALIAS`, `ENCX_PEPPER`, `AWS_REGION`, `AWS_ACCESS_KEY_ID`,
  `AWS_SECRET_ACCESS_KEY`, `S3_*`, `SESSION_SECRET`, `BOOKING_TOKEN_SECRET`,
  `FRONTEND_ORIGIN`, `STRIPE_*`, `TWILIO_*`, `SMTP_*`, `LOG_FORMAT`.
  Never `ADMIN_*`, never `AWS_ENDPOINT_URL` (LocalStack only).
- **`frontend.env`**: `NODE_ENV`, `PORT=3000`, `API_URL=http://backend:3500`,
  `PROTOCOL_HEADER=x-forwarded-proto`, `HOST_HEADER=host`,
  `SESSION_COOKIE_NAME`, `CLIENT_IP_HEADER`, `COMING_SOON`, `USE_MOCK_DATA=false`
  (and the optional `PUBLIC_PLAUSIBLE_*`). Never `ORIGIN`, never a backend secret.
- **`admin.env`**: `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_FIRST_NAME`,
  `ADMIN_LAST_NAME`, `ADMIN_TELEPHONE`, `ADMIN_POSTAL_CODE`, `ADMIN_CITY`,
  `ADMIN_ADDRESS1`, `ADMIN_ADDRESS2`.
- **`postgres.env`**, **`rabbitmq.env`**: applied when the data directory is
  first initialised only; rotating those passwords is a manual change inside
  the service (`ALTER ROLE`, `rabbitmqctl change_password`).

`test/fixtures/prod/env/` has one fake example of each.

### Network and DNS aliases

The stack joins the external `leviosa_network` (created by Ansible's docker
role, shared with staging), so a `down` here never removes the network staging
depends on. Compose registers every service name as a DNS alias on it: the
frontend reaches the backend as `backend:3500`, the backend reaches
`postgres`, `rabbitmq` and `redis` by name, and Caddy proxies to `frontend`
(and, issue 020, `backend` for `/webhooks/*`). Staging's services are therefore
prefixed `staging_` (see [Staging stack root](#staging-stack-root)) so they
never answer for those names. Without `container_name`, containers are named
`leviosa-<service>-1`.

### Image pins

`henga/leviosa-backend:prod` and `henga/leviosa-frontend:prod` are local tags,
never pushed. Each deploy pulls `:<sha>` and re-points them before anything
runs, so the one-offs and any later recreate resolve production's release and
nothing else (a shared mutable tag would let a staging deploy leak into a
production recreate). A running container keeps its image even when the pin
moves.

### Recreating a service

Anything that recreates a single service (the Infisical Agent on a config
change, an operator) must use `--no-deps`, or Compose also re-evaluates the
service's dependencies and may recreate Postgres or RabbitMQ:

```sh
docker compose --project-directory /opt/leviosa up -d --no-deps --force-recreate backend
```

## Staging stack root

Staging runs on the same server, beside production, and cannot reach
production's data (ADR-0016). It brings only its backend, frontend and its own
Redis; it uses production's Postgres and RabbitMQ with credentials of its own.

```
/opt/leviosa-staging/             compose project `leviosa-staging`
├── docker-compose.yml            docker-compose.staging.yml from the deployed commit
├── env/                          Infisical Agent renderings (Infisical `staging`), root 0600
│   ├── backend.env               staging_backend + staging_migrate + staging_admin
│   ├── frontend.env              staging_frontend only
│   ├── admin.env                 ADMIN_* only — staging_admin alone
│   └── redis.env                 REDIS_PASSWORD (staging's own)
└── data/redis/                   staging Redis AOF + RDB
```

| Service | Image | Env files | Notes |
| --- | --- | --- | --- |
| `staging_backend` | `henga/leviosa-backend:staging` | `backend.env` | depends on `staging_redis` only |
| `staging_frontend` | `henga/leviosa-frontend:staging` | `frontend.env` | `API_URL=http://staging_backend:3500` |
| `staging_migrate` | `henga/leviosa-backend:staging` | `backend.env` | one-off (`oneoff` profile) |
| `staging_admin` | `henga/leviosa-backend:staging` | `backend.env`, `admin.env` | one-off (`oneoff` profile) |
| `staging_redis` | `redis:7-alpine` | `redis.env` | staging's own; Redis has no per-database access control |

No staging service publishes a host port; Caddy (production's stack) proxies
to `staging_frontend` (and, issue 020, `staging_backend` for `/webhooks/*`).

What staging's `backend.env` points at — same keys as production's, other
values (and never a production credential):

- `ENVIRONMENT=staging`;
- `DATABASE_URL=postgres://<staging role>:…@postgres:5432/<staging db>`:
  production's Postgres, staging's own role and database;
- `RABBITMQ_URL=amqp://<staging user>:…@rabbitmq:5672/<staging vhost>`:
  production's RabbitMQ, staging's own vhost and user;
- `REDIS_ADDR=staging_redis:6379` with staging's `REDIS_PASSWORD`;
- its own `ENCX_KMS_KEY_ALIAS` and `ENCX_PEPPER`, AWS keys, bucket, session and
  booking-token secrets.

The role, database, vhost and user are created on production's services by
Ansible's shared services role (issue 026), from Infisical `staging/db` and
`staging/rabbitmq`: the role owns the staging database, `CONNECT` on
production's database is revoked from `PUBLIC` (production's role is the
instance's superuser), and the RabbitMQ user has no tags and permissions on
its own vhost only. `test/fixtures/create-staging-services.sh` does exactly
that for the harness and is the reference for that role.

Isolation rests on two rules:

- **Prefixed service names.** Both stacks join `leviosa_network`, where every
  service name is a DNS alias. A staging service named `backend` would answer
  for production's `backend` too, and Docker would round-robin production's
  traffic to staging. Staging's names all start with `staging_`.
- **Separate credentials, not separate networks.** Staging can reach
  production's Postgres, RabbitMQ and Redis containers on the network; it
  holds no credential that opens production's database, vhost or Redis.

A production outage of Postgres or RabbitMQ also stops staging; a `down` of
either stack never touches the other (separate compose projects, external
network).

## `deploy <stack> <sha>`

The single release entry point: staging deploys, promotions and rollbacks
(`deploy prod <older-sha>`), for both stacks: `deploy prod <sha>` and
`deploy staging <sha>`. It runs as root (the env files are root-only) and each
step gates the next:

1. pull `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>`, then
   re-point the stack's pins (`:prod` or `:staging`);
2. `up -d --no-recreate --wait` the stack's own services: `postgres rabbitmq
   redis` for prod, `staging_redis` alone for staging. It starts what is
   missing (first deploy on empty data directories) and never recreates a
   running one; a staging deploy never starts, stops or recreates a production
   container;
3. `run --rm migrate` (`staging_migrate`): on failure the pins move back to the
   running images and the deploy stops, so the running containers keep serving
   and a later recreate cannot start the failed release. Staging needs
   production's Postgres and RabbitMQ: when they are not running, a staging
   deploy fails here and says so;
4. `run --rm admin` (`staging_admin`; `bootstrap-admin`, idempotent: creates
   the superadmin once, never modifies an existing account);
5. recreate the backend, then the frontend (`--no-deps --force-recreate --wait`);
6. require `/healthz` (asked inside the stack's frontend container) to report
   `<sha>` for both apps; a different SHA fails at once, no answer fails after
   the timeout.

It refuses anything but a 7-to-40-character hex SHA, checks the stack root,
compose file, env files and network before touching anything, and holds a
per-stack lock (`<root>/.deploy.lock`) so two deploys of a stack cannot
interleave (a staging and a production deploy may run side by side: their pins
and projects are distinct).

Knobs (production defaults; the root runner starts `deploy` with an empty
environment, so CI can set none of them):

| Variable | Default | Purpose |
| --- | --- | --- |
| `LEVIOSA_PROD_ROOT` | `/opt/leviosa` | prod stack root |
| `LEVIOSA_STAGING_ROOT` | `/opt/leviosa-staging` | staging stack root |
| `LEVIOSA_BACKEND_REPO` | `henga/leviosa-backend` | backend image repository |
| `LEVIOSA_FRONTEND_REPO` | `henga/leviosa-frontend` | frontend image repository |
| `LEVIOSA_NETWORK` | `leviosa_network` | external network the preflight checks |
| `LEVIOSA_HEALTH_TIMEOUT` | `180` | seconds for each health wait |
| `LEVIOSA_SKIP_PULL=1` | unset | use local images, no pull — local harness only |

Compose's own `COMPOSE_FILE` / `COMPOSE_PROJECT_NAME` are honoured as usual;
the harness uses them to layer its override and project name.

## Local harness

```sh
infra/compose/test/run-harness.sh        # or: make deploy-harness
```

One command, real Docker, no secrets, nothing pushed. It builds both images
with fake 40-hex `COMMIT_SHA`s, lays out temporary prod and staging stack
roots with the fixture env files (`fixtures/prod/env/`, `fixtures/staging/env/`),
runs LocalStack (KMS only) with one encx KEK alias per environment, and drives
the real `deploy` through the acceptance cases.

Production: first deploy (migrate, admin, both SHAs), second deploy (data
kept, admin byte-for-byte unchanged even with different `ADMIN_*` values), a
failing migration (an image built from a temporary copy of `backend/` plus
`fixtures/broken-migration.sql`: nothing recreated, previous release still
serving, pins restored), a SHA mismatch on either app (deploy fails;
redeploying a good SHA recovers), no `ADMIN_*` in the running containers, data
surviving `down` and a fresh deploy.

Staging beside production: before production exists, a staging deploy fails
at migrate and starts nothing of production's. Then
`fixtures/create-staging-services.sh` creates the staging role, database,
vhost and user on production's services (twice: it is idempotent), staging is
deployed at another SHA than production, and the harness proves that each
`/healthz` reports its own SHA (also through Caddy, by host); that every
production name (`backend`, `frontend`, `redis`, …) resolves from production's
frontend to production's container only and no DNS name on the network
answers for two containers; that the staging role is refused on production's
database for lack of `CONNECT` (and connects to its own); that the staging
RabbitMQ user is refused production's vhost (`530 NOT_ALLOWED`, from a raw
AMQP handshake, and opens its own); that no value in staging's env files or
running containers contains a production Redis, Postgres or RabbitMQ password
(only key names are printed) and production's Redis refuses staging's
password; and that two staging deploys leave every production container's
id, start time and restart count, and the `:prod` pins, unchanged.

Finally it checks that nothing is left behind. It ends with a PASS/FAIL line
per case and `RESULT: OK` or `RESULT: FAILED` (non-zero exit).

Isolation, so it can run beside a developer's stacks (`make local-dev`) and
other harness runs:

- every name carries a random run id: compose projects
  `leviosa-harness-<id>-prod` and `-staging`, network `leviosa-harness-<id>`,
  LocalStack container, and image repositories `leviosa-harness-<id>-backend` /
  `-frontend` (the `henga/leviosa-*` pins are never touched);
- the only deviations from the real files are generated overrides (those image
  repositories with the same `:prod` / `:staging` pin tags, the network name,
  Caddy on a random `127.0.0.1` port) and the `LEVIOSA_*` knobs; no fixed host
  port;
- everything on disk lives under one `mktemp -d`; an EXIT trap removes the
  containers (with their anonymous volumes), the network, the built images,
  the public images the run had to pull, and the temp dir (wiping files the
  services wrote as other users from a throwaway container), on success and
  failure.

`KEEP=1` skips the teardown for inspection; `HARNESS_KEEP_PULLED=1` keeps the
public images it pulled. BuildKit's build cache is left alone.
