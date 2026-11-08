<script lang="ts" module>
	import type { Component } from "svelte";

	export interface NavItem {
		href: string;
		label: string;
		icon: Component;
	}

	/** Navigation items under a small heading. */
	export interface NavGroup {
		label: string;
		items: NavItem[];
	}

	/** A link of the account menu (another area, the public site). */
	export interface AccountLink {
		href: string;
		label: string;
		icon: Component;
		external?: boolean;
	}

	/** The name the account shows: first and last name, else the email. */
	export function accountName(user: App.User): string {
		return [user.firstname, user.lastname].filter(Boolean).join(" ") || user.email;
	}

	export function accountInitial(user: App.User): string {
		return (user.firstname?.[0] ?? user.email?.[0] ?? "L").toUpperCase();
	}

	export function roleLabel(user: App.User): string {
		return user.role === "administrator" ? "Administrateur" : user.role === "partner" ? "Praticien" : user.role;
	}
</script>

<script lang="ts">
	// The desktop sidebar of the admin and practitioner areas: the same look,
	// each area gives its own navigation and account links.
	import { page } from "$app/state";
	import { ChevronLeft, ChevronRight, EllipsisVertical, LogOut } from "@lucide/svelte";
	import ThemeToggle from "#lib/components/admin/ThemeToggle.svelte";

	interface Props {
		/** Under the brand name, e.g. "Administration". */
		area: string;
		/** The area's home page: active only on itself, not on its subpages. */
		home: string;
		groups: NavGroup[];
		accountLinks: AccountLink[];
		user: App.User;
	}

	let { area, home, groups, accountLinks, user }: Props = $props();

	let isCollapsed = $state(false);
	let userMenuOpen = $state(false);

	function closeUserMenu() {
		userMenuOpen = false;
	}

	function isActive(href: string): boolean {
		const path = page.url.pathname;
		if (href === home) return path === home || path === `${home}/`;
		return path.startsWith(href);
	}

	const name = $derived(accountName(user));
	const initial = $derived(accountInitial(user));
	const role = $derived(roleLabel(user));
</script>

<aside
	class="hidden lg:flex lg:flex-col lg:border-r lg:border-border-card bg-background sticky top-0 h-screen transition-all duration-300 {isCollapsed
		? 'lg:w-20'
		: 'lg:w-64'}"
	aria-label="Navigation"
