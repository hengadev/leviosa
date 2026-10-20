# Leviosa Terraform
#
# Every command that touches real infrastructure runs through ./tf.sh, which
# wraps it in `infisical run` (credentials from the `leviosa-infra` project)
# and hands the Infisical provider your own login. Run `infisical login`
# first. `check` and `test` are offline and need no credentials.

TF := ./tf.sh

.PHONY: help init plan apply output show state-list server-info server-ssh \
        check fmt-check validate test fmt

help:
	@echo "Leviosa Terraform (needs: infisical login)"
	@echo "=========================================="
	@echo ""
	@echo "  make init         - Initialize (S3 state backend, providers)"
	@echo "  make plan         - Show planned changes"
	@echo "  make apply        - Apply (also writes /aws and prod/backup into Infisical)"
	@echo "  make output       - Show outputs"
	@echo "  make show         - Show the state"
	@echo "  make state-list   - List resources in the state"
	@echo "  make server-info  - Server name and addresses"
	@echo "  make server-ssh   - SSH into the server as root"
	@echo ""
	@echo "Offline (no credentials):"
	@echo "  make check        - fmt -check, validate, tests, no terraform.workspace"
	@echo "  make test         - Only the tests (tests/*.tftest.hcl, mocked providers)"
	@echo "  make fmt          - Format the configuration"
	@echo ""
	@echo "Any other command: ./tf.sh <command> [args]"
	@echo "  e.g. ./tf.sh apply -replace=hcloud_server.main"

init:
	$(TF) init

plan:
	$(TF) plan

apply:
	$(TF) apply

output:
	$(TF) output

show:
	$(TF) show

state-list:
	$(TF) state list

server-info:
	@$(TF) output -json | jq -r '"Server: \(.server_name.value)\nIPv4:   \(.server_ipv4_address.value)\nIPv6:   \(.server_ipv6_address.value)"'

server-ssh:
	@ssh root@$$($(TF) output -raw server_ipv4_address)

# ============================================
# Offline checks: no credentials, no backend, no API call
# ============================================

check: fmt-check test
	@! grep -rn 'terraform\.workspace' --include='*.tf' --include='*.tfvars' --include='*.hcl' --include='*.sh' --exclude-dir=.terraform . || { echo "terraform.workspace must not be used: one state, default workspace"; exit 1; }
	@echo "check: OK"

fmt-check:
	terraform fmt -check -recursive

# Installs the providers pinned in .terraform.lock.hcl without touching the
# state backend, then validates.
validate:
	terraform init -backend=false -input=false >/dev/null
	terraform validate

# tests/offline.tftest.hcl, every provider mocked: the IAM policies (own
# bucket and key only, no key deletion, backup user cannot delete), the
# Infisical keys and paths, the DNS records, and that plan fails without the
# credentials tf.sh provides.
test: validate
	terraform test

fmt:
	terraform fmt -recursive
