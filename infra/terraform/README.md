# Leviosa Terraform

Infrastructure as code for Leviosa: the Hetzner server, Cloudflare DNS, and
the AWS side (media buckets, encx KMS keys, IAM users, production backup
bucket). Terraform also writes the credentials and identifiers it creates
into Infisical, so nobody copies them by hand.

One state in S3 holds everything, in the default workspace. Any operator
computer can plan and apply after `infisical login`, with no flags, no AWS
profile and no local tfvars file.

## What this manages

| Area | Resources |
| --- | --- |
| Hetzner | One `cpx22` in `nbg1` (`hcloud_server.main`), minimal cloud-init (Ansible does the rest), and its **primary IPs** (IPv4 + IPv6), which survive a server rebuild |
| Cloudflare | Proxied records for `leviosa.care`, `www`, `admin`, `staff`, `staging`, `admin-staging`, `staff-staging`. Mail records (MX, SPF, DKIM, verification, DMARC) from the `email_*` inputs, empty until Zoho provides them (issue 036) |
| AWS, per environment (`staging`, `prod`) | Private media bucket `leviosa-care-<env>-media` (SSE, versioning, non-current versions expire after 30 days); encx KEK: symmetric KMS key + `alias/leviosa-<env>`, `prevent_destroy`, 30-day deletion window; app IAM user `leviosa-<env>-app` with an access key and one policy: its own bucket and its own key only |
| AWS, production backups | Bucket `leviosa-care-prod-backups` (SSE, versioning, lifecycle expiry); IAM user `leviosa-prod-backup` that can list, read and write it but **not delete** |
| AWS, shared | The state bucket `leviosa-terraform-state` (imported, `prevent_destroy`) |
| Infisical | `/aws` in `staging` and `prod`, `/backup` in `prod` of the `leviosa` project (below) |

### IAM, exactly

| User | Allowed | On |
| --- | --- | --- |
| `leviosa-<env>-app` | `s3:ListBucket` | its media bucket |
| | `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` | its media bucket's objects |
| | `kms:Encrypt`, `kms:Decrypt`, `kms:GenerateDataKey*`, `kms:DescribeKey` | its KMS key |
| `leviosa-prod-backup` | `s3:ListBucket` | the backup bucket |
| | `s3:GetObject`, `s3:PutObject`, `s3:AbortMultipartUpload` | the backup bucket's objects |

