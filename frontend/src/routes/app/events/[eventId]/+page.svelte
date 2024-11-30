<script lang="ts">
	import type { PageData } from './$types';
	console.log(event);
	import { page } from '$app/stores';
	const eventId = $page.params.eventId;
	import ShoppingCart from 'lucide-svelte/icons/shopping-cart';
	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
	// TODO: use the event to make the component
	let { handleCheckout, event } = $derived(data);
</script>

<div class="container mb-12">
	<div class="content">
		<h2 class="title">Reservation du ...</h2>
		<div class="separator"></div>
		<h3 class="subtitle">Explication du deroule de l'evenement</h3>
		<p>
			Lorem ipsum dolor sit amet, consectetur adipisicing elit. Voluptate, amet rerum modi culpa
			quisquam natus quis excepturi mollitia non quo, nulla nostrum eveniet sit eos nisi officia
		</p>
		<h3 class="subtitle mt-8">Localisation</h3>
		<p>Lien vers le google maps pour vous y retrouver</p>
		<h3 class="subtitle mt-8">A prevoir</h3>
		<ul>
			<li>Un certificat medical</li>
			<li>Des vetement de sport style basket + short etc..</li>
			<li>De l'eau</li>
			<li>De la bonne humeur teehee X)</li>
		</ul>
		<button class="mt-8" onclick={() => handleCheckout(eventId)}>
			<ShoppingCart />
			Checkout
		</button>
	</div>
</div>

<style>
	.container {
		height: 100vh;
		display: grid;
		place-content: center;
		padding: 2rem;
		/* TODO: fix the width, it work well with the small screen but not for the big ones. */
		width: calc(450px - 4rem);
		width: max(60%, 450px);
	}

	.content {
		background-color: #202127;
		padding: 2rem;
		padding-block: 3rem;
		border-radius: 0.5rem;
		outline: 1px solid #bdbdbd;
		box-shadow:
			rgba(0, 0, 0, 0.02) 0px 1px 3px 0px,
			rgba(27, 31, 35, 0.15) 0px 0px 0px 1px;
	}

	.title,
	.subtitle {
		font-weight: 800;
	}

	.title {
		font-size: clamp(1.15rem, 4vw, 2rem);
		color: #f67373;
	}

	.subtitle {
		color: rgba(255, 255, 245, 0.86);
		margin-bottom: 1rem;
		font-size: clamp(1.15rem, 3vw, 1.75rem);
	}

	.separator {
		margin-block: 1rem;
		opacity: 0.3;
		width: 100%;
		height: 2px;
		background-color: #bdbdbd;
	}

	li {
		margin-left: 2rem;
		list-style-type: circle;
	}

	button {
		color: white;
		background-color: #f67373;
		font-size: 1.1rem;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
	}

	button:is(:global(:hover, :focus)) {
		opacity: 0.9;
	}
</style>
