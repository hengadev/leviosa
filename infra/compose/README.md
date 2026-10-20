# Compose deployment

Everything that releases Leviosa to the VPS, as plain files in the repository
(ADR-0016, ADR-0017; PRD `docs/prd/deployment-infisical-compose.md`): no
Ansible templating, no `${VAR}` substitution, no top-level `.env` on the server.

| File | On the server |
| --- | --- |
| `docker-compose.prod.yml` | installed by each production deploy as `/opt/leviosa/docker-compose.yml`, from the deployed commit (root runner, issue 022) |
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
prefixed `staging_` (issue 019) so they never answer for those names. Without
`container_name`, containers are named `leviosa-<service>-1`.

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

## `deploy <stack> <sha>`

The single release entry point: staging deploys, promotions and rollbacks
(`deploy prod <older-sha>`). This version implements `prod`; `staging` is one
more branch of the script's `case` (issue 019). It runs as root (the env files
are root-only) and each step gates the next:

1. pull `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>`, then
   re-point the `:prod` pins;
2. `up -d --no-recreate --wait postgres rabbitmq redis`: starts what is missing
   (first deploy on empty data directories) and never recreates a running one;
3. `run --rm migrate`: on failure the pins move back to the running images and
   the deploy stops, so the running containers keep serving and a later
   recreate cannot start the failed release;
4. `run --rm admin` (`bootstrap-admin`, idempotent: creates the superadmin once,
   never modifies an existing account);
5. recreate the backend, then the frontend (`--no-deps --force-recreate --wait`);
6. require `/healthz` (asked inside the frontend container) to report `<sha>`
   for both apps; a different SHA fails at once, no answer fails after the
   timeout.

It refuses anything but a 7-to-40-character hex SHA, checks the stack root,
compose file, env files and network before touching anything, and holds a
per-stack lock (`<root>/.deploy.lock`) so two deploys cannot interleave.

Knobs (production defaults; the root runner starts `deploy` with an empty
environment, so CI can set none of them):

| Variable | Default | Purpose |
| --- | --- | --- |
| `LEVIOSA_PROD_ROOT` | `/opt/leviosa` | stack root |
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
with fake 40-hex `COMMIT_SHA`s, lays out a temporary stack root with the
fixture env files, runs LocalStack (KMS only) with the encx KEK alias, and
drives the real `deploy` through the acceptance cases: first deploy (migrate,
admin, both SHAs), second deploy (data kept, admin byte-for-byte unchanged
even with different `ADMIN_*` values), a failing migration (an image built
from a temporary copy of `backend/` plus `fixtures/broken-migration.sql`:
nothing recreated, previous release still serving, pins restored), a SHA
mismatch on either app (deploy fails; redeploying a good SHA recovers), no
`ADMIN_*` in the running containers, data surviving `down` and a fresh deploy,
and finally that nothing is left behind. It ends with a PASS/FAIL line per
case and `RESULT: OK` or `RESULT: FAILED` (non-zero exit).

Isolation, so it can run beside a developer's stacks (`make local-dev`) and
other harness runs:

- every name carries a random run id: compose project `leviosa-harness-<id>-prod`,
  network `leviosa-harness-<id>`, LocalStack container, and image repositories
  `leviosa-harness-<id>-backend` / `-frontend` (the `henga/leviosa-*` pins are
  never touched);
- the only deviations from production are a generated override (those image
  repositories with the same `:prod` pin tag, the network name, Caddy on a
  random `127.0.0.1` port) and the `LEVIOSA_*` knobs; no fixed host port;
- everything on disk lives under one `mktemp -d`; an EXIT trap removes the
  containers (with their anonymous volumes), the network, the built images,
  the public images the run had to pull, and the temp dir (wiping files the
  services wrote as other users from a throwaway container), on success and
  failure.

`KEEP=1` skips the teardown for inspection; `HARNESS_KEEP_PULLED=1` keeps the
public images it pulled. BuildKit's build cache is left alone.
