# Deployment

How Leviosa runs in staging and production, how a release gets there, and
what to do when something breaks. This page is the overview; each component
README has the details and is linked where it matters.

- [environment.md](environment.md): every configuration value, its Infisical folder, who reads it and who writes it
- [setup.md](setup.md): a new operator computer, from zero to running Terraform and `make setup`
- [ci.md](ci.md): CI checks, image builds, the staging deploy, the CI secrets
- [promote.md](promote.md): production releases and rollbacks

The decisions behind this design:
[ADR-0015](../adr/0015-encx-keys-in-aws-kms-and-infisical.md) (encx KEK in AWS
KMS, pepper in Infisical; no HashiCorp Vault),
[ADR-0016](../adr/0016-compose-stacks-on-a-single-vps.md) (two Compose stacks
on one VPS, not Swarm) and
[ADR-0017](../adr/0017-infisical-sole-source-of-config.md) (Infisical is the
only source of configuration).

## Architecture

```
                             Cloudflare (proxied DNS)
                                      │
   ┌──────────────────────────────────▼──────────────────────────────────────┐
   │ Hetzner cpx22, nbg1 (Terraform; primary IPs survive a rebuild)          │
   │                                                                         │
   │  caddy ── TLS via Cloudflare DNS-01, lb_try_duration 30 s               │
   │   ├─ leviosa.care, www., admin., staff.                                 │
   │   │     /webhooks/* ─► backend:3500     else ─► frontend:3000           │  /opt/leviosa
   │   └─ staging., admin-staging., staff-staging.   (basic auth, noindex)   │  (prod stack)
   │         /webhooks/* ─► staging_backend  else ─► staging_frontend        │  /opt/leviosa-staging
   │                                                                         │  (staging stack)
   │  shared, owned by prod:  postgres (a role + db per env)                 │
   │                          rabbitmq (a vhost + user per env)              │
   │  per stack:              redis, staging_redis                           │
   │                                                                         │
   │  infisical-agent (systemd) ── renders env/*.env, root 0600 ─────────────┼◄── Infisical `leviosa`
   │  leviosa-backup / -snapshot (systemd timers) ───────────────────────────┼──► S3 backup bucket
   └─────────────────────────────────────────────────────────────────────────┘
        backend ──► AWS KMS (encx KEK, one key per env), S3 media bucket per env
```

- **One server, two stacks** ([ADR-0016](../adr/0016-compose-stacks-on-a-single-vps.md)).
  Production (`/opt/leviosa`, compose project `leviosa`) runs its backend,
  frontend and Redis, plus the services both stacks share: Postgres,
  RabbitMQ and Caddy. Staging (`/opt/leviosa-staging`) runs only its backend,
  frontend and its own Redis, under `staging_`-prefixed service names. Both
  join the external `leviosa_network`. Staging can reach production's
  containers on that network but holds no credential that opens production's
  database, vhost or Redis. A Postgres, RabbitMQ or Caddy outage takes both
  stacks down. Details: [infra/compose/README.md](../../infra/compose/README.md).
- **Configuration comes only from Infisical**
  ([ADR-0017](../adr/0017-infisical-sole-source-of-config.md)). The Infisical
  Agent renders one env file per consumer, polls every 60 s, and recreates
  only the service whose file changed. The frontend never receives a backend
  secret, and `ADMIN_*` reaches only the `admin` one-off. **Postgres and
  RabbitMQ are never restarted by a value change.** No secret is in the
  repository, in Ansible, in GitHub or in a hand-written server file.
  Details: [roles/infisical_agent](../../infra/ansible/roles/infisical_agent/README.md),
  [infra/infisical/README.md](../../infra/infisical/README.md).
- **Encryption keys** ([ADR-0015](../adr/0015-encx-keys-in-aws-kms-and-infisical.md)):
  the encx KEK is an AWS KMS key per environment (`alias/leviosa-<env>`,
  `prevent_destroy`, 30-day deletion window); the pepper is `ENCX_PEPPER` in
  Infisical `/encx`. Nothing needs unsealing. A missing pepper stops the
  backend instead of creating a new one.
- **Images**: `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>`,
  built by CI from `main`. Each stack refers to a local pin (`:prod`,
  `:staging`) that the deploy re-points at `<sha>`. The backend's `/health`
  and the frontend's `/healthz` report the SHA they run.
- **Provisioning**: Terraform ([infra/terraform](../../infra/terraform/README.md))
  creates the server, DNS, buckets, KMS keys and IAM users, and writes the
  credentials it creates into Infisical. Ansible
  ([infra/ansible](../../infra/ansible/README.md)) sets the server up.
  **Neither releases the app.** Both run from any operator computer with its
  own SSH key and an `infisical login` ([setup.md](setup.md)).

