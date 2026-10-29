<script lang="ts">
    import { env } from "$env/dynamic/public";

    // Self-hosted, cookieless analytics (Plausible) — public routes only. See
    // docs/issues/008-cookieless-analytics.md: client/staff/admin dashboards
    // must stay analytics-free, so only the public layouts render this.
    // Both vars come from PUBLIC_* env vars (see frontend/.env.example); unset
    // in any environment without an approved, deployed Plausible instance —
    // in that case nothing renders and $lib/utils/analytics.ts calls are no-ops.
    const plausibleDomain = env.PUBLIC_PLAUSIBLE_DOMAIN;
    const plausibleSrc = env.PUBLIC_PLAUSIBLE_SRC;
</script>

<svelte:head>
    {#if plausibleDomain && plausibleSrc}
        <script defer data-domain={plausibleDomain} src={plausibleSrc}></script>
    {/if}
</svelte:head>
