# The IAM user each environment's backend runs with. Its only policy allows
# that environment's media bucket and that environment's KMS key, so staging
# can read nothing of production's (PRD user story 13). Its access key is
# written to Infisical <env>/aws (infisical.tf), never copied by hand.

locals {
  # The only KMS actions an app user gets: what encx needs to wrap and unwrap
  # data keys. No key administration: no ScheduleKeyDeletion, DisableKey,
  # PutKeyPolicy, kms:* ... (checked by tests/policies.tftest.hcl).
  app_kms_actions = ["kms:Encrypt", "kms:Decrypt", "kms:GenerateDataKey*", "kms:DescribeKey"]

  app_policies = {
    for env in local.environments : env => {
      Version = "2012-10-17"
      Statement = [
        {
          Sid      = "ListOwnMediaBucket"
          Effect   = "Allow"
          Action   = ["s3:ListBucket"]
          Resource = [aws_s3_bucket.media[env].arn]
        },
        {
          Sid      = "ReadWriteOwnMediaObjects"
          Effect   = "Allow"
          Action   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
          Resource = ["${aws_s3_bucket.media[env].arn}/*"]
        },
        {
          Sid      = "UseOwnKek"
          Effect   = "Allow"
          Action   = local.app_kms_actions
          Resource = [aws_kms_key.encx[env].arn]
        },
      ]
    }
  }
}

resource "aws_iam_user" "app" {
  for_each = local.environments
  name     = local.app_user_names[each.key]

  tags = {
    Environment = each.key
  }
}

resource "aws_iam_user_policy" "app" {
  for_each = local.environments
  name     = "${local.app_user_names[each.key]}-media-and-kek"
  user     = aws_iam_user.app[each.key].name
  policy   = jsonencode(local.app_policies[each.key])
}

# Rotate with: ./tf.sh apply -replace='aws_iam_access_key.app["prod"]'
resource "aws_iam_access_key" "app" {
  for_each = local.environments
  user     = aws_iam_user.app[each.key].name
}
