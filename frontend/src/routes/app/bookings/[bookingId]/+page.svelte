<script lang="ts">
	// import type { PageData } from './$types';
	// export let data: PageData;
	// $: ({ bookings } = data);
	import VoteComponent from '$lib/components/events/VoteComponent.svelte';
</script>

<div class="content">
	<VoteComponent />
	<VoteComponent />
	<VoteComponent />
	<VoteComponent />
</div>

<style>
	.content {
		display: grid;
		place-content: center;
		height: 100vh;
	}
</style>
