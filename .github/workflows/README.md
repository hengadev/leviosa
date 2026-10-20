# GitHub Actions Workflows

This directory contains the CI workflow for the Leviosa project. Job graph,
secrets and the deploy path in detail: [docs/deployment/ci.md](../../docs/deployment/ci.md).

## Workflow

### CI (`ci.yaml`)

**Trigger:** Pull requests to `main` and pushes to `main`
**Purpose:** Validate every change; ship every push to `main` that passes to staging

#### Check jobs (pull requests and pushes):
- **frontend-checks**: Install deps, build the SvelteKit app
- **backend-checks**: Build the Go binary, run unit tests (`go test ./...`)
- **backend-integration-tests**: Integration tests via `make test-integration` (dependencies spun up per-test with testcontainers)
- **security-scan**: Go vulnerability scanning with `govulncheck`
- **deploy-scripts**: Deploy gate and root runner tests (`infra/ansible/roles/deploy_gate/tests/run-tests.sh`), `bash -n` on `infra/compose/deploy` and the harness
- **summary**: Aggregates the above; fails the check if any job failed

#### Release jobs (pushes to `main` only, after every check job passed):
- **build-images**: Build and push `henga/leviosa-backend:<sha>` and `henga/leviosa-frontend:<sha>` with `COMMIT_SHA=<sha>` baked in
- **deploy-staging**: SSH to the server with the CI key (pinned host key) and run `deploy staging <sha>` through the deploy gate; fails if the deploy fails

Pull requests never build an image, push or deploy.

**Failure Behavior:** PR cannot merge until all check jobs pass; a push to `main` with a failing check is neither built nor deployed

**Merge Strategy:** Rebase and merge only (preserves full commit history)

## Deployment

Staging is deployed by `ci.yaml` on every push to `main` that passes the
checks. The deploy gate on the server (`infra/ansible/roles/deploy_gate/`)
accepts only `deploy <staging|prod> <sha>` for a commit on `main` and runs the
deploy script and compose file from that commit (`infra/compose/`). Production
is not deployed by this workflow.
