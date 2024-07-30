<script>
	import { page } from '$app/stores';
	const role = $page.data.user.role;
</script>

{#if role === 'admin'}
	<slot />
{/if}
