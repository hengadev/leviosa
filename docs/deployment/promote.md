# Promote: production releases and rollbacks

`.github/workflows/promote.yaml` is the only way a release reaches production
(PRD `docs/prd/deployment-infisical-compose.md`, issue 030). By default it
ships the exact release staging serves. Given a SHA, it ships that commit of
`main` instead, which is how a rollback is done. It deploys through the same
server-side command as the staging deploy ([ci.md](ci.md)):
`deploy prod <sha>` through the deploy gate.

## Running it

Actions → **🚢 Promote to production** → **Run workflow**, with *Use workflow
from* set to **Branch: main**:

- **`sha` empty**: ship what staging serves. This is the normal release.
- **`sha` set**: ship that commit. It must be the full 40-character SHA
  (`git rev-parse <commit>`) of a commit on `main` that CI built. Use it for a
  rollback, or to ship one specific release again.

## What a run does

```
preflight (no SSH secret)                        deploy (environment production, group deploy-prod)
  refuse unless run from main                      configure SSH: CI key, pinned host key
  check out main with full history       ──▶       ssh leviosa@<SSH_HOST> 'deploy prod <sha>'
  log in to Docker Hub                             job summary: the deployed SHA
  promote-preflight.sh: resolve the SHA,
    check it is on main, check both images
```

The deploy job starts only after the preflight job succeeded and handed it a
SHA, so every check below stops the run before anything reaches the server.
The SHA is fixed when the preflight runs. If staging moves on while the deploy
waits (for a reviewer or for another Promote), the run still ships the SHA it
checked.

### What stops a run before any deploy

| Check | Refused when |
| --- | --- |
| Run from `main` | the run was dispatched from another branch |
| `sha` input | it is not exactly 40 lowercase hex characters: a short SHA, uppercase, spaces, a branch name |
| Staging's `/healthz` (no input only) | it cannot be reached, or answers anything but 200 (502: staging's backend is down; 401: the basic-auth exemption for `/healthz` is missing), or the answer is not JSON, or a value is missing or not a full SHA |
| One release on staging (no input only) | the frontend and the backend report different SHAs: a staging deploy is in progress or stopped halfway. Wait for it or fix staging, then run Promote again |
| On `main` | the SHA is not a commit on `main`: another branch, a fork's commit, or a SHA this repository does not have |
| Images | `henga/leviosa-backend:<sha>` or `henga/leviosa-frontend:<sha>` is not on Docker Hub. CI builds and pushes only the head commit of each push to `main` whose checks all passed. An older commit inside a multi-commit push, or one whose checks failed, has no images |

Each refusal is an error annotation on the run that says what failed.
`.github/scripts/promote-preflight.sh` makes these checks.

### Once the deploy runs

