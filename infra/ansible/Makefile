# Leviosa Ansible Makefile
# Usage: make <target>
#
# This Makefile sets the server up; it never releases. Releases are
# `deploy <stack> <sha>` through the deploy gate (roles/deploy_gate,
# infra/compose/README.md), run by CI.

.PHONY: help setup infisical-agent check-clean lint \
        ssh-info status staging-status agent-status logs staging-logs \
        install-deps terraform-outputs

# Default target
help:
	@echo "Leviosa Ansible Commands"
	@echo "============================"
	@echo ""
	@echo "Setup (needs only this computer's SSH key, listed in the keys file, and 'infisical login'):"
	@echo "  make setup           - Complete VPS setup (fresh server, and after any change here)"
	@echo "                         KEYS_FILE=<path> overrides ~/leviosa-deploy-keys.yml"
	@echo "                         With INFISICAL_VPS_CLIENT_ID/SECRET exported, the Infisical"
	@echo "                         Agent is installed too; otherwise that role is skipped"
	@echo ""
	@echo "  make infisical-agent - Install or rotate the Infisical Agent alone (issue 027):"
	@echo "                         the leviosa-vps credential from INFISICAL_VPS_CLIENT_ID /"
	@echo "                         INFISICAL_VPS_CLIENT_SECRET, or the playbook's prompts"
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
	@echo "  make agent-status    - Check the Infisical Agent (systemd)"
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
# twice: the second run must report changed=0. Both setup targets first check,
# locally, that this computer's operator key is in the SSH agent
# (check-agent-key.sh): failed connections get this computer banned by fail2ban.
setup:
	@echo "==========================================="
	@echo "Leviosa VPS Setup"
	@echo "==========================================="
	@echo ""
	@./check-agent-key.sh "$(KEYS_FILE)"
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
# Infisical Agent (issue 027)
# ============================================

# Install or rotate the Infisical Agent alone: renders the env files in both
# stack roots from issue 024's templates. The leviosa-vps credential is read
# at run time only (ADR-0017): export INFISICAL_VPS_CLIENT_ID /
# INFISICAL_VPS_CLIENT_SECRET and press Enter at both prompts, or answer them
# (the secret is not echoed). Re-running this with a NEW client secret is the
# rotation procedure (every three months; roles/infisical_agent/README.md):
# no app container is restarted, and the old secret is revoked afterwards.
infisical-agent:
	@echo "==========================================="
	@echo "Leviosa Infisical Agent - install or rotate"
	@echo "==========================================="
	@echo ""
	@./check-agent-key.sh "$(KEYS_FILE)"
	@vps_ip="$(VPS_IP)"; \
	if [ -z "$$vps_ip" ]; then \
		echo "Error: could not read the VPS IP from Terraform."; \
		echo "Run 'infisical login', then 'cd ../terraform && ./tf.sh output server_ipv4_address'."; \
		exit 1; \
	fi; \
	echo "VPS IP: $$vps_ip"; \
	if [ -n "$$INFISICAL_VPS_CLIENT_ID" ] && [ -n "$$INFISICAL_VPS_CLIENT_SECRET" ]; then \
		echo "Credential: exported (press Enter at both prompts)"; \
	else \
		echo "Credential: not exported (answer the prompts; the secret is not echoed)"; \
	fi; \
	echo ""; \
	printf 'Continue? [y/N] '; read confirm; [ "$$confirm" = "y" ] || exit 1; \
	echo ""; \
	ansible-playbook -i inventory/hosts.yml playbooks/infisical-agent.yml \
		-e "ansible_host=$$vps_ip"

# ============================================
# Verification
# ============================================

check-clean:
	@echo "Checking playbooks for syntax errors..."
	@ansible-playbook playbooks/site.yml --syntax-check
	@ansible-playbook playbooks/infisical-agent.yml --syntax-check
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

agent-status:
	@echo "Checking the Infisical Agent..."
	@ssh root@$(VPS_IP) "systemctl status infisical-agent --no-pager; ls -la /opt/leviosa/env /opt/leviosa-staging/env"

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
