#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# Local Docker harness for docker-compose.prod.yml, docker-compose.staging.yml,
# `deploy` and caddy/Caddyfile (issues 018, 019 and 020).
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
#   staging beside production (019) — after ../setup-shared-services.sh, the
#   script Ansible runs on the server (issue 026), created staging's Postgres
#   role and database and each environment's RabbitMQ vhost and user on
#   production's services (idempotent: a second run reports no change;
#   restarts nothing; no guest user; each user on its own vhost only)
#   S1. both stacks deployed with different SHAs, each /healthz reports its own
#   S2. production's frontend only reaches production's backend (no alias
#       collision on the shared network)
#   S3. the staging Postgres role cannot connect to production's database
#   S4. the staging RabbitMQ user cannot access production's vhost
#   S5. staging's env files hold no production Redis/Postgres/RabbitMQ
#       credential
#   S6. a staging deploy restarts no production container
#   (and, before production exists, a staging deploy fails at migrate)
#   Caddy (020) — the real caddy/Caddyfile, with Caddy's internal CA in place
#   of the Cloudflare DNS challenge (LEVIOSA_CADDY_TLS=internal); requests go
#   to its random loopback ports with each real host name as SNI and Host
#   C1. POST /webhooks/stripe and /webhooks/stripe/connect on each stack reach
#       that stack's backend (its signature error), staging's without auth
#   C2. any staging page without credentials is a 401; with them, the page
#   C3. staging /healthz is a 200 without credentials
#   C4. staging responses carry X-Robots-Tag: noindex, production's don't
#   C5. requests sent while the backend (or the frontend) is being recreated
#       get no 502
#   (and the production variant, Cloudflare DNS challenge, validates offline)
#   7. nothing left behind: containers, networks, volumes, images, temp files
#
# Isolation: every name carries a random run id (leviosa-harness-<id>): its
# own compose projects (<name>-prod, <name>-staging), network, LocalStack
# container and image repositories (<name>-backend / <name>-frontend, so the
# henga/leviosa-* pins are never touched). Nothing binds a fixed host port
# (Caddy gets random loopback ports) and nothing resolves the real host names
# (curl --connect-to). Everything on disk lives under one
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
SETUP_SHARED_SERVICES=$COMPOSE_DIR/setup-shared-services.sh

LOCALSTACK_IMAGE=localstack/localstack:4.6
AWS_REGION=eu-central-1             # as in fixtures/{prod,staging}/env/backend.env
KMS_ALIAS=alias/leviosa-harness     # ENCX_KMS_KEY_ALIAS in the prod file
KMS_ALIAS_STAGING=alias/leviosa-harness-staging # ... and in the staging file
BROKEN_MIGRATION=29991231235959_harness_broken_migration.sql

# caddy/Caddyfile's sites, and the hosts it no longer serves.
CADDYFILE=$COMPOSE_DIR/caddy/Caddyfile
CADDY_IMAGE=serfriz/caddy-cloudflare:2.11 # docker-compose.prod.yml's caddy image
PROD_HOSTS="leviosa.care www.leviosa.care admin.leviosa.care staff.leviosa.care"
STAGING_HOSTS="staging.leviosa.care admin-staging.leviosa.care staff-staging.leviosa.care"
DROPPED_HOSTS="analytics.leviosa.care www.staging.leviosa.care"
STAGING_HOST=staging.leviosa.care
# The plaintext of fixtures/prod/env/caddy.env's STAGING_BASIC_AUTH_HASH.
STAGING_AUTH_PASSWORD=harness-staging-password
# Stripe's webhook routes. The fixtures' webhook secrets are empty, so no
# signature ever verifies: the backend answers each POST with this error.
WEBHOOK_PATHS="/webhooks/stripe /webhooks/stripe/connect"
WEBHOOK_REJECTED='{"error":"webhook signature verification failed"}'

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
CADDY_HTTPS=""    # Caddy's random loopback address for :443, once it runs
STREAM_PID=""    # a background request stream through Caddy, while one runs
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

