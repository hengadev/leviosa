<script lang="ts">
	import type { PageData, ActionData } from './$types';
	export let data: PageData;
	$: ({ formControls } = data);
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	export let form: ActionData;
	if (form?.success) {
		console.log('the success status  : ', form.success);
	}
	// TODO: Do some conditional styling based on the result of the form value
	// import { page } from '$app/stores';
	// console.log($page.user);
</script>

<div class="container">
	<form
		class="content"
		method="POST"
		action="?/register"
		use:enhance={() => {
			return async ({ result }) => {
				invalidateAll();
				await applyAction(result);
			};
		}}
	>
		<h1 class="title">Email et mot de passe</h1>
		<h2 class="subtitle">Pour s'identifier !</h2>
		<div class="separator"></div>
		<div class="fields">
			{#each formControls as formcontrol}
				<div>
					<label for={formcontrol.name}>
						{formcontrol.label}
						<input
							id={formcontrol.name}
							name={formcontrol.name}
							class="input"
							type={formcontrol.type}
							placeholder={formcontrol.placeholder}
							value={formcontrol.value}
							required
						/>
					</label>
					{#if formcontrol.name === 'password'}
						<button
							class="forgotten__password"
							on:click={() => console.log('go to the forgotten password thing')}
						>
							<a class="link" href="/recoverpassword"> Mot de passe oublie ? </a>
						</button>
					{/if}
				</div>
			{/each}
			<button type="submit" class="signin">Se Connecter</button>
			<div class="separator"></div>
			<div class="else">
				<p class="paragraph">Tu ne possèdes pas de compte ?</p>
				<a href="/signup" class="link">Créé un compte !</a>
			</div>
		</div>
	</form>
</div>

<style>
	.container {
		display: grid;
		place-content: center;
		padding: 2rem;
		height: 100vh;
	}

	.content {
		background-color: #202127;
		padding: 4rem;
		padding-block: 3rem;
		border-radius: 0.5rem;
		outline: 1px solid #bdbdbd;
		box-shadow:
			rgba(0, 0, 0, 0.02) 0px 1px 3px 0px,
			rgba(27, 31, 35, 0.15) 0px 0px 0px 1px;
	}

	.title,
	.subtitle {
		/* font-size: 2rem; */
		font-size: clamp(1.15rem, 4vw, 2rem);
		font-weight: 800;
	}

	.title {
		color: #f67373;
	}
	.subtitle {
		color: #3c3c43;
		color: rgba(255, 255, 245, 0.86);
		margin-bottom: 2rem;
	}

	label {
		display: flex;
		flex-direction: column;
		font-weight: 500;
		font-size: 1.1rem;
		font-size: clamp(0.8rem, 3vw, 1.1rem);
		color: rgba(255, 255, 245, 0.86);
	}

	::placeholder {
		font-weight: 300;
	}

	.fields {
		display: grid;
		gap: clamp(1rem, 4vw, 2rem);
	}

	.input {
		width: 100%;
		/* border: 1px solid #bdbdbd; */
		padding-block: 0.5rem;
		padding-inline: 1rem;
		border-radius: 0.5rem;
		background-color: #1b1b1f;
		margin-top: 0.5rem;
	}

	input:is(:focus) {
		outline: 1px solid #bdbdbd;
	}

	.forgotten__password {
		margin: 0;
		padding: 0;
		border: none;
		background: transparent;
		font-size: clamp(0.75rem, 3vw, 0.85rem);
		margin-top: 0.5rem;
		display: block;
		opacity: 0.5;
		width: 100%;
		text-align: right;
	}

	.forgotten__password:is(:hover, :focus) {
		cursor: pointer;
		text-decoration: underline;
	}

	button {
		color: white;
		background-color: #171717;
		background-color: #f67373;
		font-size: 1.1rem;
	}

	.signin:is(:hover, :focus) {
		opacity: 0.9;
	}

	.separator {
		margin-block: 1rem;
		opacity: 0.3;
		width: 100%;
		height: 2px;
		background-color: #bdbdbd;
	}

	.else {
		text-align: center;
	}

	.paragraph {
		/* color: #171717; */
		/* opacity: 0.6; */
		color: rgba(255, 255, 245, 0.86);
	}

	.link {
		opacity: 1;
		text-decoration: none;
	}
	.link:is(:hover, :focus) {
		text-decoration: underline;
	}
</style>
