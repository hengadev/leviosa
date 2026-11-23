# ADR-0015: encx Keys in AWS KMS and Infisical, Not HashiCorp Vault (Supersedes ADR-0003's Key Management)

**Status**: Accepted
**Date**: 2026-10-20

---

## Context

ADR-0003 put the encx KEK (a Vault transit key) and the pepper in a self-hosted HashiCorp Vault container on the application server. In practice that Vault was the most fragile part of the deployment: it had to be unsealed after every restart, its unseal keys and root token lived only in a file on the server, its storage was never backed up, and the app authenticated with the root token. Losing the server meant losing every key — and with them every encrypted field and every password hash.

The deployment is being rebuilt on the Germinal model (Infisical as the only source of configuration, Compose stacks, Terraform + Ansible from any operator computer). The old server is gone and there is no data to carry over, so the key store can be chosen freely.

## Decision

- **KEK → AWS KMS**, one symmetric customer-managed key per environment (staging, production) in `eu-central-1`, created by Terraform with `prevent_destroy` and a 30-day deletion window. Each environment's app IAM user may use only its own key. encx's existing `providers/keys/aws` provider is used.
- **Pepper → Infisical**, one per environment, rendered into the app's env file by the Infisical Agent like any other config value. The backend reads it through a small env-backed `SecretManagementService` that refuses `StorePepper`, so a missing pepper stops the app instead of silently generating a new one.
- **HashiCorp Vault is removed** from every environment: no Vault container, no init/unseal automation, no `Dockerfile.init`.
- Local development uses LocalStack KMS (LocalStack already runs for S3).

## Considered Options

- **Keep Vault on the VPS.** Rejected: unseal custody, storage backups and a stateful critical service for a single-operator project; the failure mode is permanent data loss.
- **Infisical's built-in KMS on the operator's self-hosted instance.** Free, and encx could gain a provider for it, but: KMS may be a paid feature even when self-hosted; every encrypted read would depend on the operator's personal VPS being up (the app should survive Infisical being down); the app itself would need Infisical credentials (and inherit Universal Auth lockouts); and recovery would depend on backing up that instance's database plus its root encryption key.

## Consequences

- Cost: about $1/month per key plus $0.03 per 10,000 requests beyond the 20,000/month free tier — roughly $2–3/month in total.
- encx v0.10.0 calls the key provider once per encrypted record read or written, with no data-key cache. Against a remote KMS this adds a network round trip per record (≈5–10 ms from Hetzner to Frankfurt), so encx should gain a short-lived in-memory data-key cache.
- The AWS credentials that can use a KMS key are now as sensitive as the old Vault root token; they live only in Infisical and in the app's root-only env file.
- The pepper's durability rests on Infisical and its encrypted backup snapshot. Losing it invalidates all password hashes and searchable hashes.
- Integration tests that provision keys via `SetupServiceVault()` move to a KMS-backed (LocalStack) helper.
