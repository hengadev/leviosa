#!/bin/sh
# Check that docs/deployment/environment.md lists every key the Infisical
# Agent templates render:
#   - the keys a template writes itself (lines `KEY='...'` in
#     infra/infisical/templates/*.env.tmpl);
#   - every key of the folders the templates render whole, taken from the
#     template fixture infra/infisical/envcheck/testdata/infisical.json
#     (fake values; only the key names are read).
# A key counts as listed when it appears in backticks: `KEY`.
#
# Usage: docs/deployment/check-environment.sh   (from anywhere; no network)
set -eu

root=$(cd "$(dirname "$0")/../.." && pwd)
doc=$root/docs/deployment/environment.md
templates=$root/infra/infisical/templates
fixture=$root/infra/infisical/envcheck/testdata/infisical.json

[ -f "$doc" ] || { echo "check-environment: $doc is missing" >&2; exit 1; }

keys=$(
	{
		grep -hoE "^[A-Z][A-Z0-9_]*='" "$templates"/*.env.tmpl | sed "s/='\$//"
		grep -oE '^ +"[A-Z][A-Z0-9_]*":' "$fixture" | tr -d ' ":'
	} | sort -u
)

missing=0
for key in $keys; do
	if ! grep -qF "\`$key\`" "$doc"; then
		echo "missing from environment.md: $key"
		missing=$((missing + 1))
	fi
done

count=$(printf '%s\n' "$keys" | wc -l | tr -d ' ')
if [ "$missing" -gt 0 ]; then
	echo "check-environment: $missing of $count template keys are not listed" >&2
	exit 1
fi
echo "check-environment: all $count template keys are listed in docs/deployment/environment.md"
