<script lang="ts">
	interface Props {
		options?: any;
		name: string;
	}

	let {
		options = [
			{ id: 'fweg0USD/agwru|awf', text: 'this is the second option' },
			{ id: 'ega0rgea34OIHF-fwe', text: 'this is the first option' },
			{
				id: 'awega9wgruH>Faw4F1',
				text: 'this is some option that is hello long and I do not know if that thing is going to be weird but who care the point is to have the longest text possible to test the limit of that thing right ? But now I want my text to be bigger than 100vw so that I get two lines even on full screen.'
			}
		],
		name
	}: Props = $props();

	function toggleChecked(e: MouseEvent) {
		const target = e.currentTarget as HTMLButtonElement;
		const input = target.querySelector("input[type='radio']") as HTMLInputElement;
		if (input) input.click();
	}
	// TODO: init so that the first button is checked initially
	// -> just did it with the checked thing inside the input, hope it works fine right ?
</script>

<div class="component grid" style="--gap: 1rem;">
	{#each options as option, index}
		<button type="button" onclick={toggleChecked} class="option flex">
			<div class="radiobutton-container">
				<div class="radiobutton-outer stroke">
					<div class="radiobutton-inner"></div>
				</div>
			</div>
			<input class="radioinput" checked={index === 0} type="radio" {name} id={option.id} />
			<label for={option.id} class="text">
				{option.text}
			</label>
		</button>
	{/each}
</div>

<style>
	.option:has(:global(input[type='radio']:checked)) .radiobutton-outer {
		background-color: hsl(var(--clr-dark-primary));
	}
	input[type='radio'] {
		visibility: hidden;
		display: none;
	}
	.radiobutton-inner {
		width: calc(var(--dimension) / 2);
		height: calc(var(--dimension) / 2);
		border-radius: 100%;
		background-color: hsl(var(--clr-light-primary));
	}
	.radiobutton-outer {
		width: var(--dimension);
		height: var(--dimension);
		border-radius: 100%;
		background-color: hsl(var(--clr-light-primary));
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.radiobutton-container {
		--dimension: calc(1.5 * var(--p));
		position: absolute;
		right: 0;
		top: 0;
		height: 100%;
		width: 2.5rem;
		width: 4.5rem;
		display: flex;
		align-items: center;
		justify-content: center;
		pointer-events: none;
	}
	.component {
		border-radius: 0.5rem;
	}
	.option {
		padding-block: 0.75rem;
		position: relative;
		width: 100%;
		background: transparent;
		align-items: center;
		font-weight: 500;
		color: hsl(var(--clr-dark-primary));
	}
	.text {
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: 1;
		line-clamp: 1;
		overflow: hidden;
		text-align: left;
	}
</style>
