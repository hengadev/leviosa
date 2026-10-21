# Infisical Agent env templates

One Go template per consumer turns the `leviosa` project's Infisical folders
into env files (ADR-0017). The per-file filter is the security boundary: each
container gets only what it needs, and the frontend never sees a backend
secret.

The same files are rendered in three places, by the same engine (the Agent and
`infisical export --template` share the CLI's template code):

- **on the server**, by the Infisical Agent: issue 027 installs one copy per
  file, made by [`render-template.sh`](render-template.sh);
- **locally**, by `make env` ([`make-env.sh`](make-env.sh)), from `dev`;
- **for a check**, by `make env-check-<env>` ([`check-env.sh`](check-env.sh)),
  read-only, key names only.

```
infra/infisical/
├── templates/*.env.tmpl   one template per consumer (below)
├── render-template.sh     placeholders -> a concrete copy for dev|staging|prod
├── make-env.sh            make env
├── check-env.sh           make env-check-dev|staging|prod
└── envcheck/              Go module: the checker and the fixture test
```

## Files

| Template | Reads | Consumer | Rendered for | Agent on change |
| --- | --- | --- | --- | --- |
| `backend.env` | one recursive call on `/`: `/app`, `/encx`, `/aws`, `/stripe`, `/smtp`, `/twilio` whole; `/db` → `DATABASE_URL`; `/rabbitmq` → `RABBITMQ_URL`; `/redis` → `REDIS_ADDR`, rest whole | backend + its `migrate` and `admin` one-offs | prod, staging, dev | recreate that stack's backend |
| `frontend.env` | `/app`, through an allowlist | frontend | prod, staging, dev | recreate that stack's frontend |
| `admin.env` | `/admin` | the `admin` one-off only | prod, staging, dev | rewrite only |
| `redis.env` | `/redis` (`REDIS_PASSWORD`) | Redis | prod, staging | recreate that stack's Redis |
| `postgres.env` | `/db` (user, password, database) | Postgres | prod | rewrite only |
| `rabbitmq.env` | `/rabbitmq` → `RABBITMQ_DEFAULT_*` | RabbitMQ | prod | rewrite only |
| `caddy.env` | `staging` `/caddy`, then `prod` `/caddy` | Caddy (serves both stacks) | prod | recreate Caddy |
| `backup.env` | one recursive call on `/`: `/backup` whole + `POSTGRES_USER`, `POSTGRES_DB` | backup job (issue 028) | prod | rewrite only |

Server paths: `/opt/leviosa/env/<file>` (prod) and
`/opt/leviosa-staging/env/<file>` (staging), root-owned, `0600`.
`make env` writes `backend/development.env` (backend.env followed by
admin.env, so local `bootstrap-admin` runs as on the server) and
`frontend/.env`. `dev` is also checked for redis, postgres and rabbitmq,
which nothing consumes locally, to prove those templates on real folders. The
list per environment lives in one place, `envcheck.Files`
(`go run ./cmd/envcheck files <env>` from `envcheck/`).

Notes for the Agent role (issue 027): install the substituted copies with
`copy`, never `template` (Jinja uses the same `{{ }}` braces); recreate a
single service with `docker compose up -d --no-deps --force-recreate
<service>` (see `infra/compose/README.md`); Postgres, RabbitMQ, backup and
admin files get no command.

## Placeholders

The engine has no `env` function, so a template cannot know which
environment or project it renders. Each one carries placeholders that
`render-template.sh <template> <dev|staging|prod>` substitutes on stdout:

- `__LEVIOSA_ENV__`: the environment slug. Only `dev`, `staging` and `prod`
  are accepted.
- `__LEVIOSA_PROJECT_ID__`: the `leviosa` project's ID, read from
  **`infisical_project_id` in `infra/terraform/terraform.tfvars`**. That is
  the one place the operator sets it (issue 008 creates the project; issue 023
  already asks for it there). Until then every render stops with a message
  naming that file. `LEVIOSA_INFISICAL_PROJECT_ID` overrides it (the tests use
  a fake ID).

Rendering one file by hand (dev):

```sh
infra/infisical/render-template.sh infra/infisical/templates/backend.env.tmpl dev > /tmp/backend.env.tmpl
infisical export --template=/tmp/backend.env.tmpl --silent   # prints values: dev only
```

For staging or prod, use `make env-check-<env>`, which never prints a value.

## Template rules

Checked against Infisical CLI v0.43.138 (`packages/cmd/agent.go`,
`packages/templates/templates.go`) and Infisical's server
(`backend/src/server/app.ts`):

- **Functions**: `listSecrets <project> <env> <path> [<json options>]`
  (options `{"recursive": true}`, `expandSecretReferences`; imports are always
  included), `getSecretByName`, `listSecretsByProjectSlug`, `dynamicSecret`,
  plus Go built-ins. That CLI version also adds sprig (without `env` and
  `expandenv`); the templates use **Go built-ins only** so they do not depend
  on the Agent version, and the fixture test does not register sprig. Each
  secret has `.Key`, `.Value`, `.SecretPath`, `.Type`, `.Comment`, `.Tags`.
- **One secret call per template.** Each call overwrites the template's
  ETag, which is the `etag` header of the API response (`@fastify/etag`, a
  hash of the body). The Agent re-renders a file, and runs its command, only
  when that ETag differs from the previous poll's. So only the **last** call
  is change-detected. Multi-folder templates make one recursive call and
  filter on `.SecretPath`, which covers a change in any folder. A change in
  a folder a template reads but does not print (e.g. `/admin` for
  backend.env) also changes the ETag and recreates the consumer with the same
  values. That is harmless.
- **An Agent restart rewrites every file but runs no command** (the first
  render never executes it). After a restart, recreate the consumers whose
  values changed yourself.
- The secrets' order is not guaranteed; the test renders every template with
  the order reversed and expects the same values.

## Quoting: every value is single-quoted

Docker Compose interpolates unquoted `env_file` values. Checked with Compose
v5.5.0 (`docker compose config`):

| Line in the env file | Value the container gets |
| --- | --- |
| `H=$2a$14$ajq8Q7...` (a bcrypt hash) | `$2a$14.Rx1h...` (part of the hash dropped) |
| `P=Pa$word123` | `Pa` |
| `P=abc #def` | `abc` |
| `H='$2a$14$ajq8Q7...'` | unchanged |

So the templates write `KEY='value'`. Compose, systemd `EnvironmentFile`, `sh`,
godotenv (the backend in development) and dotenv (Vite) all read single-quoted
values literally. Two consequences:

- **No value may contain a single quote or a line break.** It cannot be
  written safely; `envcheck` reports it, by key name.
- **Never load these files with `docker run --env-file`**, which keeps the
  quotes as part of the value. Use `docker compose run` (or `env_file:`).

The fixture test has `docker compose config` read every rendered file back
and expects each value unchanged.

## Composite values

Each password exists once, in its own folder, and the URLs are built from the
atoms:

| Key | Format | From |
| --- | --- | --- |
| `DATABASE_URL` | `postgres://<user>:<password>@<host>:<port>/<db>?sslmode=disable` | `/db` |
| `RABBITMQ_URL` | `amqp://<user>:<password>@<host>:<port>/<vhost>` | `/rabbitmq` |
| `REDIS_ADDR` | `<host>:<port>` | `/redis` |
| `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_REGION` | copies of `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` | `/aws` |

- User, password and vhost go through `urlquery`. It writes a space as
  `+`, which is wrong inside userinfo, so **passwords in `/db` and `/rabbitmq`
  are letters and digits only**, and users are letters, digits and
  underscores (`envcheck` checks both).
- The vhost is always explicit. The default vhost `/` becomes `%2F`; a bare
  trailing slash would be an empty vhost name for amqp091-go.
- `sslmode=disable`: the Postgres container has no TLS.
- The S3 client reads `S3_*`, KMS reads `AWS_*`, and Terraform writes `/aws`
  once. The template repeats the three atoms instead of storing them twice.
- Staging and production hosts are service names on `leviosa_network`:
  `postgres`, `rabbitmq`, and `redis` (prod) or `staging_redis` (staging).
  `envcheck` refuses any other host there, so staging cannot point at
  production's Redis.

## frontend.env is an allowlist

`/app` also holds backend-only secrets (`SESSION_SECRET`,
`BOOKING_TOKEN_SECRET`). Rendering it whole would hand them to the frontend.
`frontend.env` therefore prints only the `/app` keys the SvelteKit server
reads, found by grepping `frontend/src` for `$env/dynamic/*`: `COMING_SOON`,
`CLIENT_IP_HEADER`, `SESSION_COOKIE_NAME` (listed in `.env.example`, not read
yet), and the optional `PUBLIC_PLAUSIBLE_DOMAIN` and `PUBLIC_PLAUSIBLE_SRC`.
Nothing from `/aws`, `/encx`, `/stripe`, `/twilio`, `/smtp`, `/db`, `/admin`,
`/redis`, `/rabbitmq`, `/caddy`, `/backup` or `/host` reaches it. A new
frontend key needs a template line and an entry in `envcheck.FrontendKeys`.
That is a deliberate review point.

The template sets the stack's wiring itself: `NODE_ENV` (`production`, or
`development` for dev), `PORT=3000`, `API_URL` (`http://backend:3500`,
`http://staging_backend:3500`, `http://localhost:3500`),
`PROTOCOL_HEADER=x-forwarded-proto` and `HOST_HEADER=host` (Caddy serves
several hostnames, so `ORIGIN` is never set), and `USE_MOCK_DATA=false`. Mock
mode takes the user's role from a client cookie, so no rendered file ever
enables it.

