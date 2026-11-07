<script lang="ts">
	import { page } from "$app/state";
	import {
		Home,
		Users,
		CalendarCheck,
		Package,
		HandCoins,
		CalendarClock,
		Settings,
		ExternalLink,
	} from "@lucide/svelte";
	import type { Component } from "svelte";
	import type { Permissions } from "$lib/security/permissions";
	import BackOfficeSidebar, { type AccountLink, type NavGroup } from "./BackOfficeSidebar.svelte";

	interface SidebarProps {
		user: App.User;
		permissions: Permissions;
		unreadCount?: number;
		/** The admin is also the practitioner: links to their practitioner area. */
		selfPractitioner?: boolean;
	}

	let { user, permissions, unreadCount = 0, selfPractitioner = false }: SidebarProps = $props();

	interface NavItem {
		href: string;
		label: string;
		icon: Component;
	}

	// v1 (single-practitioner launch): planning, messages, buildings and rooms,
	// analytics, reviews and waitlist are hidden from the admin navigation.
	// The routes stay in code.
	const navigation: NavGroup[] = [
		{
			label: "Au quotidien",
			items: [
				{ href: "/admin", label: "Tableau de bord", icon: Home },
				{ href: "/admin/bookings/consultations", label: "Réservations", icon: CalendarCheck },
				{ href: "/admin/users", label: "Utilisateurs", icon: Users },
			],
		},
		{
			label: "Activité",
			items: [
				{ href: "/admin/catalog", label: "Catalogue", icon: Package },
				{ href: "/admin/compta", label: "Comptabilité", icon: HandCoins },
			],
		},
		// The practice settings: company details, public page, booking rules.
		// Desktop only: the mobile bar is full and these are set up once.
		{
			label: "Administration",
			items: [{ href: "/admin/settings", label: "Paramètres", icon: Settings }],
		},
	];

	const accountLinks: AccountLink[] = $derived([
		// The admin's agenda and availabilities, as the practitioner
		...(selfPractitioner ? [{ href: "/staff", label: "Espace praticien", icon: CalendarClock }] : []),
		{ href: "https://leviosa.care", label: "Voir le site", icon: ExternalLink, external: true },
	]);

	/**
	 * Mobile Navigation Structure
	 * Limited to 5 items for bottom bar
	 */
	const mobileNavigation: NavItem[] = $derived([
		{
			href: "/admin",
			label: "Accueil",
			icon: Home,
		},
		{
			href: "/admin/users",
			label: "Utilisateurs",
			icon: Users,
		},
		{
			href: "/admin/bookings/consultations",
			label: "Réservations",
			icon: CalendarCheck,
		},
		{
			href: "/admin/catalog",
			label: "Catalogue",
			icon: Package,
		},
		// No user menu on mobile: the admin who is also the practitioner
		// reaches their agenda from the bar.
		...(selfPractitioner
			? [
					{
						href: "/staff",
						// Five labels share a phone's width
						label: "Praticien",
						icon: CalendarClock,
					},
				]
			: []),
	]);

	/**
	 * Check if a route is active
	 */
	function isActive(href: string): boolean {
		const currentPath = page.url.pathname;
		if (href === "/admin") {
			return currentPath === "/admin" || currentPath === "/admin/";
		}
		return currentPath.startsWith(href);
	}
</script>

{#if permissions.canAccessOps}
	<BackOfficeSidebar area="Administration" home="/admin" groups={navigation} {accountLinks} {user} />

	<!-- Mobile Bottom Navigation -->
	<nav
		class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-background border-t border-border-card pb-[env(safe-area-inset-bottom)]"
		aria-label="Bottom navigation"
	>
		<ul class="flex items-center justify-around px-1 py-1">
			{#each mobileNavigation as item (item.href)}
				{@const active = isActive(item.href)}
				<li class="flex-1">
					<a
						href={item.href}
						class="flex flex-col items-center gap-1.5 py-3 px-2 rounded-lg transition-all duration-200 {active
							? 'text-foreground'
							: 'text-muted-foreground'}"
						aria-current={active ? "page" : undefined}
					>
						<item.icon size={20} strokeWidth={active ? 2 : 1.5} />
						<span class="text-xs font-medium">{item.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>
{/if}
