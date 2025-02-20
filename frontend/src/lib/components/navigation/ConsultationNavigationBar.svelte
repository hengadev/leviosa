<script lang="ts">
    import ConsultationNavigationBarIcon from "./atoms/ConsultationNavigationBarIcon.svelte";
    import { consultationstate } from "$lib/stores/consultationbar";

    import type { ConsultationTabs } from "$lib/types";
    interface Props {
        role: import("$lib/types").Role;
    }
    let { role }: Props = $props();

    let tabs: ConsultationTabs = {
        user: [],
        userPremium: [
            { name: "Consultations a venir", href: "/app/reservations" },
            {
                name: "Reserve ta consultation",
                href: "/app/reservations/consultations/book",
            },
        ],
        freelance: [],
        helper: [],
        admin: [],
    };

    // {
    //     name: "Creer un type de consulation",
    //     href: "/app/reservations/events/new",
    // },
</script>

<nav class="container navigation-bar snaps-inline grid" style="--gap: 1.5rem;">
    {#each tabs[role] as { name, href }}
        <div>
            <ConsultationNavigationBarIcon
                {name}
                {href}
                active={$consultationstate === name}
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
