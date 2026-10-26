# roles/deploy_gate

The CI deploy key can run exactly one thing: a deploy of a commit on `main`
(issue 022, PRD `docs/prd/deployment-infisical-compose.md`). The repository is
public and GitHub serves commits pushed to forks under this repository's URL,
so "on `main`" is checked on the server, against `main` fetched from GitHub.

```
CI ── ssh leviosa@host 'deploy prod <sha>' ──▶ sshd
        authorized_keys: restrict,command="/usr/local/sbin/leviosa-deploy-gate" <CI key>
                         │  (runs as the deploy user, request in SSH_ORIGINAL_COMMAND)
                         ▼
     leviosa-deploy-gate ── sudo -n /usr/local/sbin/leviosa-deploy prod <sha>
                         │  (the deploy user's only sudoers rule)
                         ▼
     leviosa-deploy (root, empty environment)
        fetch main ─▶ <sha> on main? ─▶ read deploy + compose file at <sha>
        ─▶ install /opt/leviosa/docker-compose.yml ─▶ env -i … deploy prod <sha>
```

## What it installs

| Path | Owner, mode | What |
| --- | --- | --- |
| `/usr/local/sbin/leviosa-deploy-gate` | root, 0755 | the CI keys' forced command (`templates/leviosa-deploy-gate.j2`) |
| `/usr/local/sbin/leviosa-deploy` | root, 0755 | the root runner (`templates/leviosa-deploy.j2`) |
| `/etc/sudoers.d/leviosa-deploy` | root, 0440 | `leviosa ALL=(root) NOPASSWD: /usr/local/sbin/leviosa-deploy`, nothing else; checked with `visudo -cf` before it is written |
| `/var/lib/leviosa-deploy/` | root, 0700 | the runner's state: `repo.git` (bare mirror of `main`), `release-<stack>/` (the files read from the deployed commit), locks |
| `~leviosa/.ssh/authorized_keys` | leviosa, 0600 | a managed block at the top: one `restrict,command="/usr/local/sbin/leviosa-deploy-gate" <key>` line per CI key |

It also installs `git` and refuses to finish when its settings are not plain
(see the first task), when a CI key is not a single bare public key, or when
the deploy user is in the `docker`, `sudo`, `admin` or `wheel` group: the
sudoers rule only means "deploy a commit of `main`" if the deploy user has no
other way to root.

Both scripts change only with a setup run. CI can trigger a deploy but cannot
change what runs as root, and it uploads nothing.

## Variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `deploy_gate_user` | `{{ deploy_user \| default('leviosa') }}` | the deploy user (created by `roles/user`, which must run first) |
| `deploy_gate_ci_keys` | `[]` | CI public keys, one bare key per entry (`"ssh-ed25519 AAAA… leviosa-ci"`); an empty list removes the block |
| `deploy_gate_command` | `/usr/local/sbin/leviosa-deploy-gate` | the gate |
| `deploy_gate_runner` | `/usr/local/sbin/leviosa-deploy` | the runner |
| `deploy_gate_sudoers_file` | `/etc/sudoers.d/leviosa-deploy` | the sudoers drop-in (no dot in the name: `includedir` skips those) |
| `deploy_gate_repo_url` | `https://github.com/hengadev/leviosa.git` | where the runner fetches `main` (public, no credential) |
| `deploy_gate_state_dir` | `/var/lib/leviosa-deploy` | the runner's root-only state |
| `deploy_gate_stack_roots` | `{prod: /opt/leviosa, staging: /opt/leviosa-staging}` | where each stack's compose file is installed; must match `infra/compose/deploy`'s defaults |

The keys file that feeds `deploy_gate_ci_keys` (operator keys and CI keys, outside
the repo) is defined by the Ansible rewrite (issue 025). Map its CI list onto
`deploy_gate_ci_keys`, for example
`deploy_gate_ci_keys: "{{ deploy_ci_ssh_public_keys }}"`.

None of these can be changed at deploy time: Ansible renders them into the
scripts, and the runner reads no environment variable.

## The gate

`leviosa-deploy-gate` reads `SSH_ORIGINAL_COMMAND` and accepts exactly

```
deploy staging <sha>
deploy prod <sha>
```

with `<sha>` a full 40-character lowercase hex commit SHA and single spaces
between the words (matched as one string with
`^deploy (staging|prod) ([0-9a-f]{40})$` in the C locale). It then execs
`sudo -n /usr/local/sbin/leviosa-deploy <stack> <sha>` and returns its exit
status. Anything else is refused with exit 1, a one-line message on stderr and
a `leviosa-deploy-gate: refused: …` syslog line (auth facility). That covers
no command (an interactive login), shells, other commands, `scp`, `sftp`,
`rsync`, extra arguments or options, shell metacharacters, short, uppercase or
non-hex SHAs, branch names, and leading, trailing, doubled, tab, newline,
carriage-return or non-breaking spaces. The request is never evaluated by a
shell. Accepted requests are logged too.

## The runner

`leviosa-deploy <staging|prod> <40-hex sha>`, as root:

1. **Empty environment.** The shebang
   (`#!/usr/bin/env -S -i PATH=… HOME=/root /bin/bash`) starts the script with
   nothing but `PATH` and `HOME`, whoever runs it and whatever sudo keeps.
   git reads no system or user configuration.
