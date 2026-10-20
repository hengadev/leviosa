#!/bin/sh
# `make env` — render the local development env files from Infisical `dev`,
# with the SAME templates the server's Infisical Agent renders (issue 024):
#
#   backend/development.env   backend.env + admin.env (the admin bootstrap
#                             runs locally the way it runs on the server)
#   frontend/.env             frontend.env
#
# Uses your own Infisical login (`infisical login`); never a machine
# identity. Both files are OVERWRITTEN — local edits are lost; values you
# need locally (a Stripe CLI webhook secret, say) belong in Infisical `dev`.
# Nothing is written unless both renderings succeed.
#
# Without Infisical, `make local-dev` copies the committed example files
# instead (integrations off).
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)
here="$repo_root/infra/infisical"
env_slug=dev

die() {
	echo "make env: $*" >&2
	exit 1
}

command -v infisical >/dev/null 2>&1 || die "the Infisical CLI is not installed.
  Install it first: https://infisical.com/docs/cli/overview
  Without Infisical, \`make local-dev\` uses the committed example files."

# The token itself is never printed: it only proves there is a session.
token=$(infisical user get token --plain --silent 2>/dev/null) || token=
[ -n "$token" ] || die "you are not logged in to Infisical. Run: infisical login
  (to https://secrets.henga.dev, as a member of the leviosa project)"
unset token

umask 077
work=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-make-env.XXXXXX")
trap 'rm -rf "$work"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

render() { # <file> -> $work/<file>
	"$here/render-template.sh" "$here/templates/$1.tmpl" "$env_slug" >"$work/$1.tmpl" &&
		infisical export --template="$work/$1.tmpl" --silent >"$work/$1"
}

for name in backend.env admin.env frontend.env; do
	render "$name" || die "rendering $name from Infisical '$env_slug' failed (the CLI's message is above).
  Nothing was written. If your session expired, run: infisical login"
done

# The CLI fails on auth and template errors; a rendering without its
# backbone means the dev folders are incomplete — refuse to overwrite.
for key in DATABASE_URL ENCX_PEPPER ENCX_KMS_KEY_ALIAS AWS_REGION; do
	grep -q "^$key='..*'\$" "$work/backend.env" ||
		die "the rendered backend.env has no $key (is Infisical dev complete? see infra/infisical/README.md). Nothing was written."
done
grep -q "^ADMIN_EMAIL='..*'\$" "$work/admin.env" ||
	die "the rendered admin.env has no ADMIN_EMAIL (Infisical dev/admin). Nothing was written."
grep -q "^API_URL='..*'\$" "$work/frontend.env" ||
	die "the rendered frontend.env has no API_URL. Nothing was written."

cat "$work/backend.env" "$work/admin.env" >"$work/development.env"
mv -f "$work/development.env" "$repo_root/backend/development.env"
mv -f "$work/frontend.env" "$repo_root/frontend/.env"

echo "make env: OVERWROTE backend/development.env and frontend/.env with a fresh rendering of Infisical '$env_slug'"
echo "  (previous contents replaced; both files are mode 0600 and gitignored)."
echo "  Next: make local-dev"
