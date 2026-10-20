# infisical_agent

Turns the server into an [Infisical Agent](https://infisical.com/docs) host
(ADR-0017, issue 027): the Agent runs as a systemd service, renders one env
file per consumer from the `leviosa` Infisical project every 60 s — with
[issue 024's templates](../../../../infisical/README.md) — into the two
compose stack roots, and recreates the affected service when a rendered file
changes. It keeps the last render when Infisical is unreachable, so a
configuration change never depends on Infisical being up and an Infisical
outage never touches the running containers.

## What it installs

| Path | Mode | Purpose |
| --- | --- | --- |
| `/usr/bin/infisical` (apt, pinned `0.43.138`, held) | — | the CLI; ships the Agent (`infisical agent`) |
| `/etc/systemd/system/infisical-agent.service` | 0644 | `Restart=always`, `UMask=0077`, post-start hook |
| `/etc/infisical-agent/config.yaml` | 0644 | Agent config (no secrets: paths only) |
| `/etc/infisical-agent/credentials/` | 0700 | root-only |
| `/etc/infisical-agent/credentials/client-id`, `client-secret` | 0400 | the `leviosa-vps` machine identity |
| `/etc/infisical-agent/credentials/token-sink.json` | (0600 via umask) | the Agent's refreshable access token |
| `/etc/infisical-agent/templates/<env>-<name>.env.tmpl` | 0644 | per-stack copies of issue 024's templates, substituted by `render-template.sh` |
| `/etc/infisical-agent/post-start.sh` | 0750 | covers the Agent's first-render gap (below) |
| `<stack root>/env/*.env` | 0600 (root) | the rendered files (`/opt/leviosa/env`, `/opt/leviosa-staging/env`) |

## Render matrix

From `infisical_agent_templates` in `defaults/main.yml` (`name` = template
and rendered file, `env` = Infisical environment, `service` = the compose
service the file feeds — they differ on staging, where services carry a
`staging_` prefix so their names never collide with production's DNS aliases
on `leviosa_network`):

| Env | Template | Rendered to | Recreates (`up -d --no-deps --force-recreate …`) |
| --- | --- | --- | --- |
| prod | backend | `/opt/leviosa/env/backend.env` | `backend` in `/opt/leviosa` |
| prod | frontend | `/opt/leviosa/env/frontend.env` | `frontend` in `/opt/leviosa` |
| prod | redis | `/opt/leviosa/env/redis.env` | `redis` in `/opt/leviosa` |
| prod | caddy | `/opt/leviosa/env/caddy.env` | `caddy` in `/opt/leviosa` |
| prod | postgres | `/opt/leviosa/env/postgres.env` | nothing — `POSTGRES_PASSWORD` applies at first init only; rotation is a manual `ALTER ROLE` (ADR-0016) |
| prod | rabbitmq | `/opt/leviosa/env/rabbitmq.env` | nothing — `RABBITMQ_DEFAULT_*` apply at first init only; rotation is a manual `rabbitmqctl change_password` |
| prod | backup | `/opt/leviosa/env/backup.env` | nothing — the backup job (issue 028) reads the file fresh each run |
| prod | admin | `/opt/leviosa/env/admin.env` | nothing — the next deploy's `bootstrap-admin` one-off reads it |
| staging | backend | `/opt/leviosa-staging/env/backend.env` | `staging_backend` in `/opt/leviosa-staging` |
| staging | frontend | `/opt/leviosa-staging/env/frontend.env` | `staging_frontend` in `/opt/leviosa-staging` |
| staging | redis | `/opt/leviosa-staging/env/redis.env` | `staging_redis` in `/opt/leviosa-staging` |
| staging | admin | `/opt/leviosa-staging/env/admin.env` | nothing |

Staging renders only backend, frontend, admin and its own Redis: Postgres,
RabbitMQ and Caddy are production's (ADR-0016). `--no-deps` guarantees a
change can never bounce the shared services; `--force-recreate` is the
single-service recreate rule of `infra/compose/README.md` (`restart` does
not reload `env_file`). The stack roots and service names **must match** the
compose layout installed by the deploy runner (`infra/compose/`) — they are
role variables (`infisical_agent_stack_roots`, the `service` field) so a
layout change is one edit.

Because the rendered env files are root-only `0600`, anything that runs
`docker compose` on the stacks must do so as root: the deploy runner does
(`roles/deploy_gate`), and so does the Agent through its `execute` commands
and the post-start helper.

### A fresh server

`make setup` usually runs before the first deploy, so no
`docker-compose.yml` and no pinned image exist yet. Nothing in this role
starts half a stack in that state:

- the Agent never runs a template's `execute` command on the first render
  after a start, and setup is a first render;
- the post-start helper skips any stack root whose `docker-compose.yml` is
  not there yet ("the first deploy installs it");
- the helper only ever recreates services whose rendered file changed.

## The templates

The role copies the templates from `infra/infisical/templates` (issue 024)
after running [`render-template.sh`](../../../../infisical/render-template.sh)
on the control node — the same substitution `make env` performs, so there is
exactly one implementation of it. The script replaces `__LEVIOSA_ENV__` and
`__LEVIOSA_PROJECT_ID__`, reading the project ID from
`infisical_project_id` in `infra/terraform/terraform.tfvars`: the one place
the operator sets it (Terraform refuses to run until it is real either). If
it is still the placeholder, the role fails with the script's message
naming that file — the server must not be set up with a half-rendered Agent.

The copy uses the `copy` module, never `template`: the files contain Go
`{{ }}` braces, which Jinja would try to interpret.

## The credential

The `leviosa-vps` client ID and secret (Universal Auth, Viewer on `leviosa`
only) are **read from the operator at run time only** — never from a file in
the repo (ADR-0017):

- export `INFISICAL_VPS_CLIENT_ID` / `INFISICAL_VPS_CLIENT_SECRET` on the
  control machine (the role's defaults pick them up), or
- answer the prompts of [`playbooks/infisical-agent.yml`](../../playbooks/infisical-agent.yml)
  (`make infisical-agent`; the secret is not echoed).

Every task that touches the values runs with `no_log: true`, so they never
appear in Ansible output at any verbosity —
[tests/run-tests.sh](tests/run-tests.sh) re-checks that against `localhost`
in `--check` mode with a fake secret. On the server they live only in the
two `0400` root-owned files above.

**site.yml** includes the role after `app`: it runs when both environment
variables are set, and is **skipped with a note** otherwise, so a `make
setup` without the credential still succeeds; add the Agent later with
`make infisical-agent`.

**Excluded from backups.** `/etc/infisical-agent/credentials` (client ID,
client secret, token sink) must never leave the server: issue 028's
configuration archive copies the Agent *config* — which holds these paths,
not the values — and must exclude that directory (noted here for issue 028).

### Rotation (every three months)

ADR-0017 rotates the client secret every three months:

1. Create a **new client secret** for `leviosa-vps` in Infisical (Machine
   Identities); the old one stays valid until revoked.
2. Re-run the role with it: `INFISICAL_VPS_CLIENT_ID=…
   INFISICAL_VPS_CLIENT_SECRET=… make infisical-agent` (or answer the
   prompt). The Agent restarts with the new secret; the rendered files are
   unchanged, so the post-start helper recreates **no app container**.
3. Check the Agent: `systemctl status infisical-agent`, and that the env
   files' mtimes advance (`ls -l /opt/leviosa/env`).
4. **Revoke the old secret** in Infisical. If the Agent logs
   authentication failures afterwards, the new secret did not reach it —
   re-run step 2 before the identity locks out.

`--tags credential` re-places just the two files without the rest of the
role, and still restarts the Agent (the role's internal handler flush is
not tag-filtered) — a quick rotation path on an otherwise configured
server.

### Lockout recovery

Repeated failed authentications lock the identity out, and the Agent's own
retries keep renewing the lockout, so it never clears on its own while the
Agent runs. Recover:

```sh
ssh root@<server-ip>
systemctl stop infisical-agent          # stop the retries
# Infisical > Machine Identities > leviosa-vps: clear the lockout
systemctl start infisical-agent
```

### When Infisical is down

The Agent keeps the **last render** of every file and keeps retrying
(`Restart=always`, `RestartSec=10`); running containers, deploys and
one-offs read the files from disk, so none of them notice the outage.
Changes made during the outage apply on the next successful poll — within
about a minute of recovery — or, if the Agent was restarted in between,
through the post-start helper. The rendered files are never deleted.

## Agent behaviours the role compensates for

Findings from issue 005-style verification, recorded in
[infra/infisical/README.md](../../../../infisical/README.md) (checked against
infisical 0.43.138):

- **Go `{{ }}` braces**: the env templates are installed with `copy` after
  `render-template.sh` substitutes the placeholders on the control node —
  never the `template` module.
- **60 s minimum polling interval**: shorter values make the Agent shut
  down; the role asserts the configured interval is at least 60 s and
  defaults to `60s`.
- **Config errors exit with status 0**: the unit uses `Restart=always`, not
  `on-failure`.
- **Rendered files are created 0644**: the unit sets `UMask=0077`, so
  everything the Agent writes (env files, token sink) lands root-owned
  `0600`; the role additionally sweeps pre-existing rendered files to
  `0600`.
- **The `execute` command never runs on the first render after a start**:
  `ExecStartPre` snapshots per-file checksums before the Agent starts, and
  `ExecStartPost` (`post-start.sh apply`) waits for the first render, then
  recreates the service of each file whose content changed — only those.
  This covers *every* Agent restart (role re-run, rotation, systemd
  auto-restart, reboot), is selective (a changed `backend.env` does not
  touch the frontend), and is a no-op when nothing changed.
- **One `listSecrets` call per template**: a property of issue 024's
  templates (the ETag of the last call is what change-detects); the role
  adds none. See that README for the `caddy.env` call-order caveat.

## Usage

```sh
cd infra/ansible

# standalone install / rotation (prompts for the credential)
make infisical-agent

# or non-interactively
INFISICAL_VPS_CLIENT_ID=… INFISICAL_VPS_CLIENT_SECRET=… \
  ansible-playbook -i inventory/hosts.yml playbooks/infisical-agent.yml \
  -e "ansible_host=<server-ip>"
```

`playbooks/site.yml` runs the same role as part of `make setup` (after
`app`, before the shared services that consume the env files). Tags:
`infisical` / `agent` for the whole role, `credential` for the credential
files alone, `templates` for the render pipeline alone.

`infisical_agent_manage_service` (default `true`) is the single switch for
touching the systemd service: the enable/start task and both handlers skip
when it is false. No real run should set it; the role's tests use it to
exercise the file tasks against localhost, and it doubles as a repair
escape hatch (refresh files without restarting anything).

The role reads no secret from the repo: its only inputs are the run-time
credential and the committed templates.