The deploy gate accepts `deploy prod <sha>` only for a commit on `main`, and
runs the deploy script and compose file read from that commit
([roles/deploy_gate](../../infra/ansible/roles/deploy_gate/README.md),
[infra/compose/README.md](../../infra/compose/README.md#deploy-stack-sha)).
The job's result is the deploy's:

- **A failure before the backend is recreated** (pull, shared services,
  migration, admin bootstrap): the deploy restores the `:prod` pins and
  recreates nothing, so production keeps serving the release it had.
- **A failure after it** (a container that does not become healthy, or
  `/healthz` not reporting `<sha>` in time): production may run the new
  backend, or both new apps, in a broken state. Fix forward (push a fix, let
  it reach staging, promote it) or roll back to the previous SHA.
- **Success**: the deploy finished only once production's `/healthz` reported
  `<sha>` for both apps.

### Job summary

- **Preflight**: *Release to promote: `<sha>`*, with its source (staging's
  `/healthz` or the `sha` input) and the two images it found.
- **Deploy**: *Deployed `<sha>` to production* on success, or *Production
  deploy of `<sha>` failed*.

The `production` environment (repository → Environments) lists every run
that reached the deploy job, so it is also the history of production releases.

### One production deploy at a time

The deploy job runs in the concurrency group `deploy-prod` with
`cancel-in-progress: false`. A running production deploy is never cancelled.
A second Promote waits for it, and GitHub keeps only one waiting run per group,
so a third Promote cancels the second one's wait. Staging deploys use their
own group (`deploy-staging`) and do not wait for production. On the server,
the root runner also allows only one deploy per stack at a time.

## Rollback

A rollback is a Promote with the SHA of an earlier release.

1. **Find the SHA.** It is usually the release production ran before the
   current one. The `production` environment's history lists every promoted
   SHA, and each run's summary shows its SHA. Any commit on `main` that CI built
   works. The preflight refuses the others.
2. **Check the migrations in between**, from a checkout with `main` up to date:

   ```sh
   git diff --stat <rollback-sha> <current-sha> -- backend/internal/common/migrations/
   ```

   No output means the rollback crosses no migration. If files are listed,
   read [Data and migrations](#data-and-migrations) first.
3. **Run Promote** with `sha` set to that SHA. It goes through the same checks
   and the same deploy as any release.
4. **Check the result.** The job summary shows the deployed SHA, and
   `https://leviosa.care/healthz` reports it for both apps.

**What rolls back:** both images, plus the compose file and the deploy script,
because the server reads them from the deployed commit.

**What does not roll back:**

- the database schema and its data;
- the configuration in Infisical: an older release reads today's values;
- Postgres, RabbitMQ and Redis data, and media in S3.

If a value was renamed or removed in Infisical since the older release, put it
back before rolling back.

To return to the newer release, promote it again (or let the next release
reach staging and promote that).

### Data and migrations

**Migrations are forward-only.** No deploy ever runs a down migration. When an
older release is deployed over a newer schema:

- its `migrate` one-off applies nothing: the database already has every
  migration that release embeds, and more;
- its `serve` accepts the newer schema. It refuses only a schema *older* than
  its newest migration;
- so the older release runs against the newer schema, together with the data
  the newer release wrote.

This is safe only for migrations that follow the **migration rule** (backend
`CLAUDE.md`, *Database Migrations*: every migration must work with the previous
release; add first, remove later). Before rolling back across a migration,
read it:

- **An added table, or an added column that is nullable or has a default:**
  safe. The older release ignores it. Rows that the older release writes leave
  the column null or at its default, and the newer release must handle those
  rows when it is promoted again.
- **A migration the older release cannot work with** (it drops or renames
  something that release reads, adds a constraint its writes violate, or
  rewrites data it depends on): the older release fails on the affected
  queries, or as soon as it starts. Do not roll back across it. Roll forward
  instead: push a fix to `main`, let it reach staging, and promote it.
- **To undo a migration**, write a new forward migration on `main`. Never run a
  down migration or edit the schema by hand on production. Restoring the
  daily dump is the last resort, and it loses every write made since the dump.

## Setting it up (once)

- **The `production` environment** (repository Settings → Environments). The
  first run creates it with no protection rules. Then set:
  - *Deployment branches and tags*: **Selected branches and tags** → `main`.
    GitHub then refuses to start the deploy job of a run from any other
    branch, even if a branch has modified the workflow.
  - *Required reviewers* (optional): each deploy job then waits for approval.
    The preflight's summary shows the SHA it would ship before you approve.
- **Secrets**: the repository secrets in [ci.md](ci.md#ci-secrets). The
  preflight job reads `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`; read access
  is enough. The deploy job reads `SSH_HOST`, `SSH_PRIVATE_KEY` and
  `SSH_KNOWN_HOSTS`, the same ones the staging deploy uses, since both stacks
  live on one server (ADR-0016). Promote holds no Infisical credential.

**Trust boundary.** The workflow's own "run from `main`" check only stops a
run started from the wrong branch by mistake. The environment's branch limit
and the deploy gate are what enforce `main`. Repository secrets are available
to workflows on any branch. The gate is why the CI key can deploy only commits
on `main`.

## Tests

```sh
.github/scripts/tests/promote-preflight-tests.sh      # or: make promote-preflight-test
```

These run in CI's `deploy-scripts` job. No Docker, no secrets, and no network
beyond loopback. They need bash, git, curl, jq and python3, and they run the
preflight against:

- a fake staging `/healthz` on a random loopback port;
- a fake `docker` that only knows the images listed for each case;
- a fixture repository with `main` and a `feature` branch.

The cases:

- **Accepted:** staging's SHA when both apps agree; an input SHA from `main`,
  older or latest, which never reads staging.
- **Refused:** a frontend/backend mismatch, a 502 or a 401, non-JSON or
  missing values, short or uppercase SHAs, a commit off `main` or unknown, an
  unreachable staging, either image missing, a malformed input (including
  shell metacharacters) and a shallow checkout. Every refusal exits non-zero
  and hands no SHA to the deploy job.
- **Injection:** a value from `/healthz` cannot inject a workflow command.

The tests cannot cover what happens on GitHub (the dispatch, the environment,
the concurrency group) or the SSH deploy itself. Those are checked on the first
production deploy and rollback rehearsal (issue 035).
