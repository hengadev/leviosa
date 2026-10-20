#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# create-staging-services.sh <postgres container> <rabbitmq container>
#
# HARNESS FIXTURE, and the reference for Ansible's shared services role
# (issue 026): give staging its own Postgres role and database and its own
# RabbitMQ vhost and user on production's services, so that staging cannot
# reach production's data (ADR-0016). Idempotent: a second run changes
# nothing but re-applies the passwords (which is how a rotation lands).
#
#   Postgres  CREATE ROLE <staging user> LOGIN (no superuser, createdb,
#             createrole, replication or bypassrls); CREATE DATABASE
#             <staging db> OWNER <staging user>; REVOKE CONNECT ON DATABASE
#             <production db> FROM PUBLIC — production's role is the
#             instance's superuser and keeps connecting, any other role
#             (staging's) is refused.
#   RabbitMQ  add_vhost <staging vhost>; add_user <staging user> (no tags, so
#             no management access); set_permissions on that vhost only.
#
# Values come from the environment, as Infisical holds them (`prod/db`,
# `staging/db`, `staging/rabbitmq`); the harness reads them from the fixture
# env files:
#   PROD_DB_USER          production's POSTGRES_USER (the superuser)
#   PROD_DB_NAME          production's database
#   STAGING_DB_USER, STAGING_DB_PASSWORD, STAGING_DB_NAME
#   STAGING_MQ_USER, STAGING_MQ_PASSWORD, STAGING_MQ_VHOST
#
# Passwords never reach the docker CLI's command line, nor `docker events`,
# which records every exec's command: psql reads them from its stdin script,
# and a shell inside the RabbitMQ container reads them from stdin. rabbitmqctl
# itself only takes a password as an argument when it has no terminal, so it
# sits in that process's arguments for the second it runs (Ansible: no_log).
# Identifiers must be lowercase letters, digits and underscores and passwords
# letters and digits (they also go into URLs), so nothing needs escaping.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

die() {
	printf 'create-staging-services: ERROR: %s\n' "$*" >&2
	exit 1
}

[ $# -eq 2 ] || die "usage: create-staging-services.sh <postgres container> <rabbitmq container>"
postgres=$1
rabbitmq=$2

for var in PROD_DB_USER PROD_DB_NAME STAGING_DB_USER STAGING_DB_NAME STAGING_MQ_USER STAGING_MQ_VHOST; do
	[[ ${!var:-} =~ ^[a-z_][a-z0-9_]*$ ]] || die "$var must be a lowercase identifier (got '${!var:-}')"
done
for var in STAGING_DB_PASSWORD STAGING_MQ_PASSWORD; do
	[[ ${!var:-} =~ ^[A-Za-z0-9]{16,}$ ]] || die "$var must be at least 16 letters and digits"
done
[ "$STAGING_DB_NAME" != "$PROD_DB_NAME" ] || die "staging and production must not share a database"
[ "$STAGING_DB_USER" != "$PROD_DB_USER" ] || die "staging and production must not share a Postgres role"

# ── Postgres ─────────────────────────────────────────────────────────────────
# CREATE DATABASE cannot run inside a function or transaction, hence psql's
# \gexec (run each generated statement) instead of a DO block.
docker exec -i "$postgres" psql -X -q -v ON_ERROR_STOP=1 -U "$PROD_DB_USER" -d postgres <<SQL
\set role '$STAGING_DB_USER'
\set password '$STAGING_DB_PASSWORD'
\set db '$STAGING_DB_NAME'
\set prod_db '$PROD_DB_NAME'
SELECT format('CREATE ROLE %I', :'role')
	WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = :'role') \gexec
SELECT format('ALTER ROLE %I WITH LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD %L',
	:'role', :'password') \gexec
SELECT format('CREATE DATABASE %I OWNER %I', :'db', :'role')
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'db') \gexec
SELECT format('ALTER DATABASE %I OWNER TO %I', :'db', :'role') \gexec
SELECT format('REVOKE CONNECT ON DATABASE %I FROM PUBLIC', :'prod_db') \gexec
SQL
echo "create-staging-services: Postgres role and database $STAGING_DB_NAME ready; PUBLIC has no CONNECT on $PROD_DB_NAME"

# ── RabbitMQ ─────────────────────────────────────────────────────────────────
# add_vhost is a no-op on an existing vhost; add_user is not, so an existing
# user only gets its password re-applied.
printf '%s\n' "$STAGING_MQ_PASSWORD" | docker exec -i "$rabbitmq" sh -ec '
	IFS= read -r password
	rabbitmqctl -q add_vhost "$2"
	if rabbitmqctl -q list_users --no-table-headers | cut -f1 | grep -qx "$1"; then
		rabbitmqctl -q change_password "$1" "$password"
	else
		rabbitmqctl -q add_user "$1" "$password"
	fi
	rabbitmqctl -q set_user_tags "$1"
	rabbitmqctl -q set_permissions -p "$2" "$1" ".*" ".*" ".*"
' sh "$STAGING_MQ_USER" "$STAGING_MQ_VHOST" >/dev/null
echo "create-staging-services: RabbitMQ vhost $STAGING_MQ_VHOST and user $STAGING_MQ_USER ready"
