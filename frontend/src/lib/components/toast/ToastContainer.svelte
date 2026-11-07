<script lang="ts">
	import { onMount } from 'svelte';
	import Toast from './Toast.svelte';
	import { toastStore } from './store.svelte';

	let mounted = $state(false);

	onMount(() => {
		mounted = true;
	});
</script>

{#if mounted && toastStore.toasts.length > 0}
	<div class="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
		{#each toastStore.toasts as toastItem (toastItem.id)}
			<Toast toast={toastItem} onClose={toastStore.remove} />
		{/each}
	</div>
{/if}