>
	<!-- Collapse Toggle Button -->
	<button
		onclick={() => (isCollapsed = !isCollapsed)}
		class="absolute -right-4 bottom-24 z-10 w-8 h-8 rounded-full bg-background flex items-center justify-center text-foreground-alt hover:text-foreground hover:bg-muted transition-all duration-200 border border-border-card shadow-mini"
		aria-label={isCollapsed ? "Agrandir la barre latérale" : "Réduire la barre latérale"}
		title={isCollapsed ? "Agrandir la barre latérale" : "Réduire la barre latérale"}
	>
		{#if isCollapsed}
			<ChevronRight size={16} strokeWidth={2} />
		{:else}
			<ChevronLeft size={16} strokeWidth={2} />
		{/if}
	</button>

	<!-- Sidebar Header -->
	<div class="flex items-center justify-between border-b border-border-card py-6 {isCollapsed ? 'px-3' : 'px-6'}">
		{#if !isCollapsed}
			<div>
				<p class="text-sm font-semibold tracking-tight text-foreground uppercase">Leviosa</p>
				<p class="text-xs text-muted-foreground">{area}</p>
			</div>
			<ThemeToggle />
		{:else}
			<div class="mx-auto flex flex-col items-center gap-2">
				<span class="text-lg font-bold text-foreground" title="Leviosa · {area}">L</span>
				<ThemeToggle />
			</div>
		{/if}
	</div>

	<!-- Navigation Items -->
	<nav class="flex-1 overflow-y-auto py-6 {isCollapsed ? 'px-3' : 'px-4'}">
		{#each groups as group, groupIndex (group.label)}
			{#if isCollapsed}
				{#if groupIndex > 0}
					<div class="my-3 mx-3 border-t border-border-card" role="separator"></div>
				{/if}
			{:else}
				<p class="px-3 {groupIndex > 0 ? 'pt-4' : ''} pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
					{group.label}
				</p>
			{/if}
			<ul class="space-y-0.5" aria-label={group.label}>
				{#each group.items as item (item.href)}
					{@const active = isActive(item.href)}
					<li>
						<a
							href={item.href}
							class="flex items-center text-sm font-medium transition-all duration-200 rounded-lg {isCollapsed
								? 'justify-center px-3 py-3'
								: 'gap-3 px-3 py-2'} {active
								? 'bg-foreground text-background hover:bg-foreground/85'
								: 'text-foreground-alt hover:text-foreground hover:bg-muted'}"
							aria-current={active ? "page" : undefined}
							title={isCollapsed ? item.label : undefined}
						>
							<item.icon size={20} strokeWidth={active ? 2 : 1.5} />
							{#if !isCollapsed}
								<span class="tracking-tight">{item.label}</span>
							{/if}
						</a>
					</li>
				{/each}
			</ul>
		{/each}
	</nav>

	<!-- Sidebar Footer: the same account menu, expanded or collapsed -->
	{#snippet accountMenu(placement: string)}
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
		<div class="fixed inset-0 z-10" onclick={closeUserMenu}></div>
		<div
			class="absolute z-20 bg-background border border-border-card rounded-lg shadow-popover py-1 overflow-hidden {placement}"
			role="menu"
		>
			{#if isCollapsed}
				<div class="px-3 py-2 border-b border-border-card mb-1">
					<p class="text-sm font-medium text-foreground truncate tracking-tight">{name}</p>
					<p class="text-xs text-muted-foreground">{role}</p>
				</div>
			{/if}
			{#each accountLinks as link (link.href)}
				<a
					href={link.href}
					target={link.external ? "_blank" : undefined}
					rel={link.external ? "noopener noreferrer" : undefined}
					onclick={closeUserMenu}
					role="menuitem"
					class="flex items-center gap-2.5 px-3 py-2 text-sm text-foreground-alt hover:text-foreground hover:bg-muted transition-colors"
				>
					<link.icon size={15} />
					<span>{link.label}</span>
				</a>
			{/each}
			<div class="my-1 border-t border-border-card"></div>
			<form method="POST" action="/logout">
				<button
					type="submit"
					role="menuitem"
					class="w-full flex items-center gap-2.5 px-3 py-2 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950/20 transition-colors cursor-pointer"
				>
					<LogOut size={15} />
					<span>Déconnexion</span>
				</button>
			</form>
		</div>
	{/snippet}

	<div class="py-5 border-t border-border-card {isCollapsed ? 'px-3' : 'px-6'}">
		<div class="relative">
			{#if userMenuOpen}
				{@render accountMenu(isCollapsed ? "left-full bottom-0 ml-3 w-60" : "bottom-full left-0 right-0 mb-2")}
			{/if}
			{#if !isCollapsed}
				<div class="flex items-center gap-3">
					<div class="w-9 h-9 flex-shrink-0 rounded-full flex items-center justify-center bg-muted border border-border-card">
						<span class="text-xs font-semibold text-foreground-alt">{initial}</span>
					</div>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-foreground truncate tracking-tight" title={user.email}>{name}</p>
						<p class="text-xs text-muted-foreground">{role}</p>
					</div>
					<button
						onclick={() => (userMenuOpen = !userMenuOpen)}
						class="flex-shrink-0 p-1.5 rounded-md text-foreground-alt hover:text-foreground hover:bg-muted transition-colors"
						aria-label="Options utilisateur"
						aria-haspopup="menu"
						aria-expanded={userMenuOpen}
						title="Options"
					>
						<EllipsisVertical size={16} />
					</button>
				</div>
			{:else}
				<!-- Collapsed: the avatar opens the same menu, beside the sidebar. -->
				<button
					onclick={() => (userMenuOpen = !userMenuOpen)}
					class="mx-auto flex w-9 h-9 rounded-full items-center justify-center bg-muted border border-border-card hover:border-foreground-alt transition-colors"
					aria-label="Options utilisateur"
					aria-haspopup="menu"
					aria-expanded={userMenuOpen}
					title={name}
				>
					<span class="text-xs font-semibold text-foreground-alt">{initial}</span>
				</button>
			{/if}
		</div>
	</div>
</aside>