## caddy.env reads two environments

Caddy belongs to production's stack but needs staging's basic auth too:
`CLOUDFLARE_API_TOKEN` from `prod` `/caddy`, and `STAGING_BASIC_AUTH_USER` and
`STAGING_BASIC_AUTH_HASH` (bcrypt) from `staging` `/caddy`, rendered whole
under their own names.

A `listSecrets` call reads one environment, but a template may make several
calls with different slugs. So one template can read both environments, and
the identity (Viewer on the whole project) can read both. The cost is change
detection: only the last call's ETag counts. The calls run **staging first,
prod last**:

- **Detected**: a change in `prod` `/caddy` (the Cloudflare token). It
  re-renders the whole file, both folders fresh, and recreates Caddy within one
  poll. A forgotten token update would only show at the next certificate
  renewal, so this one must be automatic.
- **Not detected on its own**: a change in `staging` `/caddy` (the basic-auth
  user or hash). It is picked up at the next prod `/caddy` change or Agent
  restart. A missed update shows at once: the new password is refused. After
  changing it, force the re-render on the server:

  ```sh
  systemctl restart infisical-agent   # unit name from issue 027; rewrites every file, runs no command
  docker compose --project-directory /opt/leviosa up -d --no-deps --force-recreate caddy
  ```

Two alternatives give full detection, if staging's basic auth changes often:

