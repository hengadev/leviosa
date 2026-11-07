<script lang="ts">
    import type { PageData } from "./$types";
    import Settings from "./Settings.svelte";

    let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Paramètres | Admin</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
    <div class="mb-8">
        <h1 class="text-3xl lg:text-4xl font-bold mb-2 text-foreground">Paramètres</h1>
        <p class="text-muted-foreground">Configurez les informations de l'entreprise et les paramètres système</p>
    </div>
    <!-- Forms read better narrow than across the whole page. -->
    <div class="max-w-4xl">
        <Settings {data} />
    </div>
</div>
