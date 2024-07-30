<script lang="ts">
	import Vote from '$assets/Vote.svg';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import { months, vowelRegex } from '$lib/types';
	export let year: number;
	// NOTE: use the - 1 since months is indexed at 0 but I indexed the month in my backend at 1.
	export let month: number;
	// NOTE: Do I need the location thing ?
	// export let location: string | null;
	let location = '01 Impasse Hoche, 94200 Ivry-Sur-Seine';
	// let location = null;
</script>

<div class="event">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-4">
			<img src={Vote} class="w-[20px]" alt="voteIcon" />
			<div class="">
				<h3 class="event__title">
					Votes {months[month - 1].match(vowelRegex) ? "d'" : 'de'}
					<span class="month">{months[month - 1]}</span>
					{year} ouverts !
				</h3>
				<p class="text-sm">Participe a l'orgnaisation de ton evenement</p>
				{#if location !== null}
					<p class="localisation">{location}</p>
				{:else}
					<p class="text-amber-400">Aucune localisation pour le moment</p>
				{/if}
			</div>
		</div>
		<ChevronRight />
	</div>
</div>

<style>
	.event {
		background-color: #202127;
		border-radius: 0.5rem;
		padding: 1rem;
		border: 1px solid rgba(60, 60, 67, 0.78);
	}

	.event:is(:hover, :focus) {
		cursor: pointer;
	}

	.event__title {
		/* margin-bottom: 2rem; */
		font-size: 1.1rem;
		color: white;
	}
	.month {
		text-transform: capitalize;
	}
	.localisation {
		text-decoration: underline;
		/* color: #f67373; */
		/* color: white; */
	}
</style>
