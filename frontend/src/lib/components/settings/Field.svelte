<script lang="ts">
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import type { Icon } from 'lucide-svelte';
	import type { ComponentType } from 'svelte';

	export let name: string;
	export let icon: ComponentType<Icon>;
	export let value: string;
	export let missingText: string;
	const iconSize = '24';
</script>

<div>
	<input type="hidden" name="fieldName" value="fieldName" />
	<button type="submit" class="content mb-8">
		<div class="left">
			<svelte:component this={icon} size={iconSize} />
			<div class="text flex flex-col gap-1 text-left">
				<p class="field">{name}</p>
				<p class="value">{value || missingText}</p>
				{#if value.startsWith('Aucun')}
					<p class="text-amber-300">{missingText}</p>
				{/if}
			</div>
		</div>
		<ChevronRight />
	</button>
</div>

<style>
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

	.text {
		display: flex;
		flex-direction: column;
		gap: 0rem;
		/* align-items: center; */
	}

	.field {
		text-transform: capitalize;
		font-size: 0.9rem;
	}

	.value {
		font-size: 1.1rem;
		color: rgba(255, 255, 245, 0.86);
	}
</style>
