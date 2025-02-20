<script>
    import { navstate } from "$lib/stores/navbar";
    navstate.set("conversations"); // just to forget the value stored in localstore when reconecting and I had the page to another link.

    import { redirectTo } from "$lib/scripts/redirect";

    import { onMount } from "svelte";
    onMount(() => {
        if (window.matchMedia("(max-width: 500px)").matches) {
            console.log("need to redirect brother");
            redirectTo("/app/messages");
        }
    });
</script>
