<script lang="ts">
    import type { LayoutProps } from "./$types";
    import "../../../app.css";
    let { children }: LayoutProps = $props();

    import { setRegistrationContext } from "#lib/context/register.svelte.ts";
    setRegistrationContext();
</script>

{@render children()}
