# Outputs. Credentials are not outputs: Terraform writes them into Infisical
# (infisical.tf).

output "server_name" {
  value       = hcloud_server.main.name
  description = "Hetzner server name"
}

output "server_ipv4_address" {
  value       = hcloud_primary_ip.main["ipv4"].ip_address
  description = "Server IPv4 (a primary IP: kept when the server is replaced)"
}

output "server_ipv6_address" {
  value       = cidrhost(hcloud_primary_ip.main["ipv6"].ip_network, 1)
  description = "Server IPv6 (in a primary IP network: kept when the server is replaced)"
}

output "ssh_connection_string" {
  value       = "ssh root@${hcloud_primary_ip.main["ipv4"].ip_address}"
  description = "SSH command for the first connection, before the `Host leviosa` alias exists"
}

output "domain_name" {
  value       = var.domain_name
  description = "Apex domain"
}

output "web_hostnames" {
  value       = concat([var.domain_name], sort([for sub in local.web_subdomains : "${sub}.${var.domain_name}"]))
  description = "Proxied hostnames pointing at the server"
}

output "aws_region" {
  value       = var.aws_region
  description = "Region of the buckets and KMS keys"
}

output "media_bucket_names" {
  value       = { for env, bucket in aws_s3_bucket.media : env => bucket.bucket }
  description = "Media bucket per environment"
}

output "backup_bucket_name" {
  value       = aws_s3_bucket.backups.bucket
  description = "Production backup bucket"
}

output "kms_key_arns" {
  value       = { for env, key in aws_kms_key.encx : env => key.arn }
  description = "encx KEK ARN per environment"
}

output "kms_alias_names" {
  value       = { for env, alias in aws_kms_alias.encx : env => alias.name }
  description = "encx KEK alias per environment"
}

output "app_user_names" {
  value       = { for env, user in aws_iam_user.app : env => user.name }
  description = "App IAM user per environment"
}

output "backup_user_name" {
  value       = aws_iam_user.backup.name
  description = "Backup IAM user (write and read, no delete)"
}

output "terraform_state_bucket_name" {
  value       = aws_s3_bucket.terraform_state.bucket
  description = "Bucket holding this configuration's state"
}
