# Offline checks of what this configuration would create: `make test`
# (`terraform test`). Every provider is mocked, so nothing is contacted and no
# credential is needed. Values AWS would compute (ARNs, access keys) are
# pinned below so the assertions can tell staging's from production's.

mock_provider "aws" {
  override_during = plan
}

mock_provider "hcloud" {
  override_during = plan
}

mock_provider "cloudflare" {
  override_during = plan
}

mock_provider "infisical" {
  override_during = plan
}

# Mock providers cannot import: stand in for the state bucket (backend.tf).
override_resource {
  target          = aws_s3_bucket.terraform_state
  override_during = plan
  values          = { arn = "arn:aws:s3:::leviosa-terraform-state" }
}

override_resource {
  target          = aws_s3_bucket_versioning.terraform_state
  override_during = plan
}

override_resource {
  target          = aws_s3_bucket_server_side_encryption_configuration.terraform_state
  override_during = plan
}

override_resource {
  target          = aws_s3_bucket_public_access_block.terraform_state
  override_during = plan
}

override_resource {
  target          = aws_s3_bucket.media["staging"]
  override_during = plan
  values          = { arn = "arn:aws:s3:::leviosa-care-staging-media" }
}

override_resource {
  target          = aws_s3_bucket.media["prod"]
  override_during = plan
  values          = { arn = "arn:aws:s3:::leviosa-care-prod-media" }
}

override_resource {
  target          = aws_s3_bucket.backups
  override_during = plan
  values          = { arn = "arn:aws:s3:::leviosa-care-prod-backups" }
}

override_resource {
  target          = aws_kms_key.encx["staging"]
  override_during = plan
  values = {
    arn    = "arn:aws:kms:eu-central-1:111111111111:key/00000000-0000-0000-0000-00000000000a"
    key_id = "00000000-0000-0000-0000-00000000000a"
  }
}

override_resource {
  target          = aws_kms_key.encx["prod"]
  override_during = plan
  values = {
    arn    = "arn:aws:kms:eu-central-1:111111111111:key/00000000-0000-0000-0000-00000000000b"
    key_id = "00000000-0000-0000-0000-00000000000b"
  }
}

override_resource {
  target          = aws_iam_access_key.app["staging"]
  override_during = plan
  values          = { id = "AKIASTAGINGAPP", secret = "staging-app-secret" }
}

override_resource {
  target          = aws_iam_access_key.app["prod"]
  override_during = plan
  values          = { id = "AKIAPRODAPP", secret = "prod-app-secret" }
}

override_resource {
  target          = aws_iam_access_key.backup
  override_during = plan
  values          = { id = "AKIAPRODBACKUP", secret = "prod-backup-secret" }
}

variables {
  # What tf.sh provides from Infisical leviosa-infra.
  hcloud_token     = "test-hcloud-token"
  cloudflare_token = "test-cloudflare-token"
  # terraform.tfvars holds a placeholder until the operator fills it in.
  infisical_project_id = "00000000-0000-0000-0000-000000000000"
}

# ---------------------------------------------------------------------------
# IAM: each app user gets its own bucket and its own key, nothing else
# ---------------------------------------------------------------------------

