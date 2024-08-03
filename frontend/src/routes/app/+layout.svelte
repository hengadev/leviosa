<script lang="ts">
	import { type navState, navstate } from '$lib/stores/navbar';

	function setState(event: MouseEvent) {
		let targetElement = event.currentTarget as HTMLButtonElement;
		let id = targetElement.id as navState;
		navstate.set(id);
	}

	import type { ComponentType } from 'svelte';
	import type { Icon } from 'lucide-svelte';
	import CalendarDays from 'lucide-svelte/icons/calendar-days';
	import Vote from 'lucide-svelte/icons/vote';
	import Home from 'lucide-svelte/icons/house';
	import Mail from 'lucide-svelte/icons/mail';
	import CircleUser from 'lucide-svelte/icons/circle-user';

	type NavItem = {
		name: string;
		id: string;
		href: string;
		icon: ComponentType<Icon>;
	};

	const NavItems: NavItem[] = [
		{
			name: 'Events',
			id: 'events',
			href: '/app/events',
			icon: CalendarDays
		},
		{
			name: 'Votes',
			id: 'votes',
			href: '/app/votes/',
			icon: Vote
		},
		{
			name: 'Accueil',
			id: 'home',
			href: '/app/',
			icon: Home
		},
		{
			name: 'Messagerie',
			id: 'messagerie',
			href: '/app/mails',
			icon: Mail
		},

		{
			name: 'Profile',
			id: 'profile',
			href: '/app/settings/profile',
			icon: CircleUser
		}
	];

	import Dock from 'lucide-svelte/icons/dock';
	import MapPin from 'lucide-svelte/icons/map-pin';
	const opacityStrength = 0.4;
</script>

<div class="header">
	<div class="header__left">
		<Dock color="#2ea6ff" size={36} />
		<p>MonApp</p>
	</div>
	<div class="header__right">
		<MapPin color="white" size={20} />
		<p>Paris</p>
	</div>
</div>
<slot />
<nav>
	<ul>
		{#each NavItems as item}
			<li class="">
				<button id={item.id} on:click={setState}>
					<a href={item.href} style:opacity={$navstate === item.id ? 1 : opacityStrength}>
						<svelte:component this={item.icon} />
						<p>{item.name}</p>
					</a>
				</button>
			</li>
		{/each}
	</ul>
</nav>

<style>
	.header {
		padding: 2rem 1.5rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.header__left,
	.header__right {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	.header__left {
		gap: 1rem;
	}

	a {
		display: flex;
		flex-direction: column;
		align-items: center;
		opacity: 0.4;
	}

	p {
		font-size: 0.8rem;
	}

	nav {
		z-index: 9999;
		position: fixed;
		bottom: 0;
		left: 0;
		right: 0;
		color: white;
		padding-inline: 2rem;
		background-color: #202127;
	}

	ul {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
</style>
