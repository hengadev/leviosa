<script lang="ts">
	import type { PageData, ActionData } from './$types';
	export let data: PageData;
	// $: ({ formControls, oauthButtons, url } = data);
	$: ({ formControls, oauthButtons } = data);
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { route } from '$lib/ROUTES';

	export let form: ActionData;
	if (form?.success) {
		console.log('the success status  : ', form.success);
	}

	const action = route('GET /auth/oauth/google');

	// TODO: use that to call the golang backend or use the href button
	// const handleOAuthClick = (providerName: string) =>
	// 	(window.location.href = `${url}/api/v1/auth/${providerName}`);
</script>

<div class="container">
	<div class="content">
		<div class="logo"></div>
		<h1 class="title">Connectez vous a votre compte</h1>
		<form class="oauths" method="GET" use:enhance {action}>
			{#each oauthButtons as btn}
				{@const specificClass = `${btn.providerName}-icon`}
				<!-- <button on:click={() => handleOAuthClick(btn.providerName)} class="oauth"> -->
				<!-- <div class="oauth-icon {specificClass}"> -->
				<!--     <svelte:component this={btn.IconComponent} /> -->
				<!-- </div> -->
				<button type="submit" class="oauth">
					<div class="oauth-icon {specificClass}">
						<svelte:component this={btn.IconComponent} />
					</div>
				</button>
			{/each}
		</form>
		<div class="separator-block">
			<div class="separator"></div>
			<p class="or">ou</p>
			<div class="separator"></div>
		</div>
		<form
			method="POST"
			action="?/register"
			use:enhance={() => {
				return async ({ result }) => {
					invalidateAll();
					await applyAction(result);
				};
			}}
		>
			<div class="fields">
				{#each formControls as formcontrol}
					<div>
						<div class="field">
							<label for={formcontrol.name}>{formcontrol.label}</label>
							{#if formcontrol.name === 'password'}
								<button
									class="forgotten__password"
									on:click={() => console.log('go to the forgotten password thing')}
								>
									<a class="link" href="/recoverpassword"> Mot de passe oublie ? </a>
								</button>
							{/if}
						</div>
						<input
							id={formcontrol.name}
							name={formcontrol.name}
							class="input"
							type={formcontrol.type}
							placeholder={formcontrol.placeholder}
							value={formcontrol.value}
							required
						/>
					</div>
				{/each}
				<button type="submit" class="signin">Se Connecter</button>
				<div class="else">
					<p class="paragraph">Vous decouvrez Leviosa ?</p>
					<a href="/auth/signup" class="link">Crééz un compte !</a>
				</div>
			</div>
		</form>
	</div>
</div>

<style>
	.logo {
		width: 50px;
		background-color: white;
		border-radius: 100%;
		margin-bottom: 0.75rem;
		aspect-ratio: 1;
		margin-inline: auto;
	}
	.container {
		display: grid;
		place-content: center;
		padding: 2rem;
		height: 100vh;
	}
	.content {
		min-width: 32vw;
		background-color: #202127;
		padding: 3rem;
		padding-top: 2rem;
		border-radius: 0.5rem;
		border-radius: 1rem;
		background-color: #171717;
		box-shadow:
			rgba(60, 60, 67, 0.78) 0px 1px 3px 0px,
			rgba(60, 60, 67, 0.78) 0px 0px 0px 1px;
	}

	.oauths {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
	}

	.oauth {
		display: flex;
		align-items: center;
		gap: 1.2rem;
		width: 100%;
		background-color: #1b1b1f;
		padding-block: 0.5rem;

		font-weight: 300;
		background-color: #26272c;
	}
	.oauth:is(:focus, :hover) {
		/* opacity: 0.9; */
		/* outline: 1px solid #bdbdbd; */
		background-color: rgba(255, 255, 255, 0.16);
	}
	.oauth-label {
		padding-bottom: -0.1rem;
		font-size: clamp(0.75rem, 5vw, 1rem);
	}

	.oauth-icon {
		width: 2rem;
		aspect-ratio: 1;
		background-color: white;
		border-radius: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-inline: auto;
	}
	.apple-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.3rem;
		padding-top: 0.1rem;
	}

	/* <svg viewBox="-56.24 0 608.728 608.728" xmlns="http://www.w3.org/2000/svg"><path d="M273.81 52.973C313.806.257 369.41 0 369.41 0s8.271 49.562-31.463 97.306c-42.426 50.98-90.649 42.638-90.649 42.638s-9.055-40.094 26.512-86.971zM252.385 174.662c20.576 0 58.764-28.284 108.471-28.284 85.562 0 119.222 60.883 119.222 60.883s-65.833 33.659-65.833 115.331c0 92.133 82.01 123.885 82.01 123.885s-57.328 161.357-134.762 161.357c-35.565 0-63.215-23.967-100.688-23.967-38.188 0-76.084 24.861-100.766 24.861C89.33 608.73 0 455.666 0 332.628c0-121.052 75.612-184.554 146.533-184.554 46.105 0 81.883 26.588 105.852 26.588z" fill="#999"/></svg> */
	.google-icon {
		background: none;
		padding: 0.3rem;
	}
	.title {
		margin-block: 2rem;
		text-align: center;
		color: rgba(255, 255, 245, 0.86);
		font-size: 1.1rem;
	}
	.separator-block {
		--or-width: 3.6ch;
		display: flex;
		margin-block: 2rem;
	}
	.separator-block .separator {
		flex-shrink: 1;
		flex-grow: 1;
		background-color: #bdbdbd;
	}

	.or {
		font-weight: 600;
		box-shadow:
			rgba(60, 60, 67, 0.78) 0px 1px 3px 0px,
			rgba(60, 60, 67, 0.78) 0px 0px 0px 1px;
		border-radius: 100%;
		width: var(--or-width);
		aspect-ratio: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.8rem;
		flex-grow: 0;
		flex-shrink: 0;
	}

	label {
		display: flex;
		flex-direction: column;
		font-weight: 500;
		font-size: 1.1rem;
		font-size: clamp(0.8rem, 2vw, 1rem);
		color: rgba(255, 255, 245, 0.86);
	}

	::placeholder {
		font-weight: 300;
	}

	.fields {
		display: grid;
		gap: clamp(0.5rem, 3vw, 1.5rem);
	}

	.field {
		display: flex;
		align-items: center;
		/* justify-content: space-between; */
		text-wrap: nowrap;
	}

	.input {
		width: 100%;
		padding-block: 0.5rem;
		padding-inline: 1rem;
		background-color: #1b1b1f;
		margin-top: 0.2rem;
		border-radius: 0.5rem;

		background-color: #111111;
		color: #bdbdbd;
		font-weight: 200;
	}

	input:is(:focus, :hover) {
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
		width: 100%;
		text-align: right;
	}

	.forgotten__password:is(:hover, :focus) {
		cursor: pointer;
	}

	button {
		color: white;
		font-size: 1.1rem;
	}

	.signin {
		font-size: 0.9rem;
		background-color: #171717;
		background-color: white;
		font-weight: 500;
		color: #171717;
	}

	.signin:is(:hover, :focus) {
		opacity: 0.9;
	}

	.separator {
		margin-block: 0.8rem;
		opacity: 0.1;
		width: 100%;
		height: 1px;
		background-color: #bdbdbd;
	}

	.else {
		margin-top: 1rem;
		margin-inline: auto;

		display: flex;
		align-items: center;
		/* justify-content: space-between; */

		gap: 0.5rem;
	}

	.paragraph {
		/* color: #171717; */
		/* opacity: 0.6; */
		color: rgba(255, 255, 245, 0.86);
	}

	.link {
		text-decoration: none;
		color: #f67373;
	}
	.link:is(:hover, :focus) {
		text-decoration: underline;
	}
</style>