### Backend run modes

One binary (`backend/cmd/app`, `./leviosa <mode>` in the image):

| Mode | Used by | Does |
| --- | --- | --- |
| `serve` (default) | the `backend` service | serves the API; never migrates; refuses to start on a schema **older** than its newest migration, accepts a newer one (rollback) |
| `migrate` | the deploy's `migrate` one-off | applies pending migrations; non-zero exit stops the deploy |
| `bootstrap-admin` | the deploy's `admin` one-off | creates the superadmin from `ADMIN_*` if that email does not exist; never modifies an existing account |
| `verify` | `leviosa-restore-test` | read-only: decrypts one record and checks the pepper canary (exit 3 to 6 name the failure) |
| `check-sms <to>`, `check-email <to>` | the operator, inside the backend container | one test message through the real provider settings; never prints a credential |

## Releasing

| Step | What happens |
| --- | --- |
| Push to `main` | `.github/workflows/ci.yaml` runs every check, builds and pushes both images for the commit, then runs `deploy staging <sha>` ([ci.md](ci.md)). A red check stops the release. |
| Check staging | `https://staging.leviosa.care` (basic auth). `/healthz` reports both SHAs. |
| **Promote** | Actions → *Promote to production* → Run workflow, **`sha` empty**: ships the exact release staging serves, after checking both apps report the same SHA, it is on `main` and both images exist ([promote.md](promote.md)). |

