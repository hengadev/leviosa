# ADR-0016: Two Docker Compose Stacks on One VPS, Not Docker Swarm

**Status**: Accepted
**Date**: 2026-10-20

---

## Context

The previous deployment ran one Hetzner server per environment, configured and released by Ansible from a single operator computer. An earlier plan (local issues 008–025) proposed moving to Docker Swarm with stack files, overlay networks and secrets fetched inside the containers. The old servers are gone and there is no data to keep, so the deployment is rebuilt on the model already proven on Germinal.

## Decision

Leviosa runs on **its own** Hetzner server (`cpx22`, `nbg1`; not shared with another project) with **two Docker Compose stacks** joined to one external network:

- **Production** owns the shared services: Postgres (one database and role per environment), RabbitMQ (one vhost and user per environment) and Caddy. It also runs its own backend, frontend and Redis.
- **Staging** runs only its backend, frontend and its own Redis — Redis has no per-database access control, so sharing it would let staging read production sessions.

Frontend and backend are always released together as one commit SHA. Caddy routes everything to the frontend except `/webhooks/*`, which goes to the backend; staging is behind basic auth on every path except `/healthz` and `/webhooks/*`.

## Considered Options

- **Docker Swarm.** Rejected: on a single node it adds overlay networking, stack semantics and a second secrets mechanism without adding availability.
- **One server per environment.** Rejected: twice the hosts to harden and pay for, for a single-operator project with no traffic yet.
- **Sharing Germinal's server.** Rejected: the shared services would belong to another project's stack and deploys, and one machine identity would read both projects' configuration.

## Consequences

- A production outage of Postgres, RabbitMQ or Caddy also takes staging down.
- Plausible analytics and the Prometheus/Grafana monitoring stack are not on this server; they move to a central platform on the operator's personal VPS, out of scope here.
