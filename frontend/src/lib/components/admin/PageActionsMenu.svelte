<!--
	@component
	The "⋮" menu of a back-office page title. The dark mode switch is always
	its last item; a page adds its own actions above it with PageActionsItem:

	<PageActionsMenu>
		<PageActionsItem icon={Download} onSelect={exportCsv}>Exporter</PageActionsItem>
	</PageActionsMenu>
-->
<script lang="ts">
	import type { Snippet } from "svelte";
	import { DropdownMenu } from "bits-ui";
	import { EllipsisVertical, Moon, Sun } from "@lucide/svelte";
	import { adminDarkMode } from "#lib/stores/adminDarkMode.svelte.ts";
	import PageActionsItem from "./PageActionsItem.svelte";

	let { children }: { children?: Snippet } = $props();

	const isDark = $derived(adminDarkMode.get());
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class="inline-flex size-11 items-center justify-center rounded-lg transition-colors text-muted-foreground hover:text-foreground hover:bg-surface-hover"
		aria-label="Plus d'actions"
	>
		<EllipsisVertical size={20} />
	</DropdownMenu.Trigger>
	<DropdownMenu.Portal>
		<DropdownMenu.Content
			align="end"
			sideOffset={4}
			class="z-50 min-w-48 bg-background border border-border-card rounded-lg shadow-popover py-1"
		>
			{#if children}
				{@render children()}
				<DropdownMenu.Separator class="my-1 border-t border-border-card" />
			{/if}
			<PageActionsItem icon={isDark ? Sun : Moon} onSelect={() => adminDarkMode.toggle()}>
				{isDark ? "Mode clair" : "Mode sombre"}
			</PageActionsItem>
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>
