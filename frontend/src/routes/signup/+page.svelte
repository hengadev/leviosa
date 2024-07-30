<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { ActionData, PageData } from './$types';
	export let data: PageData;
	$: ({ general, address } = data);
	export let form: ActionData;
	$: console.log('change in the form : ', form);
	// TODO: add link to get back to sign in page
	// TODO: do the email is taken error.
	// TODO: Add the possibility to connect using Outh google + apple
	// TODO: add the possibility to do OAuth
</script>

<div class="container">
	<form
		class="content"
		method="POST"
		use:enhance={() => {
			return async ({ result }) => {
				invalidateAll();
				await applyAction(result);
			};
		}}
	>
		<h2 class="title">Cree ton compte</h2>
		<p class="light_text">Renseigne tous les champs suivants pour t'inscrire</p>
		<div class="separator"></div>
		<div class="fields">
			<h3 class="subtitle oneline">General informations</h3>
			{#each general as formcontrol}
				<label for={formcontrol.name} class:oneline={formcontrol.isOneLine}>
					{formcontrol.label}
					<input class="input" type={formcontrol.type} placeholder={formcontrol.placeholder} />
				</label>
			{/each}
			<div class="separator oneline"></div>
			<h3 class="subtitle oneline">Address informations</h3>
			{#each address as formcontrol}
				<label for={formcontrol.name} class:oneline={formcontrol.isOneLine}>
					{formcontrol.label}
					<input class="input" type={formcontrol.type} placeholder={formcontrol.placeholder} />
				</label>
			{/each}
			<button type="submit">Sign up</button>
		</div>
	</form>
</div>

<style>
	.container {
		height: 100vh;
		display: grid;
		place-content: center;
		padding: 1rem;
	}

	.content {
		padding: 2rem;
		border-radius: 0.5rem;
		outline: 1px solid #bdbdbd;
		box-shadow:
			rgba(0, 0, 0, 0.02) 0px 1px 3px 0px,
			rgba(27, 31, 35, 0.15) 0px 0px 0px 1px;
		overflow-y: scroll;
	}

	.title {
		font-size: 2rem;
		color: #f67373;
		font-weight: 700;
	}

	.subtitle {
		font-size: 1.2rem;
		color: rgba(60, 60, 67, 0.78);
		font-weight: 600;
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 1rem;
		margin-top: 2rem;
	}

	.input {
		border: 1px solid #bdbdbd;
		padding-block: 0.2rem;
		padding-inline: 1rem;
		border-radius: 0.5rem;
	}

	.light_text {
		opacity: 0.8;
		color: #27272d;
		/* color: #3c3c43; */
	}

	label {
		display: grid;
		font-weight: 400;
		font-size: 1.1rem;
		color: #3c3c43;
	}

	:placeholder {
		font-weight: 300;
	}

	.oneline {
		grid-column: 1 / -1;
	}

	button {
		margin-top: 2rem;
		color: white;
		background-color: #171717;
		font-weight: 500;
		font-size: 1.2rem;
		grid-column: 1 / -1;
	}

	.separator {
		margin-block: 1rem;
		opacity: 0.3;
		width: 100%;
		height: 2px;
		background-color: #bdbdbd;
	}
</style>
