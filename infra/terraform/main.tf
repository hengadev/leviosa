# Leviosa Terraform configuration: backend and providers.
#
# One state, default workspace only. Run Terraform through ./tf.sh (or the
# Makefile, which calls it): it injects the `leviosa-infra` credentials from
# the operator's Infisical login. Running `terraform` directly has no
# credentials and fails (see README.md).

terraform {
  # The state bucket already exists (it outlived the old workspaces) and is
  # declared in backend.tf. S3-native locking, no DynamoDB table.
  backend "s3" {
    bucket       = "leviosa-terraform-state"
    key          = "terraform.tfstate"
    region       = "eu-central-1"
    encrypt      = true
    use_lockfile = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.67"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.0"
    }
    infisical = {
      source  = "Infisical/infisical"
      version = "~> 0.19"
    }
  }

  # 1.11: mock providers with override_during, used by the offline tests in
  # tests/ (`make test`).
  required_version = ">= 1.11"
}

# Credentials: AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY from `leviosa-infra`,
# injected by tf.sh. The same credentials reach the S3 state backend.
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project   = var.project_name
      ManagedBy = "Terraform"
    }
  }
}

provider "hcloud" {
  token = var.hcloud_token
}

provider "cloudflare" {
  api_token = var.cloudflare_token
}

# Writes the values Terraform creates into the `leviosa` project
# (infisical.tf). Authenticated with the operator's own session token, which
# tf.sh passes as INFISICAL_TOKEN with INFISICAL_AUTH_METHOD=token: no machine
# identity, so no extra long-lived credential.
provider "infisical" {
  host = var.infisical_host
}