run "app_policy_names_only_its_own_bucket_and_key" {
  command = plan

  assert {
    condition = toset(flatten([
      for s in jsondecode(aws_iam_user_policy.app["staging"].policy).Statement : s.Resource
      ])) == toset([
      "arn:aws:s3:::leviosa-care-staging-media",
      "arn:aws:s3:::leviosa-care-staging-media/*",
      "arn:aws:kms:eu-central-1:111111111111:key/00000000-0000-0000-0000-00000000000a",
    ])
    error_message = "The staging app policy must name only the staging media bucket and the staging KMS key."
  }

  assert {
    condition = toset(flatten([
      for s in jsondecode(aws_iam_user_policy.app["prod"].policy).Statement : s.Resource
      ])) == toset([
      "arn:aws:s3:::leviosa-care-prod-media",
      "arn:aws:s3:::leviosa-care-prod-media/*",
      "arn:aws:kms:eu-central-1:111111111111:key/00000000-0000-0000-0000-00000000000b",
    ])
    error_message = "The prod app policy must name only the prod media bucket and the prod KMS key."
  }

  assert {
    condition = alltrue([
      for env in ["staging", "prod"] : toset(flatten([
        for s in jsondecode(aws_iam_user_policy.app[env].policy).Statement : s.Action
        ])) == toset([
        "s3:ListBucket", "s3:GetObject", "s3:PutObject", "s3:DeleteObject",
        "kms:Encrypt", "kms:Decrypt", "kms:GenerateDataKey*", "kms:DescribeKey",
      ])
    ])
    error_message = "An app policy may allow only list/get/put/delete on its bucket and encrypt/decrypt/generate-data-key/describe on its key."
  }

  assert {
    condition = alltrue([
      for env in ["staging", "prod"] : alltrue([
        for s in jsondecode(aws_iam_user_policy.app[env].policy).Statement :
        s.Effect == "Allow" && toset(keys(s)) == toset(["Sid", "Effect", "Action", "Resource"])
      ])
    ])
    error_message = "App policy statements must be plain Allow statements (no NotAction/NotResource, no conditions to reason about)."
  }

  assert {
    condition     = aws_iam_user_policy.app["staging"].user == "leviosa-staging-app" && aws_iam_user_policy.app["prod"].user == "leviosa-prod-app"
    error_message = "Each app policy must be attached to its own environment's app user."
  }
}

run "no_app_or_backup_user_can_delete_or_disable_a_key" {
  command = plan

  assert {
    condition = alltrue([
      for policy in [
        aws_iam_user_policy.app["staging"].policy,
        aws_iam_user_policy.app["prod"].policy,
        aws_iam_user_policy.backup.policy,
        ] : alltrue([
          for action in flatten([for s in jsondecode(policy).Statement : s.Action]) :
          !contains(["*", "kms:*"], action) && !can(regex("^kms:(ScheduleKeyDeletion|DisableKey|Delete|CancelKeyDeletion|PutKeyPolicy|Create|Update|Enable|Disable|Revoke|Retire)", action))
      ])
    ])
    error_message = "No app or backup policy may grant kms:*, key deletion/disabling or any key administration."
  }
}

run "backup_user_can_write_and_read_but_not_delete" {
  command = plan

  assert {
    condition = toset(flatten([
      for s in jsondecode(aws_iam_user_policy.backup.policy).Statement : s.Action
    ])) == toset(["s3:ListBucket", "s3:GetObject", "s3:PutObject", "s3:AbortMultipartUpload"])
    error_message = "The backup policy may allow only list, get, put and abort-multipart-upload."
  }

  assert {
    condition = alltrue([
      for action in flatten([for s in jsondecode(aws_iam_user_policy.backup.policy).Statement : s.Action]) :
      !can(regex("^s3:(\\*|Delete|PutBucket|PutLifecycle)", action)) && !startswith(action, "kms:") && action != "*"
    ])
    error_message = "The backup user must not delete objects, change bucket settings or use KMS."
  }

  assert {
    condition = toset(flatten([
      for s in jsondecode(aws_iam_user_policy.backup.policy).Statement : s.Resource
    ])) == toset(["arn:aws:s3:::leviosa-care-prod-backups", "arn:aws:s3:::leviosa-care-prod-backups/*"])
    error_message = "The backup policy must name only the production backup bucket."
  }

  assert {
    condition = alltrue([
      for s in jsondecode(aws_iam_user_policy.backup.policy).Statement :
      s.Effect == "Allow" && toset(keys(s)) == toset(["Sid", "Effect", "Action", "Resource"])
    ])
    error_message = "Backup policy statements must be plain Allow statements."
  }
}

# ---------------------------------------------------------------------------
# KMS and buckets
# ---------------------------------------------------------------------------