2. **Arguments.** Exactly two: a known stack and a full lowercase 40-hex SHA.
   The stack root must exist (Ansible's app layout creates it).
3. **One deploy per stack.** A per-stack lock in the state directory is held
   until the deploy script exits. A second deploy of the same stack is refused
   at once, and the other stack is not blocked.
4. **Fetch `main`, never a SHA.**
   `git fetch <repo> +refs/heads/main:refs/heads/main` into
   `/var/lib/leviosa-deploy/repo.git` (created 0700 on first use). The mirror's
   lock is released before the deploy starts.
5. **On `main`.** `<sha>` must be a commit in the mirror and
   `git merge-base --is-ancestor <sha> main` must succeed. Otherwise:
   `<sha> is not a commit on main of <repo>: deploy refused`. A commit on another
   branch, a pull-request or fork commit, or a commit that `main` no longer
   contains is refused, even when the mirror still holds its object.
6. **Read the release from the commit.** `infra/compose/deploy` and
   `infra/compose/docker-compose.<stack>.yml` are read as blobs
   (`git cat-file blob <sha>:<path>`, so a symlink is never followed) into
   `release-<stack>/`. A commit without that stack's compose file is refused.
7. **Install the compose file** as `<stack root>/docker-compose.yml`, root-owned
   and 0644, through a temporary file and a rename, so a concurrent reader (the
   Infisical Agent recreating a service) never sees half a file.
8. **Run that commit's deploy script**:
   `exec env -i PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin HOME=/root release-<stack>/deploy <stack> <sha>`.
   Its exit status is the runner's, and therefore the gate's and CI's.

A rollback is a deploy of an older commit of `main`. It rolls back the deploy
script and the compose file with the images, because all three come from that
commit. An operator logged in as root can run the runner directly
(`/usr/local/sbin/leviosa-deploy prod <sha>`); it applies the same checks.

## Tests

```sh
infra/ansible/roles/deploy_gate/tests/run-tests.sh      # or: make deploy-gate-test
```

No root, Ansible, Docker or network: bash, git, flock and coreutils (`env -S`).
`visudo`, `ssh-keygen` and Python's `jinja2` are used when present, and the
checks that need them are reported as `SKIP` otherwise. It renders the templates
with test values (plain `{{ variable }}` substitution; it fails if any Jinja
syntax is left, so the rendering matches Ansible's) and prints a PASS/FAIL/SKIP
line per check, ending with `RESULT: OK` (exit 0) or `RESULT: FAILED` (exit 1).
That makes it ready for CI (issue 029):

- **gate**: the accepted and refused table above, with `sudo` and `logger`
  replaced by stubs on `PATH` that record their arguments. Accepted forms reach
  `sudo -n <runner> <stack> <sha>`, the runner's exit status comes back, and
  refusals exit non-zero, never call sudo, and print and log the refusal.
- **runner**: a bare fixture "origin" served over `file://` with `main`, a
  `feature` branch and a `refs/pull/1/head` commit. Commits on `main` deploy:
  prod and staging land in their own stack roots, and a rollback installs the
  older compose file. A feature-branch commit, a pull-request commit, an
  unknown SHA, a commit without the stack's compose file, a commit that left
  `main` after a force-push, a missing stack root, a concurrent deploy of the
  same stack and malformed arguments are refused, with no deploy run and no
  compose file changed. The fixture deploy script records its environment: the
  caller sets `LEVIOSA_*`, `BASH_ENV`, `GIT_DIR`, `GIT_CONFIG_GLOBAL` and
  `HOME`, and the script sees only `PATH` and `HOME=/root`.
- **sudoers**: the drop-in's only rule is the runner, and `visudo -c` accepts it.
- **authorized_keys**: every CI key becomes
  `restrict,command="/usr/local/sbin/leviosa-deploy-gate" <key>`, and
  `ssh-keygen -l` parses each line.

What the tests cannot show (sshd applying the forced command, sudo's rule,
GitHub's `main`) is checked by hand on the server:

- [ ] `sudo -l -U leviosa` (as root) lists only `/usr/local/sbin/leviosa-deploy`
- [ ] `id leviosa` shows no `docker` or `sudo` group
- [ ] with the CI key: `ssh leviosa@<host>` and `ssh leviosa@<host> id` are
      refused, and `scp`/`sftp` fail
- [ ] with the CI key: `ssh leviosa@<host> 'deploy staging <sha of a PR branch>'`
      is refused with "not a commit on main"

## Integration notes (issue 025)

- Include the role in the new site playbook after `roles/user`, which creates
  the deploy user. Without that user the role fails, or skips the keys in
  `--check`. `roles/user` must keep the deploy user out of `docker`/`sudo` and
  must not add another sudoers rule.
- The CI keys live in a marked block at the top of the deploy user's
  `authorized_keys`. sshd uses the first line that matches a key, so a CI key
  that also appears lower down without options is still forced to the gate.
  If `roles/user` writes the whole file, it must keep this block, or write the
  same `restrict,command=` lines itself, above the operator keys. A CI key must
  never go into root's `authorized_keys`.
- Nothing under the stack roots may belong to the deploy user. The runner and
  the deploy script run `docker compose` there as root.
