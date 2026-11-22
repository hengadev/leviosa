<script lang="ts">
	// An administrator's navigation, the same in the admin area and the Partner
	// area: the grouped sidebar from lg up, the five-tab bar below it.
	import BackOfficeSidebar from "./BackOfficeSidebar.svelte";
	import PracticeTabBar from "./PracticeTabBar.svelte";
	import { accountLinks, home, sidebarGroups } from "./practiceNav.ts";

	let { user }: { user: App.User } = $props();
</script>

<BackOfficeSidebar area="Administration" {home} groups={sidebarGroups} {accountLinks} {user} />

<PracticeTabBar />
