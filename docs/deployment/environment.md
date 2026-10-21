# Environment reference

Every configuration value lives in **Infisical**, the operator's self-hosted
instance at `https://secrets.henga.dev`
([ADR-0017](../adr/0017-infisical-sole-source-of-config.md)). Nothing is copied
by hand into a server file, a GitHub secret or a committed `.env`. This page
says *where* each value lives, *who reads it* and *who writes it*. The values
themselves are only in Infisical.

The templates that turn folders into env files, their rules (quoting, one
secret call per template, the `caddy.env` caveat) and the checker are in
[infra/infisical/README.md](../../infra/infisical/README.md). The render matrix
on the server (which file each stack gets and what is recreated) is in
[roles/infisical_agent](../../infra/ansible/roles/infisical_agent/README.md#render-matrix).
`docs/deployment/check-environment.sh` checks that every key the templates
render is listed on this page.

## Two projects

| Project | Environments | Holds | Read by | Written by |
| --- | --- | --- | --- | --- |
| `leviosa` | `dev`, `staging`, `prod` | the app's and the server's values, one folder per concern (below) | the Infisical Agent on the server (machine identity `leviosa-vps`, Viewer, read-only); `make env` and `make env-check-<env>` (your own login); `make setup` (`/host`, `/db`, `/rabbitmq`, your login) | the operator; Terraform for `/aws` and `prod/backup` |
| `leviosa-infra` | `prod`, path `/` | Terraform's inputs: `TF_VAR_hcloud_token`, `TF_VAR_cloudflare_token`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` (the IAM user Terraform and its state backend run with) | `infra/terraform/tf.sh` only, through the operator's login. `leviosa-vps` has no access to it | the operator |

The project IDs are not secret: `infisical_project_id` in
`infra/terraform/terraform.tfvars` (the one place for `leviosa`; Terraform,
the Agent templates and Ansible read it there) and `infra_project_id` in
`infra/terraform/tf.sh` (`leviosa-infra`).

## Rendered files and their consumers

| File | Consumer | Stacks | Agent on change |
| --- | --- | --- | --- |
| `backend.env` | backend `serve`, and the `migrate` and `admin` one-offs | prod, staging; dev (`make env` → `backend/development.env`) | recreate that stack's backend |
| `frontend.env` | the SvelteKit server | prod, staging; dev (`frontend/.env`) | recreate that stack's frontend |
| `admin.env` | the `admin` (`bootstrap-admin`) one-off only | prod, staging; dev (appended to `backend/development.env`) | rewrite only |
| `redis.env` | the stack's Redis | prod, staging | recreate that stack's Redis |
| `postgres.env` | Postgres (shared) | prod | rewrite only: applied at first initialisation |
| `rabbitmq.env` | RabbitMQ (shared) | prod | rewrite only: applied at first initialisation |
| `caddy.env` | Caddy (serves both stacks) | prod | recreate Caddy (prod `/caddy` changes only) |
| `backup.env` | the backup timers (`leviosa-backup`, `leviosa-snapshot`) | prod | rewrite only: read fresh by each run |

On the server: `/opt/leviosa/env/<file>` and `/opt/leviosa-staging/env/<file>`,
root-owned `0600`.

## Folders in `leviosa`

**Written by** says who puts the value in Infisical. A folder rendered *whole*
passes any extra key through to its consumer with no template change.

### `/app` — backend runtime values (whole into `backend.env`; allowlist into `frontend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `FRONTEND_ORIGIN` | ✓ | ✓ | ✓ | `backend.env` → backend (CORS, links) | operator |
| `SESSION_SECRET` | ✓ | ✓ | ✓ | `backend.env` → backend. Never the frontend | operator |
| `BOOKING_TOKEN_SECRET` | ✓ | ✓ | ✓ | `backend.env` → backend (signed booking lookup links). Never the frontend | operator |
| `SESSION_COOKIE_NAME` | ✓ | ✓ | ✓ | `backend.env`, `frontend.env` | operator |
| `CLIENT_IP_HEADER` | ✓ | ✓ | ✓ | `backend.env`, `frontend.env` | operator |
| `COMING_SOON` | ✓ (`true` until launch) | ✓ | ✓ (`false`) | `backend.env`, `frontend.env` → the frontend redirects everything to `/coming-soon` | operator |
| `PUBLIC_PLAUSIBLE_DOMAIN`, `PUBLIC_PLAUSIBLE_SRC` | optional | optional | | `backend.env`, `frontend.env` → the analytics script | operator |

Other backend knobs (`LOG_LEVEL`, `RATE_LIMIT_*`, `REMINDER_*`, ...) may be
added to `/app`; they reach `backend.env` only. A new **frontend** key needs a
line in `frontend.env.tmpl` and in `envcheck.FrontendKeys`: a deliberate review
point.

### `/admin` — the first superadmin (`admin.env` only)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `ADMIN_EMAIL`, `ADMIN_PASSWORD` | ✓ | ✓ | ✓ | `admin.env` → `bootstrap-admin`, at every deploy | operator |
| `ADMIN_FIRST_NAME`, `ADMIN_LAST_NAME`, `ADMIN_TELEPHONE`, `ADMIN_POSTAL_CODE`, `ADMIN_CITY`, `ADMIN_ADDRESS1`, `ADMIN_ADDRESS2` (may be empty) | ✓ | ✓ | ✓ | `admin.env` → `bootstrap-admin` | operator |

`bootstrap-admin` creates the account only if the email does not exist and never
modifies an existing one; outside development it refuses a weak password. A
later change in `/admin` changes nothing for an existing account.

### `/db` — Postgres (assembled into `DATABASE_URL`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `POSTGRES_USER` | ✓ | ✓ (`leviosa_staging`) | ✓ | `DATABASE_URL` in `backend.env`; `postgres.env` (prod); `backup.env` (prod); `make setup` (staging role) | operator |
| `POSTGRES_PASSWORD` | ✓ | ✓ | ✓ | `DATABASE_URL`; `postgres.env` (prod); `make setup` | operator |
| `POSTGRES_DB` | ✓ | ✓ (`leviosa_staging`) | ✓ | `DATABASE_URL`; `postgres.env`; `backup.env` | operator |
| `POSTGRES_HOST` | `postgres` | `postgres` | `localhost` | `DATABASE_URL` | operator |
| `POSTGRES_PORT` | `5432` | `5432` | `5432` | `DATABASE_URL` | operator |

### `/redis`

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `REDIS_PASSWORD` | ✓ | ✓ | ✓ | `backend.env` → backend; `redis.env` → the stack's Redis | operator |
| `REDIS_HOST` | `redis` | `staging_redis` | `localhost` | `REDIS_ADDR` in `backend.env` | operator |
| `REDIS_PORT` | `6379` | `6379` | `6379` | `REDIS_ADDR` | operator |

### `/rabbitmq` (assembled into `RABBITMQ_URL`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `RABBITMQ_USER` | ✓ | ✓ | ✓ | `RABBITMQ_URL` in `backend.env`; `RABBITMQ_DEFAULT_USER` in `rabbitmq.env` (prod); `make setup` (the users) | operator |
| `RABBITMQ_PASSWORD` | ✓ | ✓ | ✓ | `RABBITMQ_URL`; `RABBITMQ_DEFAULT_PASS`; `make setup` | operator |
| `RABBITMQ_VHOST` | ✓ (`leviosa`) | ✓ (`leviosa_staging`) | `/` | `RABBITMQ_URL`; `RABBITMQ_DEFAULT_VHOST`; `make setup` (the vhosts) | operator |
| `RABBITMQ_HOST` | `rabbitmq` | `rabbitmq` | `localhost` | `RABBITMQ_URL` | operator |
| `RABBITMQ_PORT` | `5672` | `5672` | `5672` | `RABBITMQ_URL` | operator |

### `/encx` — the pepper and the DEK cache (whole into `backend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `ENCX_PEPPER` | ✓ | ✓ (different from prod) | ✓ (the public dev pepper from `backend/example.env`) | `backend.env` → backend, and `leviosa-restore-test` (prod) | operator, once; **never rotated** |
| `ENCX_DEK_CACHE_MAX_ENTRIES`, `ENCX_DEK_CACHE_TTL` | optional | optional | optional | `backend.env` → backend: the in-memory cache of decrypted data keys (issue 031), so the public partner listing does not pay one KMS call per partner per visit. Unset: encx's defaults (4096 entries, 10 minutes). `ENCX_DEK_CACHE_MAX_ENTRIES=0` disables the cache; the TTL is a duration (`10m`, `45s`). Invalid values stop the backend at startup | operator |

The KEK alias is **not** here: it is `ENCX_KMS_KEY_ALIAS` in `/aws`, written by
Terraform. A copy here would appear twice in `backend.env`.

### `/aws` — app AWS credentials, media bucket, KMS key (whole into `backend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | ✓ | ✓ | `test` / `test` | `backend.env` → backend (KMS), and copied to `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY` (S3) | **Terraform** (`leviosa-<env>-app`); by hand in dev |
| `AWS_REGION` | `eu-central-1` | `eu-central-1` | `us-east-1` | `backend.env`, and copied to `S3_REGION` | **Terraform**; by hand in dev |
| `S3_BUCKET_NAME` | `leviosa-care-prod-media` | `leviosa-care-staging-media` | `leviosa` | `backend.env` → backend | **Terraform**; by hand in dev |
| `ENCX_KMS_KEY_ALIAS` | `alias/leviosa-prod` | `alias/leviosa-staging` | `alias/leviosa-dev` | `backend.env` → backend (the KEK, by alias only) | **Terraform**; by hand in dev |
| `ENCX_KMS_KEY_ARN` | ✓ | ✓ | | `backend.env` (for reference; the backend finds its key by alias) | **Terraform** |
| `AWS_ENDPOINT_URL` | | | `http://localhost:4566` | `backend.env` (LocalStack only; never in staging or prod) | by hand in dev |

### `/stripe` (whole into `backend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `STRIPE_SECRET_KEY` | live | **test** | **test** | `backend.env` → backend | operator |
| `STRIPE_WEBHOOK_SECRET` | ✓ | ✓ | optional (Stripe CLI) | `backend.env` → backend (`/webhooks/stripe`) | operator, from the endpoint registered in Stripe |
| `STRIPE_CONNECT_WEBHOOK_SECRET` | ✓ | ✓ | optional | `backend.env` → backend (`/webhooks/stripe/connect`) | operator, from the Connect endpoint |

### `/smtp` — Zoho Mail (whole into `backend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `SMTP_HOST`, `SMTP_PORT` | ✓ | ✓ | absent | `backend.env` → backend, `check-email` | operator (issue 036) |
| `SMTP_USERNAME`, `SMTP_PASSWORD` (a Zoho app password) | ✓ | ✓ | absent | `backend.env` → backend, `check-email` | operator |
| `SMTP_FROM` (the sender address) | ✓ | ✓ | absent | `backend.env`, but **no consumer yet**: the backend has no from-address setting | operator |

### `/twilio` (whole into `backend.env`)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `TWILIO_ACCOUNT_SID` | ✓ | ✓ | absent | `backend.env` → backend, `check-sms` | operator |
| `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET` | ✓ | ✓ | absent | `backend.env` → backend, `check-sms`. An API key, never the auth token; created in the IE1 region | operator |
| `TWILIO_REGION`, `TWILIO_EDGE` | `ie1`, `dublin` | `ie1`, `dublin` | absent | `backend.env` → backend | operator |
| `TWILIO_SENDER` | ✓ | ✓ | absent | `backend.env` → backend: `Leviosa` (alphanumeric) or a French number | operator |

### `/caddy` (`caddy.env`, prod stack only)

| Key | prod | staging | dev | Rendered into → read by | Written by |
| --- | :-: | :-: | :-: | --- | --- |
| `CLOUDFLARE_API_TOKEN` | ✓ | | | `caddy.env` → Caddy (TLS through the DNS challenge). Its own token, DNS edit on `leviosa.care`, separate from Terraform's `TF_VAR_cloudflare_token`: each can be revoked without breaking the other, and the server holds only this one | operator |
| `STAGING_BASIC_AUTH_USER`, `STAGING_BASIC_AUTH_HASH` (a `caddy hash-password` bcrypt hash) | | ✓ | | `caddy.env` → Caddy (staging's basic auth) | operator. **Set both together**: with only one, Caddy refuses its config and both stacks go down |

### `/backup` (`backup.env`, prod only)

| Key | prod | Rendered into → read by | Written by |
| --- | :-: | --- | --- |
| `BACKUP_AWS_ACCESS_KEY_ID`, `BACKUP_AWS_SECRET_ACCESS_KEY` | ✓ | `backup.env` → the backup timers (the no-delete `leviosa-prod-backup` user) | **Terraform** |
| `BACKUP_S3_BUCKET` | `leviosa-care-prod-backups` | `backup.env` | **Terraform** |
| `BACKUP_AWS_REGION` | `eu-central-1` | `backup.env` | **Terraform** |

`backup.env` also gets `POSTGRES_USER` and `POSTGRES_DB` from `prod/db`.

### `/host` (prod only, not rendered)

| Key | prod | Read by | Written by |
| --- | :-: | --- | --- |
| `NOTIFY_EMAIL` | ✓ | `make setup` (`playbooks/site.yml`), through the operator's login: fail2ban, unattended-upgrades, the ACME account | operator |

## Keys the templates set themselves

Never put these in Infisical: the template writes them, and a copy in a folder
would appear twice (`make env-check-<env>` reports duplicates).

| Key | File | Value |
| --- | --- | --- |
| `ENVIRONMENT` | `backend.env` | `production`, `staging` or `development`, from the environment. Always explicit: `bootstrap-admin` falls back to development defaults when it is unset |
| `SERVER_PORT` | `backend.env` | `3500` |
| `LOG_FORMAT` | `backend.env` | `json` (`text` in dev) |
| `REDIS_DB` | `backend.env` | `0` (each stack has its own Redis) |
| `DATABASE_URL` | `backend.env` | `postgres://<user>:<password>@<host>:<port>/<db>?sslmode=disable`, from `/db` |
| `RABBITMQ_URL` | `backend.env` | `amqp://<user>:<password>@<host>:<port>/<vhost>`, from `/rabbitmq` |
| `REDIS_ADDR` | `backend.env` | `<host>:<port>`, from `/redis` |
| `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_REGION` | `backend.env` | copies of `/aws`'s `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` |
| `NODE_ENV` | `frontend.env` | `production` (`development` in dev) |
| `PORT` | `frontend.env` | `3000` |
| `API_URL` | `frontend.env` | `http://backend:3500`, `http://staging_backend:3500`, `http://localhost:3500` |
| `PROTOCOL_HEADER`, `HOST_HEADER` | `frontend.env` | `x-forwarded-proto`, `host`: Caddy serves several hostnames, so `ORIGIN` is **never** set |
| `USE_MOCK_DATA` | `frontend.env` | `false`: mock mode trusts a client cookie for the role |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | `postgres.env` | from `prod/db` |
| `RABBITMQ_DEFAULT_USER`, `RABBITMQ_DEFAULT_PASS`, `RABBITMQ_DEFAULT_VHOST` | `rabbitmq.env` | from `prod/rabbitmq` |
| `REDIS_PASSWORD` | `redis.env` | from the stack's `/redis` |
| `POSTGRES_USER`, `POSTGRES_DB` | `backup.env` | from `prod/db` |

## Rules that are easy to get wrong

- **Passwords in `/db` and `/rabbitmq` are letters and digits only**, and users
  letters, digits and underscores: they go into URLs. RabbitMQ vhosts on the
  server are lowercase identifiers, never `/`.
- **`POSTGRES_PASSWORD` and `RABBITMQ_DEFAULT_*` apply only when the data
  directory is first created.** See the rotation table in
  [README.md](README.md#credentials-and-rotation).
- **No value may contain a single quote or a line break**: every value is
  written single-quoted. Never load a rendered file with
  `docker run --env-file` (it keeps the quotes); use `docker compose`.
- **Staging is not a copy of prod**: its own Postgres role and database,
  RabbitMQ vhost and user, Redis, KMS key, pepper, bucket and AWS key, and
  Stripe test mode.
- **`dev` must never hold anything that spends money or reaches real data.**
  On the free Infisical plan `leviosa-vps` can read `dev` too.
- **`staging/caddy` changes are not picked up on their own** (only the last
  secret call of a template is change-detected, and `caddy.env` reads staging
  first): restart the Agent and recreate Caddy after changing the basic auth.

## Local development

- **No Infisical:** `make local-dev` copies `backend/example.env` to
  `backend/development.env`; `make local-start` copies `.env.example` to `.env`.
  Both are committed, hold no secret and leave the integrations off.
- **With Infisical:** `infisical login`, then `make env`. It renders
  `backend/development.env` (backend.env, then admin.env) and `frontend/.env`
  from `dev` with the server's templates, and **overwrites** both. Values you
  need locally (a Stripe CLI webhook secret, say) belong in `dev`. What `dev`
  must contain for LocalStack is listed in
  [infra/infisical/README.md](../../infra/infisical/README.md#folders-and-keys).
- `make env-check-<dev|staging|prod>` renders an environment into a private
  temporary directory and lists **key names** and problems only, never values.
