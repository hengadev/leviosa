<script lang="ts">
	import { eventstate } from '$lib/stores/eventbar';
	eventstate.set('Evenements a venir');

	// TODO: is this the thing to do ?
	import { reservationstate } from '$lib/stores/reservationtab';
	reservationstate.set('consultations');

	import EventCard from '$lib/components/ui/EventCard.svelte';
	import ConsultationCard from '$lib/components/ui/ConsultationCard.svelte';
	import NoEvent from './NoEvent.svelte';
	import NoConsultation from './NoConsultation.svelte';

	interface Props {
		data: import('./$types').PageData;
	}

	let { data }: Props = $props();
	let { cards } = data;

	// TODO: use the eventCard with the right content needed
</script>

{#if $reservationstate === 'events'}
	<div class="content grid" style="padding-inline: 0.75rem;">
		<div class="grid" style="margin-top: 1rem;">
			{#if cards.length > 0}
				<EventCard />
			{:else}
				<NoEvent />
			{/if}
		</div>
	</div>
{:else}
	<div class="content grid" style="padding-inline: 0.75rem;">
		{#if cards.length > 0}
			<ConsultationCard id="" />
		{:else}
			<NoConsultation />
		{/if}
	</div>
{/if}

<style>
	.content {
		padding-top: 1rem;
		padding-bottom: 7rem;
	}
</style>
