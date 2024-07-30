<script lang="ts">
	// import ArrowUp from 'lucide-svelte/icons/arrow-up';
	// import ArrowDown from 'lucide-svelte/icons/arrow-down';

	export let votes;
	// $: _votes = votes;
	// $: _votes = JSON.parse(JSON.stringify(votes));

	// const moveUp = (event: MouseEvent) => {
	// 	let targetEl = event.currentTarget as HTMLDivElement;
	// 	let id = Number(targetEl.id);
	// 	if (id === 0) return;
	// 	_votes[id] = _votes.splice(id - 1, 1, _votes[id])[0];
	// 	// [_votes[id - 1], _votes[id]] = [_votes[id], _votes[id - 1]];
	// 	console.log('the new votes is : ', _votes);
	// };

	// const moveDown = (event: MouseEvent) => {
	// 	let targetEl = event.currentTarget as HTMLDivElement;
	// 	let id = Number(targetEl.id);
	// 	if (id === votes.lenght - 1) return;
	// 	_votes[id] = _votes.splice(id + 1, 1, _votes[id])[0];
	// 	// [_votes[id + 1], _votes[id]] = [_votes[id], _votes[id + 1]];
	// 	console.log('the new votes is : ', _votes);
	// };

	// TODO: how to handle the selection to send the form?
	// -> I am going to toggle a checkbox or sth
	let active: string[] = [];
	function toggleActive(event: MouseEvent) {
		console.log('click');
		let targetEl = event.currentTarget as HTMLButtonElement;
		let id = targetEl.id;
		if (!active.includes(id)) {
			active = [...active, id];
		} else {
			active = active.filter((val) => val != id);
		}
	}
</script>

{#each votes as vote, i}
	{@const i_str = String(i)}
	<div class="vote" id={i_str} style:background-color={active.includes(i_str) ? '#202127' : 'none'}>
		<input type="hidden" name={i_str} value={vote.day} />
		<button type="button" class="vote__content" on:click={toggleActive}>
			the vote is for the day {vote.day}
		</button>
	</div>
{/each}

<style>
	.vote {
		border: 1px solid rgba(60, 60, 67, 0.78);
		/* border: 1px solid #bdbdbd; */
		border-radius: 0.5rem;
		width: 100%;
		display: flex;
		align-items: center;
		/* TODO: add that style on click */
		/* background-color: #202127; */
	}

	.vote__content {
		padding: 1rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
	}
</style>
