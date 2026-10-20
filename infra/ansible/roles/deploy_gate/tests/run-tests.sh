#!/usr/bin/env bash
# Tests for roles/deploy_gate: no root, no Ansible, no server, no network.
#
#   infra/ansible/roles/deploy_gate/tests/run-tests.sh      (or: make deploy-gate-test)
#
# The scripts' and the sudoers templates use plain `{{ variable }}`
# substitutions only, so this renders them exactly as Ansible would, with test
# values, and runs them:
#
#   gate             the CI key's forced command against a table of SSH
#                    original commands; `sudo` and `logger` are stubs on PATH
#                    that record their arguments
#   runner           a fixture repository served over file:// (main, a feature
#                    branch, a pull-request ref): commits on main deploy,
#                    anything else is refused; the fixture `deploy` records its
#                    arguments and environment; the compose file lands in the
#                    stack root
#   sudoers          the only rule is the runner; `visudo -c` accepts the file
#   authorized_keys  every CI key line is `restrict,command="<gate>" <key>`
#                    (rendered with Python's jinja2, skipped without it)
#
# Needs bash, git, flock, coreutils (`env -S`). Prints a PASS/FAIL/SKIP line
# per check and ends with `RESULT: OK` (exit 0) or `RESULT: FAILED` (exit 1).
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
templates=$here/../templates

work=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-deploy-gate-test.XXXXXX")
trap 'rm -rf "$work"' EXIT

passed=0
failed=0
skipped=0
pass() {
	passed=$((passed + 1))
	printf 'PASS  %s\n' "$*"
}
fail() {
	failed=$((failed + 1))
	printf 'FAIL  %s\n' "$*"
}
skip() {
	skipped=$((skipped + 1))
	printf 'SKIP  %s\n' "$*"
}

