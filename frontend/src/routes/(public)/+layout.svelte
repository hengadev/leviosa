<script lang="ts">
    import { env } from "$env/dynamic/public";
    import type { LayoutProps } from "./$types";
    import "../../app.css";
    import Navigation from "./navigation.svelte";
    import Footer from "./footer.svelte";
    let { data, children }: LayoutProps = $props();
    let { permissions } = data;

    // Self-hosted, cookieless analytics (Plausible) — public routes only. See
    // docs/issues/008-cookieless-analytics.md: client/staff/admin dashboards
    // must stay analytics-free, so this script tag lives here and nowhere else.
    // Both vars come from PUBLIC_* env vars (see frontend/.env.example); unset
    // in any environment without an approved, deployed Plausible instance —
    // in that case the layout renders no script and $lib/utils/analytics.ts
    // calls are no-ops.
    const plausibleDomain = env.PUBLIC_PLAUSIBLE_DOMAIN;
    const plausibleSrc = env.PUBLIC_PLAUSIBLE_SRC;
</script>

<svelte:head>
    {#if plausibleDomain && plausibleSrc}
        <script defer data-domain={plausibleDomain} src={plausibleSrc}></script>
    {/if}
</svelte:head>

<div class="min-h-screen flex flex-col">
    <Navigation {permissions} />
    <main class="flex-1">
        {@render children()}
    </main>
    <Footer />
</div>
