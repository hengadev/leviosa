<script lang="ts">
	// data
	import type { PageData } from './$types';
	// user data
	import { page } from '$app/stores';
	const name = $page.data.user.firstname;
	// components
	import * as Accordion from '$lib/components/ui/accordion';
	import CardNextEvent from '$lib/components/home/CardNextEvent.svelte';
	import NextVote from '$lib/components/home/NextVote.svelte';
	import { navstate } from '$lib/stores/navbar';
	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
	navstate.set('home'); // just to forget the value stored in localstore when reconecting and I had the page to another link.
	let { accordionItems, nextVotes } = $derived(data);
	let _nextVotes = $derived(JSON.parse(JSON.stringify(nextVotes)));
</script>

{#if $page.data.user}
	<div class="content">
		<h2 class="title">Bienvenue, {name}</h2>
		<h3 class="subtitle">Explication du concept</h3>
		<div class="accordion">
			<Accordion.Root>
				{#each accordionItems as item, i}
					<Accordion.Item value={`item-${i}`}>
						<Accordion.Trigger>{item.trigger}</Accordion.Trigger>
						<Accordion.Content>{item.content}</Accordion.Content>
					</Accordion.Item>
				{/each}
			</Accordion.Root>
		</div>
		<h3 class="subtitle">Votre prochain evenement</h3>
		<CardNextEvent eventId="some eventId" />
		<h3 class="subtitle">Les derniers votes ouverts.</h3>
		<div class="votes">
			{#each _nextVotes as vote}
				<NextVote {...vote} />
			{/each}
		</div>
	</div>
{/if}

<style>
	/* TODO: make the fonts responsive */
	.content {
		/* padding: 2rem 1.5rem; */
		padding-inline: 1.5rem;
		display: grid;
		overflow-y: scroll;
		margin-bottom: 12rem;
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

	.accordion {
		border-radius: 0.5rem;
		border: 1px solid rgba(60, 60, 67, 0.78);
		padding: 1rem;
		padding-bottom: 0;
	}
	.accordion,
	.votes {
		margin-top: 1rem;
	}

	.votes {
		display: flex;
		align-items: center;
		justify-content: space-evenly;
		justify-content: space-around;
	}
</style>
