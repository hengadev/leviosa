# Values Terraform creates, written into the Infisical `leviosa` project so
# nobody copies an IAM key or a key ARN by hand. The Infisical Agent on the
# server renders them into the env files; the key names are defined once, in
# locals.tf (infisical_aws_keys, infisical_backup_keys).
#
#   leviosa/staging/aws, leviosa/prod/aws  (that environment's backend)
#     AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY  <- aws_iam_access_key.app[<env>]
#     AWS_REGION                                <- var.aws_region
#     S3_BUCKET_NAME                            <- aws_s3_bucket.media[<env>]
#     ENCX_KMS_KEY_ARN                          <- aws_kms_key.encx[<env>]
#     ENCX_KMS_KEY_ALIAS                        <- aws_kms_alias.encx[<env>] ("alias/<name>")
#
#   leviosa/prod/backup  (the server's backup job)
#     BACKUP_AWS_ACCESS_KEY_ID, BACKUP_AWS_SECRET_ACCESS_KEY <- aws_iam_access_key.backup
#     BACKUP_S3_BUCKET                                       <- aws_s3_bucket.backups
#     BACKUP_AWS_REGION                                      <- var.aws_region
#
# Nothing is written to `dev`: it uses LocalStack and must reach no real data.
# The folders (/aws in staging and prod, /backup in prod) must already exist.

locals {
  # field => value, per environment. Indexed by the field names of
  # local.infisical_aws_keys.
  infisical_aws_values = {
    for env in local.environments : env => {
      access_key_id     = aws_iam_access_key.app[env].id
      secret_access_key = aws_iam_access_key.app[env].secret
      region            = var.aws_region
      media_bucket      = aws_s3_bucket.media[env].bucket
      kms_key_arn       = aws_kms_key.encx[env].arn
      kms_key_alias     = aws_kms_alias.encx[env].name
    }
  }

  # "<env>/<NAME>" => { env, field }, built from static names only (no
  # resource attribute, nothing sensitive) so for_each is known at plan time.
  infisical_aws_secrets = {
    for pair in setproduct(local.environments, keys(local.infisical_aws_keys)) :
    "${pair[0]}/${local.infisical_aws_keys[pair[1]]}" => { env = pair[0], field = pair[1] }
  }

  infisical_backup_values = {
    access_key_id     = aws_iam_access_key.backup.id
    secret_access_key = aws_iam_access_key.backup.secret
    bucket            = aws_s3_bucket.backups.bucket
    region            = var.aws_region
  }
}

resource "infisical_secret" "aws" {
  for_each = local.infisical_aws_secrets

  workspace_id = var.infisical_project_id
  env_slug     = each.value.env
  folder_path  = "/aws"
  name         = local.infisical_aws_keys[each.value.field]
  value        = local.infisical_aws_values[each.value.env][each.value.field]
}

resource "infisical_secret" "backup" {
  # NAME => field
  for_each = { for field, name in local.infisical_backup_keys : name => field }

  workspace_id = var.infisical_project_id
  env_slug     = "prod"
  folder_path  = "/backup"
  name         = each.key
  value        = local.infisical_backup_values[each.value]
}
