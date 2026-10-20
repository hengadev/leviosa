#!/usr/bin/env bash
# Tests for roles/infisical_agent: no root, no server, no Infisical, no
# network.
#
#   infra/ansible/roles/infisical_agent/tests/run-tests.sh
#   (or: make infisical-agent-test, from the repository root)
#
#   templates   the role's Jinja templates (Agent config, systemd unit,
#               post-start helper) are rendered by ansible itself, with the
#               role's own defaults and sample stack roots, then checked:
#               the config's YAML structure and render matrix (which file
#               recreates which service, always --no-deps --force-recreate),
#               the unit with systemd-analyze verify (stubbed binaries) and
#               grep, the helper's syntax
#   helper      the rendered post-start helper runs against fake stack roots
#               with a stub docker: a changed backend.env recreates only that
#               stack's backend, a changed admin.env nothing, an unchanged
#               re-render nothing, and a first render on a fresh server
#               (no compose file yet) skips every recreate
#   pipeline    the role's render pipeline (--tags templates) runs against
#               localhost in --check with a fake project ID: one
#               render-template.sh call per expected (env, template) file,
#               each byte-identical to running the script directly
#   credential  the credential tasks (--tags credential) run against
#               localhost in --check with a fake secret at -vv --diff: the
#               output must not contain it (no_log), and a static walk of
#               every role/playbook task proves no task touches the secret
#               variables without no_log (an assert on their length excepted)
#   wiring      playbooks/site.yml includes the role with a when on both
#               INFISICAL_VPS_* variables and a skip note
#
# Needs bash, ansible-playbook, python3 (PyYAML ships with ansible),
# coreutils. Prints a PASS/FAIL/SKIP line per check and ends with
# `RESULT: OK` (exit 0) or `RESULT: FAILED` (exit 1).
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
role=$(cd "$here/.." && pwd)
ansible_dir=$(cd "$role/../.." && pwd)

work=$(mktemp -d "${TMPDIR:-/tmp}/leviosa-infisical-agent-test.XXXXXX")
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

# The fake project ID render-template.sh accepts (LEVIOSA_INFISICAL_PROJECT_ID
# overrides infra/terraform/terraform.tfvars, which holds the placeholder
# until the operator sets the real one).
FAKE_PID="11111111-2222-3333-4444-555555555555"

# Run ansible-playbook from infra/ansible (ansible.cfg: roles_path), against
# the implicit localhost, never becoming root.
run_play() {
	(cd "$ansible_dir" && LEVIOSA_INFISICAL_PROJECT_ID="$FAKE_PID" \
		ansible-playbook -i localhost, "$@" )
}

# ===========================================================================
# 1. Render the role's templates with the role's defaults and sample roots
# ===========================================================================

mkdir -p "$work/out" "$work/opt/leviosa/env" "$work/opt/leviosa-staging/env"

cat > "$work/render.yml" <<YAML
- name: Render the infisical_agent templates with sample vars
  hosts: localhost
  connection: local
  gather_facts: false
  become: false
  vars:
    app_prod_root: $work/opt/leviosa
    app_staging_root: $work/opt/leviosa-staging
  roles:
    # when: false runs no task; the role's defaults still load, so the
    # rendering below sees exactly what a real run would.
    - role: infisical_agent
      when: false
  tasks:
    - name: render the Agent config
      ansible.builtin.template:
        src: $role/templates/config.yaml.j2
        dest: $work/out/config.yaml
        mode: "0644"
    - name: render the systemd unit
      ansible.builtin.template:
        src: $role/templates/infisical-agent.service.j2
        dest: $work/out/infisical-agent.service
        mode: "0644"
    - name: render the post-start helper
      ansible.builtin.template:
        src: $role/templates/post-start.sh.j2
        dest: $work/out/post-start.sh
        mode: "0750"
YAML

if run_play "$work/render.yml" >"$work/render.log" 2>&1; then
	pass "templates render with the role's defaults (config, unit, helper)"
else
	fail "templates render with the role's defaults (config, unit, helper)"
	sed -n '1,30p' "$work/render.log"
fi