No app or backup user can delete, disable or administer a KMS key, and the
backup user cannot delete an object (old backups go only through the
lifecycle rule; versioning keeps an overwritten object's previous version
for 30 days). `make check` proves this offline (see [Checks](#checks)).

## Credentials and inputs

- **Non-secret inputs** are in the committed [`terraform.tfvars`](terraform.tfvars):
  domain, zone ID, bucket prefix, server settings, mail records, the
  `leviosa` project ID. Never put a secret in it; it is the only tfvars file
  git accepts here.
- **Secret inputs** live in the Infisical project **`leviosa-infra`**
  (environment `prod`, path `/`) and reach Terraform as environment
  variables through `tf.sh`:

  | `leviosa-infra` key | Used for |
  | --- | --- |
  | `TF_VAR_hcloud_token` | Hetzner provider |
  | `TF_VAR_cloudflare_token` | Cloudflare provider (DNS edit on `leviosa.care`). Not the token Caddy uses: that one is `prod/caddy` `CLOUDFLARE_API_TOKEN` in `leviosa`, created separately |
  | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | AWS provider **and** the S3 state backend (an IAM user that can manage S3, IAM and KMS) |

  The server's `leviosa-vps` identity has no access to `leviosa-infra`.

### `tf.sh`, and why plain `terraform` fails

Every Makefile target that touches infrastructure, here and at the
repository root (`make infra-init|infra-plan|infra-apply|infra-output`), runs
`./tf.sh`, which:

1. runs `infisical run --projectId=<leviosa-infra> --env=prod --path=/`,
   injecting the secret inputs above (after dropping any `AWS_PROFILE` or
   session token from your shell, so these are the only AWS credentials used);
2. hands the Infisical **provider** your own session token
   (`infisical user get token`) as `INFISICAL_TOKEN` with
   `INFISICAL_AUTH_METHOD=token`. There is no machine identity for
   Terraform, so no extra long-lived credential exists. You need write
   access to the `leviosa` project's `/aws` and `/backup` folders.

**Running `terraform` directly does not work**: it has no Hetzner or
Cloudflare token (plan stops with "hcloud_token is empty: run Terraform
through ./tf.sh ..."), no AWS credentials for the state backend, and no
Infisical session for the provider. That is deliberate: no computer keeps a
tfvars file or profile with these credentials. Offline commands (`fmt`,
`validate`, `test`, see [Checks](#checks)) need none of them.

When your Infisical session expires, `tf.sh` stops with "Run: infisical login".

## What Terraform writes into Infisical

Into the **`leviosa`** project. The key names are defined once, in
[`locals.tf`](locals.tf) (`infisical_aws_keys`, `infisical_backup_keys`):
the backend reads them and the Agent templates render them.

| Environment / folder | Key | Value |
| --- | --- | --- |
| `staging/aws`, `prod/aws` | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | that environment's `leviosa-<env>-app` access key |
| | `AWS_REGION` | `eu-central-1` |
| | `S3_BUCKET_NAME` | `leviosa-care-<env>-media` |
| | `ENCX_KMS_KEY_ARN` | that environment's KMS key ARN (for reference) |
| | `ENCX_KMS_KEY_ALIAS` | `alias/leviosa-<env>`: the full alias name the backend finds its KEK by (it refuses key IDs and ARNs) |
| `prod/backup` | `BACKUP_AWS_ACCESS_KEY_ID`, `BACKUP_AWS_SECRET_ACCESS_KEY` | the `leviosa-prod-backup` access key |
| | `BACKUP_S3_BUCKET` | `leviosa-care-prod-backups` |
| | `BACKUP_AWS_REGION` | `eu-central-1` |

Nothing is written to `dev`: local development uses LocalStack and must
reach no real data. Every other folder (`/encx` pepper, `/db`, `/stripe`, ...)
is the operator's (issue 008).

## Before the first apply

The real IDs are not known to the repository yet. Fill in, commit, then run:

1. **`tf.sh`**: set `infra_project_id` to the `leviosa-infra` project ID
   (Infisical > leviosa-infra > Project settings). Until then `tf.sh` refuses
   to run. Check that its secrets are in environment `prod`, path `/`
   (otherwise adjust `infra_env` / `infra_path`).
2. **`terraform.tfvars`**: set `infisical_project_id` to the `leviosa`
   project ID. Until then plan refuses to run. This is the one place it is
   set: `infra/infisical/render-template.sh` (the Agent's env templates) and
   Ansible's `playbooks/site.yml` read the same value.
3. **`terraform.tfvars`**: confirm `cloudflare_zone_id` is the `leviosa.care`
   zone (Cloudflare > leviosa.care > Overview > Zone ID).
4. **Infisical `leviosa-infra`** holds the four keys listed above, and the
   `leviosa` project has the folders `/aws` in `staging` and `prod` and
   `/backup` in `prod` (issue 008). Terraform writes keys, not folders.
5. **Hetzner**: an SSH key named `terraform-leviosa` (`hcloud_ssh_key_name`)
   exists in the project. It only seeds root's `authorized_keys`; Ansible
   replaces it with the keys file.
6. **Old state gone** (issue 007): both old workspaces destroyed, the
   `production` workspace deleted, the `default` state empty, and the state
   bucket removed from the old state but still existing. The `import` blocks
   in [`backend.tf`](backend.tf) adopt it (and its versioning, encryption
   and public access block) on the first apply; afterwards they are no-ops.
7. **Bucket names**: S3 names are global. If `plan`/`apply` reports a name
   as taken, change `bucket_prefix` before anything holds data.

Then:

```bash
cd infra/terraform
infisical login        # once per computer / session
make init
make plan              # review: everything to add, the 4 state-bucket resources to import, nothing to destroy
make apply
```

After the apply, check in Infisical that `staging/aws`, `prod/aws` and
`prod/backup` are populated (issue 033).

## Usage

Prerequisites: Terraform >= 1.11, the Infisical CLI, `infisical login`, and
membership of both `leviosa-infra` and `leviosa`.

```bash
cd infra/terraform
make init        # S3 backend + providers (versions pinned in .terraform.lock.hcl)
make plan
make apply       # also updates /aws and prod/backup in Infisical
make output
./tf.sh <any terraform command>    # e.g. ./tf.sh state list
```

From the repository root: `make infra-init`, `make infra-plan`,
`make infra-apply`, `make infra-output`.

### Rebuilding the server

```bash
./tf.sh apply -replace=hcloud_server.main
```

The primary IPs are separate resources (delete protection, `prevent_destroy`),
so the new server keeps its addresses and the DNS records don't change. Then
re-run Ansible (`make setup`).

### Rotating AWS credentials

```bash
./tf.sh apply -replace='aws_iam_access_key.app["prod"]'    # or ["staging"]
./tf.sh apply -replace=aws_iam_access_key.backup
```

The new key is written to Infisical in the same apply; the Agent re-renders
the env file and recreates the consumer.

### The KMS keys

A KMS key cannot be recovered once deleted, and every encrypted field of its
environment would be lost. Terraform refuses to destroy them
(`prevent_destroy`). If a deletion is ever scheduled by hand, it waits 30
days and can be cancelled (`aws kms cancel-key-deletion`). Never remove the
`prevent_destroy` to "clean up".

## Checks

```bash
make check     # terraform fmt -check, validate, test, no terraform.workspace
```

Offline: no credentials, no state backend, no API call. `make test` runs
[`tests/offline.tftest.hcl`](tests/offline.tftest.hcl) with every provider
mocked, and asserts:

- each app policy names only its own bucket and its own key, with exactly the
  actions in the table above;
- no app or backup policy grants `kms:*`, key deletion, disabling or any key
  administration; the backup policy grants no `s3:Delete*` and no KMS;
- the Infisical keys, folders and environments listed above, and that each
  environment receives its own credentials, bucket and key;
- KMS deletion window, bucket privacy, versioning and expiry, the DNS
  hostnames, the server type and location;
- `plan` fails when the tokens `tf.sh` provides are missing.

## Files

```
main.tf            backend (S3, eu-central-1, lockfile) and providers
locals.tf          environments, resource names, Infisical key names
variables.tf       inputs (the two secrets default to empty and refuse to plan)
terraform.tfvars   non-secret input values (committed)
hetzner.tf         server + primary IPs
cloud-init.yml.tftpl  minimal cloud-init (python3 for Ansible)
cloudflare.tf      web records, mail records (empty until Zoho)
s3.tf              media bucket per environment
kms.tf             encx KEK + alias per environment
iam.tf             app IAM user, policy and access key per environment
backups.tf         production backup bucket and backup IAM user
infisical.tf       values written into the leviosa project
backend.tf         the state bucket itself (imported)
outputs.tf         non-secret outputs
tests/             offline tests (mocked providers)
tf.sh, Makefile    run Terraform through Infisical
```

## Troubleshooting

- **"set infra_project_id in ./tf.sh"**: step 1 of [Before the first apply](#before-the-first-apply).
- **"no Infisical session"**: run `infisical login`.
- **"hcloud_token is empty"**: you ran `terraform` directly, or the key in
  `leviosa-infra` is not named exactly `TF_VAR_hcloud_token` (same for
  `TF_VAR_cloudflare_token`).
- **"infisical_project_id must be the UUID ..."**: step 2 above.
- **Infisical "folder not found"** on apply: create `/aws` (staging, prod) or
  `/backup` (prod) in the `leviosa` project.
- **State lock held** after an interrupted run: `./tf.sh force-unlock <LOCK_ID>`.
