#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# check-agent-key.sh <keys file>
#
# Run by the Makefile before anything connects to the server: refuses unless
# the SSH agent holds one of the keys file's operator keys
# (deploy_ssh_public_keys). The check is local only, it never connects.
#
# Why: Ansible connects to the server's IP, so ~/.ssh/config's IdentityFile
# for a host alias does not apply, and it cannot type a passphrase. Without
# the key in the agent every attempt fails, and fail2ban's SSH jails (3
# failures → 2 h, the DDoS jail 2 → 24 h; playbooks/group_vars/all.yml) ban
# this computer's IP after one or two runs.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

me=check-agent-key
die() {
	printf '%s: ERROR: %s\n' "$me" "$*" >&2
	exit 1
}

[ $# -eq 1 ] || die "usage: check-agent-key.sh <keys file>"
keys_file=$1
[ -f "$keys_file" ] || die "keys file $keys_file not found (see deploy-keys.example.yml and README.md)"

# The base64 field of each entry under deploy_ssh_public_keys, up to the next
# top-level key. Entries are one quoted line each: "<type> <base64> <comment>".
operator_keys=$(awk '
	/^[A-Za-z_]+:/ { in_list = ($1 == "deploy_ssh_public_keys:"); next }
	in_list && /^[[:space:]]*-/ { gsub(/["'\'']/, ""); print $3 }
' "$keys_file")
[ -n "$operator_keys" ] || die "no operator key under deploy_ssh_public_keys in $keys_file"

# ssh-add -L: 0 = keys listed, 1 = agent running with no key, 2 = no agent.
set +e
agent_keys=$(ssh-add -L 2>/dev/null)
status=$?
set -e
case $status in
0) ;;
1) die "the SSH agent holds no key. Run: ssh-add ~/.ssh/leviosa-\$(hostname)" ;;
*) die "no SSH agent in this terminal. Run: eval \"\$(ssh-agent -s)\" && ssh-add ~/.ssh/leviosa-\$(hostname)" ;;
esac

if ! awk '{ print $2 }' <<<"$agent_keys" | grep -qxF -f <(printf '%s\n' "$operator_keys"); then
	die "none of the agent's keys is an operator key in $keys_file. Load this computer's key: ssh-add ~/.ssh/leviosa-\$(hostname)"
fi
echo "$me: an operator key from $keys_file is in the SSH agent"
