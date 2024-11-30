<script lang="ts">
	import type { PageData, ActionData } from './$types';
	export let data: PageData;
	$: ({ formControls, oauthButtons } = data);
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { route } from '$lib/ROUTES';

	import FormInput from '$lib/components/FormInput.svelte';

    // TODO: Here I make all the modification on this new page brother

	export let form: ActionData;
	if (form?.success) {
		console.log('the success status  : ', form.success);
	}
	const action = route('GET /auth/oauth/google');
</script>

<div class="container grid" style="--gap: 0rem;">
	<div class="content flow stroke" style="--flow-space: 4rem;">
		<!-- <div class="logo"></div> -->
        <div class="header">
            <h1 class="fs-h1 dark-primary-content">👋 Content de te revoir</h1>
            <p class="subtitle fs-paragraph dark-ternary-content">Rejoins notre communaute bien etre physique et mental et blablabla</p>
        </div>
		<form class="oauths" method="GET" use:enhance {action}>
			{#each oauthButtons as btn}
				{@const specificClass = `${btn.providerName}-icon`}
				<button type="submit" class="oauth">
					<div class="oauth-icon {specificClass}">
						<svelte:component this={btn.IconComponent} />
					</div>
                    <p class="fs-paragraph oauth-cta">Se connecter avec <span class="capitalize">{btn.providerName}</span></p>
				</button>
			{/each}
		</form>
		<div class="separator-block">
			<div class="separator"></div>
			<p class="or fs-paragraph stroke">ou</p>
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
			<div class="grid flow" style="--flow-space: 4rem;">
				{#each formControls as formcontrol}
                    <div class="grid flow" style="--flow-space: 0.5rem;">
						<div class="field">
							{#if formcontrol.name === 'password'}
								<button
									class="forgotten__password"
									on:click={() => console.log('go to the forgotten password thing')}
								>
									<a class="link" href="/recoverpassword"> Mot de passe oublie ? </a>
								</button>
							{/if}
						</div>
                        <FormInput
                            id={formcontrol.name}
                            name={formcontrol.name}
                            type={formcontrol.type}
                            placeholder={formcontrol.placeholder}
                            value={formcontrol.value}
                            forInput={formcontrol.name}
                            label={formcontrol.label}
                        />
					</div>
				{/each}
				<button type="submit" class="signin">Se Connecter</button>
				<div class="else fs-paragraph">
					<p class="paragraph">Tu decouvres Leviosa ?</p>
					<a href="/auth/signup" class="link">Créés un compte !</a>
				</div>
			</div>
		</form>
	</div>
    <div class="image stroke"></div> </div>

<style>
	.logo {
		width: 80px;
		background-color: hsl(var(--clr-dark-secondary));
		border-radius: 100%;
		margin-bottom: 0.75rem;
		aspect-ratio: 1;
		margin-inline: auto;
	}
	.container {
        outline: 5px solid limegreen;
        place-content: center;
        height: 100vh;
        width: 100vw;
        align-items: center;
        grid-template-columns: repeat(2, 1fr);
        grid-template-rows: repeat(1, 100%);
        /* padding-block: 4rem; */
	}

	.content {
        border: 2px solid orangered;
        grid-column: 1;
		padding: clamp(1rem, 3vw, 3rem);
		border-radius: 0.5rem;
        height: 100%;
        border-top-right-radius: 0;
        border-bottom-right-radius: 0;
        background-color: hsl(var(--clr-light-primary));
	}

    .image {
        grid-column: 2;
		border-radius: 0.5rem;
        height: 100%;
        background-color: #bdbdbd;
        border-top-left-radius: 0;
        border-bottom-left-radius: 0;
    }

    /* TODO: fix that media query so that I have a fix width up until a certain size  */
    /* and do that media query to be mobile first */
    /* @media (max-width: 1200px) { */
    /*     .container { */
    /*         grid-template-columns: 1fr; */
    /*     } */
    /*     .content { */
    /*         grid-column: 1 / 2; */
    /*     } */
    /*     .image { */
    /*         display: none; */
    /*     }  */
    /* } */


    .subtitle {
        width: 70%;
    }

	.oauths {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
	}

	.oauth {
        padding: 1rem;
        padding: calc(var(--p) - 0.5rem);
        padding: var(--p);
		display: flex;
		align-items: center;
        border-radius: 0.5rem;
        margin-inline: auto;
		gap: 1rem;
		width: 100%;
		background-color: hsl(var(--clr-dark-primary));
		font-weight: 400;
	}
	.oauth:is(:focus, :hover) { opacity: 0.9; }

	.oauth-icon {
        width: calc(var(--p) + 0.2rem);
		aspect-ratio: 1;
	}

    .oauth-cta {
        text-wrap: nowrap;
        color: hsl(var(--clr-light-primary));
    }

	.apple-icon { margin-top: -0.2rem; }

	.separator-block {
		--or-width: 3.6ch;
		display: flex;
        align-items: center;
	}

	.separator {
		width: 100%;
		height: 1px;
        background-color: hsl(var(--clr-stroke));
	}

	.or {
		font-size: 0.8rem;
        color: hsl(var(--clr-stroke));
		width: var(--or-width);
		aspect-ratio: 1;
		border-radius: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-grow: 0;
		flex-shrink: 0;
	}

	::placeholder {
		font-weight: 300;
	}

	.field {
		display: flex;
		align-items: center;
		text-wrap: nowrap;
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

	.signin {
        /* padding: 1rem; */
        padding: calc(var(--p) - 0.5rem);
        border-radius: 0.5rem;
		font-weight: 500;
        background-color: hsl(var(--clr-dark-primary));
        color: hsl(var(--clr-light-primary));
	}

	.signin:is(:hover, :focus) {
		opacity: 0.9;
	}

	.else {
		/* margin-top: 1rem; */
		margin-inline: auto;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.link {
		text-decoration: none;
        color: hsl(var(--clr-accent));
	}
	.link:is(:hover, :focus) {
		text-decoration: underline;
	}
</style>
