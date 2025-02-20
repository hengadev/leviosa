<script lang="ts">
    import EventNavigationBarIcon from "./atoms/EventNavigationBarIcon.svelte";
    import { eventstate } from "$lib/stores/eventbar";

    import type { EventTabs } from "$lib/types";

    interface Props {
        role: import("$lib/types").Role;
    }
    let { role }: Props = $props();

    // NOTE: here I put all the possible links possible, I need to split that by roles
    let tabs: EventTabs = {
        user: [],
        userPremium: [
            { name: "Evenements a venir", href: "/app/reservations" },
            { name: "Reserve ta place", href: "/app/reservations/events/book" },
            {
                name: "Creer un evenement",
                href: "/app/reservations/events/new",
            },
        ],
        freelance: [],
        helper: [],
        admin: [],
    };

    // { name: "Reservez votre place", href: "/app/reservations/events/book" },
    // {
    //     name: "Inscris toi a un evenement",
    //     href: "/app/reservations/events/new",
    // },
    // {
    //     name: "Evenements en attente",
    //     href: "/app/reservations/events/waiting",
    // }
</script>

<nav class="container navigation-bar snaps-inline grid" style="--gap: 1.5rem;">
    {#each tabs[role] as { name, href }}
        <div>
            <EventNavigationBarIcon
                {name}
                {href}
                active={$eventstate === name}
            />
        </div>
    {/each}
</nav>

<style>
    .navigation-bar {
        grid-auto-flow: column;
        /* grid-auto-columns: 16%; */
        gap: 2rem;
        border-bottom: 1px solid hsl(var(--clr-stroke));
        text-wrap: nowrap;

        overflow-x: auto;
        scrollbar-width: none;
        overscroll-behavior-inline: contain;
    }
    .snaps-inline {
        scroll-snap-type: inline mandatory;
        scroll-padding-inline: 1.5rem;
    }
    .snaps-inline > * {
        scroll-snap-align: start;
    }
</style>
