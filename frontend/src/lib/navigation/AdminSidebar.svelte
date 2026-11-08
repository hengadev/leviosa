<script lang="ts">
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
	import type { Permissions } from "#lib/security/permissions.ts";
	import BackOfficeSidebar, { type AccountLink, type NavGroup } from "./BackOfficeSidebar.svelte";
	import BackOfficeMobileNav from "./BackOfficeMobileNav.svelte";

	interface SidebarProps {
		user: App.User;
		permissions: Permissions;
		unreadCount?: number;
		/** The admin is also the practitioner: links to their practitioner area. */
		selfPractitioner?: boolean;
	}

	let { user, permissions, unreadCount = 0, selfPractitioner = false }: SidebarProps = $props();

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

	// The phone bar: the daily pages; the rest is in its Menu drawer.
	const mobileBar = [
		{ href: "/admin", label: "Accueil" },
		{ href: "/admin/bookings/consultations", label: "Réservations" },
		{ href: "/admin/users", label: "Utilisateurs" },
		{ href: "/admin/catalog", label: "Catalogue" },
	];
</script>

{#if permissions.canAccessOps}
	<BackOfficeSidebar area="Administration" home="/admin" groups={navigation} {accountLinks} {user} />

	<BackOfficeMobileNav home="/admin" groups={navigation} bar={mobileBar} {accountLinks} {user} />
{/if}
