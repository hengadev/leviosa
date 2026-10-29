<script lang="ts">
    import PracticePage from "./_page.svelte";
    import type { PageProps } from "./$types";

    let { data }: PageProps = $props();
</script>

<PracticePage locale="fr" {data} />
