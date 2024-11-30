<script lang="ts">
	import type { PageData, ActionData } from './$types';
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { route } from '$lib/ROUTES';
    import Google from '../../../assets/google.svelte';
    import Apple from '../../../assets/apple.svelte';

	import FormInput from '$lib/components/FormInput.svelte';

	interface Props {
		data: PageData;
		form: ActionData;
	}

	let { data, form }: Props = $props();
	if (form?.success) {
		console.log('the success status  : ', form.success);
	}
	const action = route('GET /auth/oauth/google');
    // TODO: change the value of flow-space to make it responsive maybe use some vh units ?
	let { formControls} = $derived(data);
</script>

<div class="container grid" style="--gap: 1rem;">
	<!-- <div class="content flow" style="--flow-space: var(--h1);">  -->
	<div class="content flow" style="--flow-space: 3rem;"> 
        <div class="header grid" style="--gap: 1rem;">
            <h1 class="fs-h1 fw-800 dark-primary-content">👋 Content de te revoir</h1>
            <p class="subtitle fs-h3 dark-ternary-content">Rejoins notre communaute bien etre physique et mental et blablabla</p>
        </div>
		<form class="oauths" method="GET" use:enhance {action}>
            <button type="submit" class="oauth oauth-full">
                <div class="oauth-icon google-icon">
                    <Google />
                </div>
                <p class="fs-paragraph oauth-cta">Se connecter avec Google</p>
            </button>
            <button type="submit" class="oauth">
                <div class="oauth-icon apple-icon">
                    <Apple />
                </div>
            </button>
            <button type="submit" class="oauth">
                <div class="oauth-icon apple-icon">
                    <Apple />
                </div>
            </button>
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
                    <div class="grid flow" style="--flow-space: 0rem;">
						<div class="field">
							{#if formcontrol.name === 'password'}
								<button
									class="forgotten__password"
									onclick={() => console.log('go to the forgotten password thing')}
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
    <div class="image stroke"></div>
</div>

<style>
    .container {
        border: 4px solid green;
        place-content: center;
        height: 100vh;
        min-width: 100vw;
        /* padding-block: var(--p); */
        /* padding-block: 4rem; */
        box-sizing: border-box;
    }

	.content {
        max-width: 700px;
        grid-column: 1 / 3;
        /* TODO: add that part on the big screen when there is the stroke brother */
		/* padding: clamp(1rem, 3vw, 3rem); */
        /* TODO: change that value of height to min-height so that it does not go past the screen */
        /* height: calc(100vh - 4rem); */
        /* height: 100%; */
        border-top-right-radius: 0;
        border-bottom-right-radius: 0;
        background-color: hsl(var(--clr-light-primary));
	}

    .image {
        display: none;
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
        outline: 2px solid orangered;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--p);
	}

	.oauth {
        padding: var(--p);
		display: flex;
		align-items: center;
        justify-content: center;
        border-radius: 0.5rem;
		gap: calc(var(--p) / 2 );
		background-color: hsl(var(--clr-dark-primary));
		font-weight: 400;
	}

	.oauth:is(:global(:focus, :hover)) { opacity: 0.9; }

    .oauth-full { width: 100%; }
	.oauth-icon { width: calc(var(--p) * 1.2); }

    .oauth-cta {
        padding: 0;
        text-wrap: nowrap;
        color: hsl(var(--clr-light-primary));
    }

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

	.forgotten__password:is(:global(:hover, :focus)) {
		cursor: pointer;
	}

	.signin {
        /* padding: 1rem; */
        padding: calc(var(--p) - 0.5rem);
        padding: var(--p);
        border-radius: 0.5rem;
		font-weight: 500;
        background-color: hsl(var(--clr-dark-primary));
        color: hsl(var(--clr-light-primary));
	}

	.signin:is(:global(:hover, :focus)) {
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
	.link:is(:global(:hover, :focus)) {
		text-decoration: underline;
	}
</style>
