#!/bin/sh
# Read-only check of the Agent templates against real Infisical values
# (issue 024). Renders every env file one Infisical environment produces,
# with the same templates and render script as the server and `make env`,
# into a private temporary directory; lists each file's key names and any
# problem (missing keys, frontend exclusions, URL formats...) — never a
# value; then deletes the renderings.
#
# Usage: check-env.sh <dev|staging|prod>      (or: make env-check ENV=<env>)
#
# Requires the Infisical CLI, a session (`infisical login`, as a member of
# the leviosa project) and Go (the checker is ./envcheck). It only reads
# Infisical and writes nothing outside its temporary directory.
# Exit status: 0 no problem, 1 problems (listed), 2 usage.
set -eu

die() {
	echo "check-env: $*" >&2
	exit 1
}

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <dev|staging|prod>" >&2
	exit 2
fi
env_slug=$1
case $env_slug in
dev | staging | prod) ;;
*) die "invalid environment: '$env_slug' (expected dev, staging or prod)" ;;
esac

here=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)

command -v infisical >/dev/null 2>&1 ||
	die "the Infisical CLI is not installed: https://infisical.com/docs/cli/overview"
command -v go >/dev/null 2>&1 || die "Go is not installed (the checker is $here/envcheck)"

# The token itself is never printed: it only proves there is a session.
token=$(infisical user get token --plain --silent 2>/dev/null) || token=
[ -n "$token" ] || die "no Infisical session. Run: infisical login"
unset token

files=$(cd "$here/envcheck" && go run ./cmd/envcheck files "$env_slug")

umask 077
work=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-env-check.XXXXXX")
trap 'rm -rf "$work"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

for name in $files; do
	"$here/render-template.sh" "$here/templates/$name.tmpl" "$env_slug" >"$work/$name.tmpl"
	if infisical export --template="$work/$name.tmpl" --silent >"$work/$name.out"; then
		mv "$work/$name.out" "$work/$name"
	else
		# Reported as "not rendered" by the checker below.
		rm -f "$work/$name.out"
		echo "check-env: rendering $name from Infisical $env_slug failed (the CLI's message is above)" >&2
	fi
done

cd "$here/envcheck"
go run ./cmd/envcheck check "$env_slug" "$work"
