#!/usr/bin/env bash
# Tests for roles/backup (issue 028): no root, no server, no AWS, no
# Infisical.
#
#   infra/ansible/roles/backup/tests/run-tests.sh [--static]
#   (or: make backup-test, from the repository root)
#
#   role      the role runs unprivileged against localhost into a temporary
#             tree (test knobs: backup_systemd_dir, backup_owner, no apt, no
#             systemd): with the committed placeholder key the weekly job is
#             refused (warning, no units, no snapshot template); with a real
#             key it is installed, the snapshot template substituted; back to
#             the placeholder removes it again; a second run changes nothing
#   units     systemd-analyze verify on the rendered services and timers, and
#             the OnCalendar expressions (03:00 daily, Sunday 04:00, Paris)
#   scripts   bash -n and shellcheck (koalaman/shellcheck:stable)
#
#   e2e       (skipped with --static) real Docker, isolated names, no host
#             port: LocalStack (KMS + S3), the backend image built from this
#             checkout, production's compose file with the harness fixtures:
#             migrate, bootstrap-admin (an encrypted auth.users record), serve
#             (the pepper canary). Then the scripts the role installed, with a
#             test configuration pointing them at LocalStack:
#     B1  leviosa-backup uploads db/<date>/…dump and config/<date>/…tar.gz
#     B2  the configuration archive holds the allowlist and no secret
#     B3  an env directory or a secret value in the allowlist is refused
#     S1  leviosa-snapshot uploads infisical/<date>/…json.age (stub CLI):
#         unreadable without the age private key, readable with it; the
#         client secret never on a command line
#     S2  the placeholder recipient is refused
#     R1  leviosa-restore-test passes on the dump (encrypted record + canary)
#     R2  … and fails with exit 6 given a wrong pepper
#     R3  … --local (latest.dump) passes too; nothing left behind
#     L1  no secret in any script output
#
# Prints PASS/FAIL/SKIP per check and ends with `RESULT: OK` (exit 0) or
# `RESULT: FAILED` (exit 1). Everything it starts carries a random run id and
# is removed on exit (KEEP=1 keeps it for inspection).
#
# shellcheck disable=SC2016
# ^ single-quoted bash -c scripts expand their own arguments.
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
role=$(cd "$here/.." && pwd)
ansible_dir=$(cd "$role/../.." && pwd)
repo=$(cd "$ansible_dir/../.." && pwd)
fixtures=$repo/infra/compose/test/fixtures

static_only=0
[ "${1:-}" = --static ] && static_only=1

run_id=$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')
NAME=leviosa-backup-test-$run_id
LABEL=leviosa.backup-test=$run_id
LOCALSTACK=$NAME-localstack
PROJECT=$NAME-prod
IMAGE=$NAME-backend:prod
LOCALSTACK_IMAGE=localstack/localstack:4.6
AWS_REGION=eu-central-1          # fixtures/prod/env/backend.env
KMS_ALIAS=alias/leviosa-harness  # ENCX_KMS_KEY_ALIAS there
BUCKET=leviosa-backup-test
FAKE_PID=11111111-2222-4333-8444-555555555555

T=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-backup-test.XXXXXX")

passed=0 failed=0 skipped=0
pass() { passed=$((passed + 1)); printf 'PASS  %s\n' "$*"; }
fail() { failed=$((failed + 1)); printf 'FAIL  %s\n' "$*"; }
skip() { skipped=$((skipped + 1)); printf 'SKIP  %s\n' "$*"; }
check() { # check <description> <command…>
	local d=$1
	shift
	if "$@" >>"$T/check.log" 2>&1; then pass "$d"; else fail "$d"; fi
}
show() { sed -n "1,${2:-30}p" "$1" | sed 's/^/        | /'; }

