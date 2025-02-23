<script lang="ts">
	import { navstate } from '$lib/stores/navbar';
	navstate.set('accueil'); // just to forget the value stored in localstore when reconecting and I had the page to another link.

	import NoEventCard from '$lib/components/ui/NoEventCard.svelte';
	import EventCard from '$lib/components/ui/EventCard.svelte';
	import InstallHeader from './InstallHeader.svelte';
	import Header from './Header.svelte';
	import Section from './Section.svelte';
	import ServiceCarousel from './ServiceCarousel.svelte';
	import QRCode from './QRCode.svelte';

	import type { PageData } from './$types';
	interface Props {
		data: PageData;
	}
	let { data }: Props = $props();
	const { name, qrcode } = data;

	import Drawer from '$lib/components/Drawer.svelte';

	let isDrawerOpen = $state(false);
	function toggleDrawer() {
		console.log('toggling the drawer');
		return () => (isDrawerOpen = !isDrawerOpen);
	}
</script>

<InstallHeader />
<div class="content flow relative" style="--flow-space: 3rem;">
	<Header {name} {toggleDrawer} />
	<Section title="decouvrez nos services" cta="Voir tout">
		<ServiceCarousel />
	</Section>
	<Section title="votre prochain evenement">
		<EventCard />
	</Section>
	<Section title="votre prochain evenement">
		<NoEventCard />
	</Section>
	<div class="next-event container grid" style="--gap: 1rem;">
		<h3 class="fs-h3 subtitle">Tu peux aussi revoir...</h3>
		<p>Jean Dupont, ton dernier prestataire massage</p>
	</div>
</div>
<Drawer bind:isOpen={isDrawerOpen} closeDrawer={() => (isDrawerOpen = false)}>
	<QRCode {qrcode} />
</Drawer>

<style>
	.content {
		padding-bottom: 7rem;
	}
	.subtitle {
		color: black;
		font-size: var(--fs-1);
		font-weight: 600;
	}
</style>
