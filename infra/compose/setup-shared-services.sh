#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# setup-shared-services.sh <postgres container> <rabbitmq container>
#
# Each environment's access to production's Postgres and RabbitMQ, the
# services both stacks share (ADR-0016). Ansible's shared services role
# (issue 026) copies this file to the server and runs it; the local harness
# (test/run-harness.sh) runs the same file against its own stack, so what
# production runs is what the harness proves.
#
#   Postgres  CREATE ROLE <staging user> LOGIN (no superuser, createdb,
#             createrole, replication or bypassrls); CREATE DATABASE
#             <staging db> OWNER <staging user>; then CONNECT revoked from
#             PUBLIC and the staging role on every other database
#             (production's, the `postgres` maintenance database, template1).
#             Production's role is the instance's superuser and keeps
#             connecting everywhere.
#   RabbitMQ  per environment (production, staging): its vhost, and its user
#             with no tags (no management access) and permissions on that
#             vhost only (any other permission is cleared); the default
#             `guest` user is deleted.
#
# Idempotent, and it says what it did: each change prints a "changed: ..."
# line, and the last line is "setup-shared-services: changed" or
# "setup-shared-services: unchanged" (Ansible's changed_when). Passwords are
# checked first, by logging in with them, and re-applied only when they
# differ: that is how a rotation lands. It never restarts a service.
#
# Values come from the environment, as Infisical holds them (`prod/db`,
# `staging/db`, `prod/rabbitmq`, `staging/rabbitmq`); the harness reads them
# from the fixture env files:
#   PROD_DB_USER          production's POSTGRES_USER (the superuser)
#   PROD_DB_NAME          production's database
#   STAGING_DB_USER, STAGING_DB_PASSWORD, STAGING_DB_NAME
#   PROD_MQ_USER, PROD_MQ_PASSWORD, PROD_MQ_VHOST
#                         production's /rabbitmq: the image's default user and
#                         vhost (env/rabbitmq.env)
#   STAGING_MQ_USER, STAGING_MQ_PASSWORD, STAGING_MQ_VHOST
#
# Passwords never reach the docker CLI's command line, nor `docker events`,
# which records every exec's command: psql reads them from its stdin script,
# and a shell inside each container reads them from stdin. rabbitmqctl itself
# only takes a password as an argument when it has no terminal, so it sits in
# that process's arguments for the second it runs (Ansible: no_log). psql
# reports errors tersely, without the failing statement, so nothing this
# script prints holds a password (the harness checks it).
# Identifiers must be lowercase letters, digits and underscores and passwords
# letters and digits (they also go into URLs), so nothing needs escaping.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

me=setup-shared-services
die() {
	printf '%s: ERROR: %s\n' "$me" "$*" >&2
	exit 1
}

[ $# -eq 2 ] || die "usage: setup-shared-services.sh <postgres container> <rabbitmq container>"
postgres=$1
rabbitmq=$2

for var in PROD_DB_USER PROD_DB_NAME STAGING_DB_USER STAGING_DB_NAME \
	PROD_MQ_USER PROD_MQ_VHOST STAGING_MQ_USER STAGING_MQ_VHOST; do
	[[ ${!var:-} =~ ^[a-z_][a-z0-9_]*$ ]] || die "$var must be a lowercase identifier (got '${!var:-}')"
done
for var in STAGING_DB_PASSWORD PROD_MQ_PASSWORD STAGING_MQ_PASSWORD; do
	[[ ${!var:-} =~ ^[A-Za-z0-9]{16,}$ ]] || die "$var must be at least 16 letters and digits"
done
[ "$STAGING_DB_NAME" != "$PROD_DB_NAME" ] || die "staging and production must not share a database"
[ "$STAGING_DB_USER" != "$PROD_DB_USER" ] || die "staging and production must not share a Postgres role"
[ "$STAGING_MQ_VHOST" != "$PROD_MQ_VHOST" ] || die "staging and production must not share a RabbitMQ vhost"
[ "$STAGING_MQ_USER" != "$PROD_MQ_USER" ] || die "staging and production must not share a RabbitMQ user"
[ "$PROD_MQ_USER" != guest ] && [ "$STAGING_MQ_USER" != guest ] || die "guest is deleted: no environment may use it"

changes=0
report() { # stdin: "changed: ..." lines, printed and counted
	local line
	while IFS= read -r line; do
		[ -n "$line" ] || continue
		printf '%s: %s\n' "$me" "$line"
		changes=$((changes + 1))
	done
}

# ── Postgres ─────────────────────────────────────────────────────────────────
# Does <role> log in to <db> with <password>? Over TCP to the container's own
# address: the image trusts the socket and 127.0.0.1 without any password.
pg_login_ok() { # $1 role, $2 db, $3 password (on the container shell's stdin)
	printf '%s\n' "$3" | docker exec -i "$postgres" sh -c '
		IFS= read -r PGPASSWORD
		export PGPASSWORD
		exec psql -X -q -tA -h "$(hostname)" -U "$1" -d "$2" -c "SELECT 1"
	' sh "$1" "$2" >/dev/null 2>&1
}
set_password=yes
if pg_login_ok "$STAGING_DB_USER" "$STAGING_DB_NAME" "$STAGING_DB_PASSWORD"; then
	set_password=no
fi

# CREATE DATABASE cannot run inside a function or transaction, hence psql's
# \gexec (run each generated statement) instead of a DO block. Each generated
# row is the statement, then a SELECT of its "changed: ..." line, so only
# what ran is reported.
pg_out=$(docker exec -i "$postgres" psql -X -q -tA -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d postgres <<SQL
\set VERBOSITY terse
\set SHOW_CONTEXT never
\set role '$STAGING_DB_USER'
\set password '$STAGING_DB_PASSWORD'
\set set_password '$set_password'
\set db '$STAGING_DB_NAME'
SELECT format('CREATE ROLE %I', :'role'),
	format('SELECT %L', 'changed: created the Postgres role ' || :'role')
	WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'role') \gexec
SELECT format('ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS', :'role'),
	format('SELECT %L', 'changed: set the attributes of the Postgres role ' || :'role')
	FROM pg_roles WHERE rolname = :'role'
	AND NOT (rolcanlogin AND NOT (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)) \gexec
SELECT format('ALTER ROLE %I WITH PASSWORD %L', :'role', :'password'),
	format('SELECT %L', 'changed: set the password of the Postgres role ' || :'role')
	WHERE :'set_password' = 'yes' \gexec
SELECT format('CREATE DATABASE %I OWNER %I', :'db', :'role'),
	format('SELECT %L', 'changed: created the database ' || :'db')
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'db') \gexec
SELECT format('ALTER DATABASE %I OWNER TO %I', :'db', :'role'),
	format('SELECT %L', 'changed: gave the database ' || :'db' || ' to ' || :'role')
	FROM pg_database WHERE datname = :'db'
	AND datdba <> (SELECT oid FROM pg_roles WHERE rolname = :'role') \gexec
