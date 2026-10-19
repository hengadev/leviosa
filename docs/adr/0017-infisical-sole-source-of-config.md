# ADR-0017: Infisical Is the Only Source of Configuration

**Status**: Accepted
**Date**: 2026-10-20

---

## Context

Configuration used to live in Ansible Vault and in `.env` files on the servers, so only one computer could operate the project, and the HashiCorp Vault unseal keys existed only on the server. The goal is that any of the operator's computers can provision, deploy and recover Leviosa.

## Decision

Every configuration value lives in the operator's self-hosted Infisical (`secrets.henga.dev`):

- Project **`leviosa`**, environments `dev`, `staging`, `prod`, one folder per concern (`/app`, `/admin`, `/db`, `/redis`, `/rabbitmq`, `/encx`, `/aws`, `/stripe`, `/smtp`, `/twilio`, `/caddy`, `/backup`, `/host`). Terraform writes the values it creates (IAM keys, buckets, KMS key ARNs) into it.
- Project **`leviosa-infra`** holds Terraform's own credentials (Hetzner, Cloudflare, AWS admin); the server can never read it. Terraform runs only through `tf.sh`, which fetches them with the operator's own login.
- On the server, the **Infisical Agent** (systemd, read-only machine identity `leviosa-vps`) renders one root-only env file per consumer — backend and frontend per environment, Postgres, RabbitMQ, Caddy, backup, and the one-off admin bootstrap — so the frontend never sees backend secrets. It keeps the last render if Infisical is unreachable.
- Local development renders its env files from `dev` with the same templates (`make env`); committed example files cover the no-Infisical path.
- CI never holds Infisical credentials: only the Docker Hub token and an SSH key forced to the deploy gate.

## Consequences

- On the free plan the server's identity can also read `dev`, so `dev` must never hold anything that spends money or reaches real data.
- The identity cannot be restricted by IP; its client secret is rotated every three months.
- The encx pepper lives here (ADR-0015), so the weekly encrypted Infisical snapshot is part of the data's recovery path.
