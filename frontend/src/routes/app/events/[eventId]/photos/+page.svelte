<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
	$: ({ getPhotos } = data);
</script>

<div class="content">
	{#await getPhotos()}
		<div>Do some loading UI for the waiting</div>
	{:then photos}
		{#each photos as photo}
			<div>{photo}</div>
		{/each}
	{/await}
</div>

<style>
	.content {
		height: 100vh;
		display: grid;
		place-content: center;
	}
</style>
