# The mail records were first created by hand in Cloudflare; these blocks
# adopt them into the state on the next apply. Delete this file once that
# apply has run (the blocks are no-ops afterwards).
import {
  to = cloudflare_dns_record.mx["mx.zoho.eu"]
  id = "${var.cloudflare_zone_id}/701a4cf4a37fbecc396180d55d042be8"
}

import {
  to = cloudflare_dns_record.mx["mx2.zoho.eu"]
  id = "${var.cloudflare_zone_id}/7d4405fb5fa180c104e50f459d3657b8"
}

import {
  to = cloudflare_dns_record.mx["mx3.zoho.eu"]
  id = "${var.cloudflare_zone_id}/721a6bedbabb51d06d93dbd61a7ecc05"
}

import {
  to = cloudflare_dns_record.spf[0]
  id = "${var.cloudflare_zone_id}/50065ab0e660a0a204ed7191d04f4ed4"
}

import {
  to = cloudflare_dns_record.dkim["zmail._domainkey"]
  id = "${var.cloudflare_zone_id}/1f32264bdc94a1e533096267a2e85bf3"
}

import {
  to = cloudflare_dns_record.dkim["brevo1._domainkey"]
  id = "${var.cloudflare_zone_id}/755203972c4dafd5b3f9281d0bf86ff9"
}

import {
  to = cloudflare_dns_record.dkim["brevo2._domainkey"]
  id = "${var.cloudflare_zone_id}/d686ae5fe52cf945f8119da26491aa33"
}

import {
  to = cloudflare_dns_record.verification["zoho-verification=zb80254448.zmverify.zoho.eu"]
  id = "${var.cloudflare_zone_id}/033b4d6d0a16fc6fabf13acf7540429d"
}

import {
  to = cloudflare_dns_record.verification["brevo-code:ab539570b000c988e949280adccc4148"]
  id = "${var.cloudflare_zone_id}/1feccdabb52e02f15101c33be7172c63"
}

import {
  to = cloudflare_dns_record.dmarc[0]
  id = "${var.cloudflare_zone_id}/0fe362e51c4e8c72579461309d3fe5ed"
}
