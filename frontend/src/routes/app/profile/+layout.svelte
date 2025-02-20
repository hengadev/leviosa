<script lang="ts">
    // make the layout for all the profile page with the right title
    // TODO: change the title style to make it more classy
    // inspiration for the layout :
    import { navstate } from "$lib/stores/navbar";
    interface Props {
        children?: import('svelte').Snippet;
    }

    let { children }: Props = $props();
    navstate.set("profil"); // just to forget the value stored in localstore when reconecting and I had the page to another link.
</script>

<div class="content">
    <div class="grid">
        <div class="event-header flex container">
            <h2 class="page-title">Profil</h2>
        </div>
    </div>
    {@render children?.()}
    <p class="copyright">© 2024 Leviosa, Inc. Tous droits reserves</p>
</div>

<style>
    .content {
        padding-bottom: 7rem;
    }
    .event-header {
        background-color: hsl(var(--clr-light-primary));
        padding-block: 2rem;
        align-items: center;
        justify-content: space-between;
    }
    .copyright {
        margin-top: 1rem;
        text-align: center;
        font-size: 0.8125rem;
        color: hsl(var(--clr-grey-400));
    }
</style>