teardown() {
	set +e
	if [ "${KEEP:-0}" = 1 ]; then
		echo "KEEP=1: left $T, containers/networks labelled $LABEL, project $PROJECT" >&2
		return
	fi
	{
		docker ps -aq --filter "label=$LABEL"
		docker ps -aq --filter "label=com.docker.compose.project=$PROJECT"
		docker ps -aq --filter "label=leviosa.restore-test"  --filter "network=$NAME"
	} 2>/dev/null | sort -u | xargs -r docker rm -fv >/dev/null 2>&1
	docker network rm "$NAME" >/dev/null 2>&1
	docker image rm "$IMAGE" >/dev/null 2>&1
	if ! rm -rf "$T" 2>/dev/null; then
		# Postgres and friends write the bind mounts as their own users.
		docker run --rm -v "$T:/wipe" --entrypoint /bin/sh redis:7-alpine \
			-c 'find /wipe -mindepth 1 -delete' >/dev/null 2>&1
		rm -rf "$T"
	fi
}
trap teardown EXIT
trap 'exit 130' INT TERM

# ===========================================================================
# Layout: the server's paths, under $T
# ===========================================================================
PROD=$T/opt/leviosa
STAGING=$T/opt/leviosa-staging
AGENT=$T/etc/infisical-agent
mkdir -p "$PROD/env" "$PROD/data/postgres" "$PROD/data/rabbitmq" "$PROD/data/redis" \
	"$PROD/caddy/etc" "$PROD/caddy/data" "$PROD/caddy/config" "$PROD/caddy/logs" \
	"$STAGING/env" "$STAGING/data/redis" "$AGENT/credentials" "$AGENT/templates" \
	"$T/bin" "$T/lib" "$T/etc/leviosa" "$T/work" "$T/systemd" "$T/stub" "$T/keys" "$T/out"
cp "$repo/infra/compose/docker-compose.prod.yml" "$PROD/docker-compose.yml"
cp "$repo/infra/compose/docker-compose.staging.yml" "$STAGING/docker-compose.yml"
cp "$fixtures/prod/env/"*.env "$PROD/env/"
cp "$fixtures/staging/env/"*.env "$STAGING/env/"
cp "$repo/infra/compose/caddy/Caddyfile" "$PROD/caddy/etc/Caddyfile"
cp "$repo/infra/infisical/templates/"*.tmpl "$AGENT/templates/"
# As the Agent renders it (issue 024): every value single-quoted.
cat >"$PROD/env/backup.env" <<'EOF'
# leviosa backup.env - TEST FIXTURE
BACKUP_AWS_ACCESS_KEY_ID='BackupFixtureKeyId0001'
BACKUP_AWS_SECRET_ACCESS_KEY='BackupFixtureSecret-not-real-0001'
BACKUP_S3_BUCKET='leviosa-backup-test'
BACKUP_AWS_REGION='eu-central-1'
POSTGRES_USER='leviosa'
POSTGRES_DB='leviosa'
EOF
chmod 600 "$PROD/env/"*.env "$STAGING/env/"*.env
printf 'FixtureClientId-0000-1111\n' >"$AGENT/credentials/client-id"
printf 'FixtureClientSecret-not-real-2222\n' >"$AGENT/credentials/client-secret"
printf '{"access_token":"FixtureAgentToken-not-real-3333"}\n' >"$AGENT/credentials/token-sink.json"
cat >"$AGENT/config.yaml" <<EOF
# fixture Agent config: paths only
auth:
  type: universal-auth
  config:
    client-id: $AGENT/credentials/client-id
    client-secret: $AGENT/credentials/client-secret
EOF

# Every fixture value that counts as a secret (L1, B2): the values of the
# fixture env files (8+ characters, not plain words the archive may hold),
# the credential files and the Agent token.
secrets_file=$T/secrets.txt
{
	for f in "$PROD/env/"*.env "$STAGING/env/"*.env; do
		sed -n -E "s/^[A-Z0-9_]*(PASSWORD|PASS|SECRET|TOKEN|PEPPER|HASH|_KEY|_KEY_ID|_SID|DATABASE_URL|RABBITMQ_URL)=//p" "$f" |
			sed -e "s/^'//" -e "s/'\$//"
	done
	cat "$AGENT/credentials/client-id" "$AGENT/credentials/client-secret"
	echo FixtureAgentToken-not-real-3333
} | awk 'length >= 8' | sort -u >"$secrets_file"
PEPPER=$(sed -n 's/^ENCX_PEPPER=//p' "$PROD/env/backend.env")
no_secret_in() { ! grep -rqF -f "$secrets_file" "$@"; }

