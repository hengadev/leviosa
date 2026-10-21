# CI: checks, images and the staging deploy

`.github/workflows/ci.yaml` checks every pull request to `main` and every push
to `main`. A push to `main` that passes every check is built and deployed to
staging with no human step (PRD `docs/prd/deployment-infisical-compose.md`,
issue 029). Production is not deployed from here: the manual **Promote**
workflow (`.github/workflows/promote.yaml`) ships the release staging serves,
and a rollback is a Promote of an older SHA ([promote.md](promote.md)).

## Jobs

| Job | Pull request to `main` | Push to `main` |
| --- | --- | --- |
| `frontend-checks`: SvelteKit build | runs | runs |
| `backend-checks`: Go build, `go test ./...` | runs | runs |
| `backend-integration-tests`: `make test-integration` (testcontainers) | runs | runs |
| `security-scan`: `govulncheck` | runs | runs |
| `deploy-scripts`: deploy gate tests, Promote's preflight tests, the Infisical Agent template tests, `bash -n` on `infra/compose/deploy` and the harness | runs | runs |
| `summary`: fails unless every check above passed | runs | runs |
| `build-images`: build and push both images | skipped | after every check passed |
| `deploy-staging`: `deploy staging <sha>` through the deploy gate | skipped | after every check passed and both images were pushed |

```
frontend-checks ───────────┐
backend-checks ────────────┤
backend-integration-tests ─┼──▶ build-images (backend, frontend) ──▶ deploy-staging
security-scan ─────────────┤                                               ▲
deploy-scripts ────────────┴───────────────────────────────────────────────┘
```

- **Pull requests never build an image, push or deploy.** `build-images` and
  `deploy-staging` run only for a push to `main` of `Leviosa-care/leviosa`
  (a fork's own pushes skip them too).
- **A red check stops the release.** Both release jobs need every check job and
  have no status override, and no check uses `continue-on-error`. Fix the
  check and push again.
- **Images**: `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>`,
  `linux/amd64`, where `<sha>` is the full 40-hex commit SHA (`github.sha`).
  It is baked in with `--build-arg COMMIT_SHA=<sha>` and reported by the
  backend's `/health` and the frontend's `/healthz`. CI pushes SHA tags only;
  the `:prod` and `:staging` pins are local tags on the server
  ([infra/compose/README.md](../../infra/compose/README.md#image-pins)).

## The staging deploy

```
ssh leviosa@<SSH_HOST> 'deploy staging <sha>'      port 22, CI key, pinned host key
```

- `leviosa` is the deploy user. The CI key's forced command is the deploy gate,
  which accepts exactly `deploy <staging|prod> <40-hex sha>` for a commit on
  `main` and runs the deploy script and compose file read from the repository
  at that SHA, so CI uploads nothing
  ([roles/deploy_gate](../../infra/ansible/roles/deploy_gate/README.md)).
- **Host key pinned.** The job trusts only the `SSH_KNOWN_HOSTS` secret
  (`StrictHostKeyChecking yes`, no global known-hosts file, no `ssh-keyscan`).
  An unknown or changed host key fails the job before anything is sent.
- **The job's result is the deploy's.** ssh exits with the remote command's
  status (255 when the connection fails), so a refused request or a failed
  deploy fails the job. The deploy itself finishes only when staging's
  `/healthz` reports `<sha>` for both apps.
- **One staging deploy at a time** (concurrency group `deploy-staging`,
  `cancel-in-progress: false`): a running deploy is never cancelled, and a newer
  push waits for it. GitHub keeps one waiting run per group, so after several
  quick pushes only the newest one deploys next.
- To retry, re-run the failed job from the workflow run. The GitHub environment
  `staging` (created on the first run) lists every deploy.

## CI secrets

Repository secrets (Settings → Secrets and variables → Actions). These five
are the only ones the workflows read (`ci.yaml`, and `promote.yaml` for
production):

| Secret | Used by | Value |
| --- | --- | --- |
| `DOCKERHUB_USERNAME` | `build-images`; Promote's `preflight` (image check) | the Docker Hub account that pushes `henga/leviosa-backend` and `henga/leviosa-frontend` |
| `DOCKERHUB_TOKEN` | `build-images`; Promote's `preflight` (image check) | a Docker Hub access token with read and write access to those repositories, not the account password |
| `SSH_PRIVATE_KEY` | `deploy-staging`; Promote's `deploy` | the private half of the `leviosa-ci` key, OpenSSH format, no passphrase. Its public half is in the keys file's CI list (`deploy_ci_ssh_public_keys`), never in root's `authorized_keys` |
| `SSH_KNOWN_HOSTS` | `deploy-staging`; Promote's `deploy` | the server's host key for `SSH_HOST`, as a known-hosts line: `203.0.113.10 ssh-ed25519 AAAA…` (a hashed line works too) |
| `SSH_HOST` | `deploy-staging`; Promote's `deploy` | the server's address, exactly as it appears in `SSH_KNOWN_HOSTS`: Terraform's `server_ipv4_address` output |

The SSH user (`leviosa`), the port (22) and the image repositories are written
in the workflows; they are not secrets. GitHub holds no Infisical, AWS, Stripe,
database or other runtime credential (ADR-0017): the server's Infisical Agent
provides those. Any other repository secret is unused by CI and can be deleted.

**Filling `SSH_KNOWN_HOSTS`.** Take the host key over a connection you already
trust (an operator computer that checked the server's fingerprint on its first
login), never from an unauthenticated scan:

```sh
ssh root@<ip> cat /etc/ssh/ssh_host_ed25519_key.pub | awk -v h=<ip> '{print h, $1, $2}'
```

A rebuilt server has a new host key: staging deploys and Promote fail with
"Host key verification failed" until this secret is updated. Both stacks live
on that one server (ADR-0016), so the same three SSH secrets serve both.

## Not run in CI

- **The compose harness** (`infra/compose/test/run-harness.sh`,
  `make deploy-harness`) builds three images and drives several full deploys
  against LocalStack, which is too slow for every pull request and push. Run it
  locally when changing `infra/compose/`. The staging deploy runs the real
  script on every push to `main`.
- The deploy gate tests run in CI as they do locally
  (`make deploy-gate-test`), except that a `SKIP` fails the CI job: every check
  must run there.
