<script lang="ts">
	import type { PageServerData } from './$types';
	interface Props {
		data: PageServerData;
	}

	let { data }: Props = $props();
	let { photos } = $derived(data);
</script>

<div class="content">
	<div>Do some loading UI for the waiting</div>
	{#each photos as photo}
		<div>{photo}</div>
	{/each}
</div>

<style>
	.content {
		height: 100vh;
		display: grid;
		place-content: center;
	}
</style>