# ===========================================================================
# 1. The role, against localhost
# ===========================================================================
age-keygen -o "$T/keys/operator.key" 2>/dev/null
age-keygen -o "$T/keys/other.key" 2>/dev/null
PUBKEY=$(age-keygen -y "$T/keys/operator.key")
printf '# test key\n%s\n' "$PUBKEY" >"$T/keys/backup-age.pub"

cat >"$T/play.yml" <<YAML
- name: Run roles/backup into a temporary tree
  hosts: localhost
  connection: local
  gather_facts: false
  become: false
  vars:
    app_prod_root: $PROD
    app_staging_root: $STAGING
    infisical_agent_config_dir: $AGENT
    infisical_agent_credentials_dir: $AGENT/credentials
    backup_bin_dir: $T/bin
    backup_lib_dir: $T/lib
    backup_conf_file: $T/etc/leviosa/backup.conf
    backup_work_dir: $T/work
    backup_systemd_dir: $T/systemd
    backup_compose_project: $PROJECT
    backup_manifest_projects: [$PROJECT, $NAME-staging]
    backup_restore_backend_image: $IMAGE
    backup_install_packages: false
    backup_manage_service: false
    backup_owner: "$(id -un)"
    backup_group: "$(id -gn)"
  roles:
    - backup
YAML
run_role() { # run_role <log> [-e …]
	local log=$1
	shift
	(cd "$ansible_dir" && ANSIBLE_FORCE_COLOR=0 \
		ansible-playbook -i localhost, "$T/play.yml" "$@") >"$T/$log" 2>&1
}
changed_of() { sed -n 's/.*localhost *: ok=[0-9]* *changed=\([0-9]*\).*/\1/p' "$T/$1"; }

if run_role role-placeholder.log -e infisical_project_id=$FAKE_PID; then
	pass "the role runs with the committed placeholder key"
else
	fail "the role runs with the committed placeholder key"
	show "$T/role-placeholder.log" 40
fi
check "placeholder: the role warns that the weekly snapshot is NOT installed" \
	grep -q "WARNING: the weekly Infisical snapshot (leviosa-snapshot) is NOT" "$T/role-placeholder.log"
check "placeholder: only the daily units are installed" \
	bash -c '[ "$(cd "$1" && ls | sort | tr "\n" " ")" = "leviosa-backup.service leviosa-backup.timer " ]' _ "$T/systemd"
check "placeholder: no snapshot template or key installed" \
	bash -c '[ ! -e "$1/snapshot.json.tmpl" ] && [ ! -e "$1/backup-age.pub" ]' _ "$T/lib"
check "scripts installed 0750, library and configuration 0644" \
	bash -c '[ "$(stat -c %a "$1/bin/leviosa-backup" "$1/bin/leviosa-snapshot" "$1/bin/leviosa-restore-test" | sort -u)" = 750 ] &&
		[ "$(stat -c %a "$1/lib/lib.sh" "$1/etc/leviosa/backup.conf" | sort -u)" = 644 ] &&
		[ "$(stat -c %a "$1/work")" = 700 ]' _ "$T"
check "the installed scripts are the role's files, byte for byte" \
	bash -c 'for f in leviosa-backup leviosa-snapshot leviosa-restore-test; do cmp -s "$1/files/$f" "$2/bin/$f" || exit 1; done; cmp -s "$1/files/leviosa-backup-lib.sh" "$2/lib/lib.sh"' _ "$role" "$T"
