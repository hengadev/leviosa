#!/usr/bin/env bash
# Promote's preflight (.github/workflows/promote.yaml, job `preflight`):
# decides which release production gets and checks it before anything is
# deployed. Exit 0 hands the SHA to the deploy job; every refusal exits 1 with
# an ::error:: annotation and writes no SHA, so the deploy job never starts.
#
#   1. The SHA. PROMOTE_SHA (the workflow's `sha` input: a pinned release or a
#      rollback) when set, which must be a full 40-character lowercase SHA.
#      Otherwise the release staging serves: STAGING_HEALTHZ_URL must answer
#      200 with the frontend and the backend reporting the same full SHA.
#   2. On main. The SHA must be a commit of MAIN_REF in this checkout, which
#      needs main's full history (actions/checkout with fetch-depth: 0). The
#      server's deploy gate refuses any other commit anyway; this stops earlier
#      and says why.
#   3. Images. BACKEND_IMAGE:<sha> and FRONTEND_IMAGE:<sha> must exist
#      (`docker buildx imagetools inspect`, after `docker login`).
#   4. Hand-off. `sha=<sha>` and `source=<input|staging>` are appended to
#      GITHUB_OUTPUT, and the release to GITHUB_STEP_SUMMARY.
#
# Environment (defaults are production's):
#   PROMOTE_SHA          empty or unset: promote what staging serves
#   STAGING_HEALTHZ_URL  https://staging.leviosa.care/healthz
#   BACKEND_IMAGE        henga/leviosa-backend
#   FRONTEND_IMAGE       henga/leviosa-frontend
#   MAIN_REF             origin/main
#
# Run it from the repository checkout. Needs bash, git, curl, jq and docker
# with buildx. Tests: .github/scripts/tests/promote-preflight-tests.sh
# (make promote-preflight-test).
set -euo pipefail

