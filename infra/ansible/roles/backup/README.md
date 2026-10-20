# Role `backup` (issue 028)

Production backups and their restore test. Everything runs as **root**
from systemd timers. The deploy user has no docker and no sudo beyond the
deploy runner (`roles/deploy_gate`), and this role gives it none.
`playbooks/site.yml` runs it last, after `shared_services` (tags `backup`,
`setup`).

| What | When | Where in the backup bucket |
| --- | --- | --- |
| `leviosa-backup`: `pg_dump -Fc` of production's database | daily, 03:00 Paris | `db/<YYYY>/<MM>/<DD>/leviosa_prod_<UTC>.dump` |
| `leviosa-backup`: configuration archive (allowlist, below) | same run | `config/<YYYY>/<MM>/<DD>/leviosa_config_<UTC>.tar.gz` |
| `leviosa-snapshot`: Infisical `prod`, every folder, `age`-encrypted | Sunday, 04:00 Paris | `infisical/<YYYY>/<MM>/<DD>/leviosa_prod_infisical_<UTC>.json.age` |
| `leviosa-restore-test`: the newest dump, restored and `verify`-ed | by hand | — |

`<YYYY>/<MM>/<DD>` and `<UTC>` are the UTC date and time of the run, so keys
sort by age. The bucket (`infra/terraform/backups.tf`) is versioned and has
one lifecycle rule: every object expires after `backup_retention_days` (90),
and an overwritten version after `backup_noncurrent_version_days` (30).
Nothing on the server deletes a backup, and it could not: the backup IAM
user may list, read and write the bucket, never delete.

The newest dump also stays on the server as `/var/backups/leviosa/latest.dump`
(root 0600), for a quick restore after a mistake. Its user fields are
encrypted (encx), like the database itself.

**Not backed up:** staging (test data, rebuilt by a deploy), Redis (sessions,
OTPs, rate limits: all short-lived) and RabbitMQ (messages in flight). The
media buckets rely on S3 versioning instead (`infra/terraform/s3.tf`).

## Where the credentials come from

- **S3:** `BACKUP_AWS_ACCESS_KEY_ID`, `BACKUP_AWS_SECRET_ACCESS_KEY`,
  `BACKUP_S3_BUCKET`, `BACKUP_AWS_REGION` (Infisical prod `/backup`, written
  by Terraform) and `POSTGRES_USER` / `POSTGRES_DB` (prod `/db`), in the
  Agent-rendered `/opt/leviosa/env/backup.env`, read fresh on every run.
  Ansible never sees them.
- **The restore test** reads production's `ENCX_KMS_KEY_ALIAS`,
  `ENCX_PEPPER` and AWS credentials from `/opt/leviosa/env/backend.env`.
- The env files are **parsed, never sourced** (a value is data), and never
  passed with `docker run --env-file`, which would keep the single quotes
  the Agent writes (issue 024). A value reaches a container by name
  (`-e NAME`, exported in a subshell), never on a command line, and no
  script prints one.
- **AWS CLI:** not packaged for Ubuntu 24.04, so it runs from the pinned
  `amazon/aws-cli` image, with no host directory mounted.
- **The snapshot** logs in with the Agent's own machine identity
  (`leviosa-vps`): `/etc/infisical-agent/credentials/client-id` and
  `client-secret` are read into the CLI's environment
  (`INFISICAL_UNIVERSAL_AUTH_CLIENT_ID` / `_SECRET`, then `INFISICAL_TOKEN`
  for the export). The server needs no second Infisical credential, and
  `ps` never shows one.

`/etc/leviosa/backup.conf` (root 0644) holds only paths and names: what the
scripts read, from this role's defaults.

## The configuration archive never contains a secret

The archive is an **allowlist** (`backup_config_paths`): both stacks'
`docker-compose.yml`, `/opt/leviosa/caddy/etc` (the Caddyfile), the Agent's
`config.yaml` and its `templates/`, plus a `manifest.txt` with the image of
every container of both compose projects (the deployed SHAs). No `env/`
file, no `/etc/infisical-agent/credentials` (client ID, client secret,
token sink) and no `caddy/data` (certificates and keys) can get in.

Two more checks before anything is uploaded: the job refuses an archive
that **lists** a secret-looking path (`env/`, `credentials/`, `*.env`,
`token-sink`, `client-id`/`client-secret`, `*.key`, `*.pem`), and one that
**contains** any value of a secret-looking key (`*PASSWORD`, `*SECRET`,
`*TOKEN`, `*PEPPER`, `*HASH`, `*_KEY`, `*_KEY_ID`, `*_SID`, `DATABASE_URL`,
`RABBITMQ_URL`) of either stack's env files, or the Agent's credential.
Every secret lives in Infisical, and the weekly snapshot covers losing
Infisical itself.

