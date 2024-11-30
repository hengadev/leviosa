<script lang="ts">
	import '../app.pcss';
	// global styles for all the components
	import '../styles/global.css';
	import '../styles/layout.css';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();
	// TODO: how to condition the layout in here.
</script>

{@render children?.()}