PROMOTE_SHA=${PROMOTE_SHA-}
STAGING_HEALTHZ_URL=${STAGING_HEALTHZ_URL:-https://staging.leviosa.care/healthz}
BACKEND_IMAGE=${BACKEND_IMAGE:-henga/leviosa-backend}
FRONTEND_IMAGE=${FRONTEND_IMAGE:-henga/leviosa-frontend}
MAIN_REF=${MAIN_REF:-origin/main}

# Exactly what the deploy gate accepts, and what /healthz reports.
sha_re='^[0-9a-f]{40}$'

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

refuse() {
	echo "::error::$*"
	exit 1
}

# show <value> [max]: a value from outside (the input, staging's answer,
# Docker's error) made safe to print: printable ASCII only, so a newline can
# never start a workflow command of its own, no `%` (annotation escapes), and
# at most <max> characters (default 80).
show() {
	local value max=${2:-80}
	value=$(printf '%s' "$1" | LC_ALL=C tr -c '[:print:]' '?' | LC_ALL=C tr '%' '?')
	if [ ${#value} -gt "$max" ]; then
		value="${value:0:max}..."
	fi
	printf '%s' "$value"
}

# staging_sha: sets sha to the SHA both of staging's apps report, or refuses.
# Called directly, not in $(...), so a refusal ends the script with its
# annotation printed.
staging_sha() {
	local body=$tmp/healthz.json status detail frontend backend
	echo "Reading the release staging serves from $STAGING_HEALTHZ_URL"
	status=$(curl --silent --show-error --proto '=http,https' --max-time 20 \
		--max-filesize 65536 --header 'Accept: application/json' \
		--output "$body" --write-out '%{http_code}' \
		"$STAGING_HEALTHZ_URL" 2>"$tmp/curl.err") ||
		refuse "could not reach staging's $STAGING_HEALTHZ_URL: $(show "$(cat "$tmp/curl.err")" 200)"

	if [ "$status" != 200 ]; then
		detail=$(jq -r '.error | strings' "$body" 2>/dev/null || true)
		refuse "staging's $STAGING_HEALTHZ_URL answered HTTP $status${detail:+ ($(show "$detail"))}." \
			"Promote ships only what a healthy staging serves: fix staging, or run Promote with a sha."
	fi
	if ! jq -e 'type == "object"' "$body" >/dev/null 2>&1; then
		refuse "staging's $STAGING_HEALTHZ_URL did not answer a JSON object with the frontend and backend SHAs."
	fi
	frontend=$(jq -r '.frontend | strings' "$body")
	backend=$(jq -r '.backend | strings' "$body")
	[ -n "$frontend" ] || refuse "staging's /healthz reports no frontend SHA."
	[ -n "$backend" ] || refuse "staging's /healthz reports no backend SHA."
	[[ $frontend =~ $sha_re ]] ||
		refuse "staging's frontend reports '$(show "$frontend")', not a full 40-character commit SHA."
	[[ $backend =~ $sha_re ]] ||
		refuse "staging's backend reports '$(show "$backend")', not a full 40-character commit SHA."
	if [ "$frontend" != "$backend" ]; then
		refuse "staging is not on one release: frontend $frontend, backend $backend." \
			"Wait for staging's deploy to finish (or fix it), then run Promote again, or run it with a sha."
	fi
	sha=$frontend
}

# on_main <sha>: refuses unless <sha> is a commit on main.
on_main() {
	local sha=$1
	if [ "$(git rev-parse --is-shallow-repository 2>/dev/null)" != false ]; then
		refuse "this checkout is shallow (or not a git repository): the preflight needs main's full history (actions/checkout with fetch-depth: 0)."
	fi
	git rev-parse --verify --quiet "$MAIN_REF^{commit}" >/dev/null ||
		refuse "$MAIN_REF is not in this checkout: the preflight needs main's full history (actions/checkout with fetch-depth: 0)."
	git cat-file -e "$sha^{commit}" 2>/dev/null ||
		refuse "$sha is not a commit of this repository: only a commit on main can be promoted."
	git merge-base --is-ancestor "$sha" "$MAIN_REF" ||
		refuse "$sha is not on main: only a commit on main can be promoted (the deploy gate refuses anything else)."
	echo "$sha is on main"
}

# images_exist <sha>: refuses unless both images exist on the registry.
images_exist() {
	local sha=$1 image missing=()
	for image in "$BACKEND_IMAGE:$sha" "$FRONTEND_IMAGE:$sha"; do
		if docker buildx imagetools inspect "$image" >/dev/null 2>"$tmp/inspect.err"; then
			echo "Found $image"
		else
			echo "Missing $image: $(show "$(cat "$tmp/inspect.err")" 200)"
			missing+=("$image")
		fi
	done
	if [ ${#missing[@]} -gt 0 ]; then
		refuse "not on Docker Hub (or not readable with DOCKERHUB_TOKEN): ${missing[*]}." \
			"CI pushes both images only for a commit it built on main after every check passed."
	fi
}

if [ -n "$PROMOTE_SHA" ]; then
	[[ $PROMOTE_SHA =~ $sha_re ]] ||
		refuse "the sha input '$(show "$PROMOTE_SHA")' is not a full 40-character commit SHA" \
			"(lowercase hex, as \`git rev-parse HEAD\` prints it). Leave it empty to promote what staging serves."
	sha=$PROMOTE_SHA
	source=input
	source_text="the \`sha\` input (a pinned release or a rollback)"
else
	staging_sha
	source=staging
	source_text="staging's \`/healthz\` (frontend and backend both report it)"
fi
echo "Release to promote: $sha (from $source)"

on_main "$sha"
images_exist "$sha"

if [ -n "${GITHUB_OUTPUT-}" ]; then
	printf 'sha=%s\nsource=%s\n' "$sha" "$source" >>"$GITHUB_OUTPUT"
fi
if [ -n "${GITHUB_STEP_SUMMARY-}" ]; then
	{
		echo "### Release to promote: \`$sha\`"
		echo
		echo "- Source: $source_text"
		echo "- On \`main\`: yes"
		echo "- Images: \`$BACKEND_IMAGE:$sha\` and \`$FRONTEND_IMAGE:$sha\` exist"
	} >>"$GITHUB_STEP_SUMMARY"
fi
