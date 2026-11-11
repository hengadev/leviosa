# Cloudflare DNS for leviosa.care.

locals {
  # The hosts the server answers, each with its own A and AAAA records:
  # production's livio, admin and staff; staging's staging, admin-staging and
  # staff-staging. Caddy routes each to its stack. None depends on the apex,
  # so the apex can be pointed elsewhere without moving them. One level under
  # the apex only: Cloudflare's universal certificate covers *.leviosa.care,
  # not *.livio.leviosa.care. No `analytics` host: Plausible is not on this
  # server.
  server_hosts = toset(["livio", "admin", "staff", "staging", "admin-staging", "staff-staging"])
}

resource "cloudflare_dns_record" "web" {
  for_each = local.server_hosts

  zone_id = var.cloudflare_zone_id
  name    = "${each.key}.${var.domain_name}"
  type    = "A"
  content = hcloud_primary_ip.main["ipv4"].ip_address
  proxied = true
  ttl     = 1
}

resource "cloudflare_dns_record" "web_aaaa" {
  for_each = local.server_hosts

  zone_id = var.cloudflare_zone_id
  name    = "${each.key}.${var.domain_name}"
  type    = "AAAA"
  content = cidrhost(hcloud_primary_ip.main["ipv6"].ip_network, 1) # the server answers on ::1 of its /64
  proxied = true
  ttl     = 1

  # Cloudflare refuses an AAAA record next to a CNAME of the same name: these
  # come once each host's CNAME (before the server_hosts records) is an A.
  depends_on = [cloudflare_dns_record.web]
}

# The apex follows Livio's site for now (Caddy redirects it there); point it
# at the apex's own site when it has one. A CNAME at the apex is flattened by
# Cloudflare, and sits next to the mail records.
resource "cloudflare_dns_record" "apex" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain_name
  type    = "CNAME"
  content = "livio.${var.domain_name}"
  proxied = true
  ttl     = 1

  # Only once the apex's AAAA record has left (it becomes livio's, below):
  # Cloudflare refuses a CNAME next to an A or AAAA of the same name.
  depends_on = [cloudflare_dns_record.web_aaaa]
}

# www goes wherever the apex goes.
resource "cloudflare_dns_record" "www" {
  zone_id = var.cloudflare_zone_id
  name    = "www.${var.domain_name}"
  type    = "CNAME"
  content = var.domain_name
  proxied = true
  ttl     = 1
}

# From the layout where every host was a CNAME to the apex's A and AAAA: the
# records keep their Cloudflare IDs. The apex's AAAA becomes livio's, which
# takes it off the apex before the CNAME arrives there (depends_on above).
moved {
  from = cloudflare_dns_record.apex_a
  to   = cloudflare_dns_record.apex
}

moved {
  from = cloudflare_dns_record.apex_aaaa
  to   = cloudflare_dns_record.web_aaaa["livio"]
}

moved {
  from = cloudflare_dns_record.web["www"]
  to   = cloudflare_dns_record.www
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
