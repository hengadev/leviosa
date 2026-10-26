# The S3 bucket that holds this configuration's own state (main.tf).
#
# It was created by the old workspace-based config and survived its teardown
# (issue 007 removed it from the old state without deleting it). The import
# blocks below adopt it into this state on the first apply; once it is in the
# state they are no-ops. The old config never set a public access block on
# it, so that one is created, not imported.

import {
  to = aws_s3_bucket.terraform_state
  id = "leviosa-terraform-state"
}

import {
  to = aws_s3_bucket_versioning.terraform_state
  id = "leviosa-terraform-state"
}

import {
  to = aws_s3_bucket_server_side_encryption_configuration.terraform_state
  id = "leviosa-terraform-state"
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = "leviosa-terraform-state"

  tags = {
    Name = "Leviosa Terraform State"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
