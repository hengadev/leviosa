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

# cax* is ARM, cx*/cpx* x86. Switching between the two cannot be done in
# place: apply with -replace=hcloud_server.main.
server_type     = "cax11"
server_location = "nbg1"
server_image    = "ubuntu-24.04"
# Off: Hetzner backups cost 20% of the server, and the server holds nothing
# that Ansible + Infisical do not rebuild; the database goes to S3 (backups.tf).
enable_backups      = false
hcloud_ssh_key_name = "terraform-leviosa"

# ============================================
# Cloudflare
# ============================================

domain_name        = "leviosa.care"
cloudflare_zone_id = "21d34caf8707b6977d1ca89537dae9a1"

# Mail: Zoho Mail receives (MX, issue 036), Brevo sends the app's emails
# (SMTP relay). Values from Zoho Mail admin > Domains and Brevo > Senders,
# domains & dedicated IPs > Domains. Empty values create no record.
email_mx_records = {
  "mx.zoho.eu"  = 10
  "mx2.zoho.eu" = 20
  "mx3.zoho.eu" = 50
}
# Zoho only: Brevo sends with its own bounce domain and passes DMARC through
# its DKIM below.
email_spf_includes = ["zohomail.eu"]
email_dkim_records = {
  "zmail._domainkey"  = { type = "TXT", content = "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCrQMAEpPpTaEbJ60VYLgDj1v0HvFfES/kMWY5JOpzd0S9/Kl8c2dQGckWlNgCmMPxhjZ5AnyWamaXWHzlsyoewPMVWn84MyvrQNx4g9BEG4yjiqsd5aeOcxa3JQy/5h6chZEQf66AIYey+YGlFhMJKlp0O7r0mQoA/ND6tPVD7awIDAQAB" }
  "brevo1._domainkey" = { type = "CNAME", content = "b1.leviosa-care.dkim.brevo.com" }
  "brevo2._domainkey" = { type = "CNAME", content = "b2.leviosa-care.dkim.brevo.com" }
}
email_verification_txt = [
  "zoho-verification=zb80254448.zmverify.zoho.eu",
  "brevo-code:ab539570b000c988e949280adccc4148",
]
email_dmarc = "v=DMARC1; p=reject"

# ============================================
# Infisical
# ============================================

infisical_host = "https://secrets.henga.dev"

# OPERATOR: the `leviosa` project's ID (Infisical > leviosa > Project
# settings). Plan refuses to run until this is a real project ID.
infisical_project_id = "43399331-2dac-4023-b49b-1abe352dbb85"
