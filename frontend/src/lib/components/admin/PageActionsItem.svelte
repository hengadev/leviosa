<!--
	@component
	One action of a PageActionsMenu.
-->
<script lang="ts">
	import type { Component, Snippet } from "svelte";
	import { DropdownMenu } from "bits-ui";

	let {
		icon: Icon,
		onSelect,
		children
	}: { icon?: Component<{ size?: number }>; onSelect: () => void; children: Snippet } = $props();
</script>

<DropdownMenu.Item
	{onSelect}
	class="flex min-h-11 items-center gap-2.5 px-3 py-2 text-sm text-foreground-alt cursor-pointer outline-none data-highlighted:text-foreground data-highlighted:bg-muted transition-colors"
>
	{#if Icon}<Icon size={15} />{/if}
	<span>{@render children()}</span>
</DropdownMenu.Item>
