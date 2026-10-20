#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# Local Docker harness for docker-compose.prod.yml, docker-compose.staging.yml
# and `deploy` (issues 018 and 019).
#
#   infra/compose/test/run-harness.sh
#
# One command, real Docker, no secrets, nothing pushed. It builds the backend
# and frontend images locally with fake COMMIT_SHAs, lays out temporary prod
# and staging stack roots with the fixture env files (fixtures/, fake
# stand-ins for the Infisical Agent's renderings), runs LocalStack (KMS only)
# for the encx KEKs, and drives the real deploy script through the issues'
# acceptance criteria:
#
#   production (018)
#   1. first deploy: migrates, bootstraps the admin, both apps report the SHA
#   2. second deploy: data kept, admin untouched
#   3. failing migration: aborts before recreate, previous containers serve
#   4. /healthz SHA mismatch (backend or frontend) fails the deploy
#   5. no ADMIN_* value in the running backend or frontend
#   6. data survives container recreation (`down`, then a deploy)
#   staging beside production (019) — after fixtures/create-staging-services.sh
#   created staging's Postgres role and database and RabbitMQ vhost and user
#   on production's services, as Ansible will (issue 026)
#   S1. both stacks deployed with different SHAs, each /healthz reports its own
#   S2. production's frontend only reaches production's backend (no alias
#       collision on the shared network)
#   S3. the staging Postgres role cannot connect to production's database
#   S4. the staging RabbitMQ user cannot access production's vhost
#   S5. staging's env files hold no production Redis/Postgres/RabbitMQ
#       credential
#   S6. a staging deploy restarts no production container
#   (and, before production exists, a staging deploy fails at migrate)
#   7. nothing left behind: containers, networks, volumes, images, temp files
#
# Isolation: every name carries a random run id (leviosa-harness-<id>): its
# own compose projects (<name>-prod, <name>-staging), network, LocalStack
# container and image repositories (<name>-backend / <name>-frontend, so the
# henga/leviosa-* pins are never touched). Nothing binds a fixed host port
# (Caddy gets a random loopback port). Everything on disk lives under one
# mktemp -d. An EXIT trap tears it all down on success and failure; public
# images the run had to pull are removed too (HARNESS_KEEP_PULLED=1 keeps
# them for faster reruns).
#
# Knobs: KEEP=1 keeps the stack and temp dir for inspection (no teardown);
# HARNESS_KEEP_PULLED=1 keeps pulled public images; TMPDIR as for mktemp.
# Requires docker with the compose plugin (>= 2.24, for !override); cold
# runs take several minutes (image builds).
#
# shellcheck disable=SC2016
# ^ single-quoted `bash -c` / `sh -c` scripts expand their own positional
#   arguments on purpose.
# ─────────────────────────────────────────────────────────────────────────────
set -euo pipefail

HARNESS_DIR=$(cd "$(dirname "$0")" && pwd)
COMPOSE_DIR=$(cd "$HARNESS_DIR/.." && pwd)
REPO_ROOT=$(cd "$COMPOSE_DIR/../.." && pwd)
DEPLOY=$COMPOSE_DIR/deploy
FIXTURES=$HARNESS_DIR/fixtures

RUN_ID=$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')
NAME=leviosa-harness-$RUN_ID
NETWORK=$NAME
PROJECT=$NAME-prod
PROJECT_STAGING=$NAME-staging
LOCALSTACK=$NAME-localstack
BACKEND_REPO=$NAME-backend
FRONTEND_REPO=$NAME-frontend
LABEL=leviosa.harness=$RUN_ID
CREATE_STAGING_SERVICES=$FIXTURES/create-staging-services.sh

LOCALSTACK_IMAGE=localstack/localstack:4.6
AWS_REGION=eu-central-1             # as in fixtures/{prod,staging}/env/backend.env
KMS_ALIAS=alias/leviosa-harness     # ENCX_KMS_KEY_ALIAS in the prod file
KMS_ALIAS_STAGING=alias/leviosa-harness-staging # ... and in the staging file
STAGING_HOST=staging.harness.leviosa.test       # fixtures/Caddyfile's staging site
BROKEN_MIGRATION=29991231235959_harness_broken_migration.sql

# Fake, obviously-not-real 40-hex commit SHAs, baked into images via COMMIT_SHA.
fake_sha() { printf 'c0ffee%034d' "$1"; }
SHA1=$(fake_sha 1)
SHA2=$(fake_sha 2)
SHA_MIGFAIL=$(fake_sha 3)      # backend with an extra, broken migration
SHA_BAD_BACKEND=$(fake_sha 4)  # frontend reports it, backend reports SHA1
SHA_BAD_FRONTEND=$(fake_sha 5) # backend reports it, frontend reports SHA1

TMP=""
TORN_DOWN=0
SUMMARY_DONE=0
PULLED=() # public images absent before this run
SECTIONS=()
declare -A SECTION_FAILS=()
CURRENT=""
TOTAL_PASS=0
TOTAL_FAIL=0
CADDY_ADDRESS="" # Caddy's random loopback address, once it runs
# Production's containers before and after the first staging deploy (S6).
PROD_BEFORE_STAGING=""
PROD_AFTER_FIRST_STAGING=""

# ── reporting ────────────────────────────────────────────────────────────────
section() {
	CURRENT=$1
	SECTIONS+=("$1")
	SECTION_FAILS[$1]=0
	printf '\n== %s ==\n' "$1"
}
ok() {
	echo "    ok   - $*"
	TOTAL_PASS=$((TOTAL_PASS + 1))
}
bad() {
	echo "    FAIL - $*"
	TOTAL_FAIL=$((TOTAL_FAIL + 1))
	if [ -n "$CURRENT" ]; then
		SECTION_FAILS[$CURRENT]=$((SECTION_FAILS[$CURRENT] + 1))
	fi
}
check() { # check <description> <command...>: ok when the command succeeds
	local description=$1
	shift
	if "$@"; then ok "$description"; else bad "$description"; fi
}
note() { echo "    ..   $*"; }
show_log() { tail -n "${2:-25}" "$1" | sed 's/^/           | /'; }
fatal() {
	bad "$*"
	echo "    the remaining cases depend on this step: stopping" >&2
	exit 1
}

# ── stack access (the harness's project name, compose override and env) ─────
compose_files() { echo "$TMP/prod/docker-compose.yml:$TMP/harness-override.yml"; }
dc() {
	env COMPOSE_FILE="$(compose_files)" COMPOSE_PROJECT_NAME="$PROJECT" \
		docker compose --project-directory "$TMP/prod" "$@"
}
deploy_prod() { # $1 sha, $2 log name; returns the deploy's exit code
	env LEVIOSA_PROD_ROOT="$TMP/prod" \
		LEVIOSA_BACKEND_REPO="$BACKEND_REPO" LEVIOSA_FRONTEND_REPO="$FRONTEND_REPO" \
		LEVIOSA_NETWORK="$NETWORK" LEVIOSA_SKIP_PULL=1 \
		COMPOSE_FILE="$(compose_files)" COMPOSE_PROJECT_NAME="$PROJECT" \
		"$DEPLOY" prod "$1" >"$TMP/logs/$2.log" 2>&1
}
# The staging stack: its own root, override and compose project.
staging_compose_files() { echo "$TMP/staging/docker-compose.yml:$TMP/harness-staging-override.yml"; }
dcs() {
	env COMPOSE_FILE="$(staging_compose_files)" COMPOSE_PROJECT_NAME="$PROJECT_STAGING" \
		docker compose --project-directory "$TMP/staging" "$@"
}
deploy_staging() { # $1 sha, $2 log name; returns the deploy's exit code
	env LEVIOSA_STAGING_ROOT="$TMP/staging" \
		LEVIOSA_BACKEND_REPO="$BACKEND_REPO" LEVIOSA_FRONTEND_REPO="$FRONTEND_REPO" \
		LEVIOSA_NETWORK="$NETWORK" LEVIOSA_SKIP_PULL=1 \
		COMPOSE_FILE="$(staging_compose_files)" COMPOSE_PROJECT_NAME="$PROJECT_STAGING" \
		"$DEPLOY" staging "$1" >"$TMP/logs/$2.log" 2>&1
}

