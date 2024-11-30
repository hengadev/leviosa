<script lang="ts">
	// get all the bookings with some nice graphics.
	// I can use the tab thing to switch between just month and month+day
	// NOTE:
	// import type { PageData } from './$types';
	// export let data: PageData;

	import CalendarIcon from 'lucide-svelte/icons/calendar';
	import { DateFormatter, type DateValue, getLocalTimeZone } from '@internationalized/date';
	import { cn } from '$lib/utils.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Calendar } from '$lib/components/ui/calendar/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';

	const df = new DateFormatter('en-US', {
		dateStyle: 'long'
	});

	let value: DateValue | undefined = $state(undefined);
	import { page } from '$app/stores';
	const year: string = $page.params.year.toUpperCase();
	const month: string = $page.params.month.toUpperCase();
	// TODO: Do all the part  where I can actually see the stat for any event
</script>

<div class="container">
	<div class="content">
		<h2 class="title">Choisis une date</h2>
		<h2 class="subtitle">Pour le mois de {month}{year}.</h2>
		<form class="grid" method="POST">
			<Popover.Root>
				<Popover.Trigger asChild >
					{#snippet children({ builder })}
										<Button
							variant="outline"
							class={cn(
								'w-[100%] justify-start text-right font-normal',
								!value && 'text-muted-foreground'
							)}
							builders={[builder]}
						>
							<CalendarIcon class="mr-2 h-4 w-4" />
							{value ? df.format(value.toDate(getLocalTimeZone())) : 'Choisis une date'}
						</Button>
														{/snippet}
								</Popover.Trigger>
				<Popover.Content class="w-auto p-0">
					<Calendar bind:value initialFocus />
				</Popover.Content>
			</Popover.Root>
			<button class="submit">Submit</button>
		</form>
	</div>
</div>

<style>
	.container {
		height: 100vh;
		padding-inline: 2rem;
		display: grid;
		place-content: center;
	}

	.content {
		padding: 2rem;
		border-radius: 0.5rem;
		outline: 1px solid #bdbdbd;
		box-shadow:
			rgba(0, 0, 0, 0.02) 0px 1px 3px 0px,
			rgba(27, 31, 35, 0.15) 0px 0px 0px 1px;
		background-color: #202127;
	}

	.grid {
		color: black;
	}

	.title,
	.subtitle {
		font-size: 2rem;
		font-weight: 800;
	}

	.title {
		color: #f67373;
	}

	.subtitle {
		color: #3c3c43;
		color: rgba(255, 255, 245, 0.86);
		margin-bottom: 2rem;
	}

	button {
		margin-top: 2rem;
		color: white;
		background-color: #171717;
		font-weight: 500;
		font-size: 1.2rem;
		grid-column: 1 / -1;
		background-color: #f67373;
	}
</style>