check "the placeholder project ID also refuses the weekly job" \
	bash -c 'run() { cd "$1" && ansible-playbook -i localhost, "$2/play.yml" -e backup_age_public_key_file="$2/keys/backup-age.pub" -e infisical_project_id=REPLACE_WITH_LEVIOSA_PROJECT_ID; }
		run "$@" >"$2/role-nopid.log" 2>&1 && grep -q "infisical_project_id is still a placeholder" "$2/role-nopid.log" && [ ! -e "$2/systemd/leviosa-snapshot.timer" ]' _ "$ansible_dir" "$T"

if run_role role-key.log -e infisical_project_id=$FAKE_PID -e backup_age_public_key_file="$T/keys/backup-age.pub"; then
	pass "the role runs with a real age key"
else
	fail "the role runs with a real age key"
	show "$T/role-key.log" 40
fi
check "real key: the daily and weekly units are installed" \
	bash -c '[ "$(cd "$1" && ls | sort | tr "\n" " ")" = "leviosa-backup.service leviosa-backup.timer leviosa-snapshot.service leviosa-snapshot.timer " ]' _ "$T/systemd"
check "real key: backup-age.pub holds exactly the public key" \
	bash -c '[ "$(cat "$1")" = "$2" ]' _ "$T/lib/backup-age.pub" "$PUBKEY"
check "real key: the snapshot template is render-template.sh's substitution, byte for byte" \
	bash -c 'diff <(LEVIOSA_INFISICAL_PROJECT_ID=$3 sh "$1/infra/infisical/render-template.sh" "$2/files/leviosa-snapshot.json.tmpl" prod) "$4"' \
	_ "$repo" "$role" "$FAKE_PID" "$T/lib/snapshot.json.tmpl"
run_role role-key-2.log -e infisical_project_id=$FAKE_PID -e backup_age_public_key_file="$T/keys/backup-age.pub" || true
check "a second run changes nothing (changed=$(changed_of role-key-2.log))" [ "$(changed_of role-key-2.log)" = 0 ]
cp -a "$T/systemd" "$T/systemd-weekly" # for systemd-analyze below
cp -a "$T/lib" "$T/lib-weekly"
run_role role-revert.log -e infisical_project_id=$FAKE_PID || true
check "back to the placeholder: the weekly units are removed" \
	bash -c '[ ! -e "$1/leviosa-snapshot.timer" ] && [ ! -e "$1/leviosa-snapshot.service" ] && [ -e "$1/leviosa-backup.timer" ]' _ "$T/systemd"
cp -a "$T/lib-weekly/." "$T/lib/" # the e2e needs the snapshot template

# The configuration the role rendered parses in bash, with the expected arrays.
check "backup.conf parses and names the allowlist, without env or credentials" \
	bash -c '. "$1"; [ ${#backup_config_paths[@]} = 5 ] && [ "$backup_require_root" = 1 ] &&
		for p in "${backup_config_paths[@]}"; do case $p in */env*|*credentials*|*caddy/data*) exit 1;; esac; done' _ "$T/etc/leviosa/backup.conf"

# ===========================================================================
# 2. systemd units and schedules
# ===========================================================================
if command -v systemd-analyze >/dev/null; then
	check "systemd-analyze verify: services and timers" \
		bash -c 'cd "$1" && systemd-analyze verify --man=no ./*.service ./*.timer' _ "$T/systemd-weekly"
	check "daily timer: OnCalendar=*-*-* 03:00:00 Europe/Paris, Persistent, small RandomizedDelaySec" \
		bash -c 'grep -qx "OnCalendar=\*-\*-\* 03:00:00 Europe/Paris" "$1" && grep -qx "Persistent=true" "$1" && grep -qx "RandomizedDelaySec=5min" "$1"' _ "$T/systemd-weekly/leviosa-backup.timer"
	check "weekly timer: OnCalendar=Sun *-*-* 04:00:00 Europe/Paris" \
		grep -qx "OnCalendar=Sun \*-\*-\* 04:00:00 Europe/Paris" "$T/systemd-weekly/leviosa-snapshot.timer"
	next=$(systemd-analyze calendar "Sun *-*-* 04:00:00 Europe/Paris" 2>/dev/null | sed -n 's/^ *Next elapse: //p')
	check "systemd parses the weekly schedule (next: $next)" bash -c '[[ $1 == Sun* ]]' _ "$next"
	check "systemd parses the daily schedule" systemd-analyze calendar "*-*-* 03:00:00 Europe/Paris"
