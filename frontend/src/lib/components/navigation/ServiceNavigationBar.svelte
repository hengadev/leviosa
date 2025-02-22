<svelte:options />

<script lang="ts">
	import ServiceNavigationBarIcon from './atoms/ServiceNavigationBarIcon.svelte';
	import type { ServiceState } from '$lib/types';
	import { servicestate } from '$lib/stores/servicebar';

	type BarType = 'light' | 'dark';
	interface Props {
		type?: BarType;
	}

	let { type = 'dark' }: Props = $props();

	// TODO: there is probably a better way to do it brother
	let names: ServiceState[] = ['A propos', 'Deroule', 'Prestataires'];

	export { type };
</script>

<nav class="navigation-bar flex">
	{#each names as name}
		<ServiceNavigationBarIcon {name} active={$servicestate === name} {type} />
	{/each}
</nav>

<style>
	.navigation-bar {
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid hsl(var(--clr-stroke));
		text-wrap: nowrap;
	}
</style>
