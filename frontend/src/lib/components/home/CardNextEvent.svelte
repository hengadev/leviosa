<script lang="ts">
	//icons
	export let eventId: string;
	import Clock from 'lucide-svelte/icons/clock';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	// TODO: get the user id for the next event for the user
</script>

<div class="block">
	<div class="block__content">
		<div class="content__top">
			<div class="content__date">
				<p class="day">30</p>
				<p class="month">Aug</p>
			</div>
			<button class="detail">
				<a href={`/events/${eventId}`}>Voir details</a>
				<ChevronRight color="rgba(235, 235, 245, 0.6)" />
			</button>
		</div>
		<div class="content__bottom">
			<div>
				<h4 class="event__title">Relaxation & detente</h4>
				<div class="content__paragraph">
					<Clock color="rgba(235, 235, 245, 0.6)" />
					<p class="block__text">Dance et art</p>
				</div>
			</div>
			<div class="content__paragraph">
				<Clock color="rgba(235, 235, 245, 0.6)" />
				<p class="block__text">07:00 PM</p>
			</div>
		</div>
	</div>
</div>

<style>
	.block {
		width: 100%;
		margin-top: 1rem;
		margin-inline: auto;
		background-color: #202127;
		border-radius: 0.5rem;
		border: 1px solid rgba(60, 60, 67, 0.78);
	}

	.block__content {
		padding: 1rem;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.content__top {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.content__date {
		padding: 0.2rem 0.8rem;
		background-color: rgba(60, 60, 67, 0.78);
		display: flex;
		flex-direction: column;
		border-radius: 0.5rem;
		text-align: center;
		/* border: 1px solid rgba(60, 60, 67, 0.78); */
	}

	.detail {
		display: flex;
		gap: 0.5rem;
		padding: 0;
		width: fit-content;
	}

	.day,
	.month {
		font-weight: 700;
	}

	.day {
		color: #f67373;
	}

	.content__bottom {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		margin-top: 3rem;
		font-size: 0.95rem;
	}

	.content__paragraph {
		display: flex;
		gap: 0.5rem;
		align-items: flex-end;
		margin-top: 0.2rem;
	}

	.event__title {
		font-size: 1.2rem;
		color: rgba(255, 255, 245, 0.86);
		color: white;
	}
</style>