run "kms_key_per_environment" {
  command = plan

  assert {
    condition     = alltrue([for env in ["staging", "prod"] : aws_kms_key.encx[env].deletion_window_in_days == 30])
    error_message = "Each KMS key must have a 30-day deletion window."
  }

  assert {
    condition     = alltrue([for env in ["staging", "prod"] : aws_kms_key.encx[env].customer_master_key_spec == "SYMMETRIC_DEFAULT" && aws_kms_key.encx[env].key_usage == "ENCRYPT_DECRYPT"])
    error_message = "Each KMS key must be a symmetric encrypt/decrypt key."
  }

  assert {
    condition     = aws_kms_alias.encx["staging"].name == "alias/leviosa-staging" && aws_kms_alias.encx["prod"].name == "alias/leviosa-prod"
    error_message = "KMS aliases must be alias/leviosa-staging and alias/leviosa-prod."
  }

  assert {
    condition     = aws_kms_alias.encx["staging"].target_key_id == aws_kms_key.encx["staging"].key_id && aws_kms_alias.encx["prod"].target_key_id == aws_kms_key.encx["prod"].key_id
    error_message = "Each alias must point at its own environment's key."
  }
}

run "buckets_private_versioned_and_expiring" {
  command = plan

  assert {
    condition = alltrue([
      for b in concat([for env in ["staging", "prod"] : aws_s3_bucket_public_access_block.media[env]], [aws_s3_bucket_public_access_block.backups]) :
      b.block_public_acls && b.block_public_policy && b.ignore_public_acls && b.restrict_public_buckets
    ])
    error_message = "Media and backup buckets must block all public access."
  }

  assert {
    condition = alltrue([
      for v in concat([for env in ["staging", "prod"] : aws_s3_bucket_versioning.media[env]], [aws_s3_bucket_versioning.backups]) :
      v.versioning_configuration[0].status == "Enabled"
    ])
    error_message = "Media and backup buckets must be versioned."
  }

  assert {
    condition     = alltrue([for env in ["staging", "prod"] : aws_s3_bucket_lifecycle_configuration.media[env].rule[0].noncurrent_version_expiration[0].noncurrent_days == 30])
    error_message = "Non-current media versions must expire after 30 days."
  }

  assert {
    condition     = aws_s3_bucket_lifecycle_configuration.backups.rule[0].expiration[0].days == var.backup_retention_days
    error_message = "Backups must expire through the lifecycle rule."
  }
}

# ---------------------------------------------------------------------------
# What Terraform writes into Infisical
# ---------------------------------------------------------------------------

run "infisical_aws_per_environment" {
  command = plan

  assert {
    condition = toset(keys(infisical_secret.aws)) == toset(flatten([
      for env in ["staging", "prod"] : [
        for name in ["AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_REGION", "S3_BUCKET_NAME", "ENCX_KMS_KEY_ARN", "ENCX_KMS_KEY_ALIAS"] :
        "${env}/${name}"
      ]
    ]))
    error_message = "Terraform must write exactly the six /aws keys into staging and prod."
  }

  assert {
    condition = alltrue([
      for k, s in infisical_secret.aws :
      s.folder_path == "/aws" && "${s.env_slug}/${s.name}" == k && s.workspace_id == "00000000-0000-0000-0000-000000000000"
    ])
    error_message = "Every /aws secret must go to its environment's /aws folder of the leviosa project."
  }

  assert {
    condition = (
      infisical_secret.aws["staging/AWS_ACCESS_KEY_ID"].value == "AKIASTAGINGAPP" &&
      infisical_secret.aws["staging/AWS_SECRET_ACCESS_KEY"].value == "staging-app-secret" &&
      infisical_secret.aws["prod/AWS_ACCESS_KEY_ID"].value == "AKIAPRODAPP" &&
      infisical_secret.aws["prod/AWS_SECRET_ACCESS_KEY"].value == "prod-app-secret"
    )
    error_message = "Each environment's /aws must hold its own app user's access key."
  }

  assert {
    condition = (
      infisical_secret.aws["staging/S3_BUCKET_NAME"].value == "leviosa-care-staging-media" &&
      infisical_secret.aws["prod/S3_BUCKET_NAME"].value == "leviosa-care-prod-media" &&
      infisical_secret.aws["staging/ENCX_KMS_KEY_ARN"].value == aws_kms_key.encx["staging"].arn &&
      infisical_secret.aws["prod/ENCX_KMS_KEY_ARN"].value == aws_kms_key.encx["prod"].arn &&
      infisical_secret.aws["staging/ENCX_KMS_KEY_ALIAS"].value == "leviosa-staging" &&
      infisical_secret.aws["prod/ENCX_KMS_KEY_ALIAS"].value == "leviosa-prod" &&
      infisical_secret.aws["staging/AWS_REGION"].value == "eu-central-1" &&
      infisical_secret.aws["prod/AWS_REGION"].value == "eu-central-1"
    )
    error_message = "Each environment's /aws must hold its own bucket, KMS key ARN and alias, and the region."
  }
}

