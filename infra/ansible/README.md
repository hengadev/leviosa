# Leviosa Ansible

Ansible sets the server up; **it never releases the app**. It runs from any
operator computer that has two things: its own SSH key, listed in the keys
file, and an `infisical login`. There is no Ansible Vault, no encrypted
variable file and no secret on the operator's disk (ADR-0017).

Releases are `deploy <stack> <sha>`, run by CI through the deploy gate
(`roles/deploy_gate`, `infra/compose/README.md`): a push to `main` deploys
staging, and production is promoted from what staging runs. Nothing is
deployed from an operator computer.

## What `make setup` does

| Role | What it sets up |
| --- | --- |
| `system` | apt upgrade, base packages (`git` and `util-linux`/`flock` for the deploy runner and script), timezone, unattended security upgrades, kernel hardening (sysctl), `/run/shm` noexec |
| `docker` | Docker Engine and the Compose plugin from Docker's repository; `daemon.json` (log rotation, metrics on `127.0.0.1:9323` only, no inter-container traffic on the default bridge, `no-new-privileges`, `live-restore`); **nobody in the `docker` group** (it is root-equivalent: the run fails if anyone is); the external `leviosa_network` both stacks join |
| `user` | the deploy user `leviosa`, in its own group only; root's and the deploy user's `authorized_keys` from the keys file (below) |
| `deploy_gate` | the CI keys' forced command, the root deploy runner and the deploy user's **only** sudoers rule (see [its README](roles/deploy_gate/README.md)) |
| `ssh` | a drop-in read before every other one (`/etc/ssh/sshd_config.d/00-leviosa-hardening.conf`): keys only, root by key only, only `root` and `leviosa` may log in, no forwarding, modern algorithms; validated with `sshd -t` before it is installed |
| `fail2ban` | SSH jails (brute force and floods); notifications to the email below. The Caddy jail stays off until Caddy writes its access log (issue 026) |
| `firewall` | UFW: deny incoming, allow 22 (rate limited), 80 and 443, nothing else |
| `app` | the stack directories (below) |

Still to come in the same playbook, in this order: `caddy` and
`shared_services` (Postgres, RabbitMQ and Caddy up, staging's database role,
the RabbitMQ vhosts; issue 026), `infisical_agent` (the env files; issue 027)
and `backup` (issue 028).

### Stack directories (`roles/app`)

```
/opt/leviosa/                     production, compose project `leviosa`
├── docker-compose.yml            installed by the deploy runner at each deploy
├── env/                          root 0700: the Infisical Agent's renderings (0600)
├── data/postgres/                70:70 0700   (postgres:17-alpine)
├── data/rabbitmq/                100:101 0750 (rabbitmq:3-management-alpine)
├── data/redis/                   999:1000 0750 (redis:7-alpine)
└── caddy/{etc,data,config,logs}/ root; data/ and config/ 0700 (certificates)
/opt/leviosa-staging/             staging, compose project `leviosa-staging`
├── docker-compose.yml            installed by the deploy runner at each deploy
├── env/                          root 0700
└── data/redis/                   999:1000 0750
```

Everything else is root's, 0750. **Nothing belongs to the deploy user**: the
runner and the Agent run `docker compose` there as root. The data
directories belong to the user each image runs as (its entrypoint chowns
them to that user anyway); change `app_*_uid`/`_gid` in
`roles/app/defaults/main.yml` if an image changes.

### The notification email

fail2ban and unattended-upgrades mail one address. It lives in Infisical,
project `leviosa`, environment `prod`, folder `/host`, key `NOTIFY_EMAIL`.
`playbooks/site.yml` reads it on **your** computer at the start of every run:

```sh
infisical secrets get NOTIFY_EMAIL --projectId=<leviosa> --env=prod --path=/host --output=dotenv --silent
```

with your own `infisical login`; nothing is written to Infisical, and the
output is hidden from Ansible's log. When the lookup fails (not logged in, no
CLI, the secret missing) the run prints a `WARNING` and uses
`root@localhost`; fix the cause and run again. Until `infisical_project_id`
in `playbooks/site.yml` holds the real project ID (the same value as
`infisical_project_id` in `infra/terraform/terraform.tfvars`), the lookup is
skipped with the same warning.

## Prerequisites (each operator computer)

```sh
# Ansible (ansible-core with the collections, or the full `ansible` package)
pipx install --include-deps ansible
cd infra/ansible && make install-deps      # ansible.posix, community.general, community.docker

# Infisical CLI, logged in to secrets.henga.dev (member of `leviosa` and `leviosa-infra`)
infisical login

# This computer's own SSH key, listed in the keys file (below)
ssh-keygen -t ed25519 -C leviosa-<computer>
```

