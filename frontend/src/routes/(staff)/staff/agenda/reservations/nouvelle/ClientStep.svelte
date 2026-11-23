<!--
	@component
	The client step of "Nouvelle réservation": a client entered by hand
	(first name, last name and phone; email optional). Its own component so
	issue 043 can add a search of existing clients above the fields.
-->
<script lang="ts">
	import { validateClient, type ClientErrors, type GuestClient } from './client.ts';

	let {
		client = $bindable(),
		onNext,
		onBack
	}: { client: GuestClient; onNext: () => void; onBack: () => void } = $props();

	let errors = $state<ClientErrors>({});

	function submit(event: SubmitEvent) {
		event.preventDefault();
		errors = validateClient(client);
		if (Object.keys(errors).length === 0) onNext();
	}

	const inputClass =
		'block h-12 w-full rounded-input border border-border-input bg-background px-3 text-base text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-foreground';
</script>

<form class="space-y-4" novalidate onsubmit={submit}>
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
		<div>
			<label for="client-first-name" class="mb-1.5 block text-sm font-medium">Prénom</label>
			<input
				id="client-first-name"
				class={inputClass}
				autocomplete="off"
				autocapitalize="words"
				aria-invalid={errors.firstName ? 'true' : undefined}
				aria-describedby={errors.firstName ? 'client-first-name-error' : undefined}
				bind:value={client.firstName}
			/>
			{#if errors.firstName}
				<p id="client-first-name-error" class="mt-1 text-sm text-destructive">{errors.firstName}</p>
			{/if}
		</div>
		<div>
			<label for="client-last-name" class="mb-1.5 block text-sm font-medium">Nom</label>
			<input
				id="client-last-name"
				class={inputClass}
				autocomplete="off"
				autocapitalize="words"
				aria-invalid={errors.lastName ? 'true' : undefined}
				aria-describedby={errors.lastName ? 'client-last-name-error' : undefined}
				bind:value={client.lastName}
			/>
			{#if errors.lastName}
				<p id="client-last-name-error" class="mt-1 text-sm text-destructive">{errors.lastName}</p>
			{/if}
		</div>
	</div>
	<div>
		<label for="client-phone" class="mb-1.5 block text-sm font-medium">Téléphone</label>
		<input
			id="client-phone"
			type="tel"
			inputmode="tel"
			class={inputClass}
			autocomplete="off"
			placeholder="06 12 34 56 78"
			aria-invalid={errors.phone ? 'true' : undefined}
			aria-describedby={errors.phone ? 'client-phone-error' : undefined}
			bind:value={client.phone}
		/>
		{#if errors.phone}
			<p id="client-phone-error" class="mt-1 text-sm text-destructive">{errors.phone}</p>
		{/if}
	</div>
	<div>
		<label for="client-email" class="mb-1.5 block text-sm font-medium">
			Email <span class="font-normal text-muted-foreground">(facultatif)</span>
		</label>
		<input
			id="client-email"
			type="email"
			inputmode="email"
			class={inputClass}
			autocomplete="off"
			autocapitalize="off"
			aria-invalid={errors.email ? 'true' : undefined}
			aria-describedby={errors.email ? 'client-email-error' : 'client-email-hint'}
			bind:value={client.email}
		/>
		{#if errors.email}
			<p id="client-email-error" class="mt-1 text-sm text-destructive">{errors.email}</p>
		{:else}
			<p id="client-email-hint" class="mt-1 text-sm text-muted-foreground">
				Gardé dans sa fiche. Aucun message ne lui est envoyé.
			</p>
		{/if}
	</div>

	<div class="flex flex-col-reverse gap-3 pt-2 sm:flex-row sm:justify-end">
		<button
			type="button"
			class="inline-flex h-12 items-center justify-center rounded-md border border-border px-6 text-sm font-medium hover:bg-muted"
			onclick={onBack}
		>
			Retour
		</button>
		<button
			type="submit"
			class="inline-flex h-12 items-center justify-center rounded-md bg-foreground px-6 text-sm font-semibold text-background hover:bg-foreground/90"
		>
			Continuer
		</button>
	</div>
</form>