## The age key pair

The weekly snapshot is encrypted to the operator's **age public key**,
committed in `files/backup-age.pub`. The private key exists **only** in the
password manager: the server can write snapshots and never read them back.

The committed file is a **placeholder**. While it is, the role does **not**
install `leviosa-snapshot` or its timer, and `make setup` prints a warning
(it also refuses while `infisical_project_id` in
`infra/terraform/terraform.tfvars` is a placeholder). To set the key up,
once, on the operator's computer:

```bash
age-keygen            # prints "# public key: age1…" and the AGE-SECRET-KEY-1… line, writes nothing
```

1. Store the whole output (both lines) in the password manager as
   "Leviosa backup age key". Do not save it to a file.
2. Replace `infra/ansible/roles/backup/files/backup-age.pub` with the single
   `age1…` line, commit it.
3. `make setup` (or `ansible-playbook … playbooks/site.yml --tags backup`).

Rotation: same steps with a new pair. Keep the old private key as long as
snapshots made with it are kept (90 days).

## Everyday commands (on the server, as root)

```bash
systemctl list-timers 'leviosa-*'                 # both timers, next runs
systemctl start leviosa-backup                    # back up now
systemctl start leviosa-snapshot                  # snapshot now
journalctl -u leviosa-backup -u leviosa-snapshot --since -2d
leviosa-restore-test                              # newest dump in the bucket
leviosa-restore-test db/2026/10/20/leviosa_prod_20261020T010000Z.dump
leviosa-restore-test --local                      # /var/backups/leviosa/latest.dump
```

