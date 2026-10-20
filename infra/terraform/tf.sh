#!/bin/sh
# Run Terraform with its credentials from Infisical. The only supported way
# to run this configuration:
#
#   ./tf.sh init | plan | apply | <any terraform command> [args]
#   (or `make init|plan|apply` here, `make infra-*` at the repository root)
#
# - `infisical run` injects the `leviosa-infra` project's secrets: the
#   TF_VAR_* tokens and the AWS credentials used by the AWS provider and the
#   S3 state backend.
# - The Infisical provider (infisical.tf) authenticates with the operator's
#   own login: the session token from `infisical user get token` is handed to
#   it as INFISICAL_TOKEN. No machine identity, no credential on disk.
#
# Requires the Infisical CLI and `infisical login` to secrets.henga.dev, as a
# member of both `leviosa-infra` and `leviosa`.
set -eu

# Where Terraform's own credentials live. Not secret.
# OPERATOR: set infra_project_id to the `leviosa-infra` project's ID
# (Infisical > leviosa-infra > Project settings) before the first run.
infra_project_id=REPLACE_WITH_LEVIOSA_INFRA_PROJECT_ID
infra_env=prod
infra_path=/

die() {
	echo "tf.sh: $*" >&2
	exit 1
}

case $infra_project_id in
REPLACE_* | '') die "set infra_project_id in $0 to the ID of the Infisical leviosa-infra project" ;;
esac

command -v infisical >/dev/null 2>&1 ||
	die "the Infisical CLI is not installed: https://infisical.com/docs/cli/overview"
command -v terraform >/dev/null 2>&1 || die "terraform is not installed"

token=$(infisical user get token --plain --silent 2>/dev/null) || token=
[ -n "$token" ] || die "no Infisical session. Run: infisical login"

cd "$(dirname -- "$0")"

# Passed under a neutral name so `infisical run` itself never treats it as
# its own INFISICAL_TOKEN; renamed for Terraform inside the child. Keeping it
# in the environment (not argv) keeps it out of `ps`.
LEVIOSA_TF_INFISICAL_TOKEN=$token
export LEVIOSA_TF_INFISICAL_TOKEN
unset token

# The leviosa-infra credentials must be the only AWS credentials Terraform
# can use, on every computer: drop any profile or session from the shell.
unset AWS_PROFILE AWS_DEFAULT_PROFILE AWS_SESSION_TOKEN

exec infisical run --projectId="$infra_project_id" --env="$infra_env" --path="$infra_path" --silent -- \
	sh -c 'INFISICAL_TOKEN=$LEVIOSA_TF_INFISICAL_TOKEN
INFISICAL_AUTH_METHOD=token
export INFISICAL_TOKEN INFISICAL_AUTH_METHOD
unset LEVIOSA_TF_INFISICAL_TOKEN
exec terraform "$@"' terraform "$@"
