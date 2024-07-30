<script lang="ts">
	export let eventCount: number;
	export let month;
</script>

<div class="vote">
	<p class="month">{month}</p>
	<p class="places">{eventCount} event{eventCount > 1 ? 's' : ''}</p>
</div>

<style>
	.vote {
		background-color: #202127;
		width: fit-content;
		border-radius: 1rem;
		border-radius: 0.5rem;
		aspect-ratio: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-direction: column;
		padding: 1rem;
		border: 1px solid #bdbdbd;
		border: 1px solid rgba(60, 60, 67, 0.78);
	}
	.month,
	.places {
		font-weight: 700;
	}
	.month {
		color: #f67373;
	}
	.places {
	}
</style>
