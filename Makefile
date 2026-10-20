# Leviosa Makefile
# Usage: make <target>
#
# Local development, the local deploy harnesses, server setup (Ansible) and
# infrastructure (Terraform). Releases never go through this Makefile: CI
# builds the images and runs `deploy <stack> <sha>` through the deploy gate
# (infra/compose/README.md, infra/ansible/roles/deploy_gate/README.md).

.PHONY: help \
		local-dev local-start local-stop local-restart stripe-listen \
		setup \
		infra-init infra-plan infra-apply infra-output infra-destroy \
		ansible-logs ansible-logs-staging ansible-status ansible-status-staging

# Default target
help:
	@echo "Leviosa Commands"
	@echo "================"
	@echo ""
	@echo "Local Development:"
	@echo "  make local-dev        - Start frontend (native) and backend deps (Docker), hot reload"
	@echo "  make local-start      - Start the whole app in Docker, mirroring the production stack"
	@echo "                           (prod Dockerfiles, migrate+admin one-offs, LocalStack S3/KMS; no hot reload)"
	@echo "  make local-stop       - Stop the local Docker stack (ephemeral: wipes its data)"
	@echo "  make local-restart    - Restart the local Docker stack"
	@echo "  make stripe-listen    - Forward Stripe test-mode webhooks to the local backend (run beside local-dev)"
	@echo "  make env              - Render backend/development.env and frontend/.env from Infisical dev (overwrites both; infisical login first)"
	@echo ""
	@echo "Infisical Agent env templates (infra/infisical):"
	@echo "  make agent-templates-test - Render every template against fake folders and check each file (no Infisical needed)"
	@echo "  make env-check-<dev|staging|prod> - Read-only: render that environment's files from Infisical, list key names (never values)"
	@echo ""
	@echo "Local deploy harnesses (Docker / shell only, nothing remote):"
	@echo "  make deploy-harness   - Drive infra/compose/deploy against the real compose file"
	@echo "  make deploy-gate-test - Test the deploy gate, root runner, sudoers rule and CI keys"
	@echo "  make promote-preflight-test - Test Promote's preflight (which SHA ships, its checks)"
	@echo ""
	@echo "Server setup (Ansible; needs this computer's key in the keys file and 'infisical login'):"
	@echo "  make setup            - Set the VPS up (run twice: the second run must report changed=0)"
	@echo "  make ansible-status   - Production containers (docker compose ps, as root)"
	@echo "  make ansible-status-staging - Staging containers"
	@echo "  make ansible-logs     - Follow production logs"
	@echo "  make ansible-logs-staging - Follow staging logs"
	@echo ""
	@echo "Infrastructure (Terraform through infra/terraform/tf.sh; needs 'infisical login'):"
	@echo "  make infra-init       - Initialize Terraform"
	@echo "  make infra-plan       - Show Terraform plan"
	@echo "  make infra-apply      - Apply Terraform changes"
	@echo "  make infra-output     - Show Terraform outputs"
	@echo "  make infra-destroy    - Destroy all infrastructure"
	@echo ""
	@echo "Releases are not make targets: a push to main deploys staging through CI,"
	@echo "and production is promoted from staging (docs/deployment/promote.md)."

# ===========================================
# Local Development
# ===========================================

local-dev:
	@if [ ! -f backend/development.env ]; then \
		echo "No backend/development.env found — copying backend/example.env (edit it to add real Stripe/Twilio/SMTP creds if needed)."; \
		cp backend/example.env backend/development.env; \
	fi
	@if [ ! -f frontend/.env ]; then \
		echo "No frontend/.env found — copying frontend/.env.example (defaults point at the local backend; set USE_MOCK_DATA=true to demo the UI without it)."; \
		cp frontend/.env.example frontend/.env; \
	fi
	@echo "Starting frontend and backend in dev mode..."
	@cd frontend && pnpm run dev &
	@cd backend && docker compose up --build

# Forward Stripe test-mode webhooks to the local backend (run alongside local-dev).
# Prints a whsec_... signing secret on startup — put it in backend/development.env
# as STRIPE_WEBHOOK_SECRET (booking/payment events) and STRIPE_CONNECT_WEBHOOK_SECRET
# (Connect account events), then restart the backend container to pick it up.
stripe-listen:
	stripe listen \
		--forward-to localhost:3500/webhooks/stripe \
		--forward-connect-to localhost:3500/webhooks/stripe/connect

