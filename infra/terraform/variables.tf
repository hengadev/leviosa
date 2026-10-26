# Input variables. Non-secret values are set in the committed terraform.tfvars;
# the two secret ones arrive as TF_VAR_* from Infisical `leviosa-infra`
# through tf.sh, and are never written to a file.

# ============================================
# Secrets (from Infisical leviosa-infra, via tf.sh)
# ============================================

# Both default to empty so that running `terraform` directly fails with a
# clear message instead of prompting for a token.

variable "hcloud_token" {
  description = "Hetzner Cloud API token (TF_VAR_hcloud_token in Infisical leviosa-infra)"
  type        = string
  sensitive   = true
  default     = ""

  validation {
    condition     = length(var.hcloud_token) > 0
    error_message = "hcloud_token is empty: run Terraform through ./tf.sh, which loads TF_VAR_hcloud_token from Infisical leviosa-infra (infisical login first)."
  }
}

variable "cloudflare_token" {
  description = "Cloudflare API token, DNS edit on the zone (TF_VAR_cloudflare_token in Infisical leviosa-infra)"
  type        = string
  sensitive   = true
  default     = ""

  validation {
    condition     = length(var.cloudflare_token) > 0
    error_message = "cloudflare_token is empty: run Terraform through ./tf.sh, which loads TF_VAR_cloudflare_token from Infisical leviosa-infra (infisical login first)."
  }
}

# ============================================
# Project and AWS
# ============================================

variable "project_name" {
  description = "Project name used in account-local resource names (IAM users, KMS aliases, server)"
  type        = string
  default     = "leviosa"
}

variable "aws_region" {
  description = "AWS region of the buckets and KMS keys"
  type        = string
  default     = "eu-central-1"
}

variable "bucket_prefix" {
  description = "Prefix of every S3 bucket name (bucket names are global across AWS accounts)"
  type        = string
  default     = "leviosa-care"

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]{1,40}[a-z0-9]$", var.bucket_prefix))
    error_message = "bucket_prefix must be lowercase letters, digits and dashes (S3 bucket naming)."
  }
}

variable "media_noncurrent_version_days" {
  description = "Days a replaced or deleted media object version is kept before it expires"
  type        = number
  default     = 30
}

variable "backup_retention_days" {
  description = "Days a backup object is kept before the lifecycle rule expires it"
  type        = number
  default     = 90

  validation {
    condition     = var.backup_retention_days >= 7 && var.backup_retention_days <= 365
    error_message = "backup_retention_days must be between 7 and 365."
  }
}

variable "backup_noncurrent_version_days" {
  description = "Days an overwritten or expired backup version is kept (time to notice a bad overwrite)"
  type        = number
  default     = 30
}

# ============================================
# Hetzner Cloud
# ============================================

variable "server_type" {
  description = "Hetzner Cloud server type (cax11, cx23, cpx22, etc.). cax* is ARM: the app images are built for linux/arm64 too"
  type        = string
  default     = "cax11"

  validation {
    condition     = can(regex("^(cax|cx|cpx)[0-9]{2}$", var.server_type))
    error_message = "server_type must be a shared-vCPU Hetzner plan (e.g., cax11, cx23, cpx22)."
  }
}

variable "server_location" {
  description = "Hetzner Cloud location"
  type        = string
  default     = "nbg1"

  validation {
    condition     = contains(["nbg1", "fsn1", "hel1", "hil", "ash", "sin"], var.server_location)
    error_message = "server_location must be a Hetzner location code."
  }
}

variable "server_image" {
  description = "Server OS image"
  type        = string
  default     = "ubuntu-24.04"
}

variable "enable_backups" {
  description = "Hetzner automatic server backups (extra cost)"
  type        = bool
  default     = true
}

variable "hcloud_ssh_key_name" {
  description = "Name of an SSH key that already exists in the Hetzner project; it becomes root's first key, before Ansible writes the keys file"
  type        = string
  default     = "terraform-leviosa"
}

# ============================================
# Cloudflare DNS
# ============================================

variable "domain_name" {
  description = "Apex domain"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID of domain_name (not a secret)"
  type        = string

  validation {
    condition     = can(regex("^[0-9a-f]{32}$", var.cloudflare_zone_id))
    error_message = "cloudflare_zone_id must be the 32-character zone ID shown on the zone's overview page in Cloudflare."
  }
}

# Mailbox provider (Zoho, issue 036). All empty until Zoho provides them;
# an empty value creates no record.

variable "email_mx_records" {
  description = "Inbound MX hosts, mapped to their priority (e.g. { \"mx.zoho.eu\" = 10 })"
  type        = map(number)
  default     = {}
}

variable "email_spf_includes" {
  description = "Domains included in the apex SPF record (no SPF record while empty)"
  type        = list(string)
  default     = []
}

variable "email_dkim_records" {
  description = "Mailbox provider DKIM records, keyed by record name relative to the domain (e.g. zmail._domainkey)"
  type        = map(object({ type = string, content = string }))
  default     = {}
}

variable "email_verification_txt" {
  description = "Apex TXT records proving domain ownership (e.g. Zoho's zoho-verification=...)"
  type        = list(string)
  default     = []
}

variable "email_dmarc" {
  description = "Content of the _dmarc TXT record, e.g. \"v=DMARC1; p=quarantine\" (no record while empty)"
  type        = string
  default     = ""
}

# ============================================
# Infisical
# ============================================

variable "infisical_host" {
  description = "Infisical instance the provider writes to"
  type        = string
  default     = "https://secrets.henga.dev"
}

variable "infisical_project_id" {
  description = "ID of the Infisical `leviosa` project, where Terraform writes /aws and prod /backup"
  type        = string

  validation {
    condition     = can(regex("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", var.infisical_project_id))
    error_message = "infisical_project_id must be the UUID of the Infisical `leviosa` project (Project settings > Project ID); set it in terraform.tfvars."
  }
}
