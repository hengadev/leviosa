<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	// NOTE: Use one of these two options for the voting system.
	import VoteOrdering from '$lib/components/votes/VoteOrdering.svelte';
	// import VoteSelect from '$lib/components/votes/VoteSelect.svelte';

	// TODO: use the isDefault to display some informations on our vote work.
	export let data;
	$: ({ month, year, isDefault, votes } = data);
	$: _votes = JSON.parse(JSON.stringify(votes));
	$: _month = JSON.parse(JSON.stringify(month));
	$: _year = JSON.parse(JSON.stringify(year));

	const reset = () => (_votes = JSON.parse(JSON.stringify(votes)));
</script>

<div class="content">
	<h2 class="title">Vote pour le mois de <span class="month">{_month}</span> {_year}</h2>
	<div class="subtitle">Description de l'evenement</div>
	<p class="text">
		Lorem ipsum dolor sit amet consectetur adipisicing elit. Saepe, natus necessitatibus impedit in
		magni veniam delectus accusamus quaerat sunt ullam sed harum temporibus incidunt tempora!
		<a href="/details">Voir details</a>
	</p>
	<div class="subtitle">Comment le vote fonctionne ?</div>
	<p class="text">
		Selectionne les dates succeptibles de t'interesser pour le mois de <span class="month"
			>{_month}. Puis clique sur le bouton enregistrer pour envoyer le formulaire.</span
		>
	</p>
	<div class="subtitle">List ordering</div>
	<form
		class="events"
		method="POST"
		use:enhance={() => {
			return async ({ result }) => {
				invalidateAll();
				await applyAction(result);
			};
		}}
	>
		<button type="button" class="" on:click={reset}>Reinitialiser</button>
		<input type="hidden" name="count" value={_votes.length} />
		<VoteOrdering votes={_votes} />
		<button class="submit" type="submit">Enregistrer</button>
	</form>
</div>

<style>
	a {
		text-decoration: underline;
		color: #2ea6ff;
	}

	.content {
		padding: 1rem;
		margin-bottom: 8rem;
	}

	.text {
		margin-top: 0.25rem;
	}

	.title {
		font-size: 1.5rem;
		color: #f67373;
		font-weight: 700;
	}

	.events {
		margin-top: 1rem;
		padding: 1rem;
		border: 1px solid rgba(60, 60, 67, 0.78);
		border: 1px solid #bdbdbd;
		border-radius: 0.5rem;
	}

	button {
		border: 1px solid rgba(60, 60, 67, 0.78);
		border-bottom-left-radius: 0;
		border-bottom-right-radius: 0;
	}

	.subtitle {
		font-size: 1.2rem;
		color: rgba(255, 255, 245, 0.86);
		font-weight: 600;
		margin-top: 1rem;
	}

	.month {
		text-transform: capitalize;
	}

	.submit {
		width: 100%;
		margin-top: 2rem;
		color: #171717;
		background-color: white;
		border-radius: 0.5rem;
		font-weight: 500;
	}

	.submit:is(:hover, :focus) {
		cursor: pointer;
		opacity: 0.9;
	}
</style>
