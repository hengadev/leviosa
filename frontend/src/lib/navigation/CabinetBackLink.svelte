<script lang="ts">
	// On a phone, the pages reached from Cabinet get a way back to it. From lg
	// up the sidebar lists them, so it doesn't show.
	import { ChevronLeft } from "@lucide/svelte";
	import { page } from "$app/state";
	import { cabinetHref, isCabinetPage } from "./practiceNav.ts";
</script>

{#if isCabinetPage(page.url.pathname)}
	<div class="lg:hidden container mx-auto px-4 pt-2 -mb-6">
		<a
			href={cabinetHref}
			class="-ml-2 inline-flex min-h-11 items-center gap-1 pl-1 pr-3 rounded-lg text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-surface-hover transition-colors"
		>
			<ChevronLeft size={20} />
			Cabinet
		</a>
	</div>
{/if}
