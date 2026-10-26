# Private media bucket per environment (product images, logos). The backend
# uploads and deletes through its app IAM user (iam.tf); nothing is public.
# Media relies on versioning instead of backups: a replaced or deleted object
# stays recoverable for var.media_noncurrent_version_days.

resource "aws_s3_bucket" "media" {
  for_each = local.environments
  bucket   = local.media_bucket_names[each.key]

  tags = {
    Name        = "Leviosa media - ${each.key}"
    Environment = each.key
  }
}

resource "aws_s3_bucket_versioning" "media" {
  for_each = local.environments
  bucket   = aws_s3_bucket.media[each.key].id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "media" {
  for_each = local.environments
  bucket   = aws_s3_bucket.media[each.key].id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "media" {
  for_each = local.environments
  bucket   = aws_s3_bucket.media[each.key].id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "media" {
  for_each = local.environments
  bucket   = aws_s3_bucket.media[each.key].id

  rule {
    id     = "expire-noncurrent-versions"
    status = "Enabled"

    filter {}

    noncurrent_version_expiration {
      noncurrent_days = var.media_noncurrent_version_days
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }

  # Lifecycle rules apply to versions, so versioning must be on first.
  depends_on = [aws_s3_bucket_versioning.media]
}