cid() { dc ps -q "$1"; }   # running container of a production service
scid() { dcs ps -q "$1"; } # running container of a staging service
health() { docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$1"; }
image_id() { docker image inspect -f '{{.Id}}' "$1"; }
container_image() { docker inspect -f '{{.Image}}' "$1"; }
container_env() { docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$1"; }

# GET <frontend service>:3000/healthz (default: production's `frontend`) from
# a throwaway container on the stacks' network (the backend image has busybox
# wget) and print "<frontend> <backend>".
healthz_shas() {
	local body
	body=$(docker run --rm --label "$LABEL" --network "$NETWORK" --entrypoint wget \
		"$BACKEND_REPO:$SHA1" -q -O - -T 5 "http://${1:-frontend}:3000/healthz" 2>/dev/null) || {
		echo "unreachable"
		return 0
	}
	healthz_body_shas <<<"$body"
}
healthz_body_shas() { # stdin: /healthz JSON bodies, one per line -> "<frontend> <backend>" each
	sed -e 's/.*"frontend":"\([^"]*\)".*"backend":"\([^"]*\)".*/\1 \2/' -e t \
		-e '/^unreachable$/b' -e 's/.*/unexpected/'
}
serving() { [ "$(healthz_shas "${2:-frontend}")" = "$1 $1" ]; } # both apps of a stack report $1
# The distinct "<frontend> <backend>" answers of $2 consecutive /healthz
# requests to $1, from one throwaway container.
healthz_repeated() {
	docker run --rm --label "$LABEL" --network "$NETWORK" --entrypoint sh "$BACKEND_REPO:$SHA1" -c '
		i=0
		while [ "$i" -lt "$2" ]; do
			wget -q -O - -T 5 "http://$1:3000/healthz" || printf unreachable
			echo
			i=$((i + 1))
		done' sh "$1" "$2" 2>/dev/null | healthz_body_shas | sort | uniq -c | sed 's/^ *//' | paste -sd ',' -
}

psql_q() { dc exec -T postgres psql -U leviosa -d leviosa -v ON_ERROR_STOP=1 -qtAc "$1"; }
# As production's superuser, in another database of the shared instance.
psql_db() { dc exec -T postgres psql -U leviosa -d "$1" -v ON_ERROR_STOP=1 -qtAc "$2"; }
user_count() { psql_q 'SELECT count(*) FROM auth.users'; }
# Everything bootstrap-admin could change on an existing account.
admin_fingerprint() {
	psql_q "SELECT string_agg(id || '|' || password_hash_secure || '|' || email_hash || '|' ||
		md5(first_name_encrypted) || '|' || md5(role_encrypted) || '|' || updated_at, ',') FROM auth.users"
}
schema_version() { psql_q 'SELECT max(version_id) FROM goose_db_version'; }
newest_migration() {
	find "$REPO_ROOT/backend/internal/common/migrations" -name '*.sql' -printf '%f\n' |
		sed -n 's/^\([0-9]*\)_.*/\1/p' | sort -n | tail -n 1
}
redis_cli() { dc exec -T redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli "$@"' redis-cli "$@"; }
rabbitmqctl_q() { dc exec -T rabbitmq rabbitmqctl -q "$@"; }
oneoff_containers() {
	docker ps -aq --filter "label=com.docker.compose.project=$PROJECT" --filter label=com.docker.compose.oneoff=True
}
log_has() { grep -q -- "$2" "$TMP/logs/$1.log"; }
mounts_are_binds() { # $1 container: it has mounts, and every one is a bind mount
	local types
	types=$(docker inspect -f '{{range .Mounts}}{{.Type}}{{"\n"}}{{end}}' "$1" | grep -v '^$')
	[ -n "$types" ] && ! grep -qvx bind <<<"$types"
}
# The compose file without its comments, and which service each line belongs to.
compose_lines_by_service() { # $1 file -> "<service> <line>" for lines under services:
	awk '/^[[:space:]]*#/ { next }
		/^[^[:space:]]/ { top = $1; next }
		top == "services:" && /^  [a-z_]+:$/ { svc = $1; sub(":", "", svc); next }
		top == "services:" { print svc, $0 }' "$1"
}

# ── staging isolation helpers ────────────────────────────────────────────────
env_value() { sed -n "s/^$2=//p" "$1" | tail -n 1; } # $1 env file, $2 key
# scheme://user:password@host[:port]/path -> "<user> <password> <host> <path>"
url_parts() {
	[[ $1 =~ ^[a-z]+://([^:@/]+):([^@/]+)@([^:/]+)(:[0-9]+)?/([^?]*) ]] || return 1
	printf '%s %s %s %s\n' "${BASH_REMATCH[1]}" "${BASH_REMATCH[2]}" "${BASH_REMATCH[3]}" "${BASH_REMATCH[5]}"
}
# Staging's /db and /rabbitmq atoms, as the Agent template would assemble
# them, read back from the staging backend.env fixture's URLs.
staging_db() { url_parts "$(env_value "$TMP/staging/env/backend.env" DATABASE_URL)"; }
staging_mq() { url_parts "$(env_value "$TMP/staging/env/backend.env" RABBITMQ_URL)"; }
field() { cut -d ' ' -f "$1"; } # field 1 user, 2 password, 3 host, 4 database or vhost

# fixtures/create-staging-services.sh with the values Ansible would read
# from Infisical (prod/db, staging/db, staging/rabbitmq).
create_staging_services() { # $1 log name
	local db mq
	db=$(staging_db) && mq=$(staging_mq) || return 1
	env PROD_DB_USER="$(env_value "$TMP/prod/env/postgres.env" POSTGRES_USER)" \
		PROD_DB_NAME="$(env_value "$TMP/prod/env/postgres.env" POSTGRES_DB)" \
		STAGING_DB_USER="$(field 1 <<<"$db")" STAGING_DB_PASSWORD="$(field 2 <<<"$db")" \
		STAGING_DB_NAME="$(field 4 <<<"$db")" \
		STAGING_MQ_USER="$(field 1 <<<"$mq")" STAGING_MQ_PASSWORD="$(field 2 <<<"$mq")" \
		STAGING_MQ_VHOST="$(field 4 <<<"$mq")" \
		"$CREATE_STAGING_SERVICES" "$(cid postgres)" "$(cid rabbitmq)" >"$TMP/logs/$1.log" 2>&1
}

# Every production container (running or not) with what a restart or a
# recreate changes: id, start time, restart count.
prod_state() {
	local id
	for id in $(docker ps -aq --filter "label=com.docker.compose.project=$PROJECT"); do
		docker inspect -f '{{.Name}} {{.Id}} {{.State.Status}} {{.State.StartedAt}} restarts={{.RestartCount}}' "$id"
	done | sort
}
pins_state() { echo "$(image_id "$BACKEND_REPO:prod") $(image_id "$FRONTEND_REPO:prod")"; }

container_ip() { docker inspect -f "{{(index .NetworkSettings.Networks \"$NETWORK\").IPAddress}}" "$1"; }
# The IPv4 addresses a name resolves to from inside a frontend container
# (distroless: node is the only tool), sorted, space-separated.
resolve_from() { # $1 container, $2 name
	docker exec "$1" /nodejs/bin/node -e '
		require("dns").promises.lookup(process.argv[1], { all: true, family: 4 })
			.then((a) => console.log(a.map((x) => x.address).sort().join(" ")))
			.catch(() => console.log("unresolved"));' "$2"
}
# DNS names (service aliases, container names, ids) that more than one
# container on the shared network answers for.
shared_dns_names() {
	local id
	for id in $(docker network inspect -f '{{range $id, $c := .Containers}}{{$id}} {{end}}' "$NETWORK"); do
		docker inspect -f "{{range (index .NetworkSettings.Networks \"$NETWORK\").DNSNames}}{{.}}{{\"\\n\"}}{{end}}" "$id" |
			sort -u
	done | grep -v '^$' | sort | uniq -d | tr '\n' ' '
}

# AMQP 0-9-1 handshake against production's `rabbitmq`, from a throwaway
# container (busybox nc): PLAIN login as $1/$2, then Connection.Open of vhost
# $3. Prints "open-ok", "refused <code> <reply text>" (the broker closed the
# connection, e.g. 530 NOT_ALLOWED) or "no answer" (e.g. bad credentials:
# RabbitMQ just drops the socket). The password travels on stdin only.
hex() { printf '%s' "$1" | od -An -tx1 -v | tr -d ' \n'; }
amqp_shortstr() { printf '%02x%s' "${#1}" "$(hex "$1")"; }
amqp_frame() { printf '010000%08x%sce' $((${#1} / 2)) "$1"; } # method frame, channel 0
hex_to_bytes() {
	local i
	for ((i = 0; i < ${#1}; i += 2)); do printf '%b' "\\x${1:i:2}"; done
}
amqp_open() {
	local response start_ok tune_ok open reply rest code len text
	local -a bytes
	response="00$(hex "$1")00$(hex "$2")"
	start_ok="000a000b00000000$(amqp_shortstr PLAIN)$(printf '%08x' $((${#response} / 2)))$response$(amqp_shortstr en_US)"
	tune_ok="000a001f0001000010000000" # channel-max 1, frame-max 4096, no heartbeat
	open="000a0028$(amqp_shortstr "$3")0000"
	reply=$(hex_to_bytes "414d515000000901$(amqp_frame "$start_ok")$(amqp_frame "$tune_ok")$(amqp_frame "$open")" |
		docker run --rm -i --label "$LABEL" --network "$NETWORK" --entrypoint sh "$BACKEND_REPO:$SHA1" \
			-c 'timeout 3 nc rabbitmq 5672' 2>/dev/null | od -An -tx1 -v | tr -s ' \n' ' ') || true
	if [[ $reply == *" 01 00 00 00 00 00 05 00 0a 00 29 00 ce"* ]]; then
		echo "open-ok"
	elif [[ $reply == *" 00 0a 00 32 "* ]]; then
		rest=${reply#*" 00 0a 00 32 "}
		read -r -a bytes <<<"$rest"
		code=$((16#${bytes[0]}${bytes[1]}))
		len=$((16#${bytes[2]}))
		text=$(printf '%b' "$(printf '\\x%s' "${bytes[@]:3:len}")")
		echo "refused $code $text"
	else
		echo "no answer"
	fi
}

# Connect to production's Postgres over TCP as role $1 (password $2) to
# database $3 and print the current user, or psql's error. The password goes
# through the environment, never a command line.
pg_connect_as() {
	env PGPASSWORD="$2" docker exec -e PGPASSWORD "$(cid postgres)" \
		psql -X -h 127.0.0.1 -U "$1" -d "$3" -tAc 'SELECT current_user' 2>&1 || true
}

# Production's Redis, Postgres and RabbitMQ passwords, from its env files:
# the dedicated files and the URLs in backend.env.
prod_secrets() {
	local env=$TMP/prod/env
	{
		env_value "$env/postgres.env" POSTGRES_PASSWORD
		env_value "$env/rabbitmq.env" RABBITMQ_DEFAULT_PASS
		env_value "$env/redis.env" REDIS_PASSWORD
		env_value "$env/backend.env" REDIS_PASSWORD
		url_parts "$(env_value "$env/backend.env" DATABASE_URL)" | field 2
		url_parts "$(env_value "$env/backend.env" RABBITMQ_URL)" | field 2
	} | grep -v '^$' | sort -u
}
# stdin: KEY=VALUE lines (env file or container environment); $1 label.
# Prints "<label>:<KEY>" for every value containing a production secret —
# key names only, never the values.
leaked_keys() {
	local label=$1 line key value secret
	local -a secrets
	mapfile -t secrets < <(prod_secrets)
	while IFS= read -r line; do
		[[ $line =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]] || continue
		key=${BASH_REMATCH[1]}
		value=${BASH_REMATCH[2]}
		[ -n "$value" ] || continue
		for secret in "${secrets[@]}"; do
			if [[ $value == *"$secret"* ]]; then
				printf '%s:%s ' "$label" "$key"
				break
			fi
		done
	done
}

# ── teardown ─────────────────────────────────────────────────────────────────
harness_containers() {
	{
		docker ps -aq --filter "label=com.docker.compose.project=$PROJECT"
		docker ps -aq --filter "label=com.docker.compose.project=$PROJECT_STAGING"
		docker ps -aq --filter "label=$LABEL"
	} | sort -u
}
harness_images() {
	docker images --format '{{.Repository}}:{{.Tag}}' "$BACKEND_REPO"
	docker images --format '{{.Repository}}:{{.Tag}}' "$FRONTEND_REPO"
}
teardown() {
	[ "$TORN_DOWN" = 1 ] && return 0
	TORN_DOWN=1
	set +e
	if [ "${KEEP:-0}" = 1 ]; then
		echo "KEEP=1: left running: projects $PROJECT and $PROJECT_STAGING, network $NETWORK, $LOCALSTACK; temp dir $TMP" >&2
		return 0
	fi
	local ids images
	mapfile -t ids < <(harness_containers)
	# -v also removes the anonymous volumes (LocalStack's) of these containers.
	[ ${#ids[@]} -gt 0 ] && docker rm -fv "${ids[@]}" >/dev/null 2>&1
	docker network rm "$NETWORK" >/dev/null 2>&1
	if [ -n "$TMP" ] && [ -d "$TMP" ]; then
		# Postgres, RabbitMQ, Redis and Caddy write into the bind mounts as
		# their own users: files the developer cannot unlink. Wipe them from
		# a throwaway root container first.
		if ! rm -rf "$TMP" 2>/dev/null; then
			docker run --rm --label "$LABEL" -v "$TMP:/wipe" --entrypoint /bin/sh redis:7-alpine \
				-c 'find /wipe -mindepth 1 -delete' >/dev/null 2>&1
			rm -rf "$TMP"
		fi
	fi
	mapfile -t images < <(harness_images)
	[ ${#images[@]} -gt 0 ] && docker image rm "${images[@]}" >/dev/null 2>&1
	if [ "${HARNESS_KEEP_PULLED:-0}" != 1 ] && [ ${#PULLED[@]} -gt 0 ]; then
		# Without -f: an image another container started using meanwhile stays.
		docker image rm "${PULLED[@]}" >/dev/null 2>&1
	fi
	set -e
}
on_exit() {
	local status=$?
	teardown
	if [ "$SUMMARY_DONE" = 0 ]; then
		summary || true
		echo "(stopped early, exit status $status)"
	fi
}
trap on_exit EXIT
trap 'exit 130' INT TERM

# ── setup ────────────────────────────────────────────────────────────────────
public_images() {
	# Services of both compose files (minus the app pins), LocalStack, and the
	# Dockerfiles' base images and syntax frontend.
	local stack
	for stack in prod staging; do
		COMPOSE_FILE="$COMPOSE_DIR/docker-compose.$stack.yml" docker compose \
			--project-directory "$FIXTURES/$stack" --profile oneoff config --images 2>/dev/null |
			grep -v '^henga/leviosa-'
	done
	echo "$LOCALSTACK_IMAGE"
	sed -n -e 's/^FROM \([^ ]*[:/][^ ]*\).*/\1/p' -e 's/^# syntax=\(.*\)$/\1/p' \
		"$REPO_ROOT/backend/Dockerfile" "$REPO_ROOT/frontend/Dockerfile"
}

setup() {
	section "setup"
	command -v docker >/dev/null || fatal "docker is required"
	docker compose version >/dev/null 2>&1 || fatal "the docker compose plugin is required"

	local image
	while read -r image; do
		docker image inspect "$image" >/dev/null 2>&1 || PULLED+=("$image")
	done < <(public_images | sort -u)
	if [ ${#PULLED[@]} -gt 0 ]; then
		note "public images not present yet (pulled now, removed at teardown): ${PULLED[*]}"
	fi

	TMP=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-harness.XXXXXX")
	# The stack root, as Ansible's app layout role and the root runner leave it.
	mkdir -p "$TMP/logs" "$TMP/prod/env" \
		"$TMP/prod/data/postgres" "$TMP/prod/data/rabbitmq" "$TMP/prod/data/redis" \
		"$TMP/prod/caddy/etc" "$TMP/prod/caddy/data" "$TMP/prod/caddy/config" "$TMP/prod/caddy/logs"
	cp "$COMPOSE_DIR/docker-compose.prod.yml" "$TMP/prod/docker-compose.yml"
	cp "$FIXTURES/prod/env/"*.env "$TMP/prod/env/"
	chmod 600 "$TMP/prod/env/"*.env # the Agent renders them 0600
	cp "$FIXTURES/Caddyfile" "$TMP/prod/caddy/etc/Caddyfile"

	# The harness's only deviations from production, layered over the
	# unmodified compose file: its own image repositories (same :prod pin
	# tag), its own network, and Caddy on a random loopback port.
	cat >"$TMP/harness-override.yml" <<EOF
# generated by run-harness.sh (run $RUN_ID)
services:
  backend:
    image: $BACKEND_REPO:prod
  migrate:
    image: $BACKEND_REPO:prod
  admin:
    image: $BACKEND_REPO:prod
  frontend:
    image: $FRONTEND_REPO:prod
  caddy:
    ports: !override
      - "127.0.0.1::80"
networks:
  leviosa_network:
    name: $NETWORK
    external: true
EOF
	check "harness override validates with the production file" dc --profile oneoff config -q

	# The staging stack root (/opt/leviosa-staging on the server), and the
	# same deviations for it: image repositories (same :staging pin tag) and
	# the network name. No port to override: staging publishes none.
	mkdir -p "$TMP/staging/env" "$TMP/staging/data/redis"
	cp "$COMPOSE_DIR/docker-compose.staging.yml" "$TMP/staging/docker-compose.yml"
	cp "$FIXTURES/staging/env/"*.env "$TMP/staging/env/"
	chmod 600 "$TMP/staging/env/"*.env
	cat >"$TMP/harness-staging-override.yml" <<EOF
# generated by run-harness.sh (run $RUN_ID)
services:
  staging_backend:
    image: $BACKEND_REPO:staging
  staging_migrate:
    image: $BACKEND_REPO:staging
  staging_admin:
    image: $BACKEND_REPO:staging
  staging_frontend:
    image: $FRONTEND_REPO:staging
networks:
  leviosa_network:
    name: $NETWORK
    external: true
EOF
	check "harness override validates with the staging file" dcs --profile oneoff config -q

	docker network create --label "$LABEL" "$NETWORK" >/dev/null || fatal "could not create network $NETWORK"
	ok "network $NETWORK created"
	note "temp dir $TMP"

	# LocalStack, KMS only, as `localstack` on the stacks' network; one KEK per
	# environment, each behind its alias, the way Terraform provisions them.
	docker run -d --name "$LOCALSTACK" --label "$LABEL" --network "$NETWORK" --network-alias localstack \
		-e SERVICES=kms -e PERSISTENCE=0 -e AWS_DEFAULT_REGION="$AWS_REGION" \
		"$LOCALSTACK_IMAGE" >/dev/null || fatal "could not start LocalStack"
	local waited=0
	until docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms list-keys >/dev/null 2>&1; do
		[ "$waited" -lt 120 ] || fatal "LocalStack KMS not ready after 120s"
		sleep 2
		waited=$((waited + 2))
	done
	local key_id alias
	for alias in "$KMS_ALIAS" "$KMS_ALIAS_STAGING"; do
		key_id=$(docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms create-key \
			--description "Leviosa harness encx KEK ($alias)" --query KeyMetadata.KeyId --output text) ||
			fatal "could not create the KMS key for $alias"
		docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms create-alias \
			--alias-name "$alias" --target-key-id "$key_id" ||
			fatal "could not create the KMS alias $alias"
	done
	ok "LocalStack KMS up with $KMS_ALIAS (prod) and $KMS_ALIAS_STAGING (staging)"
}

build() { # $1 context, $2 image, $3 COMMIT_SHA
	local log
	log="$TMP/logs/build-$(basename "$2" | tr ':' '-').log"
	note "building $2"
	if ! docker build --build-arg COMMIT_SHA="$3" --label "$LABEL" -t "$2" "$1" >"$log" 2>&1; then
		show_log "$log" 40
		fatal "build of $2 failed"
	fi
}

build_images() {
	section "build images (fake SHAs)"
	build "$REPO_ROOT/backend" "$BACKEND_REPO:$SHA1" "$SHA1"
	build "$REPO_ROOT/frontend" "$FRONTEND_REPO:$SHA1" "$SHA1"
	build "$REPO_ROOT/backend" "$BACKEND_REPO:$SHA2" "$SHA2"
	build "$REPO_ROOT/frontend" "$FRONTEND_REPO:$SHA2" "$SHA2"

	# A release whose migrate fails: the current backend plus one broken
	# migration, from a temporary copy (the frontend is just SHA2's).
	mkdir -p "$TMP/backend-migfail"
	cp -a "$REPO_ROOT/backend/." "$TMP/backend-migfail/"
	cp "$FIXTURES/broken-migration.sql" "$TMP/backend-migfail/internal/common/migrations/$BROKEN_MIGRATION"
	build "$TMP/backend-migfail" "$BACKEND_REPO:$SHA_MIGFAIL" "$SHA_MIGFAIL"
	docker tag "$FRONTEND_REPO:$SHA2" "$FRONTEND_REPO:$SHA_MIGFAIL"

	# Releases that report the wrong SHA: one app built for the release, the
	# other an older image tagged with the release's SHA.
	build "$REPO_ROOT/frontend" "$FRONTEND_REPO:$SHA_BAD_BACKEND" "$SHA_BAD_BACKEND"
	docker tag "$BACKEND_REPO:$SHA1" "$BACKEND_REPO:$SHA_BAD_BACKEND"
	build "$REPO_ROOT/backend" "$BACKEND_REPO:$SHA_BAD_FRONTEND" "$SHA_BAD_FRONTEND"
	docker tag "$FRONTEND_REPO:$SHA1" "$FRONTEND_REPO:$SHA_BAD_FRONTEND"
	ok "images built and tagged under $BACKEND_REPO and $FRONTEND_REPO"
}

# ── cases ────────────────────────────────────────────────────────────────────
case_static() {
	section "static: compose file and deploy script"
	local file=$COMPOSE_DIR/docker-compose.prod.yml
	check "deploy passes bash -n" bash -n "$DEPLOY"
	if command -v shellcheck >/dev/null; then
		check "deploy, run-harness.sh and create-staging-services.sh pass shellcheck" \
			shellcheck "$DEPLOY" "$HARNESS_DIR/run-harness.sh" "$CREATE_STAGING_SERVICES"
	else
		note "shellcheck not installed: skipped"
	fi
	check "docker-compose.prod.yml validates with the fixture env files" \
		env COMPOSE_FILE="$TMP/prod/docker-compose.yml" docker compose --project-directory "$TMP/prod" config -q
	check "no inline environment: values come only from env_file" \
		bash -c '! grep -nE "^[[:space:]]+environment:" "$1"' _ "$file"
	check "no \${VAR} substitution in the compose file" \
		bash -c '! grep -v "^[[:space:]]*#" "$1" | grep -n "\${"' _ "$file"
	local images
	images=$(env COMPOSE_FILE="$TMP/prod/docker-compose.yml" docker compose --project-directory "$TMP/prod" \
		--profile oneoff config --images | grep '^henga/' | sort -u | tr '\n' ' ')
	check "app images are the local :prod pins (got: $images)" \
		[ "$images" = "henga/leviosa-backend:prod henga/leviosa-frontend:prod " ]
	local services
	services=$(env COMPOSE_FILE="$TMP/prod/docker-compose.yml" docker compose --project-directory "$TMP/prod" \
		config --services | sort | tr '\n' ' ')
	check "a plain up starts no one-off (services: $services)" \
		[ "$services" = "backend caddy frontend postgres rabbitmq redis " ]
	local owners
	owners=$(compose_lines_by_service "$file" | awk '/env\/admin\.env/ { print $1 }' | tr '\n' ' ')
	check "only the admin one-off loads env/admin.env (loaded by: $owners)" [ "$owners" = "admin " ]
	owners=$(compose_lines_by_service "$file" | awk '$2 == "ports:" { print $1 }' | tr '\n' ' ')
	check "only caddy publishes host ports (published by: $owners)" [ "$owners" = "caddy " ]
	check "postgres, rabbitmq and redis data are bind-mounted under ./data" \
		bash -c 'grep -q "\./data/postgres:/var/lib/postgresql/data" "$1" &&
			grep -q "\./data/rabbitmq:/var/lib/rabbitmq" "$1" && grep -q "\./data/redis:/data" "$1"' _ "$file"
	check "deploy refuses a missing argument" bash -c '! "$1" prod >/dev/null 2>&1' _ "$DEPLOY"
	check "deploy refuses a non-SHA (a tag name)" bash -c '! "$1" prod prod >/dev/null 2>&1' _ "$DEPLOY"
	check "deploy refuses an unknown stack" bash -c '! "$1" nope "$2" >/dev/null 2>&1' _ "$DEPLOY" "$SHA1"

	# The staging file, under the same rules plus the prefix rule.
	local staging=$COMPOSE_DIR/docker-compose.staging.yml
	local sdc=(env COMPOSE_FILE="$TMP/staging/docker-compose.yml" docker compose --project-directory "$TMP/staging")
	check "docker-compose.staging.yml validates with the staging fixture env files" "${sdc[@]}" config -q
	check "staging: no inline environment, no \${VAR} substitution" \
		bash -c '! grep -nE "^[[:space:]]+environment:" "$1" && ! grep -v "^[[:space:]]*#" "$1" | grep -n "\${"' _ "$staging"
	images=$("${sdc[@]}" --profile oneoff config --images | grep '^henga/' | sort -u | tr '\n' ' ')
	check "staging app images are the local :staging pins (got: $images)" \
		[ "$images" = "henga/leviosa-backend:staging henga/leviosa-frontend:staging " ]
	services=$("${sdc[@]}" config --services | sort | tr '\n' ' ')
	check "staging: a plain up starts backend, frontend and its own Redis only — no postgres, rabbitmq or caddy (services: $services)" \
		[ "$services" = "staging_backend staging_frontend staging_redis " ]
	local prod_all staging_all
	prod_all=$(env COMPOSE_FILE="$TMP/prod/docker-compose.yml" docker compose --project-directory "$TMP/prod" \
		--profile oneoff config --services | sort)
	staging_all=$("${sdc[@]}" --profile oneoff config --services | sort)
	check "every staging service, one-offs included, is staging_-prefixed ($(tr '\n' ' ' <<<"$staging_all"))" \
		bash -c '! grep -qv "^staging_" <<<"$1"' _ "$staging_all"
	check "no staging service shares a name (a DNS alias) with a production service" \
		[ -z "$(comm -12 <(echo "$prod_all") <(echo "$staging_all"))" ]
	local names
	names="$(env COMPOSE_FILE="$TMP/prod/docker-compose.yml" docker compose --project-directory "$TMP/prod" config |
		sed -n 's/^name: //p') $("${sdc[@]}" config | sed -n 's/^name: //p')"
	check "the two files declare distinct compose projects (got: $names)" [ "$names" = "leviosa leviosa-staging" ]
	owners=$(compose_lines_by_service "$staging" | awk '/env\/admin\.env/ { print $1 }' | tr '\n' ' ')
	check "staging: only the staging_admin one-off loads env/admin.env (loaded by: $owners)" [ "$owners" = "staging_admin " ]
	owners=$(compose_lines_by_service "$staging" | awk '$2 == "ports:" { print $1 }' | tr '\n' ' ')
	check "staging: no service publishes a host port (published by: ${owners:-none})" [ -z "$owners" ]
	check "staging: Redis data is bind-mounted under its own ./data/redis" \
		grep -q '\./data/redis:/data' "$staging"
	check "deploy refuses a non-SHA for staging (a tag name)" bash -c '! "$1" staging staging >/dev/null 2>&1' _ "$DEPLOY"
}

case_first_deploy() {
	section "criterion 1: first deploy migrates, bootstraps the admin, both apps report the SHA"
	if deploy_prod "$SHA1" deploy-1; then
		ok "deploy prod $SHA1 exits 0"
	else
		show_log "$TMP/logs/deploy-1.log"
		fatal "deploy prod $SHA1 failed"
	fi
	check "the migrate one-off applied the migrations" log_has deploy-1 '"msg":"database migrated"'
	local newest
	newest=$(newest_migration)
	check "the schema is at the newest embedded migration ($newest)" [ "$(schema_version)" = "$newest" ]
	check "the admin one-off created the superadmin from admin.env" log_has deploy-1 'superadmin created'
	check "exactly one account exists (the admin)" [ "$(user_count)" = 1 ]
	check "backend healthy" [ "$(health "$(cid backend)")" = healthy ]
	check "frontend healthy" [ "$(health "$(cid frontend)")" = healthy ]
	check "/healthz reports $SHA1 for both apps (got: $(healthz_shas))" serving "$SHA1"
	check "pin $BACKEND_REPO:prod is $SHA1's image" [ "$(image_id "$BACKEND_REPO:prod")" = "$(image_id "$BACKEND_REPO:$SHA1")" ]
	check "pin $FRONTEND_REPO:prod is $SHA1's image" [ "$(image_id "$FRONTEND_REPO:prod")" = "$(image_id "$FRONTEND_REPO:$SHA1")" ]
	check "the one-off containers are gone" [ -z "$(oneoff_containers)" ]
	check "the backend is connected to RabbitMQ as leviosa on vhost leviosa" \
		bash -c '[[ $(cat) == *"leviosa"*"leviosa"* ]]' < <(rabbitmqctl_q list_connections user vhost 2>/dev/null)
	check "no running container publishes a host port" \
		[ -z "$(docker ps --filter "label=com.docker.compose.project=$PROJECT" --format '{{.Ports}}' | grep -- '->' || true)" ]
}

case_second_deploy() {
	section "criterion 2: a second deploy keeps the data and leaves the admin untouched"
	psql_q "CREATE TABLE harness_marker (note text); INSERT INTO harness_marker VALUES ('before the second deploy')" >/dev/null
	local fingerprint backend_before deps_before
	fingerprint=$(admin_fingerprint)
	backend_before=$(cid backend)
	deps_before="$(cid postgres) $(cid rabbitmq) $(cid redis)"
	# Different values for the same email: an existing account must not change.
	sed -i -e 's/^ADMIN_PASSWORD=.*/ADMIN_PASSWORD=HarnessChanged5678/' \
		-e 's/^ADMIN_FIRST_NAME=.*/ADMIN_FIRST_NAME=Changed/' "$TMP/prod/env/admin.env"
	if deploy_prod "$SHA2" deploy-2; then
		ok "deploy prod $SHA2 exits 0"
	else
		show_log "$TMP/logs/deploy-2.log"
		fatal "deploy prod $SHA2 failed"
	fi
	check "/healthz reports $SHA2 for both apps (got: $(healthz_shas))" serving "$SHA2"
	check "the backend container was recreated" [ "$(cid backend)" != "$backend_before" ]
	check "postgres, rabbitmq and redis were left running (not recreated)" \
		[ "$(cid postgres) $(cid rabbitmq) $(cid redis)" = "$deps_before" ]
	check "data written before the deploy is still there" \
		[ "$(psql_q 'SELECT note FROM harness_marker')" = "before the second deploy" ]
	check "the bootstrap reported the account as existing" log_has deploy-2 'already exists'
	check "still exactly one account" [ "$(user_count)" = 1 ]
	check "the admin account is byte-for-byte unchanged (password, name, role, timestamps)" \
		[ "$(admin_fingerprint)" = "$fingerprint" ]
}

case_failing_migration() {
	section "criterion 3: a failing migration aborts before recreate; the previous containers keep serving"
	local backend_before frontend_before version_before
	backend_before=$(cid backend)
	frontend_before=$(cid frontend)
	version_before=$(schema_version)
	if deploy_prod "$SHA_MIGFAIL" deploy-migfail; then
		bad "deploy prod $SHA_MIGFAIL exited 0 despite a broken migration"
	else
		ok "deploy prod $SHA_MIGFAIL exits non-zero"
	fi
	check "the deploy stopped at the migrate step" log_has deploy-migfail 'migration failed: NOT recreating'
	check "the backend named the failing migration" log_has deploy-migfail "$BROKEN_MIGRATION"
	check "the admin one-off never ran" bash -c '! grep -q "running the admin bootstrap" "$1"' _ "$TMP/logs/deploy-migfail.log"
	check "the backend container was not recreated" [ "$(cid backend)" = "$backend_before" ]
	check "the frontend container was not recreated" [ "$(cid frontend)" = "$frontend_before" ]
	check "/healthz still reports the previous release $SHA2 (got: $(healthz_shas))" serving "$SHA2"
	check "the schema version is unchanged ($version_before)" [ "$(schema_version)" = "$version_before" ]
	check "pin $BACKEND_REPO:prod moved back to the running backend's image" \
		[ "$(image_id "$BACKEND_REPO:prod")" = "$(container_image "$(cid backend)")" ]
	check "pin $FRONTEND_REPO:prod moved back to the running frontend's image" \
		[ "$(image_id "$FRONTEND_REPO:prod")" = "$(container_image "$(cid frontend)")" ]
	check "the one-off containers are gone" [ -z "$(oneoff_containers)" ]
}

case_sha_mismatch() {
	section "criterion 4: a SHA mismatch on /healthz fails the deploy"
	if deploy_prod "$SHA_BAD_BACKEND" deploy-bad-backend; then
		bad "deploy exited 0 while the backend reports another SHA"
	else
		ok "deploy exits non-zero when the backend reports another SHA"
	fi
	check "the failure names the /healthz check" log_has deploy-bad-backend "/healthz does not report $SHA_BAD_BACKEND"
	if deploy_prod "$SHA_BAD_FRONTEND" deploy-bad-frontend; then
		bad "deploy exited 0 while the frontend reports another SHA"
	else
		ok "deploy exits non-zero when the frontend reports another SHA"
	fi
	check "the failure names the /healthz check" log_has deploy-bad-frontend "/healthz does not report $SHA_BAD_FRONTEND"
	# Recovering is deploying a good release: the rollback path.
	if deploy_prod "$SHA2" deploy-recover; then
		ok "deploying $SHA2 again recovers the stack (exit 0)"
	else
		show_log "$TMP/logs/deploy-recover.log"
		fatal "deploy prod $SHA2 (recovery) failed"
	fi
	check "/healthz reports $SHA2 for both apps again (got: $(healthz_shas))" serving "$SHA2"
}

case_no_admin_env() {
	section "criterion 5: no ADMIN_* value in the running backend or frontend"
	local service env
	for service in backend frontend; do
		env=$(container_env "$(cid "$service")")
		if [ -z "$env" ]; then
			bad "could not read the $service container's environment"
		elif grep -q '^ADMIN_' <<<"$env"; then
			bad "$service has ADMIN_* values: $(grep -o '^ADMIN_[A-Z_]*' <<<"$env" | tr '\n' ' ')"
		else
			ok "$service has no ADMIN_* value ($(wc -l <<<"$env") variables inspected)"
		fi
	done
	env=$(container_env "$(cid frontend)")
	check "frontend has no backend secret either (database, Redis, RabbitMQ, encx, AWS, Stripe, Twilio, SMTP, session)" \
		bash -c '! grep -qE "^(DATABASE_URL|REDIS_|RABBITMQ_|ENCX_|AWS_|S3_|STRIPE_|TWILIO_|SMTP_|SESSION_SECRET|BOOKING_TOKEN_SECRET)=" <<<"$1"' _ "$env"
}

case_data_survives() {
	section "criterion 6: data survives container recreation"
	local fingerprint before
	fingerprint=$(admin_fingerprint)
	redis_cli SET harness:marker survived >/dev/null
	rabbitmqctl_q add_vhost harness-marker >/dev/null
	before="$(cid postgres) $(cid rabbitmq) $(cid redis) $(cid backend) $(cid frontend)"
	local service
	for service in postgres rabbitmq redis; do
		check "$service stores its data in a bind mount (no Docker volume)" mounts_are_binds "$(cid "$service")"
	done
	check "docker compose down removes every container" bash -c '"$@" >"$0" 2>&1' "$TMP/logs/down.log" \
		env COMPOSE_FILE="$(compose_files)" COMPOSE_PROJECT_NAME="$PROJECT" \
		docker compose --project-directory "$TMP/prod" down --timeout 30
	check "no container of the stack is left" [ -z "$(docker ps -aq --filter "label=com.docker.compose.project=$PROJECT")" ]
	if deploy_prod "$SHA2" deploy-after-down; then
		ok "deploy prod $SHA2 on the existing data exits 0"
	else
		show_log "$TMP/logs/deploy-after-down.log"
		fatal "deploy prod $SHA2 after down failed"
	fi
	check "every container is new" bash -c '
		for id in $2; do case " $1 " in *" $id "*) exit 1 ;; esac; done' _ \
		"$before" "$(cid postgres) $(cid rabbitmq) $(cid redis) $(cid backend) $(cid frontend)"
	check "Postgres: rows written earlier are there" [ "$(psql_q 'SELECT note FROM harness_marker')" = "before the second deploy" ]
	check "Postgres: the admin account is unchanged" [ "$(admin_fingerprint)" = "$fingerprint" ]
	check "Redis: the key written earlier is there" [ "$(redis_cli GET harness:marker)" = survived ]
	check "RabbitMQ: the vhost created earlier is there" \
		bash -c 'grep -qx harness-marker' < <(rabbitmqctl_q list_vhosts name 2>/dev/null)
	check "/healthz reports $SHA2 for both apps (got: $(healthz_shas))" serving "$SHA2"
}

case_caddy() {
	section "extra: Caddy (minimal harness Caddyfile) reaches the frontend from a random loopback port"
	if ! dc up -d --no-deps --wait --wait-timeout 60 caddy >"$TMP/logs/caddy.log" 2>&1; then
		show_log "$TMP/logs/caddy.log"
		bad "caddy did not start"
		return
	fi
	local address body
	address=$(dc port caddy 80)
	CADDY_ADDRESS=$address
	note "caddy published on $address"
	check "caddy is published on loopback only" [ "${address%%:*}" = 127.0.0.1 ]
	body=$(curl -fsS --max-time 10 "http://$address/healthz" 2>&1 || true)
	check "GET /healthz through caddy reports $SHA2 for both apps ($body)" \
		[ "$body" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]
}

# ── staging cases (issue 019) ────────────────────────────────────────────────
case_staging_before_prod() {
	section "staging: without production's Postgres and RabbitMQ, a staging deploy fails at migrate"
	if deploy_staging "$SHA1" deploy-staging-no-prod; then
		bad "deploy staging $SHA1 exited 0 while production's services do not exist"
	else
		ok "deploy staging $SHA1 exits non-zero"
	fi
	check "the deploy stopped at the migrate step" log_has deploy-staging-no-prod 'migration failed: NOT recreating'
	check "the migrate one-off could not reach the database" log_has deploy-staging-no-prod 'ping database'
	check "the error points at production's services" log_has deploy-staging-no-prod "production's postgres and rabbitmq"
	check "the admin one-off never ran" \
		bash -c '! grep -q "running the admin bootstrap" "$1"' _ "$TMP/logs/deploy-staging-no-prod.log"
	check "no production container was created or started" \
		[ -z "$(docker ps -aq --filter "label=com.docker.compose.project=$PROJECT")" ]
	local running
	running=$(dcs ps --services | sort | tr '\n' ' ')
	check "only staging_redis runs, in staging's own project (running: ${running:-none})" [ "$running" = "staging_redis " ]
	check "no staging pin is left behind (a first deploy has nothing to restore)" \
		bash -c '! docker image inspect "$1" >/dev/null 2>&1 && ! docker image inspect "$2" >/dev/null 2>&1' _ \
		"$BACKEND_REPO:staging" "$FRONTEND_REPO:staging"
}

case_staging_services() {
	section "staging setup: role, database, vhost and user on production's services (create-staging-services.sh, as Ansible will)"
	if create_staging_services staging-services-1; then
		ok "create-staging-services.sh exits 0"
	else
		show_log "$TMP/logs/staging-services-1.log"
		fatal "create-staging-services.sh failed"
	fi
	if create_staging_services staging-services-2; then
		ok "a second run exits 0 as well (idempotent)"
	else
		show_log "$TMP/logs/staging-services-2.log"
		bad "the second create-staging-services.sh run failed"
	fi
	local db mq role sdb mq_user mq_vhost prod_db
	db=$(staging_db)
	mq=$(staging_mq)
	role=$(field 1 <<<"$db")
	sdb=$(field 4 <<<"$db")
	mq_user=$(field 1 <<<"$mq")
	mq_vhost=$(field 4 <<<"$mq")
	prod_db=$(env_value "$TMP/prod/env/postgres.env" POSTGRES_DB)
	check "the staging role $role is no superuser and cannot create roles or databases" \
		[ "$(psql_q "SELECT rolsuper OR rolcreaterole OR rolcreatedb OR rolreplication OR rolbypassrls
			FROM pg_roles WHERE rolname = '$role'")" = f ]
	check "the staging database $sdb is owned by $role" \
		[ "$(psql_q "SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = '$sdb'")" = "$role" ]
	check "$role holds no CONNECT privilege on production's database $prod_db" \
		[ "$(psql_q "SELECT has_database_privilege('$role', '$prod_db', 'CONNECT')")" = f ]
	check "the RabbitMQ user $mq_user has permissions on vhost $mq_vhost only" \
		[ "$(rabbitmqctl_q list_user_permissions --no-table-headers "$mq_user" | cut -f1 | tr '\n' ' ')" = "$mq_vhost " ]
	check "the RabbitMQ user $mq_user has no tags (no management access)" \
		[ "$(rabbitmqctl_q list_users --no-table-headers | awk -v user="$mq_user" '$1 == user { print $2 }')" = "[]" ]
	check "the script's output holds no password" \
		bash -c '! grep -qF -e "$2" -e "$3" "$1"' _ "$TMP/logs/staging-services-1.log" "$(field 2 <<<"$db")" "$(field 2 <<<"$mq")"
}

case_staging_side_by_side() {
	section "staging criterion 1: production and staging side by side, each /healthz reports its own SHA"
	local db mq sdb pins_before fingerprint
	db=$(staging_db)
	mq=$(staging_mq)
	sdb=$(field 4 <<<"$db")
	PROD_BEFORE_STAGING=$(prod_state)
	pins_before=$(pins_state)
	fingerprint=$(admin_fingerprint)
	note "production serves $SHA2; deploying staging at $SHA1"
	if deploy_staging "$SHA1" deploy-staging-1; then
		ok "deploy staging $SHA1 exits 0"
	else
		show_log "$TMP/logs/deploy-staging-1.log"
		fatal "deploy staging $SHA1 failed"
	fi
	PROD_AFTER_FIRST_STAGING=$(prod_state)
	check "staging /healthz reports $SHA1 for both apps (got: $(healthz_shas staging_frontend))" serving "$SHA1" staging_frontend
	check "production /healthz still reports $SHA2 for both apps (got: $(healthz_shas))" serving "$SHA2"
	check "staging_backend healthy" [ "$(health "$(scid staging_backend)")" = healthy ]
	check "staging_frontend healthy" [ "$(health "$(scid staging_frontend)")" = healthy ]
	check "pin $BACKEND_REPO:staging is $SHA1's image" \
		[ "$(image_id "$BACKEND_REPO:staging")" = "$(image_id "$BACKEND_REPO:$SHA1")" ]
	check "pin $FRONTEND_REPO:staging is $SHA1's image" \
		[ "$(image_id "$FRONTEND_REPO:staging")" = "$(image_id "$FRONTEND_REPO:$SHA1")" ]
	check "production's :prod pins did not move" [ "$(pins_state)" = "$pins_before" ]
	check "the staging_migrate one-off migrated staging's own database" log_has deploy-staging-1 '"msg":"database migrated"'
	check "the staging database is at the newest embedded migration" \
		[ "$(psql_db "$sdb" 'SELECT max(version_id) FROM goose_db_version')" = "$(newest_migration)" ]
	check "the staging_admin one-off created staging's superadmin" log_has deploy-staging-1 'superadmin created'
	check "the staging database holds exactly that one account" [ "$(psql_db "$sdb" 'SELECT count(*) FROM auth.users')" = 1 ]
	check "production's accounts are untouched (one, byte-for-byte)" [ "$(admin_fingerprint)" = "$fingerprint" ]
	check "staging_backend is connected to RabbitMQ as $(field 1 <<<"$mq") on vhost $(field 4 <<<"$mq")" \
		bash -c 'grep -qxF "$1"' _ "$(field 1 <<<"$mq")	$(field 4 <<<"$mq")" \
		< <(rabbitmqctl_q list_connections --no-table-headers user vhost 2>/dev/null)
	check "the staging one-off containers are gone" \
		[ -z "$(docker ps -aq --filter "label=com.docker.compose.project=$PROJECT_STAGING" --filter label=com.docker.compose.oneoff=True)" ]
	check "no staging container publishes a host port" \
		[ -z "$(docker ps --filter "label=com.docker.compose.project=$PROJECT_STAGING" --format '{{.Ports}}' | grep -- '->' || true)" ]
	if [ -n "$CADDY_ADDRESS" ]; then
		local body
		body=$(curl -fsS --max-time 10 -H "Host: $STAGING_HOST" "http://$CADDY_ADDRESS/healthz" 2>&1 || true)
		check "through Caddy, $STAGING_HOST reaches staging_frontend by its service name ($body)" \
			[ "$body" = "{\"frontend\":\"$SHA1\",\"backend\":\"$SHA1\"}" ]
		body=$(curl -fsS --max-time 10 "http://$CADDY_ADDRESS/healthz" 2>&1 || true)
		check "through Caddy, production's hosts still reach production ($body)" \
			[ "$body" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]
	fi
}

case_staging_aliases() {
	section "staging criterion 2: production's frontend only reaches production's backend (no alias collision)"
	local prod_frontend staging_frontend name want got services
	prod_frontend=$(cid frontend)
	staging_frontend=$(scid staging_frontend)
	services="backend frontend redis postgres rabbitmq"
	if [ -n "$(cid caddy)" ]; then services="$services caddy"; fi
	for name in $services; do
		want=$(container_ip "$(cid "$name")")
		got=$(resolve_from "$prod_frontend" "$name")
		check "from production's frontend, '$name' resolves to production's $name only ($got)" \
			[ "$got" = "$want" ]
	done
	for name in staging_backend staging_frontend staging_redis; do
		want=$(container_ip "$(scid "$name")")
		got=$(resolve_from "$staging_frontend" "$name")
		check "from staging's frontend, '$name' resolves to staging's container ($got)" \
			bash -c '[ -n "$1" ] && [ "$1" = "$2" ]' _ "$want" "$got"
	done
	got=$(shared_dns_names)
	check "no DNS name on the shared network answers for two containers (shared: ${got:-none})" [ -z "$got" ]
	got=$(healthz_repeated frontend 20)
	check "20 consecutive /healthz through production's frontend all report $SHA2 ($got)" \
		[ "$got" = "20 $SHA2 $SHA2" ]
	got=$(healthz_repeated staging_frontend 20)
	check "20 consecutive /healthz through staging's frontend all report $SHA1 ($got)" \
		[ "$got" = "20 $SHA1 $SHA1" ]
}

case_staging_postgres() {
	section "staging criterion 3: the staging Postgres role cannot connect to production's database"
	local db role password prod_db got
	db=$(staging_db)
	role=$(field 1 <<<"$db")
	password=$(field 2 <<<"$db")
	prod_db=$(env_value "$TMP/prod/env/postgres.env" POSTGRES_DB)
	got=$(pg_connect_as "$role" "$password" "$(field 4 <<<"$db")")
	check "control: $role connects to its own database $(field 4 <<<"$db") with its password (got: $got)" \
		[ "$got" = "$role" ]
	got=$(pg_connect_as "$role" "$password" "$prod_db")
	check "$role is refused on production's database $prod_db ($(head -n 1 <<<"$got"))" \
		bash -c '[ "$1" != "$2" ] && [[ $1 == *"permission denied for database"* ]]' _ "$got" "$role"
	check "... for lack of the CONNECT privilege, not a bad password" \
		grep -q 'does not have CONNECT privilege' <<<"$got"
}

case_staging_rabbitmq() {
	section "staging criterion 4: the staging RabbitMQ user cannot access production's vhost"
	local mq user password vhost prod_vhost got
	mq=$(staging_mq)
	user=$(field 1 <<<"$mq")
	password=$(field 2 <<<"$mq")
	vhost=$(field 4 <<<"$mq")
	prod_vhost=$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_VHOST)
	got=$(amqp_open "$user" "$password" "$vhost")
	check "control: over AMQP, $user logs in and opens its own vhost $vhost (got: $got)" [ "$got" = open-ok ]
	got=$(amqp_open "$user" "$password" "$prod_vhost")
	check "over AMQP, $user logs in but is refused production's vhost $prod_vhost (got: $got)" \
		bash -c '[[ $1 == "refused 530 NOT_ALLOWED"* ]]' _ "$got"
	check "$user holds no permission on any vhost but $vhost" \
		[ "$(rabbitmqctl_q list_user_permissions --no-table-headers "$user" | cut -f1 | tr '\n' ' ')" = "$vhost " ]
}

case_staging_credentials() {
	section "staging criterion 5: staging's env files hold no production Redis, Postgres or RabbitMQ credential"
	local file name service leaks="" files="" got db mq prod_env=$TMP/prod/env
	for file in "$TMP/staging/env/"*.env; do
		name=${file##*/}
		files+="$name "
		leaks+=$(leaked_keys "$name" <"$file")
	done
	check "no value in ${files% } contains one of production's $(prod_secrets | wc -l) Redis/Postgres/RabbitMQ passwords (leaking keys: ${leaks:-none})" \
		[ -z "$leaks" ]
	for service in staging_backend staging_frontend staging_redis; do
		leaks+=$(container_env "$(scid "$service")" | leaked_keys "$service")
	done
	check "nor does the environment of a running staging container (leaking keys: ${leaks:-none})" [ -z "$leaks" ]
	db=$(staging_db)
	mq=$(staging_mq)
	check "staging's DATABASE_URL uses its own role and database ($(field 1 <<<"$db")/$(field 4 <<<"$db")), not production's" \
		bash -c '[ "$1" != "$3" ] && [ "$2" != "$4" ]' _ "$(field 1 <<<"$db")" "$(field 4 <<<"$db")" \
		"$(env_value "$prod_env/postgres.env" POSTGRES_USER)" "$(env_value "$prod_env/postgres.env" POSTGRES_DB)"
	check "staging's RABBITMQ_URL uses its own user and vhost ($(field 1 <<<"$mq")/$(field 4 <<<"$mq")), not production's" \
		bash -c '[ "$1" != "$3" ] && [ "$2" != "$4" ]' _ "$(field 1 <<<"$mq")" "$(field 4 <<<"$mq")" \
		"$(env_value "$prod_env/rabbitmq.env" RABBITMQ_DEFAULT_USER)" "$(env_value "$prod_env/rabbitmq.env" RABBITMQ_DEFAULT_VHOST)"
	got=$(env_value "$TMP/staging/env/backend.env" REDIS_ADDR)
	check "staging's REDIS_ADDR is its own Redis ($got)" [ "$got" = staging_redis:6379 ]
	got=$(dcs exec -T staging_redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h staging_redis ping 2>&1')
	check "control: staging's Redis password opens staging_redis ($got)" [ "$got" = PONG ]
	got=$(dcs exec -T staging_redis sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h redis ping 2>&1' |
		grep -v '^[[:space:]]*$' | paste -sd ' ' -)
	check "production's Redis refuses staging's password ($got)" \
		bash -c '[[ $1 != *PONG* ]] && [[ $1 == *WRONGPASS* || $1 == *NOAUTH* ]]' _ "$got"
	for service in staging_backend staging_frontend; do
		got=$(container_env "$(scid "$service")")
		check "$service has no ADMIN_* value" bash -c '[ -n "$1" ] && ! grep -q "^ADMIN_" <<<"$1"' _ "$got"
	done
	check "staging_frontend has no backend secret either" \
		bash -c '! grep -qE "^(DATABASE_URL|REDIS_|RABBITMQ_|ENCX_|AWS_|S3_|STRIPE_|TWILIO_|SMTP_|SESSION_SECRET|BOOKING_TOKEN_SECRET)=" <<<"$1"' _ "$got"
}

# $1 and $2: two prod_state snapshots; prints their differences when they differ.
same_prod_state() {
	[ -n "$1" ] && [ "$1" = "$2" ] && return 0
	diff <(echo "$1") <(echo "$2") | sed 's/^/           | /'
	return 1
}

case_staging_no_prod_restart() {
	section "staging criterion 6: a staging deploy restarts no production container"
	local count backend_before pins_before
	count=$(grep -c . <<<"$PROD_BEFORE_STAGING")
	check "first staging deploy: all $count production containers kept their id, start time and restart count" \
		same_prod_state "$PROD_BEFORE_STAGING" "$PROD_AFTER_FIRST_STAGING"
	backend_before=$(scid staging_backend)
	pins_before=$(pins_state)
	if deploy_staging "$SHA1" deploy-staging-2; then
		ok "a second deploy staging $SHA1 exits 0"
	else
		show_log "$TMP/logs/deploy-staging-2.log"
		bad "the second deploy staging $SHA1 failed"
		return
	fi
	check "it recreated staging_backend" [ "$(scid staging_backend)" != "$backend_before" ]
	check "second staging deploy: all $count production containers still unchanged" \
		same_prod_state "$PROD_BEFORE_STAGING" "$(prod_state)"
	check "production's :prod pins did not move" [ "$(pins_state)" = "$pins_before" ]
	check "staging /healthz still reports $SHA1, production's $SHA2" \
		bash -c '[ "$1" = "$2 $2" ] && [ "$3" = "$4 $4" ]' _ \
		"$(healthz_shas staging_frontend)" "$SHA1" "$(healthz_shas)" "$SHA2"
}

case_teardown() {
	section "criterion 7: the harness leaves nothing behind"
	local volumes id
	volumes=$(for id in $(harness_containers); do
		docker inspect -f '{{range .Mounts}}{{if eq .Type "volume"}}{{.Name}}{{"\n"}}{{end}}{{end}}' "$id"
	done | sort -u)
	local tmp=$TMP
	teardown
	check "no container left" [ -z "$(harness_containers)" ]
	check "network $NETWORK removed" bash -c '! docker network inspect "$1" >/dev/null 2>&1' _ "$NETWORK"
	check "no anonymous volume left ($(wc -w <<<"$volumes") seen during the run)" \
		bash -c 'for v in $1; do docker volume inspect "$v" >/dev/null 2>&1 && exit 1; done; exit 0' _ "$volumes"
	check "no harness image left (repositories $BACKEND_REPO, $FRONTEND_REPO)" [ -z "$(harness_images)" ]
	if [ "${HARNESS_KEEP_PULLED:-0}" = 1 ]; then
		note "HARNESS_KEEP_PULLED=1: kept pulled public images: ${PULLED[*]:-none}"
	else
		check "public images pulled by this run removed (${PULLED[*]:-none})" \
			bash -c 'for i in "$@"; do docker image inspect "$i" >/dev/null 2>&1 && exit 1; done; exit 0' _ "${PULLED[@]}"
	fi
	check "temp dir removed" [ ! -e "$tmp" ]
}

summary() {
	SUMMARY_DONE=1
	local s
	printf '\n== summary (run %s) ==\n' "$RUN_ID"
	for s in "${SECTIONS[@]}"; do
		if [ "${SECTION_FAILS[$s]}" -eq 0 ]; then
			printf '    PASS  %s\n' "$s"
		else
			printf '    FAIL  %s (%s failed)\n' "$s" "${SECTION_FAILS[$s]}"
		fi
	done
	printf '    %s checks passed, %s failed\n' "$TOTAL_PASS" "$TOTAL_FAIL"
	if [ "$TOTAL_FAIL" -ne 0 ]; then
		echo "RESULT: FAILED"
		return 1
	fi
	echo "RESULT: OK"
}

# ── main ─────────────────────────────────────────────────────────────────────
echo "leviosa deploy harness, run $RUN_ID"
setup
case_static
build_images
case_staging_before_prod
case_first_deploy
case_second_deploy
case_failing_migration
case_sha_mismatch
case_no_admin_env
case_data_survives
case_caddy
case_staging_services
case_staging_side_by_side
case_staging_aliases
case_staging_postgres
case_staging_rabbitmq
case_staging_credentials
case_staging_no_prod_restart
if [ "${KEEP:-0}" = 1 ]; then
	note "KEEP=1: teardown skipped"
else
	case_teardown
fi
summary
