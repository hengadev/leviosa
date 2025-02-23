<script lang="ts">
	import { run, stopPropagation, createBubbler } from 'svelte/legacy';

	const bubble = createBubbler();
	// =======================
	// Props and Reactive Variables
	// =======================
	import { slide, fade } from 'svelte/transition';
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment'; // Check if in browser

	let scrollPosition: number = 0;

	// =======================
	// Store and State Imports
	// =======================
	import { createVerticalSwipeHandler } from '$lib/scripts/swipe';
	interface Props {
		isOpen?: boolean;
		closeDrawer: () => void;
		children?: import('svelte').Snippet;
	}

	let { isOpen = $bindable(false), closeDrawer, children }: Props = $props();

	// =======================
	// Helper Functions
	// =======================
	function onSwipe(direction: 'top' | 'bottom') {
		if (direction === 'bottom') {
			closeDrawer();
		}
	}

	const closeSwipeAction = createVerticalSwipeHandler(onSwipe);
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			closeDrawer();
		}
	}

	function toggleBodyScroll(isOpen: boolean) {
		if (!browser) return;
		if (isOpen) {
			// save the current position
			scrollPosition = window.scrollY;
			document.body.style.position = 'fixed';
			document.body.style.top = `-${scrollPosition}px`;
			document.body.style.width = '100%';
		} else {
			document.body.style.position = '';
			document.body.style.top = '';
			document.body.style.width = '';
			window.scrollTo(0, scrollPosition);
		}
	}

	run(() => {
		toggleBodyScroll(isOpen);
	});

	onDestroy(() => {
		if (!browser) return;
		document.body.style.position = '';
		document.body.style.top = '';
		window.scrollTo(0, scrollPosition);
	});
</script>

{#if isOpen}
	<div
		transition:fade={{ duration: 300 }}
		class="overlay"
		onclick={closeDrawer}
		onkeydown={stopPropagation(handleKeydown)}
		class:visible={isOpen}
		tabindex="0"
		role="button"
	></div>
	<div
		transition:slide={{ duration: 300 }}
		class="drawer"
		class:visible={isOpen}
		onclick={stopPropagation(bubble('click'))}
		onkeydown={stopPropagation(handleKeydown)}
		use:closeSwipeAction.action
		tabindex="0"
		role="button"
	>
		{@render children?.()}
	</div>
	<!-- <div class="modal"> -->
	<!--     <slot></slot> -->
	<!-- </div> -->
{/if}

<style>
	.overlay {
		position: fixed;
		top: 0;
		left: 0;
		height: 100%;
		width: 100vw;
		background: rgba(0, 0, 0, 0.2);
		opacity: 0;
		visibility: hidden;
	}
	.overlay.visible {
		opacity: 1;
		visibility: visible;
	}

	.drawer {
		--border-top-radius: 1.2rem;
		position: fixed;
		bottom: -100%;
		left: 0;
		width: 100%;
		background: hsl(var(--clr-light-primary));
		/* padding: 2rem 1rem; */
		padding: 1rem;
		box-shadow: 0 -1px 10px rgba(0, 0, 0, 0.2);
		border-top-left-radius: var(--border-top-radius);
		border-top-right-radius: var(--border-top-radius);
		/* TODO: should the drawer be on top of the navigation, it makes sense but it is weird right? */
		z-index: 9999;
		color: hsl(var(--clr-dark-primary));
	}
	.drawer.visible {
		bottom: 0;
	}
</style>
