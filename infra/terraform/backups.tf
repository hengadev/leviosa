# Production backups (issue 028): the daily database dump, the configuration
# archive and the weekly encrypted Infisical snapshot land in this bucket.
# Staging is not backed up.
#
# The server's backup job runs with the backup IAM user below, whose key is
# written to Infisical prod/backup (infisical.tf). It can list, read and
# write, but not delete: a leaked key or a compromised server cannot erase
# the backups. Old backups go through the lifecycle rule only, and versioning
# keeps an overwritten object's previous version for
# var.backup_noncurrent_version_days.

resource "aws_s3_bucket" "backups" {
  bucket = local.backup_bucket_name

  tags = {
    Name        = "Leviosa backups (prod)"
    Environment = "prod"
  }
}

resource "aws_s3_bucket_versioning" "backups" {
  bucket = aws_s3_bucket.backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "backups" {
  bucket = aws_s3_bucket.backups.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "expire-backups"
    status = "Enabled"

    filter {}

    expiration {
      days = var.backup_retention_days
    }

    noncurrent_version_expiration {
      noncurrent_days = var.backup_noncurrent_version_days
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 1
    }
  }

  depends_on = [aws_s3_bucket_versioning.backups]
}

# ============================================
# Backup IAM user: write and read, never delete
# ============================================

locals {
  # No s3:DeleteObject*, no bucket configuration, no KMS at all (checked by
  # tests/policies.tftest.hcl). AbortMultipartUpload only discards the parts
  # of an upload that never completed.
  backup_policy = {
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "ListBackupBucket"
        Effect   = "Allow"
        Action   = ["s3:ListBucket"]
        Resource = [aws_s3_bucket.backups.arn]
      },
      {
        Sid      = "ReadWriteBackupObjects"
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:PutObject", "s3:AbortMultipartUpload"]
        Resource = ["${aws_s3_bucket.backups.arn}/*"]
      },
    ]
  }
}

resource "aws_iam_user" "backup" {
  name = local.backup_user_name

  tags = {
    Environment = "prod"
  }
}

resource "aws_iam_user_policy" "backup" {
  name   = "${local.backup_user_name}-write-no-delete"
  user   = aws_iam_user.backup.name
  policy = jsonencode(local.backup_policy)
}

resource "aws_iam_access_key" "backup" {
  user = aws_iam_user.backup.name
}