# ../setup-shared-services.sh with the values Ansible reads from Infisical
# (prod/db, staging/db, prod/rabbitmq, staging/rabbitmq). Extra VAR=value
# arguments override them (a password rotation).
setup_shared_services() { # $1 log name, [VAR=value...]
	local db mq log=$1
	shift
	db=$(staging_db) && mq=$(staging_mq) || return 1
	env PROD_DB_USER="$(env_value "$TMP/prod/env/postgres.env" POSTGRES_USER)" \
		PROD_DB_NAME="$(env_value "$TMP/prod/env/postgres.env" POSTGRES_DB)" \
		STAGING_DB_USER="$(field 1 <<<"$db")" STAGING_DB_PASSWORD="$(field 2 <<<"$db")" \
		STAGING_DB_NAME="$(field 4 <<<"$db")" \
		PROD_MQ_USER="$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_USER)" \
		PROD_MQ_PASSWORD="$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_PASS)" \
		PROD_MQ_VHOST="$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_VHOST)" \
		STAGING_MQ_USER="$(field 1 <<<"$mq")" STAGING_MQ_PASSWORD="$(field 2 <<<"$mq")" \
		STAGING_MQ_VHOST="$(field 4 <<<"$mq")" \
		"$@" "$SETUP_SHARED_SERVICES" "$(cid postgres)" "$(cid rabbitmq)" >"$TMP/logs/$log.log" 2>&1
}
# Postgres and RabbitMQ: container id, start time, restart count.
shared_services_state() {
	docker inspect -f '{{.Id}} {{.State.StartedAt}} restarts={{.RestartCount}}' "$(cid postgres)" "$(cid rabbitmq)"
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
# through the environment, never a command line. To the container's own
# address, not 127.0.0.1: the image trusts loopback without any password.
pg_connect_as() {
	env PGPASSWORD="$2" docker exec -e PGPASSWORD "$(cid postgres)" sh -c \
		'exec psql -X -h "$(hostname)" -U "$1" -d "$2" -tAc "SELECT current_user"' sh "$1" "$3" 2>&1 || true
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

# ── Caddy (issue 020) ────────────────────────────────────────────────────────
# One request through Caddy's HTTPS port as https://<host><path>: curl never
# resolves the host (--connect-to sends it to Caddy's loopback port, with the
# host as SNI and Host header) and accepts the internal CA's certificate (-k).
# Prints the status code (000: no HTTP answer, e.g. no certificate for the
# host); response_header and response_body read the rest of the answer.
via_caddy() { # $1 host, $2 path, curl options...
	local host=$1 path=$2
	shift 2
	: >"$TMP/caddy-response.headers"
	: >"$TMP/caddy-response.body"
	curl -sk --max-time 40 --connect-to "$host:443:$CADDY_HTTPS" \
		-D "$TMP/caddy-response.headers" -o "$TMP/caddy-response.body" -w '%{http_code}' \
		"$@" "https://$host$path" 2>/dev/null || true
}
response_header() { tr -d '\r' <"$TMP/caddy-response.headers" | sed -n "s/^$1: *//Ip" | paste -sd ',' -; }
response_body() { cat "$TMP/caddy-response.body"; }
# Staging's basic auth user (caddy.env) and the fixture password, for curl -u.
staging_credentials() { echo "$(env_value "$TMP/prod/env/caddy.env" STAGING_BASIC_AUTH_USER):$STAGING_AUTH_PASSWORD"; }
# A Stripe-shaped webhook POST whose signature cannot verify.
post_webhook() { # $1 host, $2 path, $3 User-Agent
	via_caddy "$1" "$2" -X POST -A "$3" -H 'Content-Type: application/json' \
		-H 'Stripe-Signature: t=1,v1=0000' --data '{"type":"harness.test"}'
}
# The backend logs every request (but /health and /healthz) with its
# User-Agent: a unique one tells which backend, if any, a request reached.
backend_saw() { docker logs "$1" 2>&1 | grep -qF "\"user_agent\":\"$2\""; } # $1 container, $2 User-Agent
# `caddy <command>` (validate, adapt) on the Caddyfile, in the compose file's
# Caddy image and without a network: nothing is contacted. The environment
# holds the fixture caddy.env's values, unquoted as Compose reads them (never
# on a command line), and no LEVIOSA_CADDY_TLS unless given: by default this
# is production's variant, with the Cloudflare DNS challenge.
caddy_offline() { # $1 caddy command, docker run options...
	local command=$1 hash
	shift
	hash=$(env_value "$TMP/prod/env/caddy.env" STAGING_BASIC_AUTH_HASH)
	hash=${hash#\'}
	env CLOUDFLARE_API_TOKEN="$(env_value "$TMP/prod/env/caddy.env" CLOUDFLARE_API_TOKEN)" \
		STAGING_BASIC_AUTH_USER="$(env_value "$TMP/prod/env/caddy.env" STAGING_BASIC_AUTH_USER)" \
		STAGING_BASIC_AUTH_HASH="${hash%\'}" \
		docker run --rm --label "$LABEL" --network none -v "$CADDYFILE:/etc/caddy/Caddyfile:ro" \
		-e CLOUDFLARE_API_TOKEN -e STAGING_BASIC_AUTH_USER -e STAGING_BASIC_AUTH_HASH "$@" \
		"$CADDY_IMAGE" caddy "$command" --config /etc/caddy/Caddyfile
}
caddy_validates() { caddy_offline validate "${@:2}" >"$TMP/logs/$1.log" 2>&1; } # $1 log name, docker run options...
# The JSON config without its tls app (the last one: Caddy sorts the keys).
without_tls_app() { echo "${1%%,\"tls\":\{*}"; }
# The hosts a JSON config's tls automation policy covers, sorted, one line.
tls_subjects() { sed -n 's/.*"subjects":\[\([^]]*\)\].*/\1/p' <<<"$1" | tr -d '"' | tr ',' '\n' | sort | paste -sd ' ' -; }
# Requests through Caddy, one every 100 ms, until the stop file $1 exists:
# "<status> <seconds>" per request into $2.
caddy_stream() { # $1 stop file, $2 output, $3 host, $4 path, curl options...
	local stop=$1 out=$2 host=$3 path=$4
	shift 4
	while [ ! -e "$stop" ] && [ -d "$TMP" ]; do
		# -w prints 000 when there is no HTTP answer
		curl -sk --max-time 40 --connect-to "$host:443:$CADDY_HTTPS" -o /dev/null \
			-w '%{http_code} %{time_total}\n' "$@" "https://$host$path" >>"$out" 2>/dev/null || true
		sleep 0.1
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
	[ -n "$STREAM_PID" ] && kill "$STREAM_PID" 2>/dev/null
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
	# The real Caddyfile, as Ansible's caddy role installs it.
	cp "$CADDYFILE" "$TMP/prod/caddy/etc/Caddyfile"

	# The harness's only deviations from production, layered over the
	# unmodified compose file: its own image repositories (same :prod pin
	# tag), its own network, and Caddy on random loopback ports with its
	# internal CA in place of the Cloudflare DNS challenge.
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
    environment:
      LEVIOSA_CADDY_TLS: internal
    ports: !override
      - "127.0.0.1::80"
      - "127.0.0.1::443"
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
		check "deploy, run-harness.sh and setup-shared-services.sh pass shellcheck" \
			shellcheck "$DEPLOY" "$HARNESS_DIR/run-harness.sh" "$SETUP_SHARED_SERVICES"
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

	# caddy/Caddyfile (issue 020): production's variant, then the harness's.
	check "the Caddyfile validates in its production variant (LEVIOSA_CADDY_TLS unset), offline" \
		caddy_validates caddy-validate-prod
	check "... and in the harness variant (LEVIOSA_CADDY_TLS=internal)" \
		caddy_validates caddy-validate-internal -e LEVIOSA_CADDY_TLS=internal
	check "the Caddyfile is caddy fmt-clean (no formatting warning when Caddy loads it)" \
		bash -c '"$@" >"$0" 2>&1' "$TMP/logs/caddy-fmt.log" \
		docker run --rm --label "$LABEL" --network none -v "$CADDYFILE:/etc/caddy/Caddyfile:ro" \
		"$CADDY_IMAGE" caddy fmt --diff /etc/caddy/Caddyfile
	local prod_json harness_json hosts want
	prod_json=$(caddy_offline adapt 2>/dev/null || true)
	harness_json=$(caddy_offline adapt -e LEVIOSA_CADDY_TLS=internal 2>/dev/null || true)
	want=$(tr ' ' '\n' <<<"$PROD_HOSTS $STAGING_HOSTS" | sort | paste -sd ' ' -)
	hosts=$(tls_subjects "$prod_json")
	check "production variant: certificates for exactly the seven hosts ($hosts)" [ "$hosts" = "$want" ]
	check "... from ACME through the Cloudflare DNS challenge, its token read at runtime from caddy.env" \
		bash -c '[[ $1 == *"\"issuers\":[{\"challenges\":{\"dns\":{\"provider\":{\"api_token\":\"{env.CLOUDFLARE_API_TOKEN}\",\"name\":\"cloudflare\"}}},\"module\":\"acme\"}]"* ]] &&
			[[ $1 != *"\"module\":\"internal\""* ]]' _ "$prod_json"
	check "harness variant: the same hosts from Caddy's internal CA, no ACME" \
		bash -c '[ "$2" = "$3" ] && [[ $1 == *"\"issuers\":[{\"module\":\"internal\"}]"* ]] && [[ $1 != *"\"module\":\"acme\""* ]]' _ \
		"$harness_json" "$(tls_subjects "$harness_json")" "$want"
	check "the two variants differ in the tls app only: routing, auth, headers and logs are the same config" \
		bash -c '[ -n "$1" ] && [ "$1" = "$2" ]' _ "$(without_tls_app "$prod_json")" "$(without_tls_app "$harness_json")"
	check "the harness runs the repository's Caddyfile unmodified" cmp -s "$CADDYFILE" "$TMP/prod/caddy/etc/Caddyfile"
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
	section "Caddy: the real Caddyfile (internal CA) serves production's hosts on random loopback ports"
	if ! dc up -d --no-deps --wait --wait-timeout 60 caddy >"$TMP/logs/caddy.log" 2>&1; then
		show_log "$TMP/logs/caddy.log"
		bad "caddy did not start"
		return
	fi
	local address https body host statuses="" bodies=0 n=0 waited=0 got
	address=$(dc port caddy 80)
	https=$(dc port caddy 443)
	CADDY_HTTPS=$https
	note "caddy published on $address (http) and $https (https)"
	check "caddy is published on loopback only" [ "${address%%:*} ${https%%:*}" = "127.0.0.1 127.0.0.1" ]
	# Caddy issues the internal CA's certificates as it starts: wait for them.
	until [ "$(via_caddy leviosa.care /healthz)" = 200 ] || [ "$waited" -ge 30 ]; do
		sleep 1
		waited=$((waited + 1))
	done
	body=$(response_body)
	check "GET https://leviosa.care/healthz through caddy reports $SHA2 for both apps ($body)" \
		[ "$body" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]
	for host in $PROD_HOSTS; do
		n=$((n + 1))
		statuses+="$(via_caddy "$host" /healthz) "
		if [ "$(response_body)" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]; then bodies=$((bodies + 1)); fi
	done
	check "every production host (${PROD_HOSTS// /, }) reaches production's frontend: /healthz reports $SHA2 ($bodies/$n; statuses ${statuses% })" \
		[ "$bodies" = "$n" ]
	got=$(curl -s --max-time 10 -o /dev/null -w '%{http_code} %{redirect_url}' -H 'Host: leviosa.care' "http://$address/healthz" || true)
	check "plain HTTP redirects to HTTPS, as in production ($got)" [ "$got" = "308 https://leviosa.care/healthz" ]
	via_caddy leviosa.care /healthz >/dev/null
	check "production responses keep the security headers (HSTS with preload, nosniff, SAMEORIGIN) and drop Server" \
		bash -c '[ "$1" = "max-age=31536000; includeSubDomains; preload" ] && [ "$2" = nosniff ] && [ "$3" = SAMEORIGIN ] &&
			[ "$4" = "strict-origin-when-cross-origin" ] && [ -z "$5" ]' _ \
		"$(response_header strict-transport-security)" "$(response_header x-content-type-options)" \
		"$(response_header x-frame-options)" "$(response_header referrer-policy)" "$(response_header server)"
	statuses=""
	for host in $DROPPED_HOSTS; do
		statuses+="$host=$(via_caddy "$host" /healthz) "
	done
	check "the dropped hosts are not served: no certificate, no answer (${statuses% })" \
		[ "$statuses" = "$(for host in $DROPPED_HOSTS; do printf '%s=000 ' "$host"; done)" ]
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
	section "shared services setup: staging's role and database, each environment's vhost and user (setup-shared-services.sh, as Ansible runs it)"
	local db mq role sdb mq_user mq_vhost prod_db prod_mq_user prod_mq_vhost before log
	local rotated_pg=HarnessRotatedStagingPg3 rotated_mq=HarnessRotatedStagingMq3
	db=$(staging_db)
	mq=$(staging_mq)
	role=$(field 1 <<<"$db")
	sdb=$(field 4 <<<"$db")
	mq_user=$(field 1 <<<"$mq")
	mq_vhost=$(field 4 <<<"$mq")
	prod_db=$(env_value "$TMP/prod/env/postgres.env" POSTGRES_DB)
	prod_mq_user=$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_USER)
	prod_mq_vhost=$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_VHOST)
	# The image creates no guest user when RABBITMQ_DEFAULT_USER is set: add
	# one, as an older data directory would have it.
	rabbitmqctl_q add_user guest HarnessGuestPassword1 >/dev/null
	before=$(shared_services_state)

	if setup_shared_services shared-services-1; then
		ok "setup-shared-services.sh exits 0"
	else
		show_log "$TMP/logs/shared-services-1.log"
		fatal "setup-shared-services.sh failed"
	fi
	check "the first run reports that it changed something" \
		[ "$(tail -n 1 "$TMP/logs/shared-services-1.log")" = "setup-shared-services: changed" ]
	if setup_shared_services shared-services-2; then
		ok "a second run exits 0 as well"
	else
		show_log "$TMP/logs/shared-services-2.log"
		bad "the second setup-shared-services.sh run failed"
	fi
	check "a second run changes nothing (reports unchanged, no changed: line)" \
		bash -c '[ "$(tail -n 1 "$1")" = "setup-shared-services: unchanged" ] && ! grep -q "changed: " "$1"' _ \
		"$TMP/logs/shared-services-2.log"

	# A rotation in Infisical: the next run re-applies exactly the passwords.
	setup_shared_services shared-services-rotate STAGING_DB_PASSWORD=$rotated_pg STAGING_MQ_PASSWORD=$rotated_mq || true
	check "a password rotation is applied, and reported as exactly two changes" \
		[ "$(grep -c 'changed: set the password of the' "$TMP/logs/shared-services-rotate.log")/$(grep -c 'changed: ' "$TMP/logs/shared-services-rotate.log")" = 2/2 ]
	check "... $role logs in to Postgres with the rotated password" [ "$(pg_connect_as "$role" "$rotated_pg" "$sdb")" = "$role" ]
	check "... $mq_user logs in to RabbitMQ (AMQP) with the rotated password" [ "$(amqp_open "$mq_user" "$rotated_mq" "$mq_vhost")" = open-ok ]
	setup_shared_services shared-services-restore || true
	check "rotating back is two changes again" \
		[ "$(grep -c 'changed: set the password of the' "$TMP/logs/shared-services-restore.log")/$(grep -c 'changed: ' "$TMP/logs/shared-services-restore.log")" = 2/2 ]
	setup_shared_services shared-services-3 || true
	check "... and the run after it changes nothing" \
		[ "$(tail -n 1 "$TMP/logs/shared-services-3.log")" = "setup-shared-services: unchanged" ]
	check "no run restarted or recreated Postgres or RabbitMQ" [ "$(shared_services_state)" = "$before" ]

	check "the staging role $role is no superuser and cannot create roles or databases" \
		[ "$(psql_q "SELECT rolsuper OR rolcreaterole OR rolcreatedb OR rolreplication OR rolbypassrls
			FROM pg_roles WHERE rolname = '$role'")" = f ]
	check "the staging database $sdb is owned by $role" \
		[ "$(psql_q "SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = '$sdb'")" = "$role" ]
	check "$role holds no CONNECT privilege on production's database $prod_db" \
		[ "$(psql_q "SELECT has_database_privilege('$role', '$prod_db', 'CONNECT')")" = f ]
	check "$role holds CONNECT on no database but $sdb (not even postgres or template1)" \
		[ "$(psql_q "SELECT string_agg(datname, ' ') FROM pg_database
			WHERE datallowconn AND has_database_privilege('$role', datname, 'CONNECT')")" = "$sdb" ]
	check "the RabbitMQ user $mq_user has permissions on vhost $mq_vhost only" \
		[ "$(rabbitmqctl_q list_user_permissions --no-table-headers "$mq_user" | cut -f1 | tr '\n' ' ')" = "$mq_vhost " ]
	check "production's RabbitMQ user $prod_mq_user has permissions on vhost $prod_mq_vhost only" \
		[ "$(rabbitmqctl_q list_user_permissions --no-table-headers "$prod_mq_user" | cut -f1 | tr '\n' ' ')" = "$prod_mq_vhost " ]
	check "the only RabbitMQ users are $prod_mq_user and $mq_user, both without tags (no management access)" \
		[ "$(rabbitmqctl_q list_users --no-table-headers | sort | tr '\t\n' ': ')" = "$(printf '%s\n' "$prod_mq_user:[]" "$mq_user:[]" | sort | tr '\n' ' ')" ]
	check "the guest user is gone" \
		bash -c '! grep -qx guest' < <(rabbitmqctl_q list_users --no-table-headers | cut -f1)
	check "the backend is still connected to RabbitMQ as $prod_mq_user on vhost $prod_mq_vhost" \
		bash -c '[[ $(cat) == *"$1"*"$2"* ]]' _ "$prod_mq_user" "$prod_mq_vhost" < <(rabbitmqctl_q list_connections user vhost 2>/dev/null)
	for log in shared-services-1 shared-services-2 shared-services-rotate shared-services-restore shared-services-3; do
		check "the output of $log holds no password" \
			bash -c '! grep -qF -e "$2" -e "$3" -e "$4" -e "$5" -e "$6" "$1"' _ "$TMP/logs/$log.log" \
			"$(field 2 <<<"$db")" "$(field 2 <<<"$mq")" "$(env_value "$TMP/prod/env/rabbitmq.env" RABBITMQ_DEFAULT_PASS)" \
			"$rotated_pg" "$rotated_mq"
	done
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
	if [ -n "$CADDY_HTTPS" ]; then
		local body
		via_caddy "$STAGING_HOST" /healthz >/dev/null
		body=$(response_body)
		check "through Caddy, $STAGING_HOST reaches staging_frontend by its service name ($body)" \
			[ "$body" = "{\"frontend\":\"$SHA1\",\"backend\":\"$SHA1\"}" ]
		via_caddy leviosa.care /healthz >/dev/null
		body=$(response_body)
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
	got=$(pg_connect_as "$role" "$password" postgres)
	check "$role is refused on the postgres maintenance database too, for lack of CONNECT ($(head -n 1 <<<"$got"))" \
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

# ── Caddy cases (issue 020) ──────────────────────────────────────────────────
# POST each Stripe webhook path on every host of a stack: each must get the
# backend's signature error and reach that stack's backend only, which the
# request's unique User-Agent in the two backends' request logs tells.
webhook_checks() { # $1 stack, $2 hosts, $3 own backend container, $4 its name, $5 other backend, $6 its name
	local stack=$1 hosts=$2 own=$3 own_name=$4 other=$5 other_name=$6 path host ua statuses want bodies seen leaked n
	local -a agents
	for path in $WEBHOOK_PATHS; do
		statuses="" want="" bodies=0 seen=0 leaked=0 agents=()
		for host in $hosts; do
			ua="leviosa-harness-$RUN_ID-webhook-$stack-${#agents[@]}-${path//\//-}"
			agents+=("$ua")
			statuses+="$(post_webhook "$host" "$path" "$ua") "
			want+="400 "
			if [ "$(response_body)" = "$WEBHOOK_REJECTED" ]; then bodies=$((bodies + 1)); fi
		done
		# The request logs reach `docker logs` asynchronously: wait up to 10 s
		# for the own backend to show every request, then count the leaks.
		n=${#agents[@]}
		for _ in $(seq 20); do
			seen=0
			for ua in "${agents[@]}"; do
				if backend_saw "$own" "$ua"; then seen=$((seen + 1)); fi
			done
			[ "$seen" -lt "$n" ] || break
			sleep 0.5
		done
		for ua in "${agents[@]}"; do
			if backend_saw "$other" "$ua"; then leaked=$((leaked + 1)); fi
		done
		check "$stack: POST $path on ${hosts// /, } gets the backend's signature error, not a frontend page (statuses ${statuses% }; error body $bodies/$n)" \
			[ "$statuses$bodies" = "$want$n" ]
		check "... and reached $own_name only (seen by $own_name: $seen/$n; by $other_name: $leaked)" [ "$seen $leaked" = "$n 0" ]
	done
}

# GET backend-only routes through Caddy on a stack's host: the frontend
# answers, and neither backend sees them.
backend_route_checks() { # $1 stack, $2 host, curl options...
	local stack=$1 host=$2 path ua statuses="" saw="" i=0
	local -a agents
	shift 2
	for path in /users/me /auth/refresh /categories /webhooks; do
		ua="leviosa-harness-$RUN_ID-route-$stack-$i"
		i=$((i + 1))
		agents+=("$ua")
		statuses+="$path=$(via_caddy "$host" "$path" -A "$ua" "$@") "
	done
	sleep 1
	for ua in "${agents[@]}"; do
		if backend_saw "$(cid backend)" "$ua"; then saw+="backend "; fi
		if backend_saw "$(scid staging_backend)" "$ua"; then saw+="staging_backend "; fi
	done
	check "$stack: other backend routes through Caddy go to the frontend; no backend saw them (${statuses% }; seen by: ${saw:-none})" \
		[ -z "$saw" ]
}

case_caddy_webhooks() {
	section "caddy criterion 1: Stripe's webhooks reach each stack's own backend, staging's without basic auth"
	local prod_backend staging_backend
	prod_backend=$(cid backend)
	staging_backend=$(scid staging_backend)
	webhook_checks production "$PROD_HOSTS" "$prod_backend" "production's backend" "$staging_backend" staging_backend
	webhook_checks staging "$STAGING_HOSTS" "$staging_backend" staging_backend "$prod_backend" "production's backend"
	backend_route_checks production leviosa.care
	backend_route_checks staging "$STAGING_HOST" -u "$(staging_credentials)"
}

case_caddy_staging_auth() {
	section "caddy criterion 2: any staging page without credentials is a 401; with them, the page"
	local host path statuses want challenges n credentials body
	local paths="/ /admin /staff /client /auth/login /robots.txt /sitemap.xml /coming-soon /api/products /healthz/ /webhooks /webhooks-x/stripe"
	credentials=$(staging_credentials)
	for host in $STAGING_HOSTS; do
		statuses="" want="" challenges=0 n=0
		for path in $paths; do
			n=$((n + 1))
			statuses+="$(via_caddy "$host" "$path") "
			want+="401 "
			if [[ $(response_header www-authenticate) == "Basic realm="* ]]; then challenges=$((challenges + 1)); fi
		done
		check "$host: every page without credentials is a 401 with a Basic challenge ($n paths: ${statuses% }; challenges $challenges/$n)" \
			[ "$statuses$challenges" = "$want$n" ]
	done
	statuses=""
	for host in $STAGING_HOSTS; do
		statuses+="$(via_caddy "$host" / -X POST --data x) $(via_caddy "$host" / -u "$(staging_credentials | cut -d: -f1):wrong-password") "
	done
	check "a POST without credentials and a wrong password are 401s too (${statuses% })" \
		[ "$statuses" = "401 401 401 401 401 401 " ]
	for host in $STAGING_HOSTS; do
		statuses=$(via_caddy "$host" /robots.txt -u "$credentials")
		body=$(response_body)
		check "$host: with credentials, /robots.txt is the frontend's page for that host ($statuses, $(grep -m1 '^Sitemap:' <<<"$body" || echo 'no Sitemap line'))" \
			bash -c '[ "$1" = 200 ] && grep -qxF "Sitemap: https://$3/sitemap.xml" <<<"$2"' _ "$statuses" "$body" "$host"
	done
	statuses=$(via_caddy "$STAGING_HOST" / -u "$credentials")
	check "with credentials, $STAGING_HOST/ returns the home page ($statuses, $(response_header content-type))" \
		bash -c '[ "$1" = 200 ] && [[ $2 == text/html* ]] && grep -qi "<html" "$3"' _ "$statuses" "$(response_header content-type)" "$TMP/caddy-response.body"
	local log
	log=$(dc exec -T caddy cat /var/log/caddy/staging-access.log 2>/dev/null || true)
	check "staging's requests are logged as JSON to staging-access.log ($(grep -c '"host":"[a-z-]*staging.leviosa.care"' <<<"$log") entries), production's are not" \
		bash -c 'grep -q "\"host\":\"staging.leviosa.care\"" <<<"$1" && ! grep -q "\"host\":\"leviosa.care\"" <<<"$1"' _ "$log"
	check "... and the log holds no credential (Authorization redacted)" \
		bash -c '! grep -qF -e "$2" -e "$3" <<<"$1"' _ "$log" "$STAGING_AUTH_PASSWORD" "$(printf '%s' "$credentials" | base64)"
	log=$(dc exec -T caddy cat /var/log/caddy/access.log 2>/dev/null || true)
	check "production's requests are logged as JSON to access.log ($(grep -c '"host":"[a-z.]*leviosa.care"' <<<"$log") entries), staging's are not" \
		bash -c 'grep -q "\"host\":\"leviosa.care\"" <<<"$1" && ! grep -q "staging.leviosa.care" <<<"$1"' _ "$log"
}

case_caddy_staging_healthz() {
	section "caddy criterion 3: staging /healthz is a 200 without credentials"
	local host statuses="" bodies=0 n=0
	for host in $STAGING_HOSTS; do
		n=$((n + 1))
		statuses+="$(via_caddy "$host" /healthz) "
		if [ "$(response_body)" = "{\"frontend\":\"$SHA1\",\"backend\":\"$SHA1\"}" ]; then bodies=$((bodies + 1)); fi
	done
	check "/healthz on ${STAGING_HOSTS// /, } answers without credentials (statuses ${statuses% })" \
		[ "$statuses" = "200 200 200 " ]
	check "... with staging's release, $SHA1 for both apps ($bodies/$n)" [ "$bodies" = "$n" ]
}

case_caddy_robots() {
	section "caddy criterion 4: staging responses carry X-Robots-Tag: noindex, production's don't"
	local host got="" credentials
	credentials=$(staging_credentials)
	for host in $STAGING_HOSTS; do
		got+="$host/=$(via_caddy "$host" /):$(response_header x-robots-tag) "
		got+="$host/robots.txt=$(via_caddy "$host" /robots.txt -u "$credentials"):$(response_header x-robots-tag) "
		got+="$host/healthz=$(via_caddy "$host" /healthz):$(response_header x-robots-tag) "
		got+="$host/webhooks/stripe=$(post_webhook "$host" /webhooks/stripe "leviosa-harness-$RUN_ID-robots"):$(response_header x-robots-tag) "
	done
	note "staging: ${got% }"
	check "every staging response (401, authenticated page, /healthz, webhook; on each host) carries X-Robots-Tag: noindex, once" \
		bash -c 'for r in $1; do [ "${r##*:}" = noindex ] || exit 1; done' _ "$got"
	got=""
	for host in $PROD_HOSTS; do
		got+="$host/=$(via_caddy "$host" /):$(response_header x-robots-tag) "
		got+="$host/robots.txt=$(via_caddy "$host" /robots.txt):$(response_header x-robots-tag) "
		got+="$host/healthz=$(via_caddy "$host" /healthz):$(response_header x-robots-tag) "
		got+="$host/webhooks/stripe=$(post_webhook "$host" /webhooks/stripe "leviosa-harness-$RUN_ID-robots"):$(response_header x-robots-tag) "
	done
	note "production: ${got% }"
	check "no production response (page, robots.txt, /healthz, webhook; on each host) carries X-Robots-Tag" \
		bash -c 'for r in $1; do [ -z "${r##*:}" ] || exit 1; done' _ "$got"
}

# Recreate a production service (as a deploy or the Agent does) while
# requests stream through Caddy; every request must get the service's usual
# answer, held by the retry window rather than failed with a 502.
recreate_under_load() { # $1 service, $2 expected status, $3 host, $4 path, curl options...
	local service=$1 want=$2 host=$3 path=$4 before stop out total summary longest
	shift 4
	before=$(cid "$service")
	stop=$TMP/stream-$service.stop
	out=$TMP/logs/stream-$service.log
	: >"$out"
	caddy_stream "$stop" "$out" "$host" "$path" "$@" &
	STREAM_PID=$!
	sleep 2
	if dc up -d --no-deps --force-recreate --wait --wait-timeout 120 "$service" >"$TMP/logs/recreate-$service.log" 2>&1; then
		ok "$service recreated (up -d --no-deps --force-recreate --wait) while requests went through Caddy every 100 ms"
	else
		show_log "$TMP/logs/recreate-$service.log"
		bad "recreating $service failed"
	fi
	sleep 2
	touch "$stop"
	wait "$STREAM_PID" || true
	STREAM_PID=""
	check "the $service container is a new one" [ "$(cid "$service")" != "$before" ]
	total=$(grep -c . "$out" || true)
	summary=$(cut -d ' ' -f 1 "$out" | sort | uniq -c | awk '{ printf "%s%sx %s", sep, $1, $2; sep = ", " }')
	longest=$(cut -d ' ' -f 2 "$out" | sort -n | tail -n 1)
	check "none of the $total requests ($host$path) got a 502, 503 or 504, or no answer ($summary)" \
		bash -c '[ "$2" -gt 0 ] && ! grep -qE "^(000|502|503|504) " "$1"' _ "$out" "$total"
	check "... every one got the usual $want" bash -c '! grep -qvE "^$2 " "$1"' _ "$out" "$want"
	check "... some waited in the retry window while $service was away (longest ${longest}s)" \
		awk -v t="$longest" 'BEGIN { exit !(t >= 0.5) }'
}

case_caddy_retry_window() {
	section "caddy criterion 5: requests sent while the backend or the frontend is recreated get no 502"
	recreate_under_load backend 400 leviosa.care /webhooks/stripe -X POST -A "leviosa-harness-$RUN_ID-stream" \
		-H 'Content-Type: application/json' -H 'Stripe-Signature: t=1,v1=0000' --data '{"type":"harness.test"}'
	recreate_under_load frontend 200 leviosa.care /healthz
	via_caddy leviosa.care /healthz >/dev/null
	check "production still serves $SHA2 for both apps through Caddy ($(response_body))" \
		[ "$(response_body)" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]
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
case_caddy_webhooks
case_caddy_staging_auth
case_caddy_staging_healthz
case_caddy_robots
case_caddy_retry_window
if [ "${KEEP:-0}" = 1 ]; then
	note "KEEP=1: teardown skipped"
else
	case_teardown
fi
summary
