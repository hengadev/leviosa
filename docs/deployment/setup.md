# Setup checklist: a new operator computer

From zero to a computer that can run Terraform, set the server up and render
local env files. Nothing here depends on a file that exists on only one
machine: the Terraform state is in S3, every value is in Infisical, and each
computer has its own SSH key.

How the result works: [README.md](README.md). Where each value goes:
[environment.md](environment.md). Building the whole thing from nothing (a new
server, a new Infisical project) follows the same steps; the server-side
checks are in [infra/ansible/README.md](../../infra/ansible/README.md#verification).

## 1. Accounts (once per person)

- [ ] **Infisical** (`https://secrets.henga.dev`): an account with **2FA on**,
  its recovery codes in the password manager.
- [ ] Membership of both projects, `leviosa` and `leviosa-infra`, with write
  access to `leviosa`'s `/aws` and `/backup` (Terraform writes them through
  your login).
- [ ] Access to Hetzner, Cloudflare, AWS, Stripe, Twilio, Zoho, Docker Hub and
  the GitHub repository, for the rare manual steps.

## 2. Tools

- [ ] The Infisical CLI, Terraform ≥ 1.11, Ansible
  (`pipx install --include-deps ansible`, then `cd infra/ansible && make install-deps`),
  `jq`, `age`, and Docker + Go + pnpm if you also develop locally.

## 3. Infisical login

- [ ] `infisical login`, choosing the self-hosted instance
  `https://secrets.henga.dev`. Check it with `make env-check-dev` (key names
  only, never values).

Every Infisical-backed command uses this session: `make env`, `tf.sh`,
`make setup`. When it expires they stop with "infisical login".

## 4. This computer's SSH key

- [ ] **Its own key**, never copied to another computer:

  ```sh
  ssh-keygen -t ed25519 -f ~/.ssh/leviosa-$(hostname) -C leviosa-$(hostname)
  ```

- [ ] **Add it to the keys file.** The keys file lists who may log in:
  public keys only, kept outside the repository as
  `~/leviosa-deploy-keys.yml` (format: [infra/ansible/deploy-keys.example.yml](../../infra/ansible/deploy-keys.example.yml)).
  Keep one canonical copy (a note in the password manager) and put this
  computer's `.pub` line under `deploy_ssh_public_keys` there.
- [ ] **From a computer that already has access**, refresh its
  `~/leviosa-deploy-keys.yml` from the canonical copy and run `make setup`
  (from the repository root). The run writes root's and the deploy user's
  `authorized_keys` from the file: the new key gets in, and a key missing from
  the file loses access, so never run it from a stale copy.
- [ ] Copy the canonical keys file to this computer's `~/leviosa-deploy-keys.yml`.
- [ ] **`Host leviosa`** in `~/.ssh/config`, so `ssh` and Ansible use this key
  (the IP is `make infra-output`'s `server_ipv4_address`):

  ```
  Host leviosa <server ip>
      HostName <server ip>
      User root
      IdentityFile ~/.ssh/leviosa-<this computer>
      IdentitiesOnly yes
  ```

  With a passphrase, load the key once per terminal before Ansible:
  `eval "$(ssh-agent -s)" && ssh-add ~/.ssh/leviosa-<this computer>`.
- [ ] Check: `ssh leviosa true`. On the first connection, compare the host key
  fingerprint with one from a computer that already trusts the server.

**First server ever:** Hetzner seeds root with Terraform's `terraform-leviosa`
key only. Keep that key in `deploy_ssh_public_keys` for the first `make setup`,
check that each computer's own key logs in as root, then remove it and run
`make setup` again. **Removing a computer:** delete its line from the
canonical file and run `make setup` from another computer.

## 5. Terraform through `tf.sh`

```sh
cd infra/terraform
make init      # S3 state backend, pinned providers
make plan      # an existing setup: "No changes"
```

- [ ] `make plan` runs with no AWS profile, no tfvars file and no flags.
  `tf.sh` takes Terraform's credentials from `leviosa-infra` and hands your
  own Infisical session to the Infisical provider. Plain `terraform` has no
  credentials and fails on purpose.
- From the repository root: `make infra-init`, `make infra-plan`,
  `make infra-apply`, `make infra-output`.
- Details (first apply, server rebuild, AWS key rotation, KMS):
  [infra/terraform/README.md](../../infra/terraform/README.md).

## 6. Server setup: `make setup`

Only needed when something under `infra/ansible/` changed, a key was added or
removed, or the server was rebuilt.

```sh
make setup     # reads the IP from Terraform and the keys from ~/leviosa-deploy-keys.yml, asks to confirm
make setup     # a second run must report changed=0 (apart from "Upgrade all packages")
```

- To (re)install the Infisical Agent in the same run, export the `leviosa-vps`
  credential for that shell only, without echoing it (zsh shown; in bash
  `read -rsp "prompt: " VAR`):

  ```sh
  read -rs "INFISICAL_VPS_CLIENT_ID?leviosa-vps client id: "; export INFISICAL_VPS_CLIENT_ID
  read -rs "INFISICAL_VPS_CLIENT_SECRET?leviosa-vps client secret: "; export INFISICAL_VPS_CLIENT_SECRET
  ```

  Without them the Agent role is skipped with a note and the rest succeeds.
  The Agent alone (install or the 3-monthly rotation):
  `cd infra/ansible && make infisical-agent` (it prompts).
- Releases are **not** run from here: CI deploys staging, Promote ships
  production ([ci.md](ci.md), [promote.md](promote.md)).

## 7. Local development (optional)

```sh
make env         # renders backend/development.env and frontend/.env from Infisical dev (overwrites both)
make local-dev   # frontend natively, backend and its dependencies in Docker
```

Without Infisical, `make local-dev` copies the committed example files instead
([environment.md](environment.md#local-development)).

## Recurring

| When | What |
| --- | --- |
| Every 3 months | rotate the `leviosa-vps` client secret ([procedure](../../infra/ansible/roles/infisical_agent/README.md#rotation-every-three-months)) |
| Monthly, and after a KMS key or pepper change | `leviosa-restore-test` on the server, as root |
| Now and then | `journalctl -u leviosa-backup -u leviosa-snapshot --since -8d` on the server |
| Before each `make setup` | refresh `~/leviosa-deploy-keys.yml` from the canonical copy |