run "infisical_backup_in_prod_only" {
  command = plan

  assert {
    condition     = toset(keys(infisical_secret.backup)) == toset(["BACKUP_AWS_ACCESS_KEY_ID", "BACKUP_AWS_SECRET_ACCESS_KEY", "BACKUP_S3_BUCKET", "BACKUP_AWS_REGION"])
    error_message = "Terraform must write exactly the four /backup keys."
  }

  assert {
    condition     = alltrue([for s in infisical_secret.backup : s.env_slug == "prod" && s.folder_path == "/backup"])
    error_message = "Backup credentials must go to prod /backup only."
  }

  assert {
    condition = (
      infisical_secret.backup["BACKUP_AWS_ACCESS_KEY_ID"].value == "AKIAPRODBACKUP" &&
      infisical_secret.backup["BACKUP_AWS_SECRET_ACCESS_KEY"].value == "prod-backup-secret" &&
      infisical_secret.backup["BACKUP_S3_BUCKET"].value == "leviosa-care-prod-backups" &&
      infisical_secret.backup["BACKUP_AWS_REGION"].value == "eu-central-1"
    )
    error_message = "prod /backup must hold the backup user's key, the backup bucket and the region."
  }
}

# ---------------------------------------------------------------------------
# Hetzner and Cloudflare
# ---------------------------------------------------------------------------

run "server_and_dns" {
  command = plan

  assert {
    condition     = hcloud_server.main.server_type == "cpx22" && hcloud_server.main.location == "nbg1"
    error_message = "The server must be a cpx22 in nbg1."
  }

  assert {
    condition     = alltrue([for ip in hcloud_primary_ip.main : !ip.auto_delete && ip.delete_protection])
    error_message = "Primary IPs must survive the server (no auto-delete, delete protection on)."
  }

  assert {
    condition = toset(concat(
      [cloudflare_dns_record.apex_a.name, cloudflare_dns_record.apex_aaaa.name],
      [for r in cloudflare_dns_record.web : r.name]
      )) == toset([
      "leviosa.care", "www.leviosa.care", "admin.leviosa.care", "staff.leviosa.care",
      "staging.leviosa.care", "admin-staging.leviosa.care", "staff-staging.leviosa.care",
    ])
    error_message = "Web records must be exactly apex, www, admin, staff, staging, admin-staging and staff-staging (no analytics)."
  }

  assert {
    condition     = cloudflare_dns_record.apex_a.proxied && cloudflare_dns_record.apex_aaaa.proxied && alltrue([for r in cloudflare_dns_record.web : r.proxied])
    error_message = "Every web record must be proxied by Cloudflare."
  }

  assert {
    condition     = length(cloudflare_dns_record.mx) == 0 && length(cloudflare_dns_record.spf) == 0 && length(cloudflare_dns_record.dkim) == 0
    error_message = "No mail record exists until Zoho's values are filled in (issue 036)."
  }
}

# ---------------------------------------------------------------------------
# Without tf.sh there are no credentials, and plan stops
# ---------------------------------------------------------------------------

run "direct_terraform_without_tf_sh_fails" {
  command = plan

  # What a bare `terraform plan` sees: no TF_VAR_* from Infisical.
  variables {
    hcloud_token     = ""
    cloudflare_token = ""
  }

  expect_failures = [
    var.hcloud_token,
    var.cloudflare_token,
  ]
}
