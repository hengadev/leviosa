<script lang="ts">
	import { page } from "$app/state";
	import {
		Home,
		Calendar,
		CalendarClock,
		CalendarCheck,
		Package,
		DollarSign,
		UserRound,
		ExternalLink,
		Shield,
	} from "@lucide/svelte";
	import type { Component } from "svelte";
	import type { Permissions } from "$lib/security/permissions";
	import BackOfficeSidebar, { type AccountLink, type NavGroup } from "./BackOfficeSidebar.svelte";

	interface SidebarProps {
		user: App.User;
		permissions: Permissions;
		unreadCount?: number;
	}

	let { user, permissions, unreadCount = 0 }: SidebarProps = $props();

	/**
	 * Navigation item definition
	 * - href: Route path
	 * - label: Display text
	 * - icon: Lucide icon component
	 * - adminOnly: Whether only admin users can see this
	 */
	interface NavItem {
		href: string;
		label: string;
		icon: Component;
		adminOnly?: boolean;
	}

	const isAdmin = $derived(user.role === "administrator");

	// v1 (single-practitioner launch): messages, statistics (analytics) and
	// reviews are hidden from the partner navigation. The routes stay in code.
	const navigation: NavGroup[] = $derived([
		{
			label: "Au quotidien",
			items: [
				{ href: "/staff", label: "Tableau de bord", icon: Home },
				{ href: "/staff/agenda/reservations", label: "Réservations", icon: CalendarCheck },
				{ href: "/staff/agenda/disponibilites", label: "Disponibilités", icon: CalendarClock },
			],
		},
		{
			label: "Mon activité",
			items: [
				{ href: "/staff/statistics/finances", label: "Finances", icon: DollarSign },
				{ href: "/staff/profile", label: "Mon profil", icon: UserRound },
			],
		},
		...(isAdmin
			? [{ label: "Administration", items: [{ href: "/staff/catalog", label: "Catalogue", icon: Package }] }]
			: []),
	]);

	// An administrator who is also the practitioner goes back to the admin
	// panel from the account menu, as the admin panel links here.
	const accountLinks: AccountLink[] = $derived([
		...(isAdmin ? [{ href: "/admin", label: "Administration", icon: Shield }] : []),
		{ href: "https://leviosa.care", label: "Voir le site", icon: ExternalLink, external: true },
	]);

	/**
	 * Mobile Navigation Structure
	 * Consolidated to 5 items max for bottom bar
	 */
	const mobileNavigation: NavItem[] = [
		{
			href: "/staff",
			label: "Accueil",
			icon: Home,
		},
		{
			href: "/staff/agenda/reservations",
			label: "Réservations",
			icon: CalendarCheck,
		},
		{
			href: "/staff/agenda/disponibilites",
			label: "Agenda",
			icon: Calendar,
		},
		{
			href: "/staff/profile",
			label: "Profil",
			icon: UserRound,
		},
		{
			href: "/admin",
			label: "Admin",
			icon: Shield,
			adminOnly: true,
		},
	];

	/**
	 * Filter navigation items based on user role
	 */
	const mobileItems = $derived(
		mobileNavigation.filter((item) => !item.adminOnly || user.role === "administrator"),
	);

	/**
	 * Check if a route is active
	 */
	function isActive(href: string): boolean {
		const currentPath = page.url.pathname;
		if (href === "/staff") {
			return currentPath === "/staff";
		}
		return currentPath.startsWith(href);
	}
</script>

{#if permissions.canAccessOps}
	<BackOfficeSidebar area="Espace praticien" home="/staff" groups={navigation} {accountLinks} {user} />

	<!-- Mobile Bottom Navigation -->
	<nav
		class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-sm border-t border-border-card"
		aria-label="Bottom navigation"
	>
		<ul class="flex items-center justify-around px-1 py-1">
			{#each mobileItems as item (item.href)}
				{@const active = isActive(item.href)}
				<li class="flex-1">
					<a
						href={item.href}
						class="flex flex-col items-center gap-1.5 py-3 px-2 rounded-lg transition-all duration-200
		                   {active ? 'text-foreground' : 'text-muted-foreground'}"
						aria-current={active ? "page" : undefined}
					>
						<item.icon size={20} strokeWidth={active ? 2 : 1.5} />
						<span class="text-xs font-medium">
							{item.label}
						</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>
{/if}
