<script lang="ts">
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import CircleUserRound from 'lucide-svelte/icons/circle-user-round';
	import Bell from 'lucide-svelte/icons/bell';
	import Field from '$lib/components/settings/Field.svelte';
	const profileIconSize = '200';
	const iconSize = '24';

	let { data } = $props();
	let { name, fields } = $derived(data);
</script>

<div class="container py-4 pb-24">
	<div class="mx-auto text-center">
		<CircleUserRound size={profileIconSize} class="mx-auto" />
		<h2 class="title">{name}</h2>
	</div>
	<h2 class="subtitle mt-12">Informations</h2>
	<form method="POST">
		{#each Object.entries(fields) as [fieldName, field]}
			<Field name={fieldName} {...field} />
		{/each}
	</form>
	<h2 class="subtitle mt-4">Reglages</h2>
	<button class="content">
		<div class="left">
			<Bell size={iconSize} />
			<p class="value">Notifications</p>
		</div>
		<ChevronRight />
	</button>
</div>

<style>
	button {
		padding: 0;
	}

	.container {
		display: grid;
		place-content: center;
		padding-inline: 5rem;
		width: 100%;
	}

	.title {
		font-size: 1.2rem;
		font-weight: 800;
		color: rgba(255, 255, 245, 0.86);
	}

	.subtitle {
		font-size: 1.5rem;
		font-weight: 600;
		color: rgba(255, 255, 245, 0.86);
		margin-bottom: 2rem;
	}

	.content {
		width: 90vw;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.value {
		font-size: 1.1rem;
		color: rgba(255, 255, 245, 0.86);
	}
</style>