# ===========================================================================
# 2. The Agent config: structure and render matrix
# ===========================================================================

python3 - "$work/out/config.yaml" "$work" <<'PY' && pass "config.yaml: address, auth paths, render matrix, recreate commands" || fail "config.yaml: address, auth paths, render matrix, recreate commands"
import sys, yaml

cfg = yaml.safe_load(open(sys.argv[1]))
assert cfg["infisical"]["address"] == "https://secrets.henga.dev", cfg["infisical"]

auth = cfg["auth"]
assert auth["type"] == "universal-auth"
assert auth["config"]["client-id"] == "/etc/infisical-agent/credentials/client-id"
assert auth["config"]["client-secret"] == "/etc/infisical-agent/credentials/client-secret"
# Paths, not values: nothing templated or env-looking may sit in the config.
blob = yaml.safe_dump(cfg)
assert "lookup(" not in blob and "{{" not in blob and "FAKE" not in blob

assert cfg["sinks"][0]["config"]["path"] == "/etc/infisical-agent/credentials/token-sink.json"

PROD, STAGING = f"{sys.argv[2]}/opt/leviosa", f"{sys.argv[2]}/opt/leviosa-staging"
# (env, name, service or None) — the issue 027 matrix.
EXPECTED = [
    ("prod", "backend", "backend"),
    ("prod", "frontend", "frontend"),
    ("prod", "redis", "redis"),
    ("prod", "caddy", "caddy"),
    ("prod", "postgres", None),
    ("prod", "rabbitmq", None),
    ("prod", "backup", None),
    ("prod", "admin", None),
    ("staging", "backend", "staging_backend"),
    ("staging", "frontend", "staging_frontend"),
    ("staging", "redis", "staging_redis"),
    ("staging", "admin", None),
]
entries = cfg["templates"]
assert len(entries) == len(EXPECTED), f"{len(entries)} template entries"
for entry, (env, name, service) in zip(entries, EXPECTED):
    root = PROD if env == "prod" else STAGING
    assert entry["source-path"] == f"/etc/infisical-agent/templates/{env}-{name}.env.tmpl"
    assert entry["destination-path"] == f"{root}/env/{name}.env"
    conf = entry["config"]
    assert conf["polling-interval"] == "60s", conf
    if service is None:
        assert "execute" not in conf, f"{env}/{name} must be rewrite-only"
    else:
        cmd = conf["execute"]["command"]
        want = (f"docker compose --project-directory {root} "
                f"up -d --no-deps --force-recreate {service}")
        assert cmd == want, cmd
        assert conf["execute"]["timeout"] == 60
PY

# ===========================================================================
# 3. The systemd unit
# ===========================================================================

unit="$work/out/infisical-agent.service"
for want in \
	"ExecStart=/usr/bin/infisical agent --config /etc/infisical-agent/config.yaml" \
	"ExecStartPre=-/etc/infisical-agent/post-start.sh snapshot" \
	"ExecStartPost=-/etc/infisical-agent/post-start.sh apply" \
	"Restart=always" \
	"UMask=0077"; do
	if grep -qxF "$want" "$unit"; then
		pass "unit: $want"
	else
		fail "unit: $want"
	fi
done

if command -v systemd-analyze > /dev/null 2>&1; then
	# verify checks that ExecStart/ExecStartPre/Post binaries exist: stub
	# them, by rewriting only those path prefixes in a copy of the unit.
	mkdir -p "$work/stub-bin" "$work/stub-etc"
	printf '#!/bin/sh\nexit 0\n' > "$work/stub-bin/infisical"
	printf '#!/bin/sh\nexit 0\n' > "$work/stub-etc/post-start.sh"
	chmod +x "$work/stub-bin/infisical" "$work/stub-etc/post-start.sh"
	sed -e "s|/usr/bin/infisical|$work/stub-bin/infisical|g" \
		-e "s|/etc/infisical-agent|$work/stub-etc|g" \
		"$unit" > "$work/unit-verify.service"
	if systemd-analyze verify "$work/unit-verify.service" 2>"$work/verify.err"; then
		pass "unit: systemd-analyze verify accepts it"
	else
		fail "unit: systemd-analyze verify accepts it"
		sed -n '1,10p' "$work/verify.err"
	fi
