<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import type { PageData } from './$types';
	export let data: PageData;
	$: ({ pastevents, futurevents } = data);
	function formatDate(date: Date) {
		return new Intl.DateTimeFormat('fr', { dateStyle: 'long' }).format(date);
	}
</script>

<form method="POST" action="">
	<h3 class="subtitle">Futur Events</h3>
	<Table.Root class="mt-4 rounded-xl border-[1px] border-[#bdbdbd] text-center">
		<Table.Caption>La liste des evements a venir.</Table.Caption>
		<Table.Header>
			<Table.Row class="text-center">
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)] text-center">Date</Table.Head>
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)] text-center">Localisation</Table.Head>
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)] text-center"
					>Places restantes</Table.Head
				>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#each futurevents as event}
				<Table.Row on:click={() => console.log('the event id : ', event.Id)}>
					<input type="hidden" value={event.id} />
					<Table.Cell class="text-center text-white">{formatDate(event.Date)}</Table.Cell>
					<Table.Cell class="text-center text-white">{event.Location}</Table.Cell>
					<Table.Cell class="text-center text-white">{event.PlaceCount}</Table.Cell>
				</Table.Row>
			{/each}
		</Table.Body>
	</Table.Root>
</form>

<div class="mt-8">
	<h3 class="subtitle">Past Events</h3>
	<Table.Root class="mt-4 table rounded-xl border-[1px] border-[#bdbdbd] text-center">
		<Table.Caption>La liste des evements passes.</Table.Caption>
		<Table.Header>
			<Table.Row class="text-center">
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)] text-center">Date</Table.Head>
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)]  head text-center"
					>Localisation</Table.Head
				>
				<Table.Head class="text-[rgba(255, 255, 245, 0.86)]  head text-center"
					>Places restantes</Table.Head
				>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#each pastevents as event}
				<Table.Row on:click={() => console.log('the event id : ', event.Id)}>
					<input type="hidden" value={event.Id} />
					<Table.Cell class="text-center text-white">{formatDate(event.Date)}</Table.Cell>
					<Table.Cell class="text-center text-white">{event.Location}</Table.Cell>
					<Table.Cell class="text-center text-white">{event.PlaceCount}</Table.Cell>
				</Table.Row>
			{/each}
		</Table.Body>
	</Table.Root>
</div>
