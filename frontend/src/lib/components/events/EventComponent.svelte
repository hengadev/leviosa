<script lang="ts">
	import CalendarDays from 'lucide-svelte/icons/calendar-days';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	// TODO: use the props to build the thing

	// TODO: add all the right component in here based on the formatting of the event type.
	export let beginat: Date;
	export let day: number;
	export let month: number;
	export let year: number;
	export let freeplace: number;
	export let sessionduration: number;
	export let id: string;
	export let location: string;
	export let placecount: number;

	function formatDate(date: Date) {
		return new Intl.DateTimeFormat('fr', { dateStyle: 'long' }).format(date);
	}
	// $: console.log(formatDate(Date.parse(beginat)));
	const isPlural = placecount > 1;
</script>

<a href={`/app/events/${id}`}>
	<div class="event">
		<div class="flex items-center justify-between">
			<div>
				<div class="flex items-center gap-2">
					<CalendarDays size="18" color="#fc7373" />
					<p class="text-sm">
						{formatDate(Date.parse(beginat))} -
						<span class="placecount">
							[{freeplace} place{isPlural ? 's' : ''} restante{isPlural ? 's' : ''}]</span
						>
					</p>
				</div>
				<h3 class="title mt-1">
					Event for {day}/{month}/{year} for a duration of {sessionduration}
				</h3>
				<p class="mt-1 text-sm">{location}</p>
			</div>
			<button type="button">
				<ChevronRight />
			</button>
		</div>
	</div>
</a>

<style>
	.placecount {
		color: #fc7373;
	}
	.event {
		background-color: #202127;
		margin-block: 1rem;
		border-radius: 0.5rem;
		/* NOTE: old border */
		border: 1px solid #bdbdbd;
		border: 1px solid rgba(60, 60, 67, 0.78);
		background-color: #202127;
		padding: 1rem;
	}

	.event:is(:hover, :focus) {
		cursor: pointer;
	}

	.title {
		font-size: 1.2rem;
		color: white;
	}
</style>
