#!/bin/sh
# Turn an Infisical Agent env template into the concrete copy for one
# Infisical environment of the `leviosa` project, on stdout.
#
# The Agent's template engine has no env function, so a template cannot
# compute its environment or project: each one carries two placeholders,
# substituted here, and every consumer renders from this same script —
# `make env` (dev), check-env.sh, and the Agent role installing the server's
# copies (issue 027).
#
#   __LEVIOSA_ENV__         the environment slug: dev, staging or prod
#   __LEVIOSA_PROJECT_ID__  the `leviosa` project's ID, read from
#                           infra/terraform/terraform.tfvars
#                           (infisical_project_id): the operator sets it
#                           once, there, for Terraform and the templates.
#                           LEVIOSA_INFISICAL_PROJECT_ID overrides it (the
#                           fixture test uses a fake ID).
#
# Usage: render-template.sh <template-file> <dev|staging|prod>
set -eu

die() {
	echo "render-template: $*" >&2
	exit 1
}

if [ "$#" -ne 2 ]; then
	echo "usage: $0 <template-file> <dev|staging|prod>" >&2
	exit 2
fi

template=$1
env_slug=$2

[ -f "$template" ] || die "template not found: $template"

case $env_slug in
dev | staging | prod) ;;
*) die "invalid environment: '$env_slug' (expected dev, staging or prod)" ;;
esac

script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
tfvars="$script_dir/../terraform/terraform.tfvars"

if [ -n "${LEVIOSA_INFISICAL_PROJECT_ID:-}" ]; then
	project_id=$LEVIOSA_INFISICAL_PROJECT_ID
	source_name="LEVIOSA_INFISICAL_PROJECT_ID"
else
	[ -f "$tfvars" ] || die "$tfvars not found (it holds infisical_project_id)"
	project_id=$(sed -n 's/^[[:space:]]*infisical_project_id[[:space:]]*=[[:space:]]*"\([^"]*\)".*$/\1/p' "$tfvars" | head -n 1)
	source_name="infisical_project_id in infra/terraform/terraform.tfvars"
fi

# A UUID, which also keeps the sed replacement below literal.
if ! printf '%s\n' "$project_id" | grep -Eqx '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'; then
	die "the Infisical project ID ($source_name) is not set to a project UUID yet.
  Set infisical_project_id in infra/terraform/terraform.tfvars to the ID of the
  leviosa project (Infisical > leviosa > Project settings > Project ID)."
fi

rendered=$(sed -e "s/__LEVIOSA_PROJECT_ID__/$project_id/g" -e "s/__LEVIOSA_ENV__/$env_slug/g" "$template")

case $rendered in
*__LEVIOSA_*) die "$template has a placeholder this script does not know" ;;
esac

printf '%s\n' "$rendered"
