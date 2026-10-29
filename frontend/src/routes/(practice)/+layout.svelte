<script lang="ts">
    import type { LayoutProps } from "./$types";
    import PlausibleScript from "$lib/components/analytics/PlausibleScript.svelte";
    let { children }: LayoutProps = $props();
</script>

<PlausibleScript />

{@render children()}