Both go through one server-side command, `deploy <stack> <sha>`
([infra/compose/README.md](../../infra/compose/README.md#deploy-stack-sha)):

1. pull both images for `<sha>` and re-point the stack's pins;
2. prod only: make sure Postgres, RabbitMQ and Redis are up;
3. `migrate` one-off. If it fails, the deploy stops and the running
   containers keep serving the previous release;
4. `bootstrap-admin` one-off;
5. recreate the backend, then the frontend;
6. wait until both are healthy and `/healthz` reports `<sha>` for both apps.

CI reaches that command through the **deploy gate**
([roles/deploy_gate](../../infra/ansible/roles/deploy_gate/README.md)): the CI
key can run `deploy <staging|prod> <40-hex sha>` for a commit on `main` and
nothing else. A root runner reads the deploy script and the compose file from
the repository at that SHA, so CI uploads nothing and a rollback rolls them
back too.

Known behaviour:

- Two quick pushes to `main` can deploy the older SHA after the newer one: the
  concurrency group only prevents overlapping deploys. Re-run the newer run's
  deploy job if that happens.
- The Agent's recreates do not take the deploy lock. A value change in the
  middle of a deploy can recreate the backend on the new pin before its
  migration ran; it then refuses to start (schema too old) and the deploy
  recreates it a moment later. It repairs itself.

### Rollback

Run **Promote** with the earlier SHA as input. The images, the compose file and
the deploy script roll back together. The database schema, its data and the
values in Infisical do **not**. That is why every migration follows the
migration rule. Steps and the migration check: [promote.md](promote.md#rollback).

### The migration rule

**Every migration must work with the previous release: add first, remove
later.** A deploy migrates while the previous release still serves, and a
rollback runs an older release against the newer schema indefinitely (down
migrations are never run). Add tables, and columns that are nullable or have a
default; rename or drop in steps across releases; never edit an applied
migration. The full rule is next to the migrations, in
[backend/CLAUDE.md](../../backend/CLAUDE.md) (*Database Migrations*).

### `COMING_SOON`

`COMING_SOON=true` in Infisical `prod/app` makes the frontend redirect every
page to `/coming-soon`, except `/healthz`, `/robots.txt` and `/sitemap.xml`;
`robots.txt` then disallows indexing. Production starts that way. Changing the
value reaches the frontend within about a minute (the Agent recreates it); no
deploy is needed. Turning it off is the launch, which waits on the encx
data-key cache release, the Zoho mailbox and a verified Twilio sender.

## When Infisical is down

- **The running stacks keep working.** The Agent keeps the last files it
  rendered and never blanks them when it cannot fetch. Deploys and one-offs
  read the files already on the server, so they keep working too.
- **A value change waits** until Infisical is back. The Agent retries on its
  own and applies it at the next successful poll.
- **Operator work that needs Infisical stops:** `make env`, `tf.sh`,
  `make setup` (it reads `/host`, `/db` and `/rabbitmq` through your login),
  `make infisical-agent`, the weekly snapshot.
- **Lockout.** Universal Auth locks the `leviosa-vps` identity out after
  repeated failed logins, and the Agent's retries keep renewing the lockout.
  Stop the Agent, clear the lockout in Infisical, start the Agent
  ([lockout recovery](../../infra/ansible/roles/infisical_agent/README.md#lockout-recovery)).
- **If Infisical is gone for good**, rebuild the project from the weekly
  encrypted snapshot (recovery C below). The pepper is in it: without it the
  encrypted data cannot be checked or re-hashed.

## Backups and recovery

Root systemd timers on the server
([roles/backup](../../infra/ansible/roles/backup/README.md)):

| What | When | Where (bucket `leviosa-care-prod-backups`) |
| --- | --- | --- |
| `pg_dump` of the production database, plus a configuration archive with no secret | daily, 03:00 Paris | `db/…`, `config/…` |
| Infisical `prod`, every folder, encrypted to the operator's `age` public key | Sunday, 04:00 Paris | `infisical/…` |
| `leviosa-restore-test`: the newest dump into a throwaway Postgres, checked with `verify` against production's KMS key and pepper | by hand: after the first backup, after a key or pepper change, monthly | — |

- **The backup credential cannot delete anything**; old backups expire through
  the bucket's lifecycle rule (90 days).
- **The server cannot read the snapshots**: the `age` private key exists only
  in the operator's password manager. Until
  `infra/ansible/roles/backup/files/backup-age.pub` holds the real public key,
  the weekly snapshot is not installed.
- **Not backed up:** staging, Redis, RabbitMQ. The media buckets rely on S3
  versioning.

The three recovery procedures are in
[roles/backup/README.md](../../infra/ansible/roles/backup/README.md#recovery):

- **A. Restore the production database** (server still there): test the dump
  with `leviosa-restore-test`, stop the apps, `pg_restore`, start the apps.
- **B. The server is lost**: `tf.sh apply -replace=hcloud_server.main` (the IPs
  and DNS stay), `make setup` with the Agent credential, deploy the release
  through CI, then restore the newest dump as in A.
- **C. Infisical is lost**: on an operator computer, decrypt the newest
  snapshot into memory-backed storage, re-create the `leviosa` project and its
  `prod` folders, point `infisical_project_id` at it, `tf.sh apply`, re-create
  `leviosa-vps`, `make setup`. Staging and dev are re-created by hand.

## Credentials and rotation

"Automatic" means: once the new value is in Infisical, the Agent re-renders the
file and recreates the consumer at its next poll (within about a minute), with
no command to run.

| Credential | Lives in | How to rotate | After the new value is in Infisical |
| --- | --- | --- | --- |
| `leviosa-vps` (the Agent's identity) | password manager; the server's root-only `/etc/infisical-agent/credentials/` | **every 3 months**, and at once if it may have leaked: new client secret in Infisical → `make infisical-agent` from `infra/ansible` (prompts for it) → check the Agent → **revoke the old secret** ([procedure](../../infra/ansible/roles/infisical_agent/README.md#rotation-every-three-months)) | manual (the Agent role re-run) |
| CI deploy key (`leviosa-ci`) | private half only in GitHub `SSH_PRIVATE_KEY`; public half in the keys file | new pair → replace the line in `deploy_ci_ssh_public_keys` → `make setup` → update the GitHub secret. Deploys fail between the last two steps | manual |
| Docker Hub token | password manager + GitHub `DOCKERHUB_TOKEN` | new read/write token in Docker Hub → update the GitHub secret → revoke the old one | manual (not in Infisical) |
| App AWS keys (`staging/aws`, `prod/aws`) | Infisical, **written by Terraform** | `./tf.sh apply -replace='aws_iam_access_key.app["prod"]'` (or `["staging"]`) | automatic: the backend is recreated |
| Backup AWS key (`prod/backup`) | Infisical, written by Terraform | `./tf.sh apply -replace=aws_iam_access_key.backup` | automatic: `backup.env` is rewritten; the next run reads it |
| Terraform's inputs (Hetzner, Cloudflare, AWS admin) | Infisical `leviosa-infra` | new token or key in the provider's console → update `leviosa-infra` → revoke the old one | read by the next `tf.sh` run |
| Stripe secret key and webhook secrets (`/stripe`) | Infisical | roll in the Stripe dashboard (a webhook secret is rolled on its endpoint) | automatic: the backend is recreated |
| Twilio API key (`/twilio`) | Infisical | new API key **in the IE1 region** → update `/twilio` → `check-sms` → delete the old key in Twilio | automatic: the backend is recreated |
| SMTP app password (`/smtp`, Zoho) | Infisical | new app password in Zoho → update `/smtp` → `check-email` → revoke the old one | automatic: the backend is recreated |
| Cloudflare DNS token for Caddy (`prod/caddy`; separate from Terraform's) | Infisical | new token (DNS edit on the zone) → update `prod/caddy` → revoke the old one | automatic: Caddy is recreated |
| Staging basic auth (`staging/caddy`) | Infisical; the password itself in the password manager | `caddy hash-password` → update `STAGING_BASIC_AUTH_USER` and `STAGING_BASIC_AUTH_HASH` **together** | **manual**: the Agent does not detect a change in `staging/caddy` on its own; restart the Agent and recreate Caddy ([procedure](../../infra/infisical/README.md#caddyenv-reads-two-environments)) |
| Postgres: production's role (`prod/db`) | Infisical | `ALTER ROLE … PASSWORD …` in the database **first**, then update `prod/db` (`POSTGRES_PASSWORD` applies only at first initialisation) | the backend gets the new `DATABASE_URL` automatically; the ALTER is manual |
| Postgres: staging's role (`staging/db`) | Infisical | update `staging/db`, then `make setup` at once | the staging backend gets the new URL automatically; the role's password is re-applied by the next `make setup` (`setup-shared-services.sh`) |
| RabbitMQ users (`prod/rabbitmq`, `staging/rabbitmq`) | Infisical | update the folder, then `make setup` at once | the backend gets the new URL automatically; the user's password is re-applied by the next `make setup` |
| Redis passwords (`/redis`) | Infisical | update the stack's `/redis` | automatic: that stack's Redis and backend are recreated |
| Session and booking-token secrets (`/app`) | Infisical | update `/app` | automatic: the backend is recreated |
| Superadmin (`/admin`) | Infisical + password manager | change the password in the app. `/admin` is read only when the account does not exist yet | — |
| Operator SSH keys | each computer's own key; public halves in the keys file | new key → keys file → `make setup` → check it logs in → remove the old key → `make setup` ([setup.md](setup.md)) | manual |
| `age` recovery key | password manager only | new pair → commit the public key → `make setup`. Keep the old private key 90 days (the snapshots it encrypted) | manual |
| Infisical 2FA recovery codes | password manager only | — | — |

**Never rotated:** the encx pepper (`/encx`: changing it makes every password
hash unverifiable) and the KMS keys (deleting one loses every encrypted field of
its environment; Terraform refuses to).

## The repository is public

The GitHub repository is public. The consequences:

- **Nothing secret is ever committed.** Values live in Infisical; the
  repository holds templates, key *names* and public identifiers: the
  Infisical project IDs, the Cloudflare zone ID, the `age` *public* key, the
  CI workflows, the server layout and hardening. None of them grants access on
  its own, and the security must not depend on any of it staying hidden.
- **Never commit:** a value from Infisical, a tfvars file other than the
  non-secret `infra/terraform/terraform.tfvars`, a rendered env file
  (`backend/development.env`, `frontend/.env`, `.env` are gitignored), the
  `leviosa-vps` credential, a private SSH key, the `age` private key.
- **Anything pushed stays public**, in forks and caches, even after the commit
  is removed. If a secret is pushed, **revoke it** at its provider. The Hetzner
  and Cloudflare API tokens once committed in
  `infra/terraform/terraform.tfvars.staging` and `.production` (in the
  history) must be treated as leaked and revoked in Hetzner and Cloudflare.
- **Anyone can fork and open pull requests.** Pull requests never build, push
  or deploy. CI deploys only on a push to `main` of this repository, and the
  deploy gate refuses any SHA not on `main`; a fork's commit is not on `main`
  even when GitHub serves it under this repository's URL.
- **`main` is protected** by a ruleset: no force-push, no deletion, no bypass
  (admins included). Whoever can push to `main` ships to staging, and then to
  production through Promote. The GitHub environment `production` should accept
  only `main`.
- **Keep secret scanning and push protection on** (Settings → Code security).
- **`docs/issues/`, `docs/prd/` and `notes/` are gitignored**: working notes
  stay local.

## Where to look

| Topic | Doc |
| --- | --- |
| Compose files, `deploy`, image pins, Caddy, the local harness | [infra/compose/README.md](../../infra/compose/README.md) |
| Env templates, `make env`, `make env-check-<env>`, folders and keys | [infra/infisical/README.md](../../infra/infisical/README.md) |
| Server setup, roles, the keys file, hardening | [infra/ansible/README.md](../../infra/ansible/README.md) |
| The deploy gate and root runner | [roles/deploy_gate/README.md](../../infra/ansible/roles/deploy_gate/README.md) |
| The Infisical Agent, its render matrix, rotation, lockout | [roles/infisical_agent/README.md](../../infra/ansible/roles/infisical_agent/README.md) |
| Backups, restore test, recovery | [roles/backup/README.md](../../infra/ansible/roles/backup/README.md) |
| Terraform, `tf.sh`, KMS, buckets, IAM | [infra/terraform/README.md](../../infra/terraform/README.md) |
| CI and its secrets | [ci.md](ci.md) |
| Promote and rollback | [promote.md](promote.md) |
