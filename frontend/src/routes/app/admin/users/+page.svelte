<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;
	$: ({ users } = data);
</script>

Here fetch all the users and put some action for said users
{#each users as user}
	<div>{user.firstname}</div>
{/each}
