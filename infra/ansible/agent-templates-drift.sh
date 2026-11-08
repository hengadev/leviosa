#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# agent-templates-drift.sh <server ip>
#
# Read-only: compares the Infisical Agent templates installed on the server
# (/etc/infisical-agent/templates) with the ones this checkout would install,
# rendered by render-template.sh exactly as roles/infisical_agent does.
#
# Why: the Agent renders the env files from its installed copies, never from
# the repository. A change under infra/infisical/templates (a key added to
# frontend.env's allowlist, say) reaches the server only through
# `make infisical-agent` or `make setup`; until then the server silently
# keeps rendering the old files.
#
# Prints file names and "same", "differs" or "missing" only, never values.
# Exits 1 when any template differs or is missing on the server.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

me=agent-templates-drift
die() {
	printf '%s: ERROR: %s\n' "$me" "$*" >&2
	exit 2
}

[ $# -eq 1 ] && [ -n "$1" ] || die "usage: agent-templates-drift.sh <server ip>"
ip=$1

here=$(cd "$(dirname "$0")" && pwd)
defaults=$here/roles/infisical_agent/defaults/main.yml
src_dir=$here/../infisical/templates
render=$here/../infisical/render-template.sh
remote_dir=/etc/infisical-agent/templates

# The role's template list: "  - { env: prod, name: backend, ... }" → "prod backend".
pairs=$(sed -n 's/^ *- { *env: *\([a-z]*\), *name: *\([a-z]*\),.*/\1 \2/p' "$defaults")
[ -n "$pairs" ] || die "no templates listed in $defaults (infisical_agent_templates)"

# One connection for every checksum of the server's copies.
remote=$(ssh -o BatchMode=yes "root@$ip" "cd $remote_dir && sha256sum -- *.env.tmpl" 2>/dev/null) ||
	die "could not read $remote_dir on root@$ip (is this computer's key in the SSH agent?)"

drift=0
while read -r env name; do
	file=$env-$name.env.tmpl
	# The role installs render-template.sh's output with its trailing newlines
	# stripped (Ansible's stdout) and one added back.
	rendered=$("$render" "$src_dir/$name.env.tmpl" "$env") ||
		die "could not render $name.env.tmpl for $env"
	local_sum=$(printf '%s\n' "$rendered" | sha256sum | cut -d' ' -f1)
	remote_sum=$(printf '%s\n' "$remote" | awk -v f="$file" '$2 == f { print $1 }')
	if [ -z "$remote_sum" ]; then
		printf '  %-28s missing on the server\n' "$file"
		drift=1
	elif [ "$remote_sum" != "$local_sum" ]; then
		printf '  %-28s differs\n' "$file"
		drift=1
	else
		printf '  %-28s same\n' "$file"
	fi
done <<<"$pairs"

echo ""
if [ "$drift" -eq 1 ]; then
	echo "The server's Agent templates are not this checkout's: run 'make infisical-agent'"
	echo "(from a checkout of main) to install them. The Agent then re-renders the env files"
	echo "and recreates the services whose file changed."
	exit 1
fi
echo "The server's Agent templates match this checkout."
