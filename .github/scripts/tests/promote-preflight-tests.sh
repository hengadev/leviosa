#!/usr/bin/env bash
# Tests for .github/scripts/promote-preflight.sh: no network beyond loopback,
# no Docker, no GitHub, no secrets.
#
#   .github/scripts/tests/promote-preflight-tests.sh      (or: make promote-preflight-test)
#
#   staging   a fake /healthz (python3 http.server on a random loopback port)
#             answers with the status and body of the current case and logs
#             every request
#   images    a fake `docker` on PATH: `docker buildx imagetools inspect <ref>`
#             succeeds only for the refs listed for the case, and records every
#             ref it was asked for
#   main      a fixture repository: an "origin" with two commits on `main` and
#             one on a `feature` branch, cloned with full history so
#             `origin/main` exists as in the workflow's checkout
#
# A refused case must exit non-zero and write no `sha=` output, so the deploy
# job (which needs that output) never runs.
#
# Needs bash, git, curl, jq, python3. Prints a PASS/FAIL line per check and
# ends with `RESULT: OK` (exit 0) or `RESULT: FAILED` (exit 1).
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
script=$here/../promote-preflight.sh

for tool in git curl jq python3; do
	if ! command -v "$tool" >/dev/null; then
		echo "promote-preflight-tests: $tool is required" >&2
		exit 1
	fi
done

work=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-promote-preflight-test.XXXXXX")
server_pid=
cleanup() {
	if [ -n "$server_pid" ]; then
		kill "$server_pid" 2>/dev/null || true
		wait "$server_pid" 2>/dev/null || true
	fi
	rm -rf "$work"
}
trap cleanup EXIT

passed=0
failed=0
pass() {
	passed=$((passed + 1))
	printf 'PASS  %s\n' "$*"
}
fail() {
	failed=$((failed + 1))
	printf 'FAIL  %s\n' "$*"
}

# ── fixture repository ────────────────────────────────────────────────────────
# The test's own git ignores the developer's configuration (signing, hooks,
# default branch).
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null GIT_TERMINAL_PROMPT=0
origin=$work/origin.git
seed=$work/seed
clone=$work/clone
g() { git -C "$seed" -c user.name=test -c user.email=test@example.invalid -c commit.gpgsign=false "$@"; }
git init --quiet --bare --initial-branch=main "$origin"
git init --quiet --initial-branch=main "$seed"
g remote add origin "$origin"
commit() {
	printf '%s\n' "$1" >"$seed/release.txt"
	g add -A
	g commit --quiet -m "$1"
	g rev-parse HEAD
}
older=$(commit "main: an older release")
latest=$(commit "main: the latest release")
g push --quiet origin main
g checkout --quiet -b feature
feature=$(commit "feature: not on main")
g push --quiet origin feature
git clone --quiet "file://$origin" "$clone"
git clone --quiet --depth 1 "file://$origin" "$work/shallow"
unknown=0123456789abcdef0123456789abcdef01234567

# ── fake staging /healthz ─────────────────────────────────────────────────────
staging=$work/staging
mkdir -p "$staging"
cat >"$work/fake-healthz.py" <<'PY'
import http.server
import os
import sys

root = sys.argv[1]


def read(name):
    with open(os.path.join(root, name), "rb") as f:
        return f.read()


class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        with open(os.path.join(root, "requests.log"), "a") as f:
            f.write(self.path + "\n")
        status = int(read("status").decode().strip())
        body = read("body")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


server = http.server.HTTPServer(("127.0.0.1", 0), Handler)
with open(os.path.join(root, "port.tmp"), "w") as f:
    f.write(str(server.server_port))
os.rename(os.path.join(root, "port.tmp"), os.path.join(root, "port"))
server.serve_forever()
PY
printf '200\n' >"$staging/status"
printf '{}' >"$staging/body"
python3 -I "$work/fake-healthz.py" "$staging" &
server_pid=$!
for _ in $(seq 1 50); do
	[ -s "$staging/port" ] && break
	sleep 0.1
done
if [ ! -s "$staging/port" ]; then
	echo "promote-preflight-tests: the fake /healthz did not start" >&2
	exit 1
fi
healthz_url="http://127.0.0.1:$(cat "$staging/port")/healthz"

# A loopback port nothing listens on: bind one, close it.
closed_port=$(python3 -I -c 'import socket; s = socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')

# serve <status> <body>: what staging's /healthz answers next.
serve() {
	printf '%s\n' "$1" >"$staging/status"
	printf '%s' "$2" >"$staging/body"
}
healthz() { printf '{"frontend":"%s","backend":"%s"}' "$1" "$2"; }

# ── fake docker ───────────────────────────────────────────────────────────────
stubs=$work/stubs
mkdir -p "$stubs"
cat >"$stubs/docker" <<EOF
#!/usr/bin/env bash
# docker buildx imagetools inspect <ref>: exit 0 when <ref> is listed in
# images.txt, 1 (with Docker's message) otherwise.
if [ "\$#" -ne 4 ] || [ "\$1 \$2 \$3" != "buildx imagetools inspect" ]; then
	echo "fake docker: unexpected arguments: \$*" >&2
	exit 64
