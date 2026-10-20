# encx KEK per environment (ADR-0015): a symmetric KMS key and its alias.
#
# Losing a key makes every field it encrypted unreadable, so the key cannot
# be destroyed by Terraform (prevent_destroy) and, if deletion is ever
# scheduled by hand, AWS waits the longest window (30 days) during which it
# can be cancelled. No app or backup credential may schedule deletion or
# disable a key (iam.tf, backups.tf).
#
# The key policy is AWS's default: the account administers it and IAM
# policies grant its use, so each app user can use only its own key.
# Automatic rotation is off: each rotation adds a monthly charge, and the key
# material never leaves KMS.

resource "aws_kms_key" "encx" {
  for_each = local.environments

  description              = "Leviosa ${each.key} encx KEK"
  key_usage                = "ENCRYPT_DECRYPT"
  customer_master_key_spec = "SYMMETRIC_DEFAULT"
  deletion_window_in_days  = 30

  tags = {
    Name        = "leviosa-${each.key}-encx-kek"
    Environment = each.key
  }

  lifecycle {
    prevent_destroy = true
  }
}

# The backend finds its key by this alias (ENCX_KMS_KEY_ALIAS in /aws).
resource "aws_kms_alias" "encx" {
  for_each = local.environments

  name          = "alias/${local.kms_alias_names[each.key]}"
  target_key_id = aws_kms_key.encx[each.key].key_id
}