local-start:
	@if [ ! -f .env ]; then \
		echo "No .env found — copying .env.example to .env (edit it to add real Stripe/Twilio/SMTP creds if needed)."; \
		cp .env.example .env; \
	fi
	@echo "Starting local docker compose..."
	docker compose -f local.compose.yaml up --build

# Down with -v: the stack is ephemeral by design (no named volumes — see the header
# of local.compose.yaml), and -v also clears the throwaway anonymous volumes
# (Postgres data, LocalStack state) so nothing lingers after a stop.
local-stop:
	@echo "Stopping local docker compose..."
	docker compose -f local.compose.yaml down -v

local-restart: local-stop local-start

# Render the local env files from Infisical dev with the server's Agent
# templates, using your own login: backend/development.env (backend.env +
# admin.env) and frontend/.env. OVERWRITES both. See infra/infisical/README.md.
.PHONY: env env-check-dev env-check-staging env-check-prod agent-templates-test
env:
	infra/infisical/make-env.sh

# Read-only check of one Infisical environment (env-check-dev, -staging,
# -prod): renders its env files into a private temp dir, lists each file's key
# names and any problem, never a value, then deletes them. Needs infisical
# login and Go.
env-check-dev env-check-staging env-check-prod: env-check-%:
	infra/infisical/check-env.sh $*

# The Agent templates' fixture test: renders every template against fake
# Infisical folders (no Infisical, no network) and asserts each file's key set,
# frontend.env's exclusions and the URL formats.
agent-templates-test:
	cd infra/infisical/envcheck && go test -count=1 ./...

# Local Docker harness for infra/compose (prod and staging compose files + deploy script):
# builds both images with fake SHAs and drives the real deploy script; isolated
# names, no fixed host ports, full teardown. See infra/compose/README.md.
.PHONY: deploy-harness
deploy-harness:
	infra/compose/test/run-harness.sh

# Deploy gate and root runner tests (roles/deploy_gate): no root, no Ansible, no
# Docker, no network. See infra/ansible/roles/deploy_gate/README.md.
.PHONY: deploy-gate-test
deploy-gate-test:
	infra/ansible/roles/deploy_gate/tests/run-tests.sh

# Promote's preflight (.github/workflows/promote.yaml): which SHA ships and the
# checks before the deploy, against a fake staging /healthz and a fake Docker
# Hub. No Docker, no network beyond loopback. See docs/deployment/promote.md.
.PHONY: promote-preflight-test
promote-preflight-test:
	.github/scripts/tests/promote-preflight-tests.sh

# ===========================================
# Infrastructure Commands (Terraform)
# ===========================================
# Every Terraform command runs through infra/terraform/tf.sh, which loads the
# leviosa-infra credentials from your Infisical login (infisical login first).

infra-init:
	@echo "Initializing Terraform..."
	infra/terraform/tf.sh init

infra-plan:
	@echo "Showing Terraform plan..."
	infra/terraform/tf.sh plan

infra-apply:
	@echo "Applying Terraform changes..."
	infra/terraform/tf.sh apply

infra-output:
	@echo "Terraform outputs:"
	infra/terraform/tf.sh output

infra-destroy:
	@echo "Destroying all Terraform resources..."
	infra/terraform/tf.sh destroy

# ===========================================
# Server Setup (Ansible)
# ===========================================
# Ansible sets the server up and never releases (infra/ansible/README.md).
# KEYS_FILE=<absolute path> overrides the keys file (~/leviosa-deploy-keys.yml).

setup:
	cd infra/ansible && $(MAKE) setup

ansible-logs:
	@echo "Streaming production logs..."
	cd infra/ansible && $(MAKE) logs

ansible-logs-staging:
	@echo "Streaming staging logs..."
	cd infra/ansible && $(MAKE) staging-logs

ansible-status:
	@echo "Checking production status..."
	cd infra/ansible && $(MAKE) status

ansible-status-staging:
	@echo "Checking staging status..."
	cd infra/ansible && $(MAKE) staging-status
