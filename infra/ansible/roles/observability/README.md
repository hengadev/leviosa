# observability role

Runs the monitoring agent ([Grafana Alloy](https://grafana.com/docs/alloy/))
in the production stack. It **pushes** to the homelab (`ingest.henga.dev`),
like Germinal's: nothing outside scrapes this server, and no port is opened.

| Collected | From | Lands in |
| --- | --- | --- |
| Backend metrics, prod and staging (`leviosa_http_requests_total`, latency, requests in flight, `leviosa_build_info`, Go process) | `backend:9090/metrics`, `staging_backend:9090/metrics` over `leviosa_network` (`backend/internal/app/metrics.go`) | VictoriaMetrics, label `env` |
| Host CPU, memory, disk, network | `/proc`, `/sys`, `/` mounted read-only | VictoriaMetrics, label `host="leviosa-vps"` |
| Backup heartbeat (`leviosa_backup_last_success_timestamp_seconds`) | `/var/lib/leviosa-metrics/*.prom`, written by `leviosa-backup` (roles/backup) | VictoriaMetrics |

Every series carries `app="leviosa"`. The homelab also checks
`https://livio.leviosa.care/healthz` and `https://staging.leviosa.care/healthz`
from outside (blackbox exporter). Dashboard and alerts (Telegram) are at
`grafana.henga.dev`, folder **Leviosa**, set up in the homelab repo
(`docker/observability/`).

## Metrics only: no logs, no error tracker

Germinal's agent also ships container and Caddy logs, and its app sends
errors to GlitchTip. Leviosa's does neither, on purpose: logs and error
events carry request URLs and context about the practice's clients, and no
client data may leave the hosting for the homelab. Metrics hold no path,
query string, IP or user input: the backend labels requests with the
**route pattern** (`GET /users/{user_id}`), never the URL, and its test
checks that.

Adding logs, traces or an error tracker is a new data flow out of the
hosting: settle it with the practice's data-protection rules first.

## Setup

1. Infisical `leviosa` → `prod` → `/observability`: `INGEST_URL` =
   `https://ingest.henga.dev`, `INGEST_TOKEN` = the homelab's `INGEST_TOKEN`.
2. `make infisical-agent` (from `infra/ansible`): the Agent learns the new
   template and renders `env/observability.env`.
3. Promote a release whose compose file has the `alloy` service, then run
   this role: `make setup` (or `ansible-playbook … playbooks/site.yml --tags
   observability`).

From then on a change in `/observability` recreates Alloy within a minute
(the Agent), and a change in `files/config.alloy` is applied by re-running
the role.

## Checking it

```sh
docker logs leviosa-alloy-1 --since 10m   # push errors show up here
```

The "Leviosa backend metrics missing" alert fires when nothing arrives for
five minutes, so a dead agent doesn't go unnoticed.
