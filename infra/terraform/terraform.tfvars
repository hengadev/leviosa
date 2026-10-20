# Leviosa Terraform inputs: NON-SECRET values only. This file is committed.
#
# Secrets never go here. They live in the Infisical project `leviosa-infra`
# and reach Terraform as environment variables through ./tf.sh:
#   TF_VAR_hcloud_token, TF_VAR_cloudflare_token,
#   AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY (AWS provider + state backend)

project_name = "leviosa"
aws_region   = "eu-central-1"

# S3 bucket names are global; change this before the first apply if a name
# is taken (the old production bucket needed a suffix for that reason).
bucket_prefix = "leviosa-care"

media_noncurrent_version_days  = 30
backup_retention_days          = 90
backup_noncurrent_version_days = 30

# ============================================
# Hetzner Cloud
# ============================================

server_type         = "cpx22"
server_location     = "nbg1"
server_image        = "ubuntu-24.04"
enable_backups      = true
hcloud_ssh_key_name = "terraform-leviosa"

# ============================================
# Cloudflare
# ============================================

domain_name        = "leviosa.care"
cloudflare_zone_id = "21d34caf8707b6977d1ca89537dae9a1"

# Zoho Mail (issue 036): fill in from Zoho Mail admin > Domains once the
# mailbox exists, then apply. Empty values create no record.
email_mx_records       = {}
email_spf_includes     = []
email_dkim_records     = {}
email_verification_txt = []
email_dmarc            = ""

# ============================================
# Infisical
# ============================================

infisical_host = "https://secrets.henga.dev"

# OPERATOR: the `leviosa` project's ID (Infisical > leviosa > Project
# settings). Plan refuses to run until this is a real project ID.
infisical_project_id = "REPLACE_WITH_LEVIOSA_PROJECT_ID"
