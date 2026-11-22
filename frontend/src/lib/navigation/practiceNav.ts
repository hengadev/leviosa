// The practice navigation: the one navigation an administrator sees on every
// back-office page, whether it belongs to the admin area or the Partner area.
// The phone bar, the desktop sidebar and the Cabinet page all render from it.
import type { Component } from "svelte";
import {
	Building2,
	CalendarCheck,
	CalendarClock,
	ExternalLink,
	HandCoins,
	Mail,
	Package,
	Settings,
	Sun,
	UserRound,
	Users,
} from "@lucide/svelte";
import type { AccountLink, NavGroup } from "./BackOfficeSidebar.svelte";

export type TabId = "today" | "bookings" | "availabilities" | "compta" | "cabinet";

export interface PracticeTab {
	id: TabId;
	href: string;
	/** The phone bar's short label. */
	label: string;
	/** The sidebar's full name. */
	name: string;
	icon: Component;
}

/** A page listed on the Cabinet page, with its one-line description. */
export interface CabinetItem {
	href: string;
	label: string;
	hint: string;
	icon: Component;
}

export interface CabinetGroup {
	label: string;
	items: CabinetItem[];
}

/** The admin home: the sidebar marks it active only on itself. */
export const home = "/admin";

export const cabinetHref = "/admin/cabinet";

/** The phone bar's five tabs, in order. */
export const tabs: PracticeTab[] = [
	{ id: "today", href: home, label: "Aujourd'hui", name: "Aujourd'hui", icon: Sun },
	{
		id: "bookings",
		href: "/staff/agenda/reservations",
		label: "Réservations",
		name: "Réservations",
		icon: CalendarCheck,
	},
	{
		id: "availabilities",
		href: "/staff/agenda/disponibilites",
		label: "Disponibilités",
		name: "Disponibilités",
		icon: CalendarClock,
	},
	{ id: "compta", href: "/admin/compta", label: "Compta", name: "Comptabilité", icon: HandCoins },
	{ id: "cabinet", href: cabinetHref, label: "Cabinet", name: "Cabinet", icon: Building2 },
];

/** The rarely used management pages, gathered on the Cabinet page. */
export const cabinetGroups: CabinetGroup[] = [
	{
		label: "Gestion",
		items: [
			{ href: "/admin/catalog", label: "Catalogue", hint: "Vos séances et leurs tarifs", icon: Package },
			{ href: "/admin/users", label: "Utilisateurs", hint: "Comptes et rôles", icon: Users },
			{
				href: "/staff/profile",
				label: "Mon profil praticien",
				hint: "Votre page publique : bio, photo",
				icon: UserRound,
			},
		],
	},
	{
		label: "Administration",
		items: [
			{ href: "/admin/emails", label: "Emails", hint: "Aperçu des emails envoyés", icon: Mail },
			{ href: "/admin/settings", label: "Paramètres", hint: "Cabinet, page publique, accès", icon: Settings },
		],
	},
];

/** The desktop sidebar: the daily pages, then the Cabinet pages under their headings. */
export const sidebarGroups: NavGroup[] = [
	{
		label: "Au quotidien",
		items: tabs.filter((tab) => tab.id !== "cabinet").map(({ href, name, icon }) => ({ href, label: name, icon })),
	},
	...cabinetGroups.map((group) => ({
		label: group.label,
		items: group.items.map(({ href, label, icon }) => ({ href, label, icon })),
	})),
];

/** The account links (the logout form comes after them). */
export const accountLinks: AccountLink[] = [
	{ href: "https://livio.leviosa.care", label: "Voir le site", icon: ExternalLink, external: true },
];

function within(path: string, href: string): boolean {
	return path === href || path.startsWith(`${href}/`);
}

/** The tab active for a path: Cabinet also on every page it lists. */
export function activeTab(path: string): TabId | null {
	if (path === home || path === `${home}/`) return "today";
	const tab = tabs.find((t) => t.href !== home && within(path, t.href));
	if (tab) return tab.id;
	if (cabinetGroups.some((group) => group.items.some((item) => within(path, item.href)))) return "cabinet";
	return null;
}

/**
 * Whether a path is one of the pages the Cabinet page lists, where the phone
 * shows a way back to Cabinet. Their sub-pages (a product form) have their
 * own way back to the page above them.
 */
export function isCabinetPage(path: string): boolean {
	const bare = path.length > 1 ? path.replace(/\/$/, "") : path;
	return cabinetGroups.some((group) => group.items.some((item) => item.href === bare));
}
