# The VPS that runs both Compose stacks (ADR-0016) and its primary IPs.

# An SSH key that already exists in the Hetzner project. It is only root's
# first key: Ansible replaces authorized_keys from the operators' keys file.
data "hcloud_ssh_key" "default" {
  name = var.hcloud_ssh_key_name
}

# Primary IPs, managed on their own so that replacing the server keeps its
# addresses: DNS and the operators' known_hosts entries stay valid across
# rebuilds. auto_delete = false and delete protection keep them when the
# server is destroyed.
resource "hcloud_primary_ip" "main" {
  for_each = toset(["ipv4", "ipv6"])

  name              = "${var.project_name}-${each.key}"
  type              = each.key
  location          = var.server_location
  auto_delete       = false
  delete_protection = true

  labels = {
    project    = var.project_name
    managed_by = "terraform"
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "hcloud_server" "main" {
  name        = var.project_name
  image       = var.server_image
  server_type = var.server_type
  location    = var.server_location
  ssh_keys    = [data.hcloud_ssh_key.default.id]
  backups     = var.enable_backups

  public_net {
    ipv4_enabled = true
    ipv4         = hcloud_primary_ip.main["ipv4"].id
    ipv6_enabled = true
    ipv6         = hcloud_primary_ip.main["ipv6"].id
  }

  labels = {
    project    = var.project_name
    managed_by = "terraform"
  }

  # Minimal: only makes the server reachable by Ansible, which does the rest.
  user_data = file("${path.module}/cloud-init.yml.tftpl")

  # The server holds nothing that Ansible + Infisical + the backups cannot
  # rebuild, so it may be replaced (`./tf.sh apply -replace=hcloud_server.main`);
  # the primary IPs above survive that. A change to user_data would replace it
  # too, so cloud-init stays minimal and is not edited casually.
  lifecycle {
    prevent_destroy = false
  }
}