`leviosa-restore-test` takes the dump (from the bucket: the backup IAM user
may list and read it; `--local` for the server's latest copy), restores it
into a throwaway `postgres:17-alpine` container on its own Docker network
(no host port), and runs `henga/leviosa-backend:prod` in `verify` mode
against it with production's KMS key and pepper. `verify` decrypts an
`auth.users` record and checks the pepper canary. The script prints
`PASS` or `FAIL (verify exit N): …` and exits with `verify`'s status:
0 OK, 3 nothing to verify, 4 canary missing, 5 wrong KMS key, 6 wrong
pepper, 1 anything else. `--backend-env <file>` takes the encx settings from
another file. The containers, the network and the downloaded dump are
removed on every exit. Production is never touched.

Run it after the first backup on a new server, after a KMS key or pepper
change, and monthly.

## Recovery

### A. Restore the production database (server still there)

1. **Pick the dump.** `/var/backups/leviosa/latest.dump` is the newest. For
   an older one, list and download from the bucket with the backup
   credential:

   ```bash
   bash -c '. /usr/local/lib/leviosa-backup/lib.sh; backup_init
     aws s3 ls "$bucket/db/" --recursive' | tail
   bash -c '. /usr/local/lib/leviosa-backup/lib.sh; backup_init
     aws s3 cp "$bucket/db/<YYYY>/<MM>/<DD>/<file>.dump" -' > /root/restore.dump
   chmod 600 /root/restore.dump
   ```

2. **Test it first:** `leviosa-restore-test db/<YYYY>/<MM>/<DD>/<file>.dump`
   must print `PASS`. A fail with exit 5 or 6 means the dump was written
   under another KMS key or pepper: restoring it would give unreadable
   user data.
3. **Stop the apps** so nothing writes during the restore:
   `docker compose --project-directory /opt/leviosa stop frontend backend`.
4. **Restore** into the running Postgres (production's roles exist there,
   so owners are kept):

   ```bash
   bash -c '. /usr/local/lib/leviosa-backup/lib.sh; backup_init
     docker exec -i "$(db_container)" pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
       --clean --if-exists --single-transaction --exit-on-error' < /root/restore.dump
   ```

5. **Start the apps** (`docker compose --project-directory /opt/leviosa
   up -d --no-deps backend frontend`), check `https://leviosa.care/healthz`,
   then `rm /root/restore.dump`.
6. A dump brings back data deleted after it was taken (accounts erased on
   request, for instance): re-apply those deletions.

### B. The server is lost

1. Rebuild it through the normal path: Terraform (`infra/terraform/tf.sh
   apply`), then `make setup` with the Agent credential
   (`INFISICAL_VPS_CLIENT_ID` / `_SECRET`). The Agent renders the env files,
   `shared_services` starts an empty Postgres with its roles, this role
   installs the timers.
2. Deploy the release (`deploy prod <sha>` through CI; the SHA is in the
   newest configuration archive's `manifest.txt` if needed). The deploy
   migrates the empty database and bootstraps the admin.
3. Restore the newest dump as in **A** (from the bucket: a new server has no
   `latest.dump`), stopping the apps first.
4. `leviosa-restore-test`, then the next morning check that the daily backup
   ran (`systemctl list-timers 'leviosa-*'`, `journalctl -u leviosa-backup`).

The configuration archive is a reference (what the compose files, Caddyfile
and Agent config were); the repository at the deployed SHA is the source of
all of them.

### C. Infisical is lost: recover from the encrypted snapshot

On the operator's computer, never on the server. You need `age`, `jq`, the
Infisical CLI, the private key from the password manager and read access to
the backup bucket (AWS console, or the backup credential if it survived).

1. **Download** the newest `infisical/<YYYY>/<MM>/<DD>/leviosa_prod_infisical_<UTC>.json.age`.
2. **Decrypt into memory-backed storage only** (`$XDG_RUNTIME_DIR` is a
   per-user tmpfs). Paste the private key from the password manager:

   ```bash
   R=$XDG_RUNTIME_DIR/leviosa-recovery; install -d -m 700 "$R"
   install -m 600 /dev/null "$R/age.key"; $EDITOR "$R/age.key"
   age -d -i "$R/age.key" leviosa_prod_infisical_<UTC>.json.age > "$R/snapshot.json"
   jq -r '.takenAt, (.secrets | length), ([.secrets[].path] | unique | join(" "))' "$R/snapshot.json"
   ```

3. **Re-create the `leviosa` project** (new instance or new project) with
   `dev`, `staging` and `prod` environments, `infisical login`, then every
   prod folder and its values. Values go through `secrets set --file` as
   YAML (JSON is valid YAML), which keeps multi-line values intact:

   ```bash
   S=$R/snapshot.json; P=<new project id>
   for path in $(jq -r '[.secrets[].path] | unique | sort[]' "$S"); do
     [ "$path" = / ] || infisical secrets folders create --projectId "$P" --env prod \
       --path "$(dirname "$path")" --name "$(basename "$path")"
     jq --arg p "$path" '[.secrets[] | select(.path == $p) | {(.key): .value}] | add' "$S" > "$R/folder.yaml"
     infisical secrets set --projectId "$P" --env prod --path "$path" --file "$R/folder.yaml" > /dev/null
   done
   rm -rf "$R"
   ```

   Values are the raw ones (secret references are kept as references).
   Comments are in the snapshot (`.comment`) but are not re-applied.
4. **Point everything at the new project:** `infisical_project_id` in
   `infra/terraform/terraform.tfvars` (the one place: Terraform, the Agent
   templates and this role read it), then `tf.sh apply` (re-writes `/aws`
   and `/backup`), re-create the `leviosa-vps` machine identity (Viewer on
   the project) and re-run `make setup` with its new credential.
5. **Staging and dev are not in the snapshot.** Re-create them by hand
   (staging uses test-mode values; dev holds nothing of production's).

The pepper (`/encx` `ENCX_PEPPER`) is in the snapshot: without it, hashed
lookups on the restored database cannot be recomputed. Rehearse step C once
into a scratch project, so the procedure is known to work before it is
needed.

## Tests

`make backup-test` (`tests/run-tests.sh`; `BACKUP_TEST_ARGS=--static` skips
Docker) runs the role unprivileged against localhost into a temporary tree
(placeholder key: no weekly job; real key: installed; back: removed; second
run: `changed=0`), checks the units with `systemd-analyze verify` and the
schedules, shellchecks the scripts, then end to end with Docker: LocalStack
(KMS, S3), the backend built from the checkout with the compose harness's
fixtures, an encrypted admin and the pepper canary; the daily backup's
objects and archive (no secret in it, refused when one is added), the
snapshot through a stub Infisical CLI (unreadable without the private key),
the restore test (passes; exit 6 with a wrong pepper), and no secret in any
output. The snapshot template itself is checked against the fake Agent
engine in `infra/infisical/envcheck/snapshot_test.go`
(`make agent-templates-test`).

What only the server can show: the timers in `systemctl list-timers`, real
uploads with the backup IAM key, the Infisical CLI's universal-auth login
and template export against the real instance, and `verify` against real
KMS.