else
	skip "systemd-analyze not installed"
fi

# ===========================================================================
# 3. Scripts
# ===========================================================================
check "bash -n on the scripts and the library" \
	bash -c 'for f in "$1"/files/leviosa-*; do case $f in *.tmpl) continue;; esac; bash -n "$f" || exit 1; done' _ "$role"
if docker info >/dev/null 2>&1; then
	check "shellcheck: scripts, library, rendered backup.conf, this test" \
		docker run --rm -v "$role:/role:ro" -v "$T/etc/leviosa:/conf:ro" -w /role/files koalaman/shellcheck:stable -x \
		leviosa-backup-lib.sh leviosa-backup leviosa-snapshot leviosa-restore-test /conf/backup.conf /role/tests/run-tests.sh
else
	skip "docker not available: shellcheck"
fi

# ===========================================================================
# 4. End to end
# ===========================================================================
if [ "$static_only" = 1 ]; then
	skip "end-to-end (--static)"
elif ! docker info >/dev/null 2>&1; then
	skip "end-to-end: docker not available"
else
	dc() { docker compose -p "$PROJECT" -f "$PROD/docker-compose.yml" -f "$T/override.yml" --project-directory "$PROD" "$@"; }
	awsl() { docker exec "$LOCALSTACK" awslocal --region "$AWS_REGION" "$@"; }
	keys_under() { awsl s3api list-objects-v2 --bucket "$BUCKET" --prefix "$1/" --query 'Contents[].Key' --output text | tr '\t' '\n' | grep -v '^None$' || true; }
	fatal() { fail "$*"; exit 1; }

	echo "..    e2e: building $IMAGE, starting LocalStack and the stack (a few minutes)"
	docker network create --label "$LABEL" "$NAME" >/dev/null
	docker run -d --name "$LOCALSTACK" --label "$LABEL" --network "$NAME" --network-alias localstack \
		-e SERVICES=kms,s3 -e PERSISTENCE=0 -e AWS_DEFAULT_REGION="$AWS_REGION" "$LOCALSTACK_IMAGE" >/dev/null
	docker build --label "$LABEL" -t "$IMAGE" "$repo/backend" >"$T/build.log" 2>&1 || { show "$T/build.log" 40; fatal "backend image build"; }
	for _ in $(seq 60); do awsl kms list-keys >/dev/null 2>&1 && break; sleep 2; done
	key_id=$(awsl kms create-key --query KeyMetadata.KeyId --output text) || fatal "LocalStack KMS"
	awsl kms create-alias --alias-name "$KMS_ALIAS" --target-key-id "$key_id"
	awsl s3 mb "s3://$BUCKET" >/dev/null || fatal "LocalStack S3 bucket"
	pass "LocalStack (KMS $KMS_ALIAS, bucket $BUCKET) and $IMAGE ready"

	cat >"$T/override.yml" <<EOF
services:
  backend: {image: $IMAGE}
  migrate: {image: $IMAGE}
  admin: {image: $IMAGE}
networks:
  leviosa_network: {name: $NAME, external: true}
