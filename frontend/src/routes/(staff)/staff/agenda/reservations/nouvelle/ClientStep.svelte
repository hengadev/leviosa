<!--
	@component
	The client step of "Nouvelle réservation": a search of the practitioner's
	existing clients (issue 043) above the manual entry (first name, last name
	and phone; email optional). Picking a result pre-fills the fields; a client
	with an account is booked under their id.
-->
<script lang="ts">
	import {
		displayPhone,
		guestFieldsOf,
		validateClient,
		type ClientErrors,
		type ClientSummary,
		type GuestClient
	} from './client.ts';

	let {
		client = $bindable(),
		partnerId,
		pickedClientId = $bindable(),
		onNext,
		onBack
	}: {
		client: GuestClient;
		partnerId: string;
		/** The account id of the client picked in the search, when they have
		 * one: the booking then carries client_id instead of guest fields. */
		pickedClientId: string | null;
		onNext: () => void;
		onBack: () => void;
	} = $props();

	let errors = $state<ClientErrors>({});

	// ── Search of existing clients ───────────────────────────────────

	let query = $state('');
	let results = $state<ClientSummary[]>([]);
	let searching = $state(false);
	let searched = $state(false);
	let searchError = $state('');
	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	let searchSeq = 0;
	let firstNameInput: HTMLInputElement | undefined = $state();

	const SEARCH_DEBOUNCE_MS = 250;

	/** Typing in the search or the fields gives up the picked client: the
	 * entry becomes manual again (booked as a guest). */
	function forgetPick() {
		pickedClientId = null;
	}

	function onQueryInput() {
		forgetPick();
		clearTimeout(searchTimer);
		if (query.trim().length < 2) {
			results = [];
			searched = false;
			searchError = '';
			return;
		}
		searchTimer = setTimeout(searchClients, SEARCH_DEBOUNCE_MS);
	}

	async function searchClients() {
		const q = query.trim();
		if (q.length < 2) return;
		const seq = ++searchSeq;
		searching = true;
		searchError = '';
		try {
			const params = new URLSearchParams({ partner_id: partnerId, q });
			const res = await fetch(`/api/partners/clients?${params}`);
			if (seq !== searchSeq) return;
			if (!res.ok) {
				searchError = 'Impossible de chercher les clients. Veuillez réessayer.';
				return;
			}
			const body = await res.json();
			results = Array.isArray(body) ? body : [];
			searched = true;
		} catch {
			if (seq === searchSeq) searchError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			if (seq === searchSeq) searching = false;
		}
	}

	function pick(summary: ClientSummary) {
		pickedClientId = summary.client_id ?? null;
		client = { ...guestFieldsOf(summary) };
		query = '';
		results = [];
		searched = false;
	}

	function newClient() {
		forgetPick();
		query = '';
		results = [];
		searched = false;
		firstNameInput?.focus();
	}

	function submit(event: SubmitEvent) {
		event.preventDefault();
		errors = validateClient(client);
		if (Object.keys(errors).length === 0) onNext();
	}

	const inputClass =
		'block h-12 w-full rounded-input border border-border-input bg-background px-3 text-base text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-foreground';
</script>

<form class="space-y-4" novalidate onsubmit={submit}>
	<!-- The search of existing clients (issue 043) -->
	<div>
		<label for="client-search" class="mb-1.5 block text-sm font-medium">
			Rechercher un client <span class="font-normal text-muted-foreground">(facultatif)</span>
		</label>
		<input
			id="client-search"
			type="search"
			class={inputClass}
			autocomplete="off"
			inputmode="search"
			placeholder="Nom ou numéro de téléphone"
			aria-describedby="client-search-hint"
			bind:value={query}
			oninput={onQueryInput}
		/>
		<p id="client-search-hint" class="mt-1 text-sm text-muted-foreground">
			Au moins 2 caractères. Vous pouvez aussi saisir un nouveau client ci-dessous.
		</p>

		{#if query.trim().length >= 2}
			{#if searching}
				<p class="mt-2 py-2 text-sm text-muted-foreground" role="status">Recherche…</p>
			{:else if searchError}
				<div class="mt-2 space-y-2">
					<p role="alert" class="text-sm text-destructive">{searchError}</p>
					<button
						type="button"
						class="inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm font-medium hover:bg-muted"
						onclick={searchClients}
					>
						Réessayer
					</button>
				</div>
			{:else if searched && results.length === 0}
				<div class="mt-2 space-y-2">
					<p class="text-sm text-muted-foreground">Aucun client trouvé.</p>
					<button
						type="button"
						class="inline-flex min-h-11 items-center rounded-md bg-foreground px-4 text-sm font-semibold text-background hover:bg-foreground/90"
						onclick={newClient}
					>
						Nouveau client
					</button>
				</div>
			{:else if results.length > 0}
				<ul class="mt-2 space-y-2" aria-label="Clients trouvés">
					{#each results as summary (summary.client_id ?? `${summary.name}|${summary.phone ?? ''}`)}
						<li>
							<button
								type="button"
								class="flex min-h-11 w-full items-center justify-between gap-3 rounded-xl border border-border px-4 py-2.5 text-left transition-colors hover:bg-muted"
								onclick={() => pick(summary)}
							>
								<span class="min-w-0 font-medium break-words">{summary.name}</span>
								<span class="shrink-0 text-sm tabular-nums text-muted-foreground">
									{summary.phone ? displayPhone(summary.phone) : summary.email || ''}
								</span>
							</button>
						</li>
					{/each}
				</ul>
			{/if}
		{/if}
	</div>

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
				bind:this={firstNameInput}
				bind:value={client.firstName}
				oninput={forgetPick}
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
				oninput={forgetPick}
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
			oninput={forgetPick}
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
			oninput={forgetPick}
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
