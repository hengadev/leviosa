<script lang="ts">
    import type { PageProps } from './$types';
    import Cards from "./_cards.svelte";
    import Agenda from "./_agenda.svelte";
    import Actions from "./_actions.svelte";
    import Activity from "./_activity.svelte";
    import Volume from "./_volume.svelte";

    import { onMount } from 'svelte';

    let { data }: PageProps = $props();
    let visible = $state(false);
    onMount(() => { visible = true; });
</script>

<svelte:head>
	<title>Accueil | Staff</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
    <div class="mb-8">
        <h1 class="text-3xl lg:text-4xl font-bold mb-1 text-foreground">Tableau de bord</h1>
        <p class="text-muted-foreground">Votre journée et l'activité de votre cabinet</p>
    </div>

    <!-- KPI row with staggered delay -->
    <div class="transition-all duration-700 delay-150 ease-out {visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'}">
        <Cards kpi={data.kpi} />
    </div>

    <!-- Two-column layout -->
    <div class="mt-10 grid grid-cols-1 xl:grid-cols-[1fr_360px] gap-8">
        <div class="space-y-6 transition-all duration-700 delay-300 ease-out {visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'}">
            <Agenda slots={data.todaySlots} />
            <Volume days={data.volumeDays} />
        </div>
        <div class="space-y-6 transition-all duration-700 delay-[400ms] ease-out {visible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'}">
            <Actions actions={data.actions} />
            <Activity events={data.activityEvents} />
        </div>
    </div>
</div>
