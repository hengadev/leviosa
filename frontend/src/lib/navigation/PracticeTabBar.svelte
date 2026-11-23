<script lang="ts">
	// The administrator's phone bar: the five tabs of the practice navigation,
	// no Menu and no drawer. The active tab gets a stronger icon, a light label on
	// the dark bar.
	import { page } from "$app/state";
	import { activeTab, tabs } from "./practiceNav.ts";

	const active = $derived(activeTab(page.url.pathname));
</script>

<nav
	class="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-foreground dark:bg-black dark:border-t dark:border-border-card pb-1 [@media(display-mode:standalone)]:pb-6"
	aria-label="Navigation"
>
	<ul class="flex">
		{#each tabs as tab (tab.id)}
			{@const current = active === tab.id}
			<li class="flex-1 min-w-0">
				<a
					href={tab.href}
					aria-current={current ? "page" : undefined}
					class="flex flex-col items-center justify-center gap-1 min-h-14 {current ? 'text-background dark:text-foreground' : 'text-zinc-500 hover:text-zinc-300'}"
				>
					<span class="flex items-center justify-center">
						<tab.icon size={20} strokeWidth={current ? 2.2 : 1.5} />
					</span>
					<span class="max-w-full truncate text-[11px] font-medium tracking-tight">{tab.label}</span>
				</a>
			</li>
		{/each}
	</ul>
</nav>
