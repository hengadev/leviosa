<script lang="ts">
	import {
		Home,
		CalendarClock,
		CalendarCheck,
		Package,
		DollarSign,
		UserRound,
		ExternalLink,
		Shield,
	} from "@lucide/svelte";
	import type { Permissions } from "#lib/security/permissions.ts";
	import BackOfficeSidebar, { type AccountLink, type NavGroup } from "./BackOfficeSidebar.svelte";
	import BackOfficeMobileNav from "./BackOfficeMobileNav.svelte";

	interface SidebarProps {
		user: App.User;
		permissions: Permissions;
		unreadCount?: number;
	}

	let { user, permissions, unreadCount = 0 }: SidebarProps = $props();

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

	// The phone bar: the daily pages; the rest is in its Menu drawer.
	const mobileBar = [
		{ href: "/staff", label: "Accueil" },
		{ href: "/staff/agenda/reservations", label: "Réservations" },
		{ href: "/staff/agenda/disponibilites", label: "Agenda" },
		{ href: "/staff/profile", label: "Profil" },
	];
</script>

{#if permissions.canAccessOps}
	<BackOfficeSidebar area="Espace praticien" home="/staff" groups={navigation} {accountLinks} {user} />

	<BackOfficeMobileNav home="/staff" groups={navigation} bar={mobileBar} {accountLinks} {user} />
{/if}