fi
printf '%s\n' "\$4" >>"$work/docker.log"
if grep -qxF -- "\$4" "$work/images.txt"; then
	echo "Name: docker.io/\$4"
	exit 0
fi
echo "ERROR: docker.io/\$4: not found" >&2
exit 1
EOF
chmod 0755 "$stubs/docker"

# images <sha> [backend|frontend|both|none]: which images exist on "Docker Hub".
images() {
	local sha=$1 which=${2:-both}
	: >"$work/images.txt"
	case $which in
		both | backend) printf 'henga/leviosa-backend:%s\n' "$sha" >>"$work/images.txt" ;;
	esac
	case $which in
		both | frontend) printf 'henga/leviosa-frontend:%s\n' "$sha" >>"$work/images.txt" ;;
	esac
}

# ── running the script ────────────────────────────────────────────────────────
# run [dir] [NAME=value ...]: runs the script from a checkout as the workflow
# does, with an empty environment apart from PATH, HOME, the GitHub files and
# the given variables (so nothing leaks in from the developer's shell).
rc=0
run() {
	local dir=$clone
	if [ $# -gt 0 ] && [ -d "$1" ]; then
		dir=$1
		shift
	fi
	rm -f "$work/docker.log" "$staging/requests.log"
	: >"$work/output"
	: >"$work/summary"
	rc=0
	(
		cd "$dir"
		env -i "PATH=$stubs:/usr/local/bin:/usr/bin:/bin" "HOME=$work" \
			GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null \
			"GITHUB_OUTPUT=$work/output" "GITHUB_STEP_SUMMARY=$work/summary" \
			"STAGING_HEALTHZ_URL=$healthz_url" \
			"$@" bash "$script"
	) >"$work/log" 2>&1 || rc=$?
}

log_excerpt() { tr '\n' ' ' <"$work/log" | cut -c1-300; }

# promoted <description> <sha> <source>: the last run passed and handed the
# deploy job exactly that SHA, after checking both images.
promoted() {
	local desc=$1 sha=$2 source=$3 want_images
	want_images=$(printf 'henga/leviosa-backend:%s\nhenga/leviosa-frontend:%s' "$sha" "$sha")
	if [ $rc -ne 0 ]; then
		fail "$desc (exit $rc: $(log_excerpt))"
	elif [ "$(cat "$work/output")" != "$(printf 'sha=%s\nsource=%s' "$sha" "$source")" ]; then
		fail "$desc (GITHUB_OUTPUT: $(tr '\n' ' ' <"$work/output"))"
	elif [ "$(cat "$work/docker.log" 2>/dev/null)" != "$want_images" ]; then
		fail "$desc (images checked: $(tr '\n' ' ' <"$work/docker.log" 2>/dev/null))"
	elif ! grep -qF "$sha" "$work/summary"; then
		fail "$desc (the job summary does not show the SHA)"
	else
		pass "$desc"
	fi
}

# refused <description> <message>: the last run stopped with a non-zero exit,
# an ::error:: annotation containing <message>, and no SHA for the deploy job.
refused() {
	local desc=$1 message=$2
	if [ $rc -eq 0 ]; then
		fail "$desc (exit 0)"
	elif grep -q '^sha=' "$work/output"; then
		fail "$desc (a SHA was handed to the deploy job: $(tr '\n' ' ' <"$work/output"))"
	elif ! grep -q '^::error::' "$work/log"; then
		fail "$desc (no ::error:: annotation: $(log_excerpt))"
	elif ! grep -qiF -- "$message" "$work/log"; then
		fail "$desc (no \"$message\" in: $(log_excerpt))"
	else
		pass "$desc (exit $rc)"
	fi
}

# no_image_check <description>: the last run stopped before asking Docker Hub.
no_image_check() {
	if [ -s "$work/docker.log" ]; then
		fail "$1 (images were checked: $(tr '\n' ' ' <"$work/docker.log"))"
	else
		pass "$1"
	fi
}

# ── without input: the release staging serves ────────────────────────────────
serve 200 "$(healthz "$latest" "$latest")"
images "$latest"
run
promoted "no input: promotes the SHA both staging apps report" "$latest" staging
if [ "$(cat "$staging/requests.log" 2>/dev/null)" = /healthz ]; then
	pass "no input: asks staging's /healthz once"
else
	fail "no input: asks staging's /healthz once (requests: $(tr '\n' ' ' <"$staging/requests.log" 2>/dev/null))"
fi

run PROMOTE_SHA=
promoted "an empty sha input reads staging" "$latest" staging

serve 200 "$(healthz "$latest" "$older")"
images "$latest"
run
refused "frontend and backend on different SHAs" "frontend $latest"
no_image_check "a SHA mismatch stops before the image check"

serve 200 "$(healthz "$older" "$latest")"
run
refused "backend ahead of frontend" "backend $latest"

serve 200 '{"frontend":"dev","backend":"dev"}'
run
refused "staging reports a non-SHA version" "not a full 40-character commit SHA"

serve 200 "$(printf '{"frontend":"%s","backend":null}' "$latest")"
run
refused "staging's backend reports no SHA" "backend"

serve 200 "$(printf '{"frontend":"%s"}' "$latest")"
run
refused "staging's answer has no backend field" "backend"

serve 200 "$(healthz "${latest:0:7}" "${latest:0:7}")"
run
refused "staging reports a short SHA" "not a full 40-character commit SHA"

upper=$(tr 'a-f' 'A-F' <<<"$latest")
serve 200 "$(healthz "$upper" "$upper")"
run
refused "staging reports an uppercase SHA" "not a full 40-character commit SHA"

serve 502 "$(printf '{"frontend":"%s","backend":null,"error":"backend unreachable"}' "$latest")"
run
refused "staging unhealthy (502)" "502"
if grep -qF "backend unreachable" "$work/log"; then
	pass "a 502 shows staging's error"
else
	fail "a 502 shows staging's error ($(log_excerpt))"
fi
no_image_check "an unhealthy staging stops before the image check"

serve 401 'Unauthorized'
run
refused "staging /healthz behind basic auth (401)" "401"

serve 200 '<html>not json</html>'
run
refused "staging answers something that is not JSON" "frontend"

serve 200 '[]'
run
refused "staging answers JSON that is not an object" "frontend"

serve 200 "$(healthz "$feature" "$feature")"
images "$feature"
run
refused "staging reports a commit that is not on main" "not on main"
no_image_check "a commit off main stops before the image check"

serve 200 "$(healthz "$unknown" "$unknown")"
images "$unknown"
run
refused "staging reports a commit the repository does not have" "not a commit"

run "STAGING_HEALTHZ_URL=http://127.0.0.1:$closed_port/healthz"
refused "staging unreachable" "could not reach"

# A value is shown in the log only after sanitising: a newline cannot start a
# workflow command of its own.
serve 200 '{"frontend":"x\n::warning::injected","backend":"x"}'
run
refused "a value with a newline is refused" "frontend"
if grep -q '^::warning::' "$work/log"; then
	fail "a value from /healthz cannot inject a workflow command ($(log_excerpt))"
else
	pass "a value from /healthz cannot inject a workflow command"
fi

# ── images ────────────────────────────────────────────────────────────────────
serve 200 "$(healthz "$latest" "$latest")"
images "$latest" frontend
run
refused "backend image missing" "henga/leviosa-backend:$latest"

images "$latest" backend
run
refused "frontend image missing" "henga/leviosa-frontend:$latest"

images "$latest" none
run
refused "both images missing" "henga/leviosa-frontend:$latest"
if grep -qF "henga/leviosa-backend:$latest" "$work/log"; then
	pass "both missing images are named"
else
	fail "both missing images are named ($(log_excerpt))"
fi

# ── with a sha input: a pinned release or a rollback ─────────────────────────
# Staging keeps reporting a mismatch: an input never reads it.
serve 200 "$(healthz "$latest" "$older")"
images "$older"
run "PROMOTE_SHA=$older"
promoted "an older SHA from main is promoted (rollback)" "$older" input
if [ -e "$staging/requests.log" ]; then
	fail "a sha input does not read staging's /healthz"
else
	pass "a sha input does not read staging's /healthz"
fi

images "$latest"
run "PROMOTE_SHA=$latest"
promoted "main's latest SHA as input" "$latest" input

images "$older"
# shellcheck disable=SC2016 # '$(id)' is a literal input, never expanded
for bad in "${older:0:7}" "${older:0:39}" "${older}0" "$(tr 'a-f' 'A-F' <<<"$older")" \
	"$older " " $older" "$older; id" "main" "HEAD~1" '$(id)'; do
	run "PROMOTE_SHA=$bad"
	refused "input refused: '$bad'" "full 40-character"
	no_image_check "input '$bad' stops before the image check"
done

images "$feature"
run "PROMOTE_SHA=$feature"
refused "input from another branch" "not on main"

images "$unknown"
run "PROMOTE_SHA=$unknown"
refused "input unknown to the repository" "not a commit"

images "$older" backend
run "PROMOTE_SHA=$older"
refused "input with a missing frontend image" "henga/leviosa-frontend:$older"

# ── checkout ──────────────────────────────────────────────────────────────────
images "$latest"
run "$work/shallow" "PROMOTE_SHA=$latest"
refused "a shallow checkout is refused" "shallow"

# ── result ────────────────────────────────────────────────────────────────────
echo
echo "$passed passed, $failed failed"
if [ "$failed" -eq 0 ]; then
	echo "RESULT: OK"
else
	echo "RESULT: FAILED"
	exit 1
fi
