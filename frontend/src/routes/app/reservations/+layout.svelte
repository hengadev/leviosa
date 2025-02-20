<script lang="ts">
    import { navstate } from "$lib/stores/navbar";
    navstate.set("reservations"); // just to forget the value stored in localstore when reconecting and I had the page to another link.

    import EventNavigationBar from "$lib/components/navigation/EventNavigationBar.svelte";
    import ConsultationNavigationBar from "$lib/components/navigation/ConsultationNavigationBar.svelte";
    import Tabs from "$lib/components/Tabs.svelte";

    import { Images } from "lucide-svelte";

    import { redirectTo } from "$lib/scripts/redirect";

    import { reservationstate } from "$lib/stores/reservationtab";
    interface Props {
        children?: import("svelte").Snippet;
        data: import("./$types").PageData;
    }

    let { children, data }: Props = $props();
    let { role } = data;
    function handleTab(): void {
        if ($reservationstate === "consultations")
            reservationstate.set("events");
        else reservationstate.set("consultations");
    }
    const offers = [
        { name: "Consultations", action: () => handleTab() },
        { name: "Evenements", action: () => handleTab() },
    ];
</script>

<div class="page">
    <div class="event-header grid">
        <div class="flex container event-header-top">
            <h2 class="page-title">Reservations</h2>
            <button onclick={() => redirectTo("galerie")}>
                <Images strokeWidth={1.5} absoluteStrokeWidth={true} />
            </button>
        </div>
        <div class="container">
            <Tabs {offers} isSecondary={true} />
        </div>
        {#if $reservationstate === "consultations"}
            <ConsultationNavigationBar {role} />
        {:else}
            <EventNavigationBar {role} />
        {/if}
    </div>
    <div class="slot">
        {@render children?.()}
    </div>
</div>

<style>
    .page {
        display: flex;
        flex-direction: column;
        min-height: 100vh;
    }
    .slot {
        flex-grow: 1;
        overflow: auto;
        /* background-color: hsl(var(--clr-light-secondary)); */
        background-color: #f7f7f9;
    }
    .event-header {
        background-color: hsl(var(--clr-light-primary));
        padding-top: 2rem;
        flex-shrink: 0;
        flex: none;
    }
    .event-header-top {
        justify-content: space-between;
        align-items: center;
    }
</style>
