# shellcheck shell=bash
# shellcheck disable=SC2154,SC2034
# ^ the backup_* settings come from backup.conf (sourced below, rendered by
#   Ansible), and the variables backup_init sets are the scripts' to use.
# Managed by Ansible (roles/backup, issue 028) — DO NOT EDIT ON THE SERVER
#
# Shared by leviosa-backup, leviosa-snapshot and leviosa-restore-test.
# Sourced, never executed. Everything runs as root.
#
# Configuration: /etc/leviosa/backup.conf (Ansible, paths and names only, no
# secret), or the file LEVIOSA_BACKUP_CONF names (the role's tests point it
# at a local stack; only root can run these scripts on the server, so only
# root can redirect them).
#
# Secrets: the Agent-rendered env files are PARSED, never sourced (a value is
# data, even if it contains `$(…)`), and only the names a script asks for are
# read. A value reaches a container by NAME (`-e VAR` with VAR exported in a
# subshell), never on a command line, and is never printed.

set -euo pipefail
umask 077
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

backup_name=$(basename "$0")

log() { printf '%s: %s\n' "$backup_name" "$*"; }
die() {
	printf '%s: ERROR: %s\n' "$backup_name" "$*" >&2
	exit 1
}

# shellcheck source=/dev/null
. "${LEVIOSA_BACKUP_CONF:-/etc/leviosa/backup.conf}"

# env_get <file> <KEY>: print KEY's value from an Agent-rendered env file
# (KEY='value', issue 024: every value single-quoted, no quote or newline
# inside). An unquoted value is read as is. Prints nothing when KEY is absent.
env_get() {
	local line value=
	while IFS= read -r line || [ -n "$line" ]; do
		case $line in "$2="*) value=${line#*=} ;; esac
	done <"$1"
	unquote "$value"
}

# unquote <raw value>: drop one pair of surrounding single quotes.
unquote() {
	local value=$1
	if [ "${#value}" -ge 2 ] && [ "${value:0:1}" = "'" ] && [ "${value: -1}" = "'" ]; then
		value=${value:1:${#value}-2}
	fi
	printf '%s' "$value"
}

# backup_init: root check, one run per script at a time, the backup.env
# values this job reads as shell variables, the bucket and the run's UTC
# timestamp and date prefix.
backup_init() {
	if [ "${backup_require_root:-1}" = 1 ] && [ "$(id -u)" != 0 ]; then
		die "must run as root"
	fi
	install -d -m 0700 "$backup_work_dir"
	exec 9>"$backup_work_dir/.$backup_name.lock"
	flock -n 9 || die "already running"
	[ -r "$backup_env_file" ] || die "$backup_env_file is missing (rendered by the Infisical Agent)"

	local key
	for key in BACKUP_AWS_ACCESS_KEY_ID BACKUP_AWS_SECRET_ACCESS_KEY BACKUP_S3_BUCKET BACKUP_AWS_REGION POSTGRES_USER POSTGRES_DB; do
		printf -v "$key" '%s' "$(env_get "$backup_env_file" "$key")"
		[ -n "${!key}" ] || die "$key is empty in $backup_env_file (Infisical prod /backup or /db)"
	done
	bucket=s3://$BACKUP_S3_BUCKET
	timestamp=$(date -u +%Y%m%dT%H%M%SZ)
	date_prefix=$(date -u -d "${timestamp:0:8}" +%Y/%m/%d)
}

# aws <args…>: the AWS CLI from its pinned image. The credentials reach the
# container by name (-e VAR), so they never appear in a command line. No host
# directory is mounted: uploads read stdin (`aws --stdin s3 cp - <url>`),
# downloads write stdout (`s3 cp <url> -`). Only --stdin attaches stdin.
aws() {
	local opts=(--rm --log-driver none -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION)
	if [ "${1:-}" = --stdin ]; then
		opts+=(-i)
		shift
	fi
	# Tests only (LocalStack): the endpoint and the network that reaches it.
	[ -n "${backup_aws_endpoint_url:-}" ] && opts+=(-e "AWS_ENDPOINT_URL=$backup_aws_endpoint_url")
	[ -n "${backup_docker_network:-}" ] && opts+=(--network "$backup_docker_network")
	(
		export AWS_ACCESS_KEY_ID=$BACKUP_AWS_ACCESS_KEY_ID
		export AWS_SECRET_ACCESS_KEY=$BACKUP_AWS_SECRET_ACCESS_KEY
		export AWS_DEFAULT_REGION=$BACKUP_AWS_REGION
		exec docker run "${opts[@]}" "$backup_aws_cli_image" "$@"
	)
}

# upload <local file> <key under the bucket>
upload() {
	aws --stdin s3 cp --only-show-errors - "$bucket/$2" <"$1"
	log "uploaded $bucket/$2 ($(stat -c %s "$1") bytes)"
}

# newest_key <prefix> <basename ERE>: the newest object under <prefix>/ whose
# name matches (keys sort by their date prefix and UTC timestamp).
newest_key() {
	aws s3api list-objects-v2 --bucket "$BACKUP_S3_BUCKET" --prefix "$1/" \
		--query 'Contents[].Key' --output text |
		tr '\t' '\n' | grep -E "^$1/[0-9]{4}/[0-9]{2}/[0-9]{2}/$2\$" | sort | tail -n 1
}

# db_container: production's running Postgres container, by its compose
# labels (project and service), so a container rename never breaks the job.
db_container() {
	local ids
	ids=$(docker ps -q --filter "label=com.docker.compose.project=$backup_compose_project" \
		--filter "label=com.docker.compose.service=$backup_db_service")
	[ -n "$ids" ] || die "no running $backup_db_service container in compose project $backup_compose_project"
	[ "$(wc -l <<<"$ids")" = 1 ] || die "more than one $backup_db_service container in compose project $backup_compose_project"
	printf '%s' "$ids"
}
