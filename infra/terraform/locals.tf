# Shared names. Everything per environment is declared once for each key of
# local.environments, and every name below is derived from it.

locals {
  # Keys are the Infisical environment slugs of the `leviosa` project, so the
  # same word names an environment in AWS, in Infisical and in the deploy
  # script (`deploy staging|prod <sha>`).
  environments = toset(["staging", "prod"])

  # S3 bucket names are global across every AWS account, hence the
  # distinctive prefix (var.bucket_prefix). IAM users and KMS aliases are
  # account-local and use the project name.
  media_bucket_names = {
    for env in local.environments : env => "${var.bucket_prefix}-${env}-media"
  }
  backup_bucket_name = "${var.bucket_prefix}-prod-backups"

  # encx reads the KEK by alias and prepends "alias/" itself, so Infisical
  # holds the bare name (ENCX_KMS_KEY_ALIAS) and AWS the full one.
  kms_alias_names = {
    for env in local.environments : env => "${var.project_name}-${env}"
  }

  app_user_names = {
    for env in local.environments : env => "${var.project_name}-${env}-app"
  }
  backup_user_name = "${var.project_name}-prod-backup"

  # ------------------------------------------------------------------------
  # Infisical key names Terraform writes (infisical.tf). This is the only
  # place they are spelled: rename a key here and the next apply moves it.
  # The backend reads them (issue 010) and the Agent templates render them
  # (issue 024); README.md lists them.
  # ------------------------------------------------------------------------

  # leviosa/<staging|prod>/aws
  infisical_aws_keys = {
    access_key_id     = "AWS_ACCESS_KEY_ID"
    secret_access_key = "AWS_SECRET_ACCESS_KEY"
    region            = "AWS_REGION"
    media_bucket      = "S3_BUCKET_NAME"
    kms_key_arn       = "ENCX_KMS_KEY_ARN"
    kms_key_alias     = "ENCX_KMS_KEY_ALIAS"
  }

  # leviosa/prod/backup
  infisical_backup_keys = {
    access_key_id     = "BACKUP_AWS_ACCESS_KEY_ID"
    secret_access_key = "BACKUP_AWS_SECRET_ACCESS_KEY"
    bucket            = "BACKUP_S3_BUCKET"
    region            = "BACKUP_AWS_REGION"
  }
}