`make setup` reads the server's IP through `infra/terraform/tf.sh output`,
which also needs the `infisical login`.

## The keys file

Who may log in is one small YAML file of **public keys**, kept outside the
repository: `~/leviosa-deploy-keys.yml` by default (`make setup
KEYS_FILE=/absolute/path.yml` for another one). The committed
[`deploy-keys.example.yml`](deploy-keys.example.yml) shows the format, with
fake keys:

```yaml
deploy_ssh_public_keys:        # one per operator computer: root + the deploy user
  - "ssh-ed25519 AAAA… leviosa-laptop"
  - "ssh-ed25519 AAAA… leviosa-desktop"
deploy_ci_ssh_public_keys:     # the CI key: the deploy user only, forced to the deploy gate
  - "ssh-ed25519 AAAA… leviosa-ci"
```

- Each entry is one bare public key, as in the `.pub` file (`<type> <base64>
  <comment>`): no options, no line break. Name the computer in the comment.
  Accepted types: ed25519, ECDSA and RSA keys, and their `sk-` (security
  key) variants. `roles/user` stops before writing any key if an entry is
  malformed, if the operator list is empty, or if a key is in both lists.
- **Operator keys** are written to root's `authorized_keys` (the whole file)
  and to a managed block of the deploy user's. They are unrestricted.
- **CI keys** go only into the deploy user's `authorized_keys`, at the top,
  as `restrict,command="/usr/local/sbin/leviosa-deploy-gate" <key>`
  (`roles/deploy_gate`): they can run `deploy <staging|prod> <sha>` for a
  commit on `main` and nothing else. Never root.
- **The file is the whole truth.** A key removed from it loses access on the
  next run, and a key added on the server by hand is removed by it (root's
  file is rewritten; in the deploy user's, every line that is not one of the
  two managed blocks and carries no key from the file is deleted).
- It holds public keys only, so it is not a secret, but it decides who gets
  in: keep **one canonical copy** (for example a note in the password
  manager) and refresh your local file from it before each `make setup`. A
  run from a stale copy removes the keys that copy lacks.
- When Hetzner creates the server, root's only key is Terraform's
  `terraform-leviosa` key. Keep it in `deploy_ssh_public_keys` for the first
  run, check that each computer's own key logs in as root, then remove it.

### Adding a computer

1. On the new computer: `ssh-keygen -t ed25519 -C leviosa-<computer>`, and
   install Ansible, the collections and the Infisical CLI (prerequisites).
2. Add its `.pub` line to `deploy_ssh_public_keys` in the canonical keys file.
3. **From a computer that already has access**, refresh its local keys file
   and run `make setup`. The new key can now log in as root and as `leviosa`.
4. On the new computer: copy the keys file to `~/leviosa-deploy-keys.yml`,
   `infisical login`, and check `ssh root@<server-ip> true`.

### Rotating or removing a key

- **Remove** (lost or retired computer): delete its line from the canonical
  keys file and run `make setup` from another computer. The key loses root
  and the deploy user in that run.
- **Rotate** an operator key: add the new key, run `make setup`, check the new
  key logs in, then remove the old one and run `make setup` again.
- **Rotate the CI key**: generate a new pair, replace the line in
  `deploy_ci_ssh_public_keys`, run `make setup`, then replace the private key
  in the GitHub Actions secret. Deploys fail between the two steps, nothing
  else.

## Running it

```sh
infisical login
make setup                     # from the repository root, or from infra/ansible
make setup                     # again: must report changed=0
```

`make setup` reads the server IP (and domain) from Terraform through
`tf.sh`, shows them with the keys file it will use, asks for confirmation,
then runs:

```sh
ansible-playbook -i inventory/hosts.yml playbooks/site.yml \
  -e "ansible_host=<server-ip>" \
  -e @$HOME/leviosa-deploy-keys.yml \
  -e "app_domain=<domain>"
```

Run it on a fresh server, after `tf.sh apply -replace=hcloud_server.main`
(a rebuilt server keeps its IP), and after any change under `infra/ansible/`.
Ansible connects as root with your computer's key.

**A second run must report `changed=0`.** Anything else is a bug in a role
(a task that is not idempotent) or a change made on the server by hand that
the run just reverted; find out which before moving on. The exception is
`Upgrade all packages` when Ubuntu published updates between the two runs.

One role only: `ansible-playbook … --tags <tag>` with the same `-e` options.
Tags: `system`, `docker`, `user` (also runs `deploy_gate`), `deploy_gate`,
`ssh`, `fail2ban`, `firewall`, `app`, plus `setup` and `security` for groups
of them. On a server that is already set up, `--check --diff` shows what a
run would change.

