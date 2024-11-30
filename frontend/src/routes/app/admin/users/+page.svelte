<script lang="ts">
	import type { PageData } from './$types';
	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
	let { users } = $derived(data);
</script>

Here fetch all the users and put some action for said users
{#each users as user}
	<div>{user.firstname}</div>
{/each}
