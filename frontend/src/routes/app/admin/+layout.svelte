<script lang="ts">
	import { page } from '$app/stores';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();
	const role = $page.data.user.role;
</script>

{#if role === 'admin'}
	{@render children?.()}
{/if}