- Two Caddy env files (`caddy.env` from prod, `caddy-staging.env` from
  staging), one call each, both listed in the `caddy` service's `env_file`.
  This needs a change to `docker-compose.prod.yml`.
- An Infisical secret import of `staging:/caddy` into `prod:/caddy`. One call
  reads both, since imported values are in the response body and so in the
  ETag. But the import is structure that only exists in Infisical, and it puts
  staging values into prod's folder and its backup snapshot.

## Folders and keys

The layout the templates expect, and the checklist for filling Infisical
(issue 008). [`envcheck/testdata/infisical.json`](envcheck/testdata/infisical.json)
has one fake example per environment. Folders a template prints whole may hold
more keys: they reach their consumer with no template change.

| Folder | Keys | Notes |
| --- | --- | --- |
| `/app` | `FRONTEND_ORIGIN`, `SESSION_SECRET`, `BOOKING_TOKEN_SECRET`, `SESSION_COOKIE_NAME`, `CLIENT_IP_HEADER`, `COMING_SOON`; optional `PUBLIC_PLAUSIBLE_DOMAIN`, `PUBLIC_PLAUSIBLE_SRC`, other backend knobs | whole into backend.env, allowlist into frontend.env; `prod` starts with `COMING_SOON=true` |
| `/admin` | `ADMIN_EMAIL`, `ADMIN_PASSWORD`, `ADMIN_FIRST_NAME`, `ADMIN_LAST_NAME`, `ADMIN_TELEPHONE`, `ADMIN_POSTAL_CODE`, `ADMIN_CITY`, `ADMIN_ADDRESS1`, `ADMIN_ADDRESS2` (may be empty) | admin.env only |
| `/db` | `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_HOST`, `POSTGRES_PORT` | password letters and digits |
| `/redis` | `REDIS_PASSWORD`, `REDIS_HOST`, `REDIS_PORT` | staging's host is `staging_redis` |
| `/rabbitmq` | `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `RABBITMQ_VHOST`, `RABBITMQ_HOST`, `RABBITMQ_PORT` | password letters and digits |
| `/encx` | `ENCX_PEPPER`; optional `ENCX_DEK_CACHE_MAX_ENTRIES`, `ENCX_DEK_CACHE_TTL` | pepper: base64 of 32 bytes. The cache keys size encx's in-memory data-key cache (issue 031); unset means encx's defaults (4096 entries, 10 minutes), `0` entries disables it. The KEK alias is in `/aws` (Terraform). Do not repeat it here, or backend.env gets it twice |
| `/aws` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, `S3_BUCKET_NAME`, `ENCX_KMS_KEY_ARN`, `ENCX_KMS_KEY_ALIAS` | written by Terraform in staging and prod; by hand in dev |
| `/stripe` | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_CONNECT_WEBHOOK_SECRET` | |
| `/smtp` | `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, the sender address | the sender has no consumer yet |
| `/twilio` | `TWILIO_ACCOUNT_SID`, `TWILIO_API_KEY_SID`, `TWILIO_API_KEY_SECRET`, `TWILIO_REGION`, `TWILIO_EDGE`, `TWILIO_SENDER` | |
| `/caddy` | prod: `CLOUDFLARE_API_TOKEN`; staging: `STAGING_BASIC_AUTH_USER`, `STAGING_BASIC_AUTH_HASH` | |
| `/backup` | `BACKUP_AWS_ACCESS_KEY_ID`, `BACKUP_AWS_SECRET_ACCESS_KEY`, `BACKUP_S3_BUCKET`, `BACKUP_AWS_REGION` | Terraform, prod only |
| `/host` | the notification email | Ansible only; no template reads it |

**Never put these in Infisical.** The templates set them, and a copy in a
folder would appear twice (`envcheck` reports duplicates):

- `ENVIRONMENT` (`production`, `staging` or `development`, from the slug;
  always explicit, because `bootstrap-admin` falls back to development
  defaults when it is unset)
- `SERVER_PORT=3500`, `LOG_FORMAT` (`json`, or `text` for dev), `REDIS_DB=0`
- `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_REGION`
- frontend.env's wiring above

**`dev` for `make local-dev`.** `backend/compose.yaml` overrides the
infrastructure values (`DATABASE_URL`, `REDIS_*`, `RABBITMQ_URL`, `S3_*`,
`AWS_ENDPOINT_URL`, `ENVIRONMENT`) inside the container. What `dev` must get
right is LocalStack's encryption setup:

- `/aws`: `AWS_REGION=us-east-1` (the region `localstack-init` creates the
  KMS key in), `AWS_ACCESS_KEY_ID=test`, `AWS_SECRET_ACCESS_KEY=test`,
  `AWS_ENDPOINT_URL=http://localhost:4566`, `S3_BUCKET_NAME=leviosa`,
  `ENCX_KMS_KEY_ALIAS=alias/leviosa-dev` (the alias `localstack-init`
  creates)
