<script lang="ts">
	import type { LayoutProps } from "./$types";
	import { setUserContext } from "#lib/context/user.svelte.ts";
	import StaffSidebar from "#lib/navigation/StaffSidebar.svelte";
	import { adminDarkMode } from "#lib/stores/adminDarkMode.svelte.ts";
	import { browser } from "$app/env";

	let { children, data }: LayoutProps = $props();

	setUserContext(data.user);

	// Apply dark mode class to html element (same pattern as admin)
	$effect(() => {
		if (browser) {
			if (adminDarkMode.get()) {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		}
	});

	// Clean up: remove dark class when leaving staff routes
	$effect(() => {
		return () => {
			if (browser) {
				document.documentElement.classList.remove('dark');
			}
		};
	});
</script>

<!-- One tree for every width: the page renders once. The sidebar shows
     from lg up, the bottom bar below it. -->
<div class="flex min-h-screen bg-muted/25">
	<StaffSidebar user={data.user} permissions={data.permissions} unreadCount={data.unreadCount} />
	<!-- Room under the content for the fixed bottom bar and the home indicator -->
	<main class="flex-1 min-w-0 bg-background overflow-x-hidden pb-[calc(5rem+env(safe-area-inset-bottom))] lg:pb-0">
		{@render children()}
	</main>
</div>