# render <template> <output> [name=value ...]
# Replaces each `{{ name }}` with value and refuses to write anything that
# still holds Jinja syntax, so a template that grows a filter, a loop or a
# stray `${#…}` (a Jinja comment opener) fails here instead of rendering
# differently under Ansible.
render() {
	local src=$1 out=$2 text kv name value
	shift 2
	text=$(<"$src")
	for kv in "$@"; do
		name=${kv%%=*}
		value=${kv#*=}
		text=${text//"{{ $name }}"/"$value"}
	done
	case $text in
		*'{{'* | *'{%'* | *'{#'*)
			echo "render: $src still holds Jinja syntax after substitution" >&2
			return 1
			;;
	esac
	printf '%s\n' "$text" >"$out"
	chmod 0755 "$out"
}

# ── gate ──────────────────────────────────────────────────────────────────────
runner_path=/usr/local/sbin/leviosa-deploy
gate=$work/bin/leviosa-deploy-gate
stubs=$work/stubs
mkdir -p "$work/bin" "$stubs"
render "$templates/leviosa-deploy-gate.j2" "$gate" "deploy_gate_runner=$runner_path"

# sudo records its arguments, one per line, and exits with $STUB_EXIT;
# logger appends its message to logger.log.
cat >"$stubs/sudo" <<EOF
#!/bin/sh
for a in "\$@"; do printf '%s\n' "\$a"; done >"$work/sudo.argv"
exit "\${STUB_EXIT:-0}"
EOF
cat >"$stubs/logger" <<EOF
#!/bin/sh
printf '%s\n' "\$*" >>"$work/logger.log"
EOF
chmod 0755 "$stubs/sudo" "$stubs/logger"

sha=0123456789abcdef0123456789abcdef01234567

# run_gate [command]: runs the gate as sshd would (an empty environment apart
# from PATH and, when given, SSH_ORIGINAL_COMMAND), with a non-C locale the
# client could have sent (Ubuntu's sshd accepts LANG and LC_*).
run_gate() {
	rm -f "$work/sudo.argv" "$work/logger.log" "$work/gate.err"
	local env=(env -i "PATH=$stubs:/usr/bin:/bin" LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 "STUB_EXIT=${STUB_EXIT:-0}")
	[ $# -eq 0 ] || env+=("SSH_ORIGINAL_COMMAND=$1")
	"${env[@]}" "$gate" 2>"$work/gate.err" >/dev/null
}

# accept <description> <command> <stack>
accept() {
	local desc=$1 cmd=$2 stack=$3 rc=0 want
	run_gate "$cmd" || rc=$?
	want=$(printf '%s\n' -n "$runner_path" "$stack" "$sha")
	if [ $rc -ne 0 ]; then
		fail "gate accepts $desc (exit $rc: $(cat "$work/gate.err"))"
	elif [ "$(cat "$work/sudo.argv" 2>/dev/null)" != "$want" ]; then
		fail "gate accepts $desc (sudo argv: $(tr '\n' ' ' <"$work/sudo.argv" 2>/dev/null))"
	else
		pass "gate accepts $desc -> sudo -n $runner_path $stack <sha>"
	fi
}

# refuse <description> [command]   (no command: SSH_ORIGINAL_COMMAND unset)
refuse() {
	local desc=$1 rc=0
	shift
	run_gate "$@" || rc=$?
	if [ $rc -eq 0 ]; then
		fail "gate refuses $desc (exit 0)"
	elif [ -e "$work/sudo.argv" ]; then
		fail "gate refuses $desc (sudo was called: $(tr '\n' ' ' <"$work/sudo.argv"))"
	elif ! grep -q 'refused' "$work/gate.err"; then
		fail "gate refuses $desc (no refusal message on stderr)"
	elif ! grep -q 'refused' "$work/logger.log" 2>/dev/null; then
		fail "gate refuses $desc (refusal not logged)"
	else
		pass "gate refuses $desc (exit $rc)"
	fi
}

accept "a staging deploy" "deploy staging $sha" staging
accept "a production deploy" "deploy prod $sha" prod

rc=0
STUB_EXIT=3 run_gate "deploy prod $sha" || rc=$?
if [ $rc -eq 3 ]; then
	pass "gate returns the runner's exit status"
else
	fail "gate returns the runner's exit status (got $rc, want 3)"
fi

refuse "no command (interactive shell)"
refuse "an empty command" ""
refuse "a shell" "bash"
refuse "an interactive login shell" "/bin/bash -il"
refuse "another command" "id"
refuse "the runner called directly" "$runner_path prod $sha"
refuse "sudo" "sudo -n $runner_path prod $sha"
refuse "a shell wrapping a deploy" "sh -c 'deploy prod $sha'"
refuse "scp upload" "scp -t /tmp"
refuse "scp download" "scp -f /etc/shadow"
refuse "sftp (OpenSSH server)" "/usr/lib/openssh/sftp-server"
refuse "sftp (internal)" "internal-sftp"
refuse "rsync" "rsync --server -vlogDtpre.iLsfxC . /tmp"
refuse "git upload-pack" "git-upload-pack '/opt/leviosa'"
refuse "deploy alone" "deploy"
refuse "deploy without a SHA" "deploy prod"
refuse "deploy without a stack" "deploy $sha"
refuse "an unknown stack" "deploy dev $sha"
refuse "the stack spelled out" "deploy production $sha"
refuse "an uppercase stack" "deploy PROD $sha"
refuse "an uppercase verb" "DEPLOY prod $sha"
refuse "both stacks" "deploy staging prod $sha"
refuse "a short SHA (7)" "deploy prod ${sha:0:7}"
refuse "a 39-character SHA" "deploy prod ${sha:0:39}"
refuse "a 41-character SHA" "deploy prod ${sha}0"
refuse "an uppercase SHA" "deploy prod 0123456789ABCDEF0123456789ABCDEF01234567"
refuse "a mixed-case SHA" "deploy prod 0123456789abcdeF0123456789abcdef01234567"
refuse "a non-hex SHA" "deploy prod 0123456789abcdeg0123456789abcdef01234567"
refuse "a branch name" "deploy prod main"
refuse "a revision expression" "deploy prod HEAD~1"
refuse "an extra argument" "deploy prod $sha extra"
refuse "an extra option" "deploy prod $sha --force"
refuse "an option before the stack" "deploy --force prod $sha"
refuse "a ; command" "deploy prod $sha; id"
refuse "a && command" "deploy prod $sha && id"
refuse "a || command" "deploy prod $sha || id"
refuse "a pipe" "deploy prod $sha | sh"
refuse "a background &" "deploy prod $sha &"
refuse "a redirection" "deploy prod $sha > /tmp/x"
refuse "a command substitution" "deploy prod \$(id)"
refuse "backticks" "deploy prod \`id\`"
refuse "a glob" "deploy prod *"
refuse "a variable" "deploy \$STACK $sha"
refuse "a quoted stack" "deploy 'prod' $sha"
refuse "a newline and a second command" "deploy prod $sha"$'\n'"id"
refuse "a trailing newline" "deploy prod $sha"$'\n'
refuse "a trailing carriage return" "deploy prod $sha"$'\r'
refuse "a leading space" " deploy prod $sha"
refuse "a trailing space" "deploy prod $sha "
refuse "two spaces between words" "deploy  prod $sha"
refuse "a tab between words" "deploy"$'\t'"prod $sha"
refuse "a non-breaking space between words" "deploy"$' '"prod $sha"

# ── runner ────────────────────────────────────────────────────────────────────
# A bare "origin" served over file:// stands in for GitHub: `main`, a
# `feature` branch, and a pull-request ref (GitHub serves fork commits under
# the parent repository's URL). The test's own git ignores the developer's
# configuration (signing, hooks, default branch).
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null GIT_TERMINAL_PROMPT=0
origin=$work/origin.git
clone=$work/clone
record=$work/deployed
state=$work/state
prod_root=$work/opt/leviosa
staging_root=$work/opt/leviosa-staging
runner=$work/bin/leviosa-deploy
mkdir -p "$state" "$prod_root"
chmod 0700 "$state"
render "$templates/leviosa-deploy.j2" "$runner" \
	"deploy_gate_repo_url=file://$origin" \
	"deploy_gate_state_dir=$state" \
	"deploy_gate_stack_roots.prod=$prod_root" \
	"deploy_gate_stack_roots.staging=$staging_root"

g() { git -C "$clone" -c user.name=test -c user.email=test@example.invalid -c commit.gpgsign=false "$@"; }
git init --quiet --bare --initial-branch=main "$origin"
git init --quiet --initial-branch=main "$clone"
g remote add origin "$origin"
mkdir -p "$clone/infra/compose"

# commit <message> <deploy exit status> <prod compose> [<staging compose>]
# The fixture `deploy` records its argv and environment in $record and exits
# with the given status. A bash script, like the real one: a leaked BASH_ENV
# would run before it.
commit() {
	local message=$1 status=$2 prod=$3 staging=${4-}
	cat >"$clone/infra/compose/deploy" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$@" >"$record/argv"
env >"$record/env"
exit $status
EOF
	chmod 0755 "$clone/infra/compose/deploy"
	printf '%s\n' "$prod" >"$clone/infra/compose/docker-compose.prod.yml"
	if [ -n "$staging" ]; then
		printf '%s\n' "$staging" >"$clone/infra/compose/docker-compose.staging.yml"
	else
		rm -f "$clone/infra/compose/docker-compose.staging.yml"
	fi
	g add -A
	g commit --quiet -m "$message"
	g rev-parse HEAD
}

c1=$(commit "main: prod only" 0 "# prod at c1")
c2=$(commit "main: prod and staging" 0 "# prod at c2" "# staging at c2")
g push --quiet origin main
g checkout --quiet -b feature
feature=$(commit "feature: not on main" 0 "# prod at feature" "# staging at feature")
g push --quiet origin feature
g checkout --quiet main
fork=$(commit "fork: a pull request" 0 "# prod at fork" "# staging at fork")
g push --quiet origin "HEAD:refs/pull/1/head"
g reset --quiet --hard "$c2"
failing=$(commit "main: deploy fails" 7 "# prod at failing" "# staging at failing")
g push --quiet origin main
unknown=fedcba9876543210fedcba9876543210fedcba98

# The caller's environment, which must reach neither the runner's tools nor
# the deploy script.
cat >"$work/bash_env.sh" <<EOF
touch "$work/bash_env.ran"
EOF
caller_env=(
	LEVIOSA_TEST_SENTINEL=leaked
	"LEVIOSA_PROD_ROOT=$work/elsewhere"
	"BASH_ENV=$work/bash_env.sh"
	"GIT_DIR=$work/not-a-repo"
	"GIT_CONFIG_GLOBAL=$work/not-a-config"
	"HOME=$work"
)

# run_runner <stack> <sha> [extra args]: runs the rendered runner as the
# deploy user's sudo would, with a polluted caller environment.
run_runner() {
	rm -rf "$record" "$work/bash_env.ran"
	mkdir -p "$record"
	env "${caller_env[@]}" "$runner" "$@" >"$work/runner.out" 2>&1
}

installed() { cat "$1/docker-compose.yml" 2>/dev/null || echo "<none>"; }

# expect_deployed <description> <stack> <sha> <stack root> <compose content>
expect_deployed() {
	local desc=$1 stack=$2 sha=$3 root=$4 compose=$5 rc=0
	run_runner "$stack" "$sha" || rc=$?
	if [ $rc -ne 0 ]; then
		fail "runner $desc (exit $rc: $(tail -n 3 "$work/runner.out"))"
	elif [ "$(cat "$record/argv" 2>/dev/null)" != "$(printf '%s\n' "$stack" "$sha")" ]; then
		fail "runner $desc (deploy argv: $(tr '\n' ' ' <"$record/argv" 2>/dev/null || echo 'not run'))"
	elif [ "$(installed "$root")" != "$compose" ]; then
		fail "runner $desc (installed compose: $(installed "$root"))"
	elif [ "$(stat -c %a "$root/docker-compose.yml")" != 644 ]; then
		fail "runner $desc (compose file mode $(stat -c %a "$root/docker-compose.yml"), want 644)"
	else
		pass "runner $desc"
	fi
}

# expect_refused <description> <message> <args...>: non-zero, the message,
# no deploy run and the stack roots untouched.
expect_refused() {
	local desc=$1 message=$2 rc=0 before_prod before_staging
	shift 2
	before_prod=$(installed "$prod_root")
	before_staging=$(installed "$staging_root")
	run_runner "$@" || rc=$?
	if [ $rc -eq 0 ]; then
		fail "runner refuses $desc (exit 0)"
	elif [ -e "$record/argv" ]; then
		fail "runner refuses $desc (deploy ran)"
	elif ! grep -qF -- "$message" "$work/runner.out"; then
		fail "runner refuses $desc (no '$message' in: $(tail -n 3 "$work/runner.out"))"
	elif [ "$(installed "$prod_root")" != "$before_prod" ] || [ "$(installed "$staging_root")" != "$before_staging" ]; then
		fail "runner refuses $desc (a compose file changed)"
	else
		pass "runner refuses $desc"
	fi
}

expect_deployed "deploys a commit on main" prod "$c2" "$prod_root" "# prod at c2"

allowed='^(PATH|HOME|PWD|OLDPWD|SHLVL|_)$'
extra=$(cut -d= -f1 "$record/env" | grep -Ev "$allowed" | tr '\n' ' ' || true)
if [ -n "$extra" ]; then
	fail "deploy runs with an empty environment (also got: $extra)"
elif ! grep -qx 'PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin' "$record/env" ||
	! grep -qx 'HOME=/root' "$record/env"; then
	fail "deploy runs with an empty environment (PATH/HOME: $(grep -E '^(PATH|HOME)=' "$record/env" | tr '\n' ' '))"
elif [ -e "$work/bash_env.ran" ]; then
	fail "deploy runs with an empty environment (the caller's BASH_ENV ran)"
else
	pass "deploy runs with an empty environment (PATH and HOME only; no caller variable)"
fi

if [ "$(stat -c %a "$state/repo.git")" = 700 ]; then
	pass "runner keeps its mirror root-only (0700)"
else
	fail "runner keeps its mirror root-only (mode $(stat -c %a "$state/repo.git"))"
fi

expect_refused "a stack whose root is missing" "stack root $staging_root missing" staging "$c2"
mkdir -p "$staging_root"
expect_deployed "deploys staging into its own root" staging "$c2" "$staging_root" "# staging at c2"
if [ "$(installed "$prod_root")" = "# prod at c2" ]; then
	pass "runner leaves the other stack's compose file alone"
else
	fail "runner leaves the other stack's compose file alone (prod now: $(installed "$prod_root"))"
fi
expect_deployed "rolls back to an older commit of main, compose file included" prod "$c1" "$prod_root" "# prod at c1"

expect_refused "a commit only on another branch" "not a commit on main" prod "$feature"
expect_refused "a pull-request commit served under the repo's URL" "not a commit on main" prod "$fork"
expect_refused "an unknown SHA" "not a commit on main" prod "$unknown"
expect_refused "a commit of main without that stack's compose file" "has no infra/compose/docker-compose.staging.yml" staging "$c1"

rc=0
run_runner prod "$failing" || rc=$?
if [ $rc -eq 7 ] && [ "$(installed "$prod_root")" = "# prod at failing" ]; then
	pass "runner returns the deploy script's exit status"
else
	fail "runner returns the deploy script's exit status (got $rc, want 7)"
fi

# main moves backwards (force-push): a commit the mirror already holds but
# main no longer contains is refused.
g reset --quiet --hard "$c2"
g push --quiet --force origin main
expect_refused "a commit that left main (force-push)" "not a commit on main" prod "$failing"

rc=0
flock "$state/prod.lock" env "$runner" prod "$c2" >"$work/runner.out" 2>&1 || rc=$?
if [ $rc -ne 0 ] && grep -q 'another prod deploy is running' "$work/runner.out"; then
	pass "runner refuses a second deploy of the same stack while one runs"
else
	fail "runner refuses a second deploy of the same stack while one runs (exit $rc: $(tail -n 2 "$work/runner.out"))"
fi

expect_refused "an unknown stack" "unknown stack" dev "$c2"
expect_refused "a short SHA" "not a full 40-character" prod "${c2:0:7}"
expect_refused "an uppercase SHA" "not a full 40-character" prod "$(tr 'a-f' 'A-F' <<<"$c2")"
expect_refused "a branch name" "not a full 40-character" prod main
expect_refused "an extra argument" "usage:" prod "$c2" extra
expect_refused "no arguments" "usage:"

# ── sudoers ───────────────────────────────────────────────────────────────────
sudoers=$work/leviosa-deploy.sudoers
render "$templates/leviosa-deploy.sudoers.j2" "$sudoers" \
	deploy_gate_user=leviosa "deploy_gate_runner=$runner_path"
rules=$(grep -Ev '^[[:space:]]*(#|$)' "$sudoers" || true)
if [ "$rules" = "leviosa ALL=(root) NOPASSWD: $runner_path" ]; then
	pass "sudoers grants the deploy user the runner and nothing else"
else
	fail "sudoers grants the deploy user the runner and nothing else (rules: $(tr '\n' '|' <<<"$rules"))"
fi
visudo=$(command -v visudo || echo /usr/sbin/visudo)
if [ -x "$visudo" ]; then
	if "$visudo" -cf "$sudoers" >/dev/null 2>&1; then
		pass "sudoers drop-in passes visudo -c"
	else
		fail "sudoers drop-in passes visudo -c ($("$visudo" -cf "$sudoers" 2>&1 | tail -n 1))"
	fi
else
	skip "sudoers drop-in passes visudo -c (no visudo here)"
fi

# ── authorized_keys ───────────────────────────────────────────────────────────
# The CI keys' lines are a Jinja loop, so this needs Python's jinja2 (present
# wherever Ansible is); rendered with trim_blocks, as Ansible's template lookup.
gate_path=/usr/local/sbin/leviosa-deploy-gate
keys=$work/ci-keys
if ! command -v ssh-keygen >/dev/null || ! python3 -I -c 'import jinja2' 2>/dev/null; then
	skip "authorized_keys forces every CI key to the gate (needs ssh-keygen and python3 jinja2)"
else
	for n in 1 2; do
		ssh-keygen -q -t ed25519 -N '' -C "leviosa-ci-$n" -f "$work/ci$n" </dev/null
		cat "$work/ci$n.pub"
	done >"$keys"
	python3 -I - "$templates/authorized_keys.ci.j2" "$gate_path" "$keys" >"$work/authorized_keys" <<'PY'
import sys
import jinja2

template, command, keys = sys.argv[1:4]
with open(template) as f:
    source = f.read()
with open(keys) as f:
    ci_keys = [line.strip() for line in f if line.strip()]
env = jinja2.Environment(trim_blocks=True, undefined=jinja2.StrictUndefined)
sys.stdout.write(env.from_string(source).render(deploy_gate_ci_keys=ci_keys, deploy_gate_command=command))
PY
	lines=$(grep -Ev '^[[:space:]]*(#|$)' "$work/authorized_keys" || true)
	want=$(sed "s|^|restrict,command=\"$gate_path\" |" "$keys")
	if [ "$lines" != "$want" ]; then
		fail "authorized_keys forces every CI key to the gate (got: $(tr '\n' '|' <<<"$lines"))"
	elif [ "$(ssh-keygen -lf "$work/authorized_keys" 2>/dev/null | wc -l)" -ne 2 ]; then
		fail "authorized_keys forces every CI key to the gate (ssh-keygen cannot parse every line)"
	else
		pass "authorized_keys forces every CI key to the gate (restrict,command=\"$gate_path\")"
	fi
fi

# ── result ────────────────────────────────────────────────────────────────────
echo
echo "$passed passed, $failed failed, $skipped skipped"
if [ "$failed" -eq 0 ]; then
	echo "RESULT: OK"
else
	echo "RESULT: FAILED"
	exit 1
fi