## Variables

Inventory (`inventory/hosts.yml`): `ansible_host` (placeholder, always passed
by `make setup`), `ansible_user: root`, `app_domain`, `ssh_port`,
`ssh_permit_root_login: prohibit-password`, `ssh_password_authentication`,
`ssh_allowed_users: [root, leviosa]`.

Group variables (`playbooks/group_vars/all.yml`, no secrets):

| Variable | Default | Purpose |
| --- | --- | --- |
| `deploy_ssh_public_keys` | `[]` | operator keys, from the keys file |
| `deploy_ci_ssh_public_keys` | `[]` | CI keys, from the keys file (fed to `deploy_gate_ci_keys`) |
| `deploy_user` | `leviosa` | the deploy user |
| `app_prod_root`, `app_staging_root` | `/opt/leviosa`, `/opt/leviosa-staging` | stack roots (also `deploy_gate_stack_roots`; must match `infra/compose/deploy`) |
| `host_notify_email` | `root@localhost` | fallback when the `/host` lookup fails |
| `docker_network_name` | `leviosa_network` | the external network of both compose files |
| `system_timezone` | `Etc/UTC` | |
| `auto_security_updates`, `auto_updates_reboot` | `true`, `false` | unattended upgrades (Docker packages excluded) |
| `fail2ban_*` | see the file | jail thresholds; `fail2ban_caddy_enabled: false` until issue 026 |

Values that are secrets are not variables: they live in Infisical and reach
the server through the Agent's env files (issue 027).

## Verification

```sh
make check-clean        # ansible-playbook --syntax-check on every playbook
make lint               # ansible-lint, when installed
make deploy-gate-test   # from the repository root: the deploy gate's tests
```

After a setup, on the server (as root):

- [ ] `sshd -T | grep -Ei '^(permitrootlogin|passwordauthentication|allowusers|kexalgorithms)'`
      shows `without-password`/`prohibit-password`, `no`, `root leviosa` and the
      hardened algorithms
- [ ] `ufw status verbose`: deny incoming; 22 (LIMIT), 80, 443 only
- [ ] `fail2ban-client status` lists `sshd` and `sshd-ddos`
- [ ] `getent group docker` has no members; `id leviosa` shows only `leviosa`
- [ ] `sudo -l -U leviosa` lists only `/usr/local/sbin/leviosa-deploy`
- [ ] `/root/.ssh/authorized_keys` holds exactly the operator keys;
      `~leviosa/.ssh/authorized_keys` the CI block, then the operator block
- [ ] with the CI key, `ssh leviosa@<host> id` is refused by the gate
- [ ] `docker network inspect leviosa_network` exists
- [ ] `ls -la /opt/leviosa /opt/leviosa-staging` matches the layout above
- [ ] a second `make setup` reports `changed=0`

## Troubleshooting

- **Permission denied (publickey):** the key your SSH client offers must be
  in `deploy_ssh_public_keys`. With many keys in ssh-agent, sshd stops after
  3 attempts: `ssh -o IdentitiesOnly=yes -i ~/.ssh/<key> root@<ip>`.
- **"The keys file is missing or malformed":** `make setup` could not find
  `~/leviosa-deploy-keys.yml` (or `KEYS_FILE`), or an entry is not a bare
  public key.
- **WARNING: could not read NOTIFY_EMAIL:** `infisical login` again; check
  `infisical_project_id` in `playbooks/site.yml` and that `prod` `/host` has
  `NOTIFY_EMAIL`.
- **"Could not read the VPS IP from Terraform":** `infisical login`, then
  `cd ../terraform && ./tf.sh output server_ipv4_address`.
- **"The docker group has members":** someone was added by hand;
  `gpasswd -d <user> docker` and run again.
- **A playbook fails:** re-run with `-vvv`.

## Directory structure

```
infra/ansible/
├── ansible.cfg
├── Makefile                    # make setup, check-clean, lint, status/logs
├── requirements.yml            # collections
├── deploy-keys.example.yml     # the keys file's format (fake keys)
├── inventory/hosts.yml
├── playbooks/
│   ├── group_vars/all.yml      # defaults, no secrets
│   └── site.yml                # the whole setup
└── roles/
    ├── system/                 # packages, unattended upgrades, sysctl
    ├── docker/                 # Docker, daemon.json, leviosa_network
    ├── user/                   # deploy user, authorized_keys from the keys file
    ├── deploy_gate/            # CI forced command, root runner, sudoers rule
    ├── ssh/                    # sshd hardening drop-in
    ├── fail2ban/               # SSH (and later Caddy) jails
    ├── firewall/               # UFW: 22, 80, 443
    └── app/                    # stack directories
```
