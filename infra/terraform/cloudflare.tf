# Cloudflare DNS for leviosa.care.

locals {
  # Web hosts, all proxied to the one server; Caddy routes each to its stack
  # (production: apex, www, admin, staff; staging: staging, admin-staging,
  # staff-staging). No `analytics` host: Plausible is not on this server.
  web_subdomains = toset(["www", "admin", "staff", "staging", "admin-staging", "staff-staging"])
}

resource "cloudflare_dns_record" "apex_a" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "A"
  content = hcloud_primary_ip.main["ipv4"].ip_address
  proxied = true
  ttl     = 1
}

resource "cloudflare_dns_record" "apex_aaaa" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "AAAA"
  content = cidrhost(hcloud_primary_ip.main["ipv6"].ip_network, 1) # the server answers on ::1 of its /64
  proxied = true
  ttl     = 1
}

resource "cloudflare_dns_record" "web" {
  for_each = local.web_subdomains

  zone_id = var.cloudflare_zone_id
  name    = "${each.key}.${var.domain_name}"
  type    = "CNAME"
  content = var.domain_name
  proxied = true
  ttl     = 1
}

# ============================================
# Mailbox provider (Zoho Mail, issue 036)
# ============================================
# Driven by the email_* inputs in terraform.tfvars; none exist until Zoho
# provides them. Mail records are never proxied.

resource "cloudflare_dns_record" "mx" {
  for_each = var.email_mx_records

  zone_id  = var.cloudflare_zone_id
  name     = var.domain_name
  type     = "MX"
  content  = each.key
  priority = each.value
  proxied  = false
  ttl      = 3600
}

resource "cloudflare_dns_record" "spf" {
  count = length(var.email_spf_includes) > 0 ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "TXT"
  content = "\"v=spf1 ${join(" ", [for d in var.email_spf_includes : "include:${d}"])} ~all\""
  proxied = false
  ttl     = 3600
}

resource "cloudflare_dns_record" "dkim" {
  for_each = var.email_dkim_records

  zone_id = var.cloudflare_zone_id
  name    = "${each.key}.${var.domain_name}"
  type    = each.value.type
  content = each.value.type == "TXT" ? "\"${each.value.content}\"" : each.value.content
  proxied = false
  ttl     = 3600
}

resource "cloudflare_dns_record" "verification" {
  for_each = toset(var.email_verification_txt)

  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "TXT"
  content = "\"${each.key}\""
  proxied = false
  ttl     = 3600
}

resource "cloudflare_dns_record" "dmarc" {
  count = var.email_dmarc != "" ? 1 : 0

  zone_id = var.cloudflare_zone_id
  name    = "_dmarc.${var.domain_name}"
  type    = "TXT"
  content = "\"${var.email_dmarc}\""
  proxied = false
  ttl     = 3600
}