- `/encx`: `ENCX_PEPPER`, the public dev pepper from `backend/example.env`

For tools run outside Docker, use the host-side addresses of the local
stack: `/db` `user` / `password` / `leviosa_dev` / `localhost` / `5432`,
`/redis` `secret` / `localhost` / `6379`, `/rabbitmq` `guest` / `guest` /
`/` / `localhost` / `5672`. `/stripe`, `/smtp` and `/twilio` may be absent
(integrations off). `dev` must never hold anything that spends money or
reaches real data (ADR-0017).

## `make env`

```sh
infisical login      # once, to https://secrets.henga.dev
make env             # OVERWRITES backend/development.env and frontend/.env
make local-dev
```

It renders `dev` only, with your own login. It stops with a clear message,
and writes nothing, when the CLI is missing, there is no session, a rendering
fails, or the rendering lacks `DATABASE_URL`, `ENCX_PEPPER`,
`ENCX_KMS_KEY_ALIAS`, `AWS_REGION`, `ADMIN_EMAIL` or `API_URL`. Otherwise it
replaces both files (mode `0600`, gitignored) and says so. Values you need
locally, such as a Stripe CLI webhook secret, belong in `dev`. Without
Infisical, `make local-dev` copies the committed example files instead.

## Checking renderings: `make env-check-<env>`

```sh
make env-check-dev       # or env-check-staging, env-check-prod
```

It renders that environment's files into a private temporary directory
(`0700`, deleted on exit), then lists each file's **key names** and every
problem: a missing or empty required key, a key outside frontend.env's
allowlist (or `ORIGIN`), a backend secret in the wrong file, a duplicate, an
unquotable value, a malformed URL, a wrong stack host, a pepper that is not 32
bytes, or a basic-auth hash that is not bcrypt. Provider keys that arrive late
in the rollout (Stripe, Twilio, SMTP) are warnings. It never prints a value,
and it only reads Infisical. It needs the Infisical CLI, `infisical login` and
Go. Exit status 0 means no problem.

## Tests

```sh
make agent-templates-test    # cd infra/infisical/envcheck && go test ./...
```

No Infisical and no network. It renders every template through
`render-template.sh`, against [fake folders](envcheck/testdata/infisical.json),
with stand-ins for the Agent's template functions. It asserts:

- each file's exact key set;
- whole folders kept verbatim (`$`, `#`, quotes, backslashes);
- the URL formats;
- frontend.env's exclusions, by key and by value;
- the number and order of secret calls per template;
- that every key of the compose harness fixtures
  (`infra/compose/test/fixtures`) is rendered;
- that `docker compose config` reads every value back unchanged (skipped
  without Docker).

It also covers the checker's own rules, `render-template.sh`'s refusals, and
`make-env.sh` / `check-env.sh` against a stub CLI. The real `infisical` is
never on the tests' `PATH`.
