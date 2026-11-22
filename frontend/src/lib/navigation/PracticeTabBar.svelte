<script lang="ts">
	// The administrator's phone bar: the five tabs of the practice navigation,
	// no Menu and no drawer. The active tab gets a stronger icon, a soft pill
	// behind it and a darker label.
	import { page } from "$app/state";
	import { activeTab, tabs } from "./practiceNav.ts";

	const active = $derived(activeTab(page.url.pathname));
</script>

<nav
	class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-background border-t border-border-card pb-[env(safe-area-inset-bottom)]"
	aria-label="Navigation"
>
	<ul class="flex">
		{#each tabs as tab (tab.id)}
			{@const current = active === tab.id}
			<li class="flex-1 min-w-0">
				<a
					href={tab.href}
					aria-current={current ? "page" : undefined}
					class="flex flex-col items-center gap-1 py-2.5 {current ? 'text-foreground' : 'text-muted-foreground'}"
				>
					<span class="flex h-7 w-14 items-center justify-center rounded-full {current ? 'bg-muted' : ''}">
						<tab.icon size={20} strokeWidth={current ? 2.2 : 1.5} />
					</span>
					<span class="max-w-full truncate text-[11px] font-medium tracking-tight">{tab.label}</span>
				</a>
			</li>
		{/each}
	</ul>
</nav>
