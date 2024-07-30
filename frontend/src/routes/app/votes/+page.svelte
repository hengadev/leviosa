<script>
	import VoteComponent from '$lib/components/votes/VoteComponent.svelte';
	export let data;
	import { months } from '$lib/types';
	$: ({ futurVotes } = data);
	$: _futurVotes = JSON.parse(JSON.stringify(futurVotes));
	// $: console.log('the year', futurVotes[0].year);
	// $: console.log('the month : ', months[futurVotes[0].month]);
</script>

<div class="content">
	<h2 class="title">Votes</h2>
	<h3 class="subtitle">Votes a venir</h3>
	{#each _futurVotes as vote}
		{@const writtenMonth = months[vote.month - 1]}
		<div class="vote">
			<a href={`/app/votes/${vote.year}/${writtenMonth}`}>
				<VoteComponent {...vote} />
			</a>
		</div>
	{/each}
</div>

<style>
	.content {
		padding: 1rem;
		margin-bottom: 4rem;
	}

	/* TODO: Put that thing as a global style friend */
	.title {
		font-size: 1.5rem;
		color: #f67373;
		font-weight: 700;
	}

	.subtitle {
		font-size: 1.2rem;
		color: rgba(255, 255, 245, 0.86);
		font-weight: 600;
		margin-top: 2rem;
	}
	.vote {
		margin-block: 1rem;
	}
</style>
