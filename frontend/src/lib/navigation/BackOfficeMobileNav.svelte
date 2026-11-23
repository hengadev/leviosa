<script lang="ts">
	// The phone navigation of the back office: a bottom bar
	// with the pages reached most often and a Menu button, whose drawer holds
	// the rest of the navigation, under the sidebar's headings, and the
	// account (other areas, the public site, logout).
	import { page } from "$app/state";
	import { afterNavigate } from "$app/navigation";
	import { LogOut, Menu } from "@lucide/svelte";
	import Drawer from "#lib/ui/Drawer.svelte";
	import {
		accountInitial,
		accountName,
		roleLabel,
		type AccountLink,
		type NavGroup,
	} from "./BackOfficeSidebar.svelte";

	interface Props {
		/** The area's home page: active only on itself. */
		home: string;
		groups: NavGroup[];
		/** The bar's pages, with the shorter label a phone's width allows. */
		bar: { href: string; label: string }[];
		accountLinks: AccountLink[];
		user: App.User;
	}

	let { home, groups, bar, accountLinks, user }: Props = $props();

	let menuOpen = $state(false);
	afterNavigate(() => (menuOpen = false));

	function isActive(href: string): boolean {
		const path = page.url.pathname;
		if (href === home) return path === home || path === `${home}/`;
		return path.startsWith(href);
	}

	const items = $derived(groups.flatMap((group) => group.items));
	const barItems = $derived(
		bar.flatMap(({ href, label }) => {
			const item = items.find((i) => i.href === href);
			return item ? [{ ...item, label }] : [];
		}),
	);
	// The drawer lists what the bar leaves out.
	const menuGroups = $derived(
		groups
			.map((group) => ({ ...group, items: group.items.filter((item) => !bar.some((b) => b.href === item.href)) }))
			.filter((group) => group.items.length > 0),
	);
	// The Menu button looks active on a page the bar does not show.
	const menuActive = $derived(menuGroups.some((group) => group.items.some((item) => isActive(item.href))));
</script>

<nav
	class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-foreground dark:bg-black dark:border-t dark:border-border-card pb-1 [@media(display-mode:standalone)]:pb-6"
	aria-label="Bottom navigation"
>
	<ul class="flex items-stretch">
		{#each barItems as item (item.href)}
			{@const active = isActive(item.href)}
			<li class="flex-1 flex">
				<a
					href={item.href}
					class="flex-1 flex flex-col items-center justify-center gap-1 min-h-14 px-2 transition-colors {active
						? 'text-background dark:text-foreground'
						: 'text-zinc-500 hover:text-zinc-300'}"
					aria-current={active ? "page" : undefined}
				>
					<item.icon size={20} strokeWidth={active ? 2 : 1.5} />
					<span class="text-xs font-medium">{item.label}</span>
				</a>
			</li>
		{/each}
		<li class="flex-1 flex">
			<button
				type="button"
				onclick={() => (menuOpen = true)}
				class="w-full flex flex-col items-center justify-center gap-1 min-h-14 px-2 transition-colors cursor-pointer {menuActive || menuOpen
					? 'text-background dark:text-foreground'
					: 'text-zinc-500 hover:text-zinc-300'}"
				aria-haspopup="dialog"
				aria-expanded={menuOpen}
			>
				<Menu size={20} strokeWidth={menuActive ? 2 : 1.5} />
				<span class="text-xs font-medium">Menu</span>
			</button>
		</li>
	</ul>
</nav>

<div class="lg:hidden">
	<Drawer bind:isOpen={menuOpen}>
		<div class="flex flex-col gap-4 pt-2 pb-[env(safe-area-inset-bottom)]">
			<div class="flex items-center gap-3 pb-4 border-b border-border-card">
				<div class="w-12 h-12 rounded-full flex items-center justify-center bg-muted border border-border-card">
					<span class="text-sm font-semibold text-foreground-alt">{accountInitial(user)}</span>
				</div>
				<div class="flex-1 min-w-0">
					<p class="text-sm font-medium text-foreground truncate">{accountName(user)}</p>
					<p class="text-xs text-muted-foreground">{roleLabel(user)}</p>
				</div>
			</div>

			{#each menuGroups as group (group.label)}
				<div>
					<p class="px-4 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{group.label}</p>
					{#each group.items as item (item.href)}
						{@const active = isActive(item.href)}
						<a
							href={item.href}
							aria-current={active ? "page" : undefined}
							class="flex items-center gap-3 px-4 py-3 rounded-lg transition-colors {active
								? 'bg-foreground text-background'
								: 'text-foreground-alt hover:text-foreground hover:bg-muted'}"
						>
							<item.icon size={20} strokeWidth={1.5} />
							<span class="text-sm font-medium">{item.label}</span>
						</a>
					{/each}
				</div>
			{/each}

			<div class="border-t border-border-card pt-4">
				{#each accountLinks as link (link.href)}
					<a
						href={link.href}
						target={link.external ? "_blank" : undefined}
						rel={link.external ? "noopener noreferrer" : undefined}
						onclick={() => (menuOpen = false)}
						class="flex items-center gap-3 px-4 py-3 text-foreground-alt hover:text-foreground hover:bg-muted rounded-lg transition-colors"
					>
						<link.icon size={20} strokeWidth={1.5} />
						<span class="text-sm font-medium">{link.label}</span>
					</a>
				{/each}
				<form method="POST" action="/logout">
					<button
						type="submit"
						class="w-full flex items-center gap-3 px-4 py-3 text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/20 rounded-lg transition-colors cursor-pointer"
					>
						<LogOut size={20} strokeWidth={1.5} />
						<span class="text-sm font-medium">Déconnexion</span>
					</button>
				</form>
			</div>
		</div>
	</Drawer>
</div>
