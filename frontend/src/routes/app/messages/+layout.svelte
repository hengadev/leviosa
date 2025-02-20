<script lang="ts">
    import { navstate } from "$lib/stores/navbar";

    import { redirectTo } from "$lib/scripts/redirect";
    import { onMount } from "svelte";
    onMount(() => {
        if (window.matchMedia("(max-width: 500px)").matches) {
            console.log("need to redirect brother");
            navstate.set("messages"); // just to forget the value stored in localstore when reconecting and I had the page to another link.
            redirectTo("/app/messages");
        } else if (window.matchMedia("(min-width: 500px)").matches) {
            if($navstate !== "conversations" && $navstate !== "notes de seances"){
                navstate.set("conversations")
            } 
        }
    });

    import MessageNavigationBar from "$lib/components/navigation/MessageNavigationBar.svelte";
    import Conversations from "./[id]/Conversations.svelte";
    import NoteDeSeance from "./[id]/NoteDeSeance.svelte";

    import type { LayoutData } from "./$types";
    interface Props {
        data: LayoutData;
        children?: import('svelte').Snippet;
    }
    let { data, children }: Props = $props();
    const { messages, notes } = data;

    import { messagestate } from "$lib/stores/messagebar";

    import { SquarePen } from "lucide-svelte";

    import NewMessage from "./NewMessage.svelte";
    import Drawer from "$lib/components/Drawer.svelte";


    let isDrawerOpen: boolean = $state(false);

    function toggleDrawer() {
        // make sur  that you have different action based on the window size
        // on small screen I want to open the drawer, on bigger screen, I want to open a modal
        isDrawerOpen = !isDrawerOpen;
    }
</script>

<div class="content">
    <div class="left">
        <div class="grid message-header">
            <div
                class="flex container message-header-content"
                style="margin-bottom: 1rem;"
            >
                <div>
                    <h2 class="page-title">Messages</h2>
                    <p>Afin de garder le contact</p>
                </div>
                {#if $messagestate === "Conversations"}
                    <div class="icons">
                        <button class="new-message" onclick={toggleDrawer}>
                            <SquarePen />
                        </button>
                    </div>
                {/if}
            </div>
            <div class="message-navigation-bar">
                <MessageNavigationBar />
            </div>
        </div>
        {@render children?.()}
    </div>
    <div class="right">
        {#if $navstate === "conversations" || $navstate === "messages"}
            <Conversations {messages} />
        {:else if $navstate === "notes de seances"}
            <NoteDeSeance {notes}/>
        {/if}
    </div>
</div>
<Drawer bind:isOpen={isDrawerOpen} closeDrawer={() => (isDrawerOpen = false)}>
    <NewMessage />
</Drawer>

<!-- TODO: make that component and then make sure that the new message display differently on modal -->
<!-- <Popup bind:isOpen={isModalOpen} closeModal={() => (isModalOpen = false)}> -->
<!--     <NewMessage /> -->
<!-- </Popup> -->

<style>
    .content {
        /* view-transition-name: pushing; */
        position: relative;
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
    }
    .left,
    .right {
        height: 100vh;
        overflow-y: auto;
    }
    /* now that thing when the screen is small brother */
    /* then get the content for the page in right, and change that class name, it is horrible */
    .right {
        display: none;
        visibility: hidden;
        height: 100vh;
        overflow-y: auto;

        /* HACK: that value need to be found with the auto fit thing, I do not see any formula that make it work otherwise */
        @media only screen and (min-width: 920px) {
            display: initial;
            visibility: visible;
        }
    }

    /* HACK: that value need to be found with the auto fit thing, I do not see any formula that make it work otherwise */
    @media only screen and (min-width: 920px) {
        .message-navigation-bar {
            display: none;
            visibility: hidden;
        }
    }

    .message-header {
        background-color: hsl(var(--clr-light-primary));
        padding-top: 2rem;
    }
    .message-header-content {
        align-items: center;
        justify-content: space-between;
    }
    .new-message {
        background: transparent;
    }
    .icons {
        display: grid;
        place-content: center;
        border-radius: 100%;
        padding: 0.75rem;
        background-color: #f7f7f9;
    }
</style>
