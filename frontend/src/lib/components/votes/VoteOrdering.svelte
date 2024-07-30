<script lang="ts">
	import ArrowUp from 'lucide-svelte/icons/arrow-up';
	import ArrowDown from 'lucide-svelte/icons/arrow-down';

	export let votes;
	$: _votes = JSON.parse(JSON.stringify(votes));

	const moveUp = (event: MouseEvent) => {
		let targetEl = event.currentTarget as HTMLDivElement;
		let id = Number(targetEl.id);
		if (id === 0) return;
		_votes[id] = _votes.splice(id - 1, 1, _votes[id])[0];
	};

	const moveDown = (event: MouseEvent) => {
		let targetEl = event.currentTarget as HTMLDivElement;
		let id = Number(targetEl.id);
		if (id === votes.lenght - 1) return;
		_votes[id] = _votes.splice(id + 1, 1, _votes[id])[0];
	};
</script>

{#each _votes as vote, i}
	{@const indice = i + 1}
	{@const i_str = String(i)}
	<input type="hidden" name={i_str} value={vote.day} />
	<div class="vote">
		<div class="indice">{indice}</div>
		<div class="vote__content">
			<button type="button" class="arrowBtn" id={i_str} on:click={moveDown}>
				<ArrowDown />
			</button>
			<div>
				the vote is for the day {vote.day}
			</div>
			<button type="button" class="arrowBtn" id={i_str} on:click={moveUp}>
				<ArrowUp />
			</button>
		</div>
	</div>
{/each}

<style>
	.indice {
		padding: 1rem 1.5rem;
		/* background-color: rgba(60, 60, 67, 0.78); */
		border: 1px solid rgba(60, 60, 67, 0.78);
		border-top: 0;
		border-bottom: 0;
		color: white;
		font-weight: 600;
		max-width: 65px;
		width: 65px;
		text-align: center;
	}

	.arrowBtn {
		all: unset;
	}

	.vote {
		border: 1px solid rgba(60, 60, 67, 0.78);
		/* border: 1px solid #bdbdbd; */
		border-radius: 0.5rem;
		display: flex;
		align-items: center;
		/* justify-content: space-between; */
	}

	.vote__content {
		padding: 1rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
	}
</style>
