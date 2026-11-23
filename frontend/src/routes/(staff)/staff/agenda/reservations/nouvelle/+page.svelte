<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { ChevronLeft } from '@lucide/svelte';
	import PageActionsMenu from '#lib/components/admin/PageActionsMenu.svelte';
	import PractitionerProfileMissing from '#lib/components/admin/PractitionerProfileMissing.svelte';
	import { formatEuros } from '#lib/onSitePayment.ts';
	import ClientStep from './ClientStep.svelte';
	import { displayPhone, emptyGuestClient, normalizePhone, type GuestClient } from './client.ts';
	import type { PageProps } from './$types';
	import type { BookableProduct } from './+page.server';

	let { data }: PageProps = $props();

	// ── Steps: product → day and free slot → client → confirm ─────────

	type Step = 'product' | 'slot' | 'client' | 'confirm';
	const STEPS: [Step, string][] = [
		['product', 'Soin'],
		['slot', 'Créneau'],
		['client', 'Client'],
		['confirm', 'Confirmation']
	];

	/** One free slot (GET /api/partners/booking-slots); instants carry the
	 * Paris offset, so their first characters are the Paris day and time. */
	interface FreeSlot {
		availability_id: string;
		start: string;
		end: string;
	}

	// A single product needs no choosing: the form opens on the slots.
	const onlyProduct = (() => (data.products.length === 1 ? data.products[0] : null))();
	let step = $state<Step>(onlyProduct ? 'slot' : 'product');
	let product = $state<BookableProduct | null>(onlyProduct);
	let slot = $state<FreeSlot | null>(null);
	let client = $state<GuestClient>(emptyGuestClient());

	// The free slots, fetched by windows of days from today.
	const WINDOW_DAYS = 21;
	const MAX_DAYS = 120;
	let slots = $state<FreeSlot[]>([]);
	let loadedDays = $state(0);
	let slotsLoading = $state(false);
	let slotsError = $state('');
	let selectedDay = $state<string | null>(null);
	/** Why the form came back to the slot step (a slot taken meanwhile). */
	let slotNotice = $state('');

	let submitting = $state(false);
	let submitError = $state('');

	const stepIndex = $derived(STEPS.findIndex(([s]) => s === step));
	const returnLabel = $derived(data.from === '/admin' ? "Aujourd'hui" : 'Réservations');

	onMount(() => {
		if (product) loadSlots(true);
	});

	// ── Dates (Europe/Paris) ─────────────────────────────────────────

	function parisToday(): string {
		return new Intl.DateTimeFormat('en-CA', {
			timeZone: 'Europe/Paris',
			year: 'numeric',
			month: '2-digit',
			day: '2-digit'
		}).format(new Date());
	}

	/** The YYYY-MM-DD `days` after `day`. */
	function addDays(day: string, days: number): string {
		const date = new Date(`${day}T12:00:00Z`);
		date.setUTCDate(date.getUTCDate() + days);
		return date.toISOString().slice(0, 10);
	}

	function dayOf(instant: string): string {
		return instant.slice(0, 10);
	}

	function timeOf(instant: string): string {
		return instant.slice(11, 16);
	}

	function capitalize(text: string): string {
		return text.charAt(0).toUpperCase() + text.slice(1);
	}

	function shortDayLabel(day: string): string {
		const today = parisToday();
		if (day === today) return "Aujourd'hui";
		if (day === addDays(today, 1)) return 'Demain';
		return capitalize(
			new Date(`${day}T12:00:00Z`).toLocaleDateString('fr-FR', {
				weekday: 'short',
				day: 'numeric',
				month: 'short',
				timeZone: 'UTC'
			})
		);
	}

	function longDayLabel(day: string): string {
		return capitalize(
			new Date(`${day}T12:00:00Z`).toLocaleDateString('fr-FR', {
				weekday: 'long',
				day: 'numeric',
				month: 'long',
				timeZone: 'UTC'
			})
		);
	}

	function durationLabel(minutes: number): string {
		if (minutes < 60) return `${minutes} min`;
		const hours = Math.floor(minutes / 60);
		const rest = minutes % 60;
		return rest ? `${hours} h ${String(rest).padStart(2, '0')}` : `${hours} h`;
	}

	// ── Slots ────────────────────────────────────────────────────────

	/** The days with a free slot, in order, and their slots. */
	const days = $derived.by(() => {
		const byDay = new Map<string, FreeSlot[]>();
		for (const s of slots) {
			const day = dayOf(s.start);
			byDay.set(day, [...(byDay.get(day) ?? []), s]);
		}
		return [...byDay.entries()];
	});
	const daySlots = $derived(days.find(([day]) => day === selectedDay)?.[1] ?? []);
	const searchedUntil = $derived(loadedDays > 0 ? addDays(parisToday(), loadedDays - 1) : '');

	/** Fetches the next window of days; `reset` starts again from today. */
	async function loadSlots(reset: boolean) {
		if (!product) return;
		const start = reset ? 0 : loadedDays;
		const from = addDays(parisToday(), start);
		const to = addDays(parisToday(), start + WINDOW_DAYS - 1);
		slotsLoading = true;
		slotsError = '';
		try {
			const query = new URLSearchParams({ product_id: product.id, from, to });
			const res = await fetch(`/api/partners/booking-slots?${query}`);
			if (!res.ok) {
				slotsError = 'Impossible de charger les créneaux. Veuillez réessayer.';
				return;
			}
			const body = await res.json();
			const fetched: FreeSlot[] = Array.isArray(body?.slots) ? body.slots : [];
			slots = reset ? fetched : [...slots, ...fetched];
			loadedDays = start + WINDOW_DAYS;
			const firstDay = fetched.length ? dayOf(fetched[0].start) : null;
			if (reset || selectedDay === null) {
				selectedDay = days.some(([day]) => day === selectedDay) ? selectedDay : firstDay;
			}
		} catch {
			slotsError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			slotsLoading = false;
		}
	}

	// ── Navigation between steps ─────────────────────────────────────

	function goTo(next: Step) {
		step = next;
		submitError = '';
		if (typeof window !== 'undefined') window.scrollTo({ top: 0 });
	}

	function chooseProduct(choice: BookableProduct) {
		if (product?.id !== choice.id) {
			product = choice;
			slot = null;
			slots = [];
			loadedDays = 0;
			selectedDay = null;
			loadSlots(true);
		}
		slotNotice = '';
		goTo('slot');
	}

	function chooseSlot(choice: FreeSlot) {
		slot = choice;
		slotNotice = '';
		goTo('client');
	}

	function back() {
		if (step === 'slot' && data.products.length > 1) goTo('product');
		else if (step === 'client') goTo('slot');
		else if (step === 'confirm') goTo('client');
	}

	// ── Booking ──────────────────────────────────────────────────────

	async function book() {
		if (!product || !slot || submitting) return;
		submitting = true;
		submitError = '';
		try {
			const email = client.email.trim();
			const res = await fetch('/api/partners/bookings', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					availability_id: slot.availability_id,
					product_id: product.id,
					slot_start_time: slot.start,
					guest_first_name: client.firstName.trim(),
					guest_last_name: client.lastName.trim(),
					guest_phone: normalizePhone(client.phone),
					...(email ? { guest_email: email } : {})
				})
			});
			if (res.status === 201) {
				const created = await res.json();
				await goto(`${data.from}?booked=${encodeURIComponent(created.id)}`);
				return;
			}
			const body = await res.json().catch(() => null);
			const message = typeof body?.error === 'string' ? body.error : '';
			if (res.status === 409 || (res.status === 400 && message.includes('slot_start_time'))) {
				// Taken meanwhile (an online booking, or another device): back
				// to the slots, fetched again.
				slot = null;
				slotNotice = 'Ce créneau vient d’être pris. Choisissez-en un autre.';
				goTo('slot');
				await loadSlots(true);
				return;
			}
			if (res.status === 400 && (message.includes('guest_phone') || message.includes('guest_email'))) {
				submitError = message.includes('guest_phone')
					? 'Le numéro de téléphone n’est pas accepté. Vérifiez-le.'
					: 'L’adresse email n’est pas acceptée. Vérifiez-la.';
				return;
			}
			if (res.status === 401) {
				submitError = 'Votre session a expiré. Reconnectez-vous puis réessayez.';
				return;
			}
			submitError = 'La réservation n’a pas pu être enregistrée. Veuillez réessayer.';
		} catch {
			submitError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Nouvelle réservation | Staff</title>
</svelte:head>

<div class="container mx-auto max-w-xl px-4 py-8 lg:py-12">
	<div class="mb-6">
		<a
			href={data.from}
			class="mb-4 inline-flex min-h-11 items-center gap-1 text-sm font-medium text-muted-foreground hover:text-foreground"
		>
			<ChevronLeft size={16} />
			{returnLabel}
		</a>
		<div class="flex items-center justify-between gap-4">
			<h1 class="text-3xl font-bold text-foreground lg:text-4xl">Nouvelle réservation</h1>
			<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
			<div class="lg:hidden">
				<PageActionsMenu />
			</div>
		</div>
		{#if !data.profileMissing && data.products.length > 0}
			<p class="mt-1 text-muted-foreground">
				Étape {stepIndex + 1} sur {STEPS.length} · {STEPS[stepIndex][1]}
			</p>
			<div class="mt-3 grid grid-cols-4 gap-1.5" aria-hidden="true">
				{#each STEPS as [s], i (s)}
					<div class="h-1 rounded-full {i <= stepIndex ? 'bg-foreground' : 'bg-muted'}"></div>
				{/each}
			</div>
		{/if}
	</div>

	{#if data.profileMissing}
		<PractitionerProfileMissing />
	{:else if data.loadFailed}
		<p class="rounded-xl border border-border p-4 text-sm text-muted-foreground">
			Impossible de charger vos soins. Veuillez réessayer.
		</p>
	{:else if data.products.length === 0}
		<p class="rounded-xl border border-border p-4 text-sm text-muted-foreground">
			Aucun soin publié à réserver. Publiez d’abord un soin dans le catalogue.
		</p>
	{:else if step === 'product'}
		<!-- 1. The product -->
		<h2 class="mb-3 text-lg font-semibold">Quel soin ?</h2>
		<ul class="space-y-2">
			{#each data.products as p (p.id)}
				<li>
					<button
						type="button"
						class="flex min-h-14 w-full items-center justify-between gap-3 rounded-xl border px-4 py-3 text-left transition-colors {product?.id === p.id
							? 'border-foreground bg-foreground text-background'
							: 'border-border hover:bg-muted'}"
						onclick={() => chooseProduct(p)}
					>
						<span class="min-w-0 font-medium break-words">{p.name}</span>
						<span class="shrink-0 text-sm tabular-nums {product?.id === p.id ? '' : 'text-muted-foreground'}">
							{durationLabel(p.durationMinutes)} · {formatEuros(p.priceCents)}
						</span>
					</button>
				</li>
			{/each}
		</ul>
	{:else if step === 'slot' && product}
		<!-- 2. The day and the free slot -->
		<div class="mb-4 flex min-h-11 items-center justify-between gap-3 rounded-xl border border-border px-4 py-2">
			<p class="min-w-0 text-sm">
				<span class="font-medium">{product.name}</span>
				<span class="text-muted-foreground"> · {durationLabel(product.durationMinutes)}</span>
			</p>
			{#if data.products.length > 1}
				<button
					type="button"
					class="inline-flex min-h-11 shrink-0 items-center text-sm font-medium underline underline-offset-4"
					onclick={() => goTo('product')}
				>
					Modifier
				</button>
			{/if}
		</div>

		{#if slotNotice}
			<p role="alert" class="mb-4 rounded-xl bg-foreground px-4 py-3 text-sm font-medium text-background">
				{slotNotice}
			</p>
		{/if}

		<h2 class="mb-3 text-lg font-semibold">Quel jour ?</h2>
		{#if slotsError && slots.length === 0}
			<div class="space-y-3">
				<p class="text-sm text-muted-foreground">{slotsError}</p>
				<button
					type="button"
					class="inline-flex h-11 items-center rounded-md border border-border px-4 text-sm font-medium hover:bg-muted"
					onclick={() => loadSlots(true)}
				>
					Réessayer
				</button>
			</div>
		{:else if (slotsLoading || loadedDays === 0) && slots.length === 0}
			<p class="py-6 text-sm text-muted-foreground">Recherche des créneaux libres…</p>
		{:else}
			{#if days.length === 0}
				<p class="mb-3 text-sm text-muted-foreground">
					Aucun créneau libre d’ici le {longDayLabel(searchedUntil).toLowerCase()}.
				</p>
			{:else}
				<!-- The days with a free slot: one row that scrolls sideways -->
				<div
					class="-mx-4 mb-5 flex gap-2 overflow-x-auto px-4 pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
					role="listbox"
					aria-label="Jours avec un créneau libre"
				>
					{#each days as [day, list] (day)}
						<button
							type="button"
							role="option"
							aria-selected={selectedDay === day}
							class="flex min-h-11 shrink-0 flex-col items-center justify-center rounded-xl border px-3 py-1.5 text-sm transition-colors {selectedDay === day
								? 'border-foreground bg-foreground text-background'
								: 'border-border hover:bg-muted'}"
							onclick={() => (selectedDay = day)}
						>
							<span class="font-medium whitespace-nowrap">{shortDayLabel(day)}</span>
							<span class="text-xs whitespace-nowrap {selectedDay === day ? 'opacity-80' : 'text-muted-foreground'}">
								{list.length} créneau{list.length > 1 ? 'x' : ''}
							</span>
						</button>
					{/each}
				</div>

				{#if selectedDay}
					<h2 class="mb-3 text-lg font-semibold">{longDayLabel(selectedDay)}</h2>
					<div class="grid grid-cols-3 gap-2 sm:grid-cols-4">
						{#each daySlots as s (s.start)}
							<button
								type="button"
								class="h-12 rounded-xl border text-base font-medium tabular-nums transition-colors {slot?.start === s.start
									? 'border-foreground bg-foreground text-background'
									: 'border-border hover:bg-muted'}"
								onclick={() => chooseSlot(s)}
							>
								{timeOf(s.start)}
							</button>
						{/each}
					</div>
				{/if}
			{/if}

			{#if slotsError}
				<p class="mt-4 text-sm text-muted-foreground">{slotsError}</p>
			{/if}
			{#if loadedDays < MAX_DAYS}
				<button
					type="button"
					class="mt-5 inline-flex h-11 w-full items-center justify-center rounded-md border border-border px-4 text-sm font-medium hover:bg-muted disabled:opacity-50 sm:w-auto"
					disabled={slotsLoading}
					onclick={() => loadSlots(false)}
				>
					{slotsLoading ? 'Recherche…' : 'Voir les jours suivants'}
				</button>
			{/if}
		{/if}

		{#if data.products.length > 1}
			<div class="mt-6">
				<button
					type="button"
					class="inline-flex h-12 w-full items-center justify-center rounded-md border border-border px-6 text-sm font-medium hover:bg-muted sm:w-auto"
					onclick={back}
				>
					Retour
				</button>
			</div>
		{/if}
	{:else if step === 'client' && product && slot}
		<!-- 3. The client -->
		<p class="mb-4 text-sm text-muted-foreground">
			{product.name} · {longDayLabel(dayOf(slot.start))} à {timeOf(slot.start)}
		</p>
		<h2 class="mb-3 text-lg font-semibold">Pour qui ?</h2>
		<ClientStep bind:client onNext={() => goTo('confirm')} onBack={back} />
	{:else if step === 'confirm' && product && slot}
		<!-- 4. Confirm -->
		<h2 class="mb-3 text-lg font-semibold">Vérifiez et réservez</h2>
		<dl class="divide-y divide-border rounded-2xl border border-border">
			<div class="flex items-start justify-between gap-4 px-4 py-3">
				<dt class="text-sm text-muted-foreground">Soin</dt>
				<dd class="min-w-0 text-right text-sm font-medium break-words">{product.name}</dd>
			</div>
			<div class="flex items-start justify-between gap-4 px-4 py-3">
				<dt class="text-sm text-muted-foreground">Quand</dt>
				<dd class="text-right text-sm font-medium">
					{longDayLabel(dayOf(slot.start))}<br />
					<span class="tabular-nums">{timeOf(slot.start)} – {timeOf(slot.end)}</span>
				</dd>
			</div>
			<div class="flex items-start justify-between gap-4 px-4 py-3">
				<dt class="text-sm text-muted-foreground">Client</dt>
				<dd class="min-w-0 text-right text-sm">
					<span class="font-medium break-words">{client.firstName.trim()} {client.lastName.trim()}</span><br />
					<span class="tabular-nums">{displayPhone(client.phone)}</span>
					{#if client.email.trim()}
						<br /><span class="break-all text-muted-foreground">{client.email.trim()}</span>
					{/if}
				</dd>
			</div>
			<div class="flex items-start justify-between gap-4 px-4 py-3">
				<dt class="text-sm text-muted-foreground">Paiement</dt>
				<dd class="text-right text-sm font-medium">À régler sur place · {formatEuros(product.priceCents)}</dd>
			</div>
		</dl>
		<p class="mt-3 text-sm text-muted-foreground">
			Confirmée tout de suite. Le client ne reçoit aucun message de l’application.
		</p>

		{#if submitError}
			<p role="alert" class="mt-4 text-sm text-destructive">{submitError}</p>
		{/if}

		<div class="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
			<button
				type="button"
				class="inline-flex h-12 items-center justify-center rounded-md border border-border px-6 text-sm font-medium hover:bg-muted disabled:opacity-50"
				disabled={submitting}
				onclick={back}
			>
				Retour
			</button>
			<button
				type="button"
				class="inline-flex h-12 items-center justify-center rounded-md bg-foreground px-6 text-sm font-semibold text-background hover:bg-foreground/90 disabled:opacity-50"
				disabled={submitting}
				onclick={book}
			>
				{submitting ? 'Réservation…' : 'Réserver'}
			</button>
		</div>
	{/if}
</div>
