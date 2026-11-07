<script lang="ts">
    import type { LayoutProps } from "./$types";
    import "../../../app.css";
    let { children }: LayoutProps = $props();
</script>

<div class="container mx-auto max-w-3xl px-4 py-16 md:py-24">
    {@render children()}
</div>
