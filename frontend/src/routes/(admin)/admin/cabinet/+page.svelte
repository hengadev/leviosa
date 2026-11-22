<script lang="ts">
	// Cabinet: the rarely used management pages, gathered behind the phone
	// bar's last tab, and the account (the public site, logout).
	import { ChevronRight, LogOut } from "@lucide/svelte";
	import PageActionsMenu from "#lib/components/admin/PageActionsMenu.svelte";
	import { accountInitial, accountName } from "#lib/navigation/BackOfficeSidebar.svelte";
	import { accountLinks, cabinetGroups } from "#lib/navigation/practiceNav.ts";
	import type { PageProps } from "./$types";

	let { data }: PageProps = $props();

	const role = $derived(data.selfPractitioner ? "Administrateur · praticien" : "Administrateur");
</script>

<svelte:head>
	<title>Cabinet | Admin</title>
</svelte:head>

<div class="container mx-auto max-w-2xl px-4 py-8 lg:py-12">
	<div class="mb-6 flex items-center justify-between gap-4">
		<h1 class="text-3xl lg:text-4xl font-bold text-foreground">Cabinet</h1>
		<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
		<div class="lg:hidden">
			<PageActionsMenu />
		</div>
	</div>

	<div class="space-y-6">
		<div class="flex items-center gap-3">
			<div class="flex size-12 shrink-0 items-center justify-center rounded-full border border-border-card bg-muted font-semibold text-foreground-alt">
				{accountInitial(data.user)}
			</div>
			<div class="min-w-0">
				<p class="truncate font-medium text-foreground">{accountName(data.user)}</p>
				<p class="text-xs text-muted-foreground">{role}</p>
			</div>
		</div>

		{#each cabinetGroups as group (group.label)}
			<section>
				<h2 class="px-1 pb-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{group.label}</h2>
				<ul class="divide-y divide-border-card overflow-hidden rounded-2xl border border-border-card">
					{#each group.items as item (item.href)}
						<li>
							<a
								href={item.href}
								class="flex min-h-11 items-center gap-3 px-4 py-3 transition-colors hover:bg-muted"
							>
								<item.icon size={20} strokeWidth={1.5} class="shrink-0 text-muted-foreground" />
								<span class="min-w-0 flex-1">
									<span class="block font-medium text-foreground">{item.label}</span>
									<span class="block truncate text-xs text-muted-foreground">{item.hint}</span>
								</span>
								<ChevronRight size={18} class="shrink-0 text-muted-foreground" />
							</a>
						</li>
					{/each}
				</ul>
			</section>
		{/each}

		<ul class="divide-y divide-border-card overflow-hidden rounded-2xl border border-border-card">
			{#each accountLinks as link (link.href)}
				<li>
					<a
						href={link.href}
						target={link.external ? "_blank" : undefined}
						rel={link.external ? "noopener noreferrer" : undefined}
						class="flex min-h-11 items-center gap-3 px-4 py-3.5 text-foreground transition-colors hover:bg-muted"
					>
						<link.icon size={20} strokeWidth={1.5} class="shrink-0 text-muted-foreground" />
						<span class="flex-1">{link.label}</span>
					</a>
				</li>
			{/each}
			<li>
				<form method="POST" action="/logout">
					<button
						type="submit"
						class="flex min-h-11 w-full cursor-pointer items-center gap-3 px-4 py-3.5 text-left text-red-600 transition-colors hover:bg-red-50 dark:hover:bg-red-950/20"
					>
						<LogOut size={20} strokeWidth={1.5} class="shrink-0" />
						<span>Déconnexion</span>
					</button>
				</form>
			</li>
		</ul>
	</div>
</div>
