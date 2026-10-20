# Leviosa Ansible Makefile
# Usage: make <target>
#
# This Makefile sets the server up; it never releases. Releases are
# `deploy <stack> <sha>` through the deploy gate (roles/deploy_gate,
# infra/compose/README.md), run by CI.

.PHONY: help setup check-clean lint \
        ssh-info status staging-status logs staging-logs \
        install-deps terraform-outputs

# Default target
help:
	@echo "Leviosa Ansible Commands"
	@echo "============================"
	@echo ""
	@echo "Setup (needs only this computer's SSH key, listed in the keys file, and 'infisical login'):"
	@echo "  make setup           - Complete VPS setup (fresh server, and after any change here)"
	@echo "                         KEYS_FILE=<path> overrides ~/leviosa-deploy-keys.yml"
	@echo ""
	@echo "Releases: NOT here. CI runs 'deploy <stack> <sha>' through the deploy gate."
	@echo ""
	@echo "Verification:"
	@echo "  make check-clean     - Check playbooks for syntax errors"
	@echo "  make lint            - Run ansible-lint"
	@echo ""
	@echo "Server Operations (read-only, as root):"
	@echo "  make ssh-info        - Display SSH connection details"
	@echo "  make status          - Check production service status"
	@echo "  make staging-status  - Check staging service status"
	@echo "  make logs            - View production logs"
	@echo "  make staging-logs    - View staging logs"
	@echo ""
	@echo "Dependencies:"
	@echo "  make install-deps    - Install Ansible and the required collections"
	@echo "  make terraform-outputs - Show Terraform outputs (through tf.sh)"

# VPS IP and domain from Terraform, through ../terraform/tf.sh (plain
# `terraform` has no credentials). Recursive variables: evaluated only by the
# targets that use them.
VPS_IP = $(shell ../terraform/tf.sh output -raw server_ipv4_address 2>/dev/null || echo "")
DOMAIN = $(shell ../terraform/tf.sh output -raw domain_name 2>/dev/null || echo "")

# Operator keys (one per computer) and the CI key, kept outside the repo:
#   deploy_ssh_public_keys: [...]      deploy_ci_ssh_public_keys: [...]
# See deploy-keys.example.yml.
KEYS_FILE ?= $(HOME)/leviosa-deploy-keys.yml

# ============================================
# Setup
# ============================================

# Complete VPS setup (fresh server, and every later setup change). Run it
# twice: the second run must report changed=0.
setup:
	@echo "==========================================="
	@echo "Leviosa VPS Setup"
	@echo "==========================================="
	@echo ""
	@test -f "$(KEYS_FILE)" || { echo "Error: keys file $(KEYS_FILE) not found (see deploy-keys.example.yml and README.md)."; exit 1; }
	@vps_ip="$(VPS_IP)"; domain="$(DOMAIN)"; \
	if [ -z "$$vps_ip" ]; then \
		echo "Error: could not read the VPS IP from Terraform."; \
		echo "Run 'infisical login', then 'cd ../terraform && ./tf.sh output server_ipv4_address'."; \
		exit 1; \
	fi; \
	echo "VPS IP:    $$vps_ip"; \
	echo "Domain:    $${domain:-(default)}"; \
	echo "Keys file: $(KEYS_FILE)"; \
	echo ""; \
	printf 'Continue? [y/N] '; read confirm; [ "$$confirm" = "y" ] || exit 1; \
	echo ""; \
	ansible-playbook -i inventory/hosts.yml playbooks/site.yml \
		-e "ansible_host=$$vps_ip" \
		-e "@$(KEYS_FILE)" \
		$${domain:+-e "app_domain=$$domain"}

# ============================================
# Verification
# ============================================

check-clean:
	@echo "Checking playbooks for syntax errors..."
	@ansible-playbook playbooks/site.yml --syntax-check
	@echo "All playbooks are clean!"

lint:
	@echo "Running ansible-lint..."
	@command -v ansible-lint > /dev/null || { echo "Install with: pipx install ansible-lint"; exit 1; }
	@ansible-lint playbooks/

# ============================================
# Server Operations (read-only, as root: the deploy user has no docker)
# ============================================

ssh-info:
	@echo "==========================================="
	@echo "SSH Connection Details"
	@echo "==========================================="
	@echo ""
	@echo "VPS IP: $(VPS_IP)"
	@echo "Domain: $(DOMAIN)"
	@echo ""
	@echo "Connect as root (this computer's own key, from the keys file):"
	@echo "  ssh root@$(VPS_IP)"
	@echo ""
	@echo "The deploy user can only run the deploy gate; use root for docker and logs."
	@echo "==========================================="

status:
	@echo "Checking production service status..."
	@ssh root@$(VPS_IP) "cd /opt/leviosa && docker compose ps"

staging-status:
	@echo "Checking staging service status..."
	@ssh root@$(VPS_IP) "cd /opt/leviosa-staging && docker compose ps"

logs:
	@echo "Streaming production logs (Ctrl+C to exit)..."
	@ssh root@$(VPS_IP) "cd /opt/leviosa && docker compose logs -f"

staging-logs:
	@echo "Streaming staging logs (Ctrl+C to exit)..."
	@ssh root@$(VPS_IP) "cd /opt/leviosa-staging && docker compose logs -f"

# ============================================
# Dependencies
# ============================================

install-deps:
	@echo "Installing Ansible dependencies..."
	@command -v ansible-playbook > /dev/null || pipx install --include-deps ansible
	@ansible-galaxy collection install -r requirements.yml
	@echo "Done! Run 'make check-clean' to verify."

# ============================================
# Terraform Integration
# ============================================

terraform-outputs:
	@../terraform/tf.sh output
