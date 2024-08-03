<script lang="ts">
	// import type { PageData } from './$types';
	// export let data: PageData;
	// $: ({ bookings } = data);
	import VoteComponent from '$lib/components/votes/VoteComponent.svelte';
</script>

<div class="content">
	<VoteComponent year={2024} month={8} />
	<VoteComponent year={2024} month={8} />
	<VoteComponent year={2024} month={8} />
	<VoteComponent year={2024} month={8} />
</div>

<style>
	.content {
		display: grid;
		place-content: center;
		height: 100vh;
	}
</style>