EOF
	# The data services first, as `deploy` starts them (the one-offs have no
	# depends_on).
	dc up -d --wait postgres rabbitmq redis >"$T/services.log" 2>&1 || { show "$T/services.log"; fatal "data services up"; }
	dc --profile oneoff run --rm migrate >"$T/migrate.log" 2>&1 || { show "$T/migrate.log"; fatal "migrate"; }
	dc --profile oneoff run --rm admin >"$T/admin.log" 2>&1 || { show "$T/admin.log"; fatal "bootstrap-admin"; }
	dc up -d backend >"$T/up.log" 2>&1 || { show "$T/up.log"; fatal "backend up"; }
	for _ in $(seq 90); do
		[ "$(docker inspect -f '{{.State.Health.Status}}' "$(dc ps -q backend)" 2>/dev/null)" = healthy ] && break
		sleep 2
	done
	users=$(dc exec -T postgres psql -U leviosa -d leviosa -Atqc 'select count(*) from auth.users')
	canary=$(dc exec -T postgres psql -U leviosa -d leviosa -Atqc 'select count(*) from encryption.canary' 2>/dev/null || echo 0)
	check "production stack up: $users encrypted auth.users record(s), pepper canary written ($canary)" \
		bash -c '[ "$1" -ge 1 ] && [ "$2" = 1 ]' _ "$users" "$canary"

	# The scripts as the role installed them, with a test configuration on
	# top of the rendered one: LocalStack, no root, the stub Infisical CLI.
	cat >"$T/test.conf" <<EOF
# shellcheck shell=bash
. "$T/etc/leviosa/backup.conf"
backup_require_root=0
backup_aws_endpoint_url=http://localstack:4566
backup_docker_network=$NAME
PATH="$T/stub:\$PATH"
EOF
	export LEVIOSA_BACKUP_CONF=$T/test.conf LEVIOSA_BACKUP_LIB=$T/lib/lib.sh
	run() { # run <log> <script> [args…]: exit status in $rc
		local log=$1
		shift
		rc=0
		"$T/bin/$1" "${@:2}" >"$T/out/$log.log" 2>&1 || rc=$?
	}
	today=$(date -u +%Y/%m/%d)

	# ── B1 ──
	run backup leviosa-backup
	[ "$rc" = 0 ] || show "$T/out/backup.log"
	check "B1 leviosa-backup exits 0" [ "$rc" = 0 ]
	dump_key=$(keys_under db)
	config_key=$(keys_under config)
	check "B1 the dump is at db/$today/leviosa_prod_<UTC>.dump ($dump_key)" \
		bash -c '[[ $1 =~ ^db/$2/leviosa_prod_[0-9]{8}T[0-9]{6}Z\.dump$ ]]' _ "$dump_key" "$today"
	check "B1 the archive is at config/$today/leviosa_config_<UTC>.tar.gz ($config_key)" \
		bash -c '[[ $1 =~ ^config/$2/leviosa_config_[0-9]{8}T[0-9]{6}Z\.tar\.gz$ ]]' _ "$config_key" "$today"
	check "B1 latest.dump kept locally, mode 0600" bash -c '[ "$(stat -c %a "$1")" = 600 ]' _ "$T/work/latest.dump"

	# ── B2 ──
	awsl s3 cp "s3://$BUCKET/$config_key" - >"$T/config.tar.gz"
	mkdir "$T/config"
	tar -xzf "$T/config.tar.gz" -C "$T/config"
	tar -tzf "$T/config.tar.gz" >"$T/config.list"
	check "B2 the archive holds both compose files, caddy/etc, the Agent config and templates, manifest.txt" \
		bash -c 'l=$1; for p in opt/leviosa/docker-compose.yml opt/leviosa-staging/docker-compose.yml opt/leviosa/caddy/etc/Caddyfile etc/infisical-agent/config.yaml etc/infisical-agent/templates/backend.env.tmpl manifest.txt; do grep -q "$p\$" "$l" || exit 1; done' _ "$T/config.list"
	check "B2 the manifest names the running backend image" grep -q "$PROJECT/backend: $IMAGE" "$T/config/manifest.txt"
	check "B2 no env file, no credentials directory, no token in the archive" \
		bash -c '! grep -E "(^|/)(env|credentials)(/|$)|\.env$|token-sink|client-(id|secret)" "$1"' _ "$T/config.list"
	check "B2 no fixture secret value anywhere in the archive" no_secret_in "$T/config"

	# ── B3 ──
	before=$(keys_under config | wc -l)
	printf '. "%s"\nbackup_config_paths+=("%s")\n' "$T/test.conf" "$PROD/env" >"$T/test-env.conf"
	rc=0; LEVIOSA_BACKUP_CONF=$T/test-env.conf "$T/bin/leviosa-backup" >"$T/out/backup-env.log" 2>&1 || rc=$?
	check "B3 an env directory in the allowlist is refused (exit $rc)" \
		bash -c '[ "$1" != 0 ] && grep -q "secret-looking path" "$2"' _ "$rc" "$T/out/backup-env.log"
	printf 'pepper=%s\n' "$PEPPER" >"$PROD/caddy/etc/leak.txt"
	run backup-leak leviosa-backup
	rm -f "$PROD/caddy/etc/leak.txt"
	check "B3 a secret value inside an allowlisted file is refused (exit $rc)" \
		bash -c '[ "$1" != 0 ] && grep -q "contains a secret value" "$2"' _ "$rc" "$T/out/backup-leak.log"
	check "B3 nothing was uploaded by the refused runs" [ "$(keys_under config | wc -l)" = "$before" ]

	# ── S1 ── a stub `infisical`: login checks the credential arrives through
	# the environment, export checks the token does, and prints a snapshot in
	# the template's format (the template itself: envcheck/snapshot_test.go).
	cat >"$T/stub/infisical" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$*" >>"$T/stub/argv.log"