SELECT format('REVOKE CONNECT ON DATABASE %I FROM PUBLIC, %I', datname, :'role'),
	format('SELECT %L', 'changed: revoked CONNECT on the database ' || datname || ' from PUBLIC and ' || :'role')
	FROM pg_database WHERE datname <> :'db' AND datallowconn
	AND has_database_privilege(:'role', datname, 'CONNECT') ORDER BY datname \gexec
SQL
) || die "the Postgres setup failed"
report <<<"$pg_out"
echo "$me: Postgres role and database $STAGING_DB_NAME ready; $STAGING_DB_USER can connect to no other database"

# ── RabbitMQ ─────────────────────────────────────────────────────────────────
# One environment: its vhost; its user (the password checked, re-applied only
# when it differs), with no tags and full permissions on that vhost only.
# add_user is not idempotent, hence the checks before each command.
rabbitmq_env() { # $1 user, $2 vhost, $3 password (on the container shell's stdin)
	printf '%s\n' "$3" | docker exec -i "$rabbitmq" sh -ec '
		IFS= read -r password
		user=$1 vhost=$2
		changed() { printf "changed: %s\n" "$*"; }
		q() { rabbitmqctl -q "$@"; }
		if ! q list_vhosts --no-table-headers name | grep -qxF "$vhost"; then
			q add_vhost "$vhost" >/dev/null
			changed "created the RabbitMQ vhost $vhost"
		fi
		if ! q list_users --no-table-headers | cut -f1 | grep -qxF "$user"; then
			q add_user "$user" "$password" >/dev/null
			changed "created the RabbitMQ user $user"
		elif ! q authenticate_user "$user" "$password" >/dev/null 2>&1; then
			q change_password "$user" "$password" >/dev/null
			changed "set the password of the RabbitMQ user $user"
		fi
		tags=$(q list_users --no-table-headers | awk -F "\t" -v u="$user" "\$1 == u { print \$2 }")
		if [ "$tags" != "[]" ]; then
			q set_user_tags "$user" >/dev/null
			changed "cleared the tags $tags of the RabbitMQ user $user"
		fi
		perms=$(q list_user_permissions --no-table-headers "$user")
		for other in $(printf "%s\n" "$perms" | cut -f1 | grep -vxF "$vhost" || true); do
			q clear_permissions -p "$other" "$user" >/dev/null
			changed "cleared the permissions of the RabbitMQ user $user on the vhost $other"
		done
		own=$(printf "%s\n" "$perms" | awk -F "\t" -v v="$vhost" "\$1 == v { print \$2, \$3, \$4 }")
		if [ "$own" != ".* .* .*" ]; then
			q set_permissions -p "$vhost" "$user" ".*" ".*" ".*" >/dev/null
			changed "gave the RabbitMQ user $user full permissions on the vhost $vhost"
		fi
	' sh "$1" "$2"
}
mq_out=$(rabbitmq_env "$PROD_MQ_USER" "$PROD_MQ_VHOST" "$PROD_MQ_PASSWORD") ||
	die "the RabbitMQ setup failed for $PROD_MQ_USER"
report <<<"$mq_out"
mq_out=$(rabbitmq_env "$STAGING_MQ_USER" "$STAGING_MQ_VHOST" "$STAGING_MQ_PASSWORD") ||
	die "the RabbitMQ setup failed for $STAGING_MQ_USER"
report <<<"$mq_out"

# The image creates `guest` only when RABBITMQ_DEFAULT_USER is unset; delete
# it whenever it is there anyway (e.g. a data directory from before).
mq_out=$(docker exec "$rabbitmq" sh -ec '
	if rabbitmqctl -q list_users --no-table-headers | cut -f1 | grep -qx guest; then
		rabbitmqctl -q delete_user guest >/dev/null
		echo "changed: deleted the RabbitMQ user guest"
	fi
') || die "could not check for, or delete, the RabbitMQ guest user"
report <<<"$mq_out"
echo "$me: RabbitMQ vhosts $PROD_MQ_VHOST and $STAGING_MQ_VHOST ready, each user on its own vhost only; no guest user"

if [ "$changes" -gt 0 ]; then
	echo "$me: changed"
else
	echo "$me: unchanged"
fi