else
	skip "unit: systemd-analyze verify (systemd-analyze not installed)"
fi

# ===========================================================================
# 4. The post-start helper: syntax, then behaviour against a stub docker
# ===========================================================================

helper="$work/out/post-start.sh"
if sh -n "$helper"; then
	pass "helper: sh -n accepts it"
else
	fail "helper: sh -n accepts it"
fi

# The helper writes its snapshot under /run; point that at the work dir.
sed "s|/run/infisical-agent/pre-start.sum|$work/run/pre-start.sum|" \
	"$helper" > "$work/helper-test.sh"
chmod +x "$work/helper-test.sh"
mkdir -p "$work/run" "$work/bin"
cat > "$work/bin/docker" <<EOF
#!/bin/sh
echo "docker \$*" >> "\$DOCKER_LOG"
EOF
chmod +x "$work/bin/docker"

helper_run() {
	DOCKER_LOG="$work/docker.log" PATH="$work/bin:$PATH" "$work/helper-test.sh" "$@"
}

check_docker_log() {
	want=$1
	if [ -f "$work/docker.log" ] && grep -qxF "$want" "$work/docker.log"; then
		pass "helper: called '$want'"
	else
		fail "helper: called '$want'"
		sed -n '1,10p' "$work/docker.log" 2> /dev/null
	fi
}

# Scenario 1: everything deployed, backend.env and staging's admin.env change.
for f in backend frontend redis caddy postgres rabbitmq backup admin; do
	echo "INITIAL-$f" > "$work/opt/leviosa/env/$f.env"
done
for f in backend frontend redis admin; do
	echo "INITIAL-staging-$f" > "$work/opt/leviosa-staging/env/$f.env"
done
touch "$work/opt/leviosa/docker-compose.yml" # prod deployed; staging not
helper_run snapshot
echo "NEW-backend-VALUE" > "$work/opt/leviosa/env/backend.env"
echo "NEW-admin-VALUE" > "$work/opt/leviosa-staging/env/admin.env"
find "$work/opt" -name '*.env' -exec touch {} +
helper_run apply >"$work/apply1.log" 2>&1
check_docker_log "docker compose --project-directory $work/opt/leviosa up -d --no-deps --force-recreate backend"
if grep -q "recreating backend" "$work/apply1.log" &&
	grep -q "rewrite only" "$work/apply1.log" &&
	[ "$(grep -c "docker " "$work/docker.log")" -eq 1 ]; then
	pass "helper: a changed backend.env recreates only prod's backend; admin.env nothing"
else
	fail "helper: a changed backend.env recreates only prod's backend; admin.env nothing"
	sed -n '1,15p' "$work/apply1.log"
fi

# Scenario 2: the Agent restarts (new snapshot), files re-rendered unchanged.
helper_run snapshot
find "$work/opt" -name '*.env' -exec touch {} +
helper_run apply >"$work/apply2.log" 2>&1
if grep -q "rendered files unchanged" "$work/apply2.log" &&
	grep -q "nothing to recreate" "$work/apply2.log"; then
	pass "helper: an unchanged re-render (rotation restart) recreates nothing"
else
	fail "helper: an unchanged re-render (rotation restart) recreates nothing"
	sed -n '1,15p' "$work/apply2.log"
fi

# Scenario 3: a fresh server — first render ever, no compose file anywhere.
rm -rf "$work/opt2" "$work/run2"
mkdir -p "$work/opt2/leviosa/env" "$work/opt2/leviosa-staging/env" "$work/run2"
sed -e "s|$work/opt/leviosa-staging|$work/opt2/leviosa-staging|" \
	-e "s|$work/opt/leviosa|$work/opt2/leviosa|" \
	-e "s|$work/run/pre-start.sum|$work/run2/pre-start.sum|" \
	"$work/helper-test.sh" > "$work/helper-fresh.sh"
chmod +x "$work/helper-fresh.sh"
DOCKER_LOG="$work/docker2.log" PATH="$work/bin:$PATH" "$work/helper-fresh.sh" snapshot
for f in backend frontend redis caddy postgres rabbitmq backup admin; do
	echo "FIRST-$f" > "$work/opt2/leviosa/env/$f.env"
