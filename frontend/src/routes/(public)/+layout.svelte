<script lang="ts">
    import type { LayoutProps } from "./$types";
    import "../../app.css";
    import Navigation from "./navigation.svelte";
    import Footer from "./footer.svelte";
    import PlausibleScript from "$lib/components/analytics/PlausibleScript.svelte";
    let { data, children }: LayoutProps = $props();
    let { permissions } = data;
</script>

<PlausibleScript />

<div class="min-h-screen flex flex-col">
    <Navigation {permissions} />
    <main class="flex-1">
        {@render children()}
    </main>
    <Footer />
</div>
