#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# Local Docker harness for docker-compose.prod.yml and `deploy` (issue 018).
#
#   infra/compose/test/run-harness.sh
#
# One command, real Docker, no secrets, nothing pushed. It builds the backend
# and frontend images locally with fake COMMIT_SHAs, lays out a temporary
# stack root with the fixture env files (fixtures/, fake stand-ins for the
# Infisical Agent's renderings), runs LocalStack (KMS only) for the encx KEK,
# and drives the real deploy script through the issue's acceptance criteria:
#
#   1. first deploy: migrates, bootstraps the admin, both apps report the SHA
#   2. second deploy: data kept, admin untouched
#   3. failing migration: aborts before recreate, previous containers serve
#   4. /healthz SHA mismatch (backend or frontend) fails the deploy
#   5. no ADMIN_* value in the running backend or frontend
#   6. data survives container recreation (`down`, then a deploy)
#   7. nothing left behind: containers, networks, volumes, images, temp files
#
# Isolation: every name carries a random run id (leviosa-harness-<id>): its
# own compose project, network, LocalStack container and image repositories
# (<name>-backend / <name>-frontend, so the henga/leviosa-* pins are never
# touched). Nothing binds a fixed host port (Caddy gets a random loopback
# port). Everything on disk lives under one mktemp -d. An EXIT trap tears it
# all down on success and failure; public images the run had to pull are
# removed too (HARNESS_KEEP_PULLED=1 keeps them for faster reruns).
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
LOCALSTACK=$NAME-localstack
BACKEND_REPO=$NAME-backend
FRONTEND_REPO=$NAME-frontend
LABEL=leviosa.harness=$RUN_ID

LOCALSTACK_IMAGE=localstack/localstack:4.6
AWS_REGION=eu-central-1             # as in fixtures/prod/env/backend.env
KMS_ALIAS=alias/leviosa-harness     # ENCX_KMS_KEY_ALIAS in the same file
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

cid() { dc ps -q "$1"; } # running container of a service
health() { docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$1"; }
image_id() { docker image inspect -f '{{.Id}}' "$1"; }
container_image() { docker inspect -f '{{.Image}}' "$1"; }
container_env() { docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' "$1"; }

# GET frontend:3000/healthz from a throwaway container on the stack's
# network (the backend image has busybox wget) and print "<frontend> <backend>".
healthz_shas() {
	local body
	body=$(docker run --rm --label "$LABEL" --network "$NETWORK" --entrypoint wget \
		"$BACKEND_REPO:$SHA1" -q -O - -T 5 http://frontend:3000/healthz 2>/dev/null) || {
		echo "unreachable"
		return 0
	}
	printf '%s %s\n' \
		"$(sed -n 's/.*"frontend":"\([^"]*\)".*/\1/p' <<<"$body")" \
		"$(sed -n 's/.*"backend":"\([^"]*\)".*/\1/p' <<<"$body")"
}
serving() { [ "$(healthz_shas)" = "$1 $1" ]; } # both apps report $1

psql_q() { dc exec -T postgres psql -U leviosa -d leviosa -v ON_ERROR_STOP=1 -qtAc "$1"; }
user_count() { psql_q 'SELECT count(*) FROM auth.users'; }
# Everything bootstrap-admin could change on an existing account.
admin_fingerprint() {
	psql_q "SELECT string_agg(id || '|' || password_hash_secure || '|' || email_hash || '|' ||
		md5(first_name_encrypted) || '|' || md5(role_encrypted) || '|' || updated_at, ',') FROM auth.users"
}
schema_version() { psql_q 'SELECT max(version_id) FROM goose_db_version'; }
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

# ── teardown ─────────────────────────────────────────────────────────────────
harness_containers() {
	{
		docker ps -aq --filter "label=com.docker.compose.project=$PROJECT"
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
		echo "KEEP=1: left running: project $PROJECT, network $NETWORK, $LOCALSTACK; temp dir $TMP" >&2
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
	# Services of the production compose file (minus the app pins), LocalStack,
	# and the Dockerfiles' base images and syntax frontend.
	COMPOSE_FILE="$COMPOSE_DIR/docker-compose.prod.yml" docker compose \
		--project-directory "$FIXTURES/prod" --profile oneoff config --images 2>/dev/null |
		grep -v '^henga/leviosa-'
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

	docker network create --label "$LABEL" "$NETWORK" >/dev/null || fatal "could not create network $NETWORK"
	ok "network $NETWORK created"
	note "temp dir $TMP"

	# LocalStack, KMS only, as `localstack` on the stack's network; the KEK
	# alias is created the way Terraform provisions it in real environments.
	docker run -d --name "$LOCALSTACK" --label "$LABEL" --network "$NETWORK" --network-alias localstack \
		-e SERVICES=kms -e PERSISTENCE=0 -e AWS_DEFAULT_REGION="$AWS_REGION" \
		"$LOCALSTACK_IMAGE" >/dev/null || fatal "could not start LocalStack"
	local waited=0
	until docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms list-keys >/dev/null 2>&1; do
		[ "$waited" -lt 120 ] || fatal "LocalStack KMS not ready after 120s"
		sleep 2
		waited=$((waited + 2))
	done
	local key_id
	key_id=$(docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms create-key \
		--description "Leviosa harness encx KEK" --query KeyMetadata.KeyId --output text) ||
		fatal "could not create the KMS key"
	docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" kms create-alias \
		--alias-name "$KMS_ALIAS" --target-key-id "$key_id" ||
		fatal "could not create the KMS alias $KMS_ALIAS"
	ok "LocalStack KMS up with $KMS_ALIAS"
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
		check "deploy and run-harness.sh pass shellcheck" shellcheck "$DEPLOY" "$HARNESS_DIR/run-harness.sh"
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
	newest=$(find "$REPO_ROOT/backend/internal/common/migrations" -name '*.sql' -printf '%f\n' |
		sed -n 's/^\([0-9]*\)_.*/\1/p' | sort -n | tail -n 1)
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
	note "caddy published on $address"
	check "caddy is published on loopback only" [ "${address%%:*}" = 127.0.0.1 ]
	body=$(curl -fsS --max-time 10 "http://$address/healthz" 2>&1 || true)
	check "GET /healthz through caddy reports $SHA2 for both apps ($body)" \
		[ "$body" = "{\"frontend\":\"$SHA2\",\"backend\":\"$SHA2\"}" ]
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
case_first_deploy
case_second_deploy
case_failing_migration
case_sha_mismatch
case_no_admin_env
case_data_survives
case_caddy
if [ "${KEEP:-0}" = 1 ]; then
	note "KEEP=1: teardown skipped"
else
	case_teardown
fi
summary