done
for f in backend frontend redis admin; do
	echo "FIRST-$f" > "$work/opt2/leviosa-staging/env/$f.env"
done
DOCKER_LOG="$work/docker2.log" PATH="$work/bin:$PATH" "$work/helper-fresh.sh" apply >"$work/apply3.log" 2>&1
if [ ! -e "$work/docker2.log" ] &&
	grep -q "not present yet" "$work/apply3.log" &&
	grep -q "skipping backend" "$work/apply3.log"; then
	pass "helper: first render on a fresh server runs no compose command"
else
	fail "helper: first render on a fresh server runs no compose command"
	sed -n '1,15p' "$work/apply3.log"
fi

# ===========================================================================
# 5. The render pipeline (--tags templates), in --check against localhost
# ===========================================================================

mkdir -p "$work/etc-infisical-agent"
cat > "$work/pipeline.yml" <<YAML
- name: Render-pipeline check (check mode, fake project ID)
  hosts: localhost
  connection: local
  gather_facts: false
  become: false
  vars:
    infisical_agent_config_dir: $work/etc-infisical-agent
    infisical_agent_manage_service: false
  roles:
    - role: infisical_agent
  tasks:
    - name: Dump the render results for inspection
      ansible.builtin.copy:
        content: "{{ infisical_agent_rendered | to_json }}"
        dest: $work/rendered.json
      no_log: false
      # always + check_mode: false: the dump must run (and really write the
      # JSON — template text with the fake project ID, no secret) under
      # --tags templates even though the play runs in --check.
      check_mode: false
      tags: ['always']
YAML

if run_play "$work/pipeline.yml" --check --tags templates >"$work/pipeline.log" 2>&1; then
	python3 - "$work/rendered.json" "$ansible_dir/../infisical/render-template.sh" "$FAKE_PID" <<'PY' \
		&& pass "pipeline: one render-template.sh call per matrix entry, byte-identical" \
		|| fail "pipeline: one render-template.sh call per matrix entry, byte-identical"
import json, subprocess, sys

results = json.load(open(sys.argv[1]))
script, fake_pid = sys.argv[2], sys.argv[3]
EXPECTED = [(e, n) for e in ("prod", "staging")
            for n in (["backend", "frontend", "redis", "caddy", "postgres",
                       "rabbitmq", "backup", "admin"] if e == "prod"
                      else ["backend", "frontend", "redis", "admin"])]
seen = [(r["item"]["env"], r["item"]["name"]) for r in results["results"]]
assert sorted(seen) == sorted(EXPECTED), seen
for r in results["results"]:
	assert r["rc"] == 0, r
	assert "__LEVIOSA_" not in r["stdout"], r["item"]
	env, name = r["item"]["env"], r["item"]["name"]
	src = f"{script.rsplit('/', 1)[0]}/templates/{name}.env.tmpl"
	direct = subprocess.run([script, src, env], capture_output=True, text=True,
	                        env={"LEVIOSA_INFISICAL_PROJECT_ID": fake_pid, "PATH": "/usr/bin:/bin"})
	assert direct.returncode == 0, direct.stderr
	assert r["stdout"] == direct.stdout.rstrip("\n"), f"{env}/{name} differs from a direct render"
	assert fake_pid in r["stdout"], f"{env}/{name} lacks the project ID"
PY
else
	fail "pipeline: one render-template.sh call per matrix entry, byte-identical"
	sed -n '1,20p' "$work/pipeline.log"
fi

# ===========================================================================
# 6. The credential: never in Ansible output, never without no_log
# ===========================================================================

fake_id="fake-client-id-0123456789"
fake_secret="FAKE-SECRET-leak-test-$(date +%s)-$$-do-not-print"
cat > "$work/leak.yml" <<YAML
- name: Credential leak check (check mode, fake values)
  hosts: localhost
  connection: local
  gather_facts: false
  become: false
  vars:
    infisical_agent_client_id: "$fake_id"
    infisical_agent_client_secret: "$fake_secret"
    infisical_agent_config_dir: $work/etc-infisical-agent
    infisical_agent_manage_service: false
  roles:
    - role: infisical_agent