case \$1 in
login)
	[ "\$INFISICAL_UNIVERSAL_AUTH_CLIENT_ID" = FixtureClientId-0000-1111 ] || exit 3
	[ "\$INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET" = FixtureClientSecret-not-real-2222 ] || exit 3
	echo stub-access-token-4444 ;;
export)
	[ "\$INFISICAL_TOKEN" = stub-access-token-4444 ] || exit 4
	tmpl=; for a in "\$@"; do case \$a in --template=*) tmpl=\${a#--template=} ;; esac; done
	grep -q 'listSecrets "$FAKE_PID" "prod" "/"' "\$tmpl" || exit 5
	printf '{"format":"leviosa-infisical-snapshot/1","project":"$FAKE_PID","environment":"prod","secrets":[{"path":"/encx","key":"ENCX_PEPPER","value":"%s","comment":""},{"path":"/backup","key":"BACKUP_AWS_SECRET_ACCESS_KEY","value":"BackupFixtureSecret-not-real-0001","comment":""}]}\n' "$PEPPER" ;;
*) exit 2 ;;
esac
EOF
	chmod 755 "$T/stub/infisical"
	run snapshot leviosa-snapshot
	[ "$rc" = 0 ] || show "$T/out/snapshot.log"
	check "S1 leviosa-snapshot exits 0" [ "$rc" = 0 ]
	snap_key=$(keys_under infisical)
	check "S1 the snapshot is at infisical/$today/leviosa_prod_infisical_<UTC>.json.age ($snap_key)" \
		bash -c '[[ $1 =~ ^infisical/$2/leviosa_prod_infisical_[0-9]{8}T[0-9]{6}Z\.json\.age$ ]]' _ "$snap_key" "$today"
	awsl s3 cp "s3://$BUCKET/$snap_key" - >"$T/snapshot.age"
	check "S1 the uploaded object is age ciphertext, without the pepper in clear" \
		bash -c 'head -c 21 "$1" | grep -q "age-encryption.org/v1" && ! grep -qF "$2" "$1"' _ "$T/snapshot.age" "$PEPPER"
	check "S1 unreadable without the private key (age -d, no identity)" bash -c '! age -d "$1" >/dev/null 2>&1' _ "$T/snapshot.age"
	check "S1 unreadable with another key" bash -c '! age -d -i "$2" "$1" >/dev/null 2>&1' _ "$T/snapshot.age" "$T/keys/other.key"
	check "S1 readable with the operator's private key: the pepper, its folder, takenAt" \
		bash -c 'age -d -i "$2" "$1" | jq -e --arg p "$3" ".takenAt and ([.secrets[] | select(.path == \"/encx\" and .key == \"ENCX_PEPPER\")][0].value == \$p)"' \
		_ "$T/snapshot.age" "$T/keys/operator.key" "$PEPPER"
	check "S1 the client secret and the token never on a command line" \
		bash -c '! grep -qE "FixtureClientSecret|FixtureClientId|stub-access-token" "$1"' _ "$T/stub/argv.log"
	check "S1 no plaintext left on disk (work dir holds latest.dump and locks only)" \
		bash -c '[ "$(cd "$1" && ls -A | grep -v "^\.leviosa-.*\.lock$" | tr "\n" " ")" = "latest.dump " ]' _ "$T/work"

	# ── S2 ──
	cp "$T/lib/backup-age.pub" "$T/keys/real.pub"
	cp "$role/files/backup-age.pub" "$T/lib/backup-age.pub"
	before=$(keys_under infisical | wc -l)
	run snapshot-placeholder leviosa-snapshot
	cp "$T/keys/real.pub" "$T/lib/backup-age.pub"
	check "S2 the placeholder recipient is refused (exit $rc)" \
		bash -c '[ "$1" != 0 ] && grep -q "holds no age public key" "$2"' _ "$rc" "$T/out/snapshot-placeholder.log"
	check "S2 … nothing uploaded" [ "$(keys_under infisical | wc -l)" = "$before" ]

	# ── R1–R3 ──
	run restore leviosa-restore-test
	[ "$rc" = 0 ] || show "$T/out/restore.log" 40
	check "R1 leviosa-restore-test passes on the newest dump (exit $rc)" [ "$rc" = 0 ]
	check "R1 … verify decrypted an auth.users record and matched the pepper canary" \
		bash -c 'grep -q "verify: OK: auth.users record .* decrypted" "$1" && grep -q "pepper canary matched" "$1" && grep -q "^leviosa-restore-test: PASS" "$1"' _ "$T/out/restore.log"
	sed "s|^ENCX_PEPPER=.*|ENCX_PEPPER=$(head -c 32 /dev/urandom | base64)|" "$PROD/env/backend.env" >"$T/wrong-pepper.env"
	run restore-wrong-pepper leviosa-restore-test --backend-env "$T/wrong-pepper.env"
	check "R2 … fails with verify's exit 6 given a wrong pepper (exit $rc)" \
		bash -c '[ "$1" = 6 ] && grep -q "FAIL (verify exit 6): wrong pepper" "$2"' _ "$rc" "$T/out/restore-wrong-pepper.log"
	run restore-local leviosa-restore-test --local
	check "R3 … --local (latest.dump) passes (exit $rc)" [ "$rc" = 0 ]
	run restore-key leviosa-restore-test "$dump_key"
	check "R3 … an explicit key passes (exit $rc)" [ "$rc" = 0 ]
	check "R3 no restore-test container or network left behind" \
		bash -c '[ -z "$(docker ps -aq --filter label=leviosa.restore-test)" ] && [ -z "$(docker network ls -q --filter name=leviosa-restore-test-)" ]'
	check "R3 the restore test never published a host port" \
		bash -c '! grep -q -- "-p \|--publish" "$1"' _ "$role/files/leviosa-restore-test"

	# ── L1 ──
	check "L1 no fixture secret in any script output (backup, snapshot, restore test)" no_secret_in "$T/out"
	check "L1 … nor in the generated pepper of R2" bash -c '! grep -rqF "$(sed -n "s/^ENCX_PEPPER=//p" "$1")" "$2"' _ "$T/wrong-pepper.env" "$T/out"
fi

echo
echo "passed=$passed failed=$failed skipped=$skipped"
if [ "$failed" = 0 ]; then
	echo "RESULT: OK"
else
	echo "RESULT: FAILED"
	exit 1
fi
