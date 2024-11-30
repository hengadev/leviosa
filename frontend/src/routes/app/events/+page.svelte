<script lang="ts">
	import EventComponent from '$lib/components/events/EventComponent.svelte';
	import type { PageData } from './$types';
	import { page } from '$app/stores';
	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
	let { pastEvents, nextEvents, incomingEvents } = $derived(data);
	let _pastEvents = $derived(JSON.parse(JSON.stringify(pastEvents)));
	let _nextEvents = $derived(JSON.parse(JSON.stringify(nextEvents)));
	let _incomingEvents = $derived(JSON.parse(JSON.stringify(incomingEvents)));
</script>

{#if $page.data.user}
	<div class="content">
		<h2 class="title">Mes evenements</h2>
		<h3 class="subtitle">Evenements a venir</h3>
		{#each _incomingEvents as event}
			<EventComponent {...event} />
		{/each}
		<h3 class="subtitle">Evenements qui pourrait vous interesser</h3>
		{#each _nextEvents as event}
			<EventComponent {...event} />
		{/each}
		<h3 class="subtitle">Evenements passes</h3>
		{#each _pastEvents as event}
			<EventComponent {...event} />
		{/each}
	</div>
{/if}

<style>
	.content {
		padding: 1rem;
		margin-bottom: 4rem;
	}

	.title {
		font-size: 1.5rem;
		color: #f67373;
		font-weight: 700;
	}

	.subtitle {
		font-size: 1.2rem;
		color: rgba(255, 255, 245, 0.86);
		font-weight: 600;
		margin-top: 3rem;
	}
</style>