YAML

if run_play "$work/leak.yml" --check --diff -vv --tags credential >"$work/leak.log" 2>&1; then
	if grep -qF "$fake_secret" "$work/leak.log"; then
		fail "credential: the fake secret never appears in -vv --diff output"
	else
		pass "credential: the fake secret never appears in -vv --diff output"
	fi
else
	fail "credential: the fake secret never appears in -vv --diff output"
	sed -n '1,20p' "$work/leak.log"
fi
if grep -q "Place the leviosa-vps client secret" "$work/leak.log"; then
	pass "credential: the credential tasks actually ran in the leak check"
else
	fail "credential: the credential tasks actually ran in the leak check"
fi

# Static: every task that mentions the secret variables is no_log (an assert
# on their length is the one exception — it prints booleans only).
python3 - "$ansible_dir" <<'PY' && pass "credential: every task touching the secret variables is no_log" || fail "credential: every task touching the secret variables is no_log"
import sys, glob, yaml

VARS = ("infisical_agent_client_secret", "infisical_agent_client_id")
problems = []

def tasks_of(play):
	for key in ("tasks", "pre_tasks", "post_tasks", "handlers"):
		yield from play.get(key) or []

def check(task):
	if not isinstance(task, dict):
		return
	dump = yaml.safe_dump(task)
	if any(v in dump for v in VARS):
		if task.get("no_log") is not True:
			action = next((k for k in task if k.startswith("ansible.builtin.")
			               or k in ("assert", "copy", "command", "debug", "set_fact", "systemd", "template", "file", "find")),
			              None)
			length_only = action in ("assert", "ansible.builtin.assert") and all(
				"| length" in that for that in task.get("that", []))
			if not length_only:
				problems.append(f"{task.get('name', action)}")
	for key in ("block", "rescue", "always"):
		for sub in task.get(key) or []:
			check(sub)

for path in (glob.glob(f"{sys.argv[1]}/roles/infisical_agent/tasks/*.yml")
             + glob.glob(f"{sys.argv[1]}/roles/infisical_agent/handlers/*.yml")
             + glob.glob(f"{sys.argv[1]}/playbooks/*.yml")):
	for doc in yaml.safe_load_all(open(path)) or []:
		for play in (doc if isinstance(doc, list) else [doc]):
			if isinstance(play, dict):
				for task in tasks_of(play):
					check(task)

assert not problems, f"tasks without no_log touching the credential: {problems}"
PY

# ===========================================================================
# 7. site.yml wiring: the role, its when, and the skip note
# ===========================================================================

python3 - "$ansible_dir/playbooks/site.yml" <<'PY' && pass "wiring: site.yml includes the role behind a when on both INFISICAL_VPS_* variables, with a skip note" || fail "wiring: site.yml includes the role behind a when on both INFISICAL_VPS_* variables, with a skip note"
import sys, yaml

plays = yaml.safe_load(open(sys.argv[1]))
var = "leviosa_infisical_credential_present"
definitions = [str(p.get("vars", {}).get(var)) for p in plays if var in p.get("vars", {})]
assert len(definitions) == 1, definitions
assert "INFISICAL_VPS_CLIENT_ID" in definitions[0]
assert "INFISICAL_VPS_CLIENT_SECRET" in definitions[0]
entry = [r for p in plays for r in (p.get("roles") or [])
         if isinstance(r, dict) and r.get("role") == "infisical_agent"]
assert len(entry) == 1, f"{len(entry)} infisical_agent role entries"
assert var in str(entry[0].get("when", "")), entry[0].get("when")
names = [t.get("name", "") for p in plays for t in (p.get("pre_tasks") or [])]
assert any("Infisical Agent role will be skipped" in n for n in names), names
PY

# ===========================================================================

echo ""
echo "RESULT: $([ $failed -eq 0 ] && echo OK || echo FAILED) (passed=$passed failed=$failed skipped=$skipped)"
[ $failed -eq 0 ]
