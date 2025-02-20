<script lang="ts">
    import { navstate } from "$lib/stores/navbar";
    navstate.set("services"); // just to forget the value stored in localstore when reconecting and I had the page to another link.
    // =======================
    // Components Imports
    // =======================

    // Global components
    // import Tabs from "$lib/components/Tabs.svelte";
    import Tabs from "./Tabs.svelte";
    import ServiceNavigationBar from "$lib/components/navigation/ServiceNavigationBar.svelte";
    import CarouselIndicator from "$lib/components/ui/CarouselIndicator.svelte";

    // Local components
    import Deroule from "./Deroule.svelte";
    import Prestataires from "./Prestataires.svelte";
    import APropos from "./APropos.svelte";

    // =======================
    // Types Imports
    // =======================

    // SvelteKit types
    import type { PageData } from "./$types";
    import { fade } from "svelte/transition";

    // Custom types
    import type { Offer, getOffersRes } from "$lib/types";

    // =======================
    // Store and State Imports
    // =======================

    import { servicestate } from "$lib/stores/servicebar";
    import { createHorizontalSwipeHandler } from "$lib/scripts/swipe";

    // ============================
    // Props and Reactive Variables
    // ============================

    interface Props {
        data: PageData;
    }

    let { data }: Props = $props();
    let { offers }: { offers: Offer[] } = data;
    let selectedServiceIndex: number = $state(0);
    let selectedOfferIndex: number = $state(0);
    let bgurl: string = $state("");
    let bgblur: number = $state(0);
    let currentOffer: Offer = $derived(offers[selectedOfferIndex]);
    let currentService = $derived(currentOffer.services[selectedServiceIndex]);
    $effect(() => {
        bgurl = currentService.bgurl;
        bgblur = currentService.bgblur;
    });

    let background: HTMLDivElement | undefined = $state();
    let serviceNavBar: any = $state();

    // =======================
    // Helper Functions
    // =======================

    // Retrieves formatted offers with actions
    function getOffers(offers: Offer[]): getOffersRes[] {
        return offers.map((offer: Offer, index: number) => ({
            name: offer.type,
            action: () => {
                selectedServiceIndex = 0;
                selectedOfferIndex = index;
            },
        }));
    }

    // Handle swipe navigation for service items
    function swipeService(direction: "left" | "right"): void {
        const totalItems = currentOffer.services.length || 0;
        servicestate.set("A propos"); // Reset to initial state
        if (direction === "left") {
            selectedServiceIndex = (selectedServiceIndex + 1) % totalItems;
            const nextservice = currentOffer.services[selectedServiceIndex];
            bgurl = nextservice.bgurl;
            bgblur = nextservice.bgblur;
            if (background) {
                background.style.background = `url(${bgurl})`;
                background.style.filter = `blur(${bgblur})px`;
            }
        } else if (direction === "right") {
            selectedServiceIndex =
                (selectedServiceIndex - 1 + totalItems) % totalItems;
            const nextservice = currentOffer.services[selectedServiceIndex];
            bgurl = nextservice.bgurl;
            bgblur = nextservice.bgblur;
            if (background) {
                background.style.background = `url(${bgurl})`;
                background.style.filter = `blur(${bgblur})px`;
            }
        }
    }

    // Handle swipe navigation for content state changes
    function swipeContent(direction: "left" | "right"): void {
        if (direction === "left") {
            if ($servicestate === "A propos") servicestate.set("Deroule");
            else if ($servicestate === "Deroule")
                servicestate.set("Prestataires");
        }
        if (direction === "right") {
            if ($servicestate === "Deroule") servicestate.set("A propos");
            else if ($servicestate === "Prestataires")
                servicestate.set("Deroule");
        }
    }

    // =======================
    // Swipe Handler Actions
    // =======================

    const swipeServiceAction = createHorizontalSwipeHandler(swipeService);
    const swipeContentAction = createHorizontalSwipeHandler(swipeContent);
    // TODO: put the bgblur to 200 and remove the isLight thing because all text are going to be black
    // TODO: remove the nav bar
</script>

<div class="content">
    <div class="overlay-background"></div>
    <div
        class="background"
        style="background:url({bgurl}); filter:blur({bgblur}px) brightness(30%);"
        bind:this={background}
    ></div>
    <div class="header container grid">
        <h2 class="page-title" style="color: hsl(var(--clr-light-primary));">
            Nos services
        </h2>
        <div class="separator"></div>
    </div>
    {#if offers.length > 0}
        <div class="tabs-container flex" style="margin-top: 1rem;">
            <Tabs offers={getOffers(offers)} />
        </div>
    {/if}
    <div
        class="flow offer-text"
        style="--flow-space: 2rem;"
        transition:fade={{ duration: 300 }}
    >
        <div class="images" use:swipeServiceAction.action>
            <img class="image" src={currentService.bgurl} alt="" />
        </div>
        <div class="carousel-indicator flex">
            <CarouselIndicator
                isLight={false}
                count={currentOffer.services.length}
                activeIndex={selectedServiceIndex}
            />
        </div>
        <div
            class="grid container"
            style="--gap: 0.1rem;"
            use:swipeServiceAction.action
        >
            <h2 class="title">{currentService.name}</h2>
            <p>
                {currentService.label}
            </p>
        </div>
        <div class="container">
            <ServiceNavigationBar type={"light"} bind:this={serviceNavBar} />
        </div>
        <div use:swipeContentAction.action>
            {#if $servicestate === "A propos"}
                <APropos
                    description={currentService.description}
                    duration={currentService.duration}
                    clients_count={currentService.clients_count}
                    positive_responses={currentService.positive_responses}
                />
            {:else if $servicestate === "Deroule"}
                <Deroule />
            {:else if $servicestate === "Prestataires"}
                <Prestataires freelancers={currentService.freelancers} />
            {/if}
        </div>
    </div>
</div>

<style>
    /* NOTE: the thing is that it moves down with the drawer opening */
    .content {
        padding-block: 2rem;
        position: relative;
        color: hsl(var(--clr-grey-200));
    }
    .background,
    .overlay-background {
        position: absolute;
        left: 0;
        top: 0;
        width: 100%;
        height: 100%;
        z-index: -1;
    }
    .background {
        background-position: center;
    }
    .overlay-background {
        background-color: hsl(var(--clr-dark-ternary));
    }
    .header {
        color: hsl(var(--clr-light-primary));
    }
    .tabs-container {
        justify-content: center;
    }
    .offer-text {
        padding-bottom: 6rem;
    }
    .images {
        margin-top: 2rem;
        display: flex;
        justify-content: center;
    }
    .image {
        /* TODO: change that this is horrible brother */
        width: 85%;
        /* height: 25vh; */
        /* height: 65%; */
        /* aspect-ratio: 8/3; */

        height: 100%;
        border-radius: 0.5rem;

        /* TODO: do the response for when the page gets bigger */
        @media only screen and (min-width: 500px) {
            width: 45%;
        }
    }
    .carousel-indicator {
        justify-content: center;
    }
    .title {
        font-size: var(--fs-2);
        font-weight: 800;
        color: hsl(var(--clr-light-primary));
    }
</style>
