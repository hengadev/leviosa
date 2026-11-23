<script lang="ts">
	import PageActionsMenu from "#lib/components/admin/PageActionsMenu.svelte";
	import type { PageProps } from './$types';
	import type { Booking } from './+page.server';
	import { invalidateAll } from '$app/navigation';
	import { Dialog } from 'bits-ui';
	import PractitionerProfileMissing from '#lib/components/admin/PractitionerProfileMissing.svelte';
	import { Clock, CheckCircle, XCircle, AlertCircle, Ban, RotateCcw } from '@lucide/svelte';
	import { changeOnSitePayment } from '#lib/onSitePayment.ts';

	let { data }: PageProps = $props();

	type Tab = 'upcoming' | 'completed' | 'no_show' | 'cancelled' | 'all';
	let activeTab = $state<Tab>((data.statusFilter as Tab) || 'upcoming');
	let actionLoading: Record<string, boolean> = $state({});
	let actionError = $state('');

	function formatDateTime(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	function formatTime(iso: string): string {
		return new Date(iso).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
	}

	function formatCents(cents: number): string {
		return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(
			cents / 100,
		);
	}

	/** Black and white only: a booking still to come stands out, a settled one recedes. */
	function statusBadge(status: string): string {
		switch (status) {
			case 'confirmed':
				return 'bg-foreground text-background';
			case 'cancelled':
				return 'bg-muted text-muted-foreground';
			default:
				return 'bg-muted text-foreground';
		}
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'confirmed':
				return 'À venir';
			case 'completed':
				return 'Terminé';
			case 'no_show':
				return 'Absent';
			case 'cancelled':
				return 'Annulé';
			default:
				return status;
		}
	}

	/** Only shown when the payment needs attention, so one outlined style is enough. */
	function paymentStatusBadge(status: string): string {
		return status === 'disputed' || status === 'failed'
			? 'border-foreground text-foreground'
			: 'border-border text-muted-foreground';
	}

	function paymentStatusLabel(status: string): string {
		switch (status) {
			case 'paid':
				return 'Payé';
			case 'pending':
				return 'En attente';
			case 'failed':
				return 'Échoué';
			case 'refunded':
				return 'Remboursé';
			case 'partially_refunded':
				return 'Remboursé en partie';
			case 'disputed':
				return 'Contesté';
			default:
				return status;
		}
	}

	/** A payment state other than the usual "paid" (or "pending" on a free booking) */
	function paymentNeedsAttention(status: string): boolean {
		return status !== 'paid' && status !== 'pending';
	}

	/** A confirmed booking is eligible for complete/no-show only if its end time has passed */
	function isActionable(booking: Booking): boolean {
		if (booking.status !== 'confirmed') return false;
		return new Date(booking.endTime) < new Date();
	}

	/** A confirmed booking that has not started yet can be cancelled by the practitioner */
	function isCancellable(booking: Booking): boolean {
		return booking.status === 'confirmed' && new Date(booking.startTime) > new Date();
	}

	/** A closed booking can be reopened only on the Paris day it was closed */
	function isReopenable(booking: Booking): boolean {
		const closedAt = booking.status === 'completed' ? booking.completedAt : booking.noShowAt;
		if (!closedAt) return false;
		if (booking.status !== 'completed' && booking.status !== 'no_show') return false;
		return parisDay(new Date(closedAt)) === parisDay(new Date());
	}

	function parisDay(date: Date): string {
		return new Intl.DateTimeFormat('en-CA', {
			timeZone: 'Europe/Paris',
			year: 'numeric',
			month: '2-digit',
			day: '2-digit'
		}).format(date);
	}

	// Practitioner cancellation: the booking awaiting confirmation in the dialog
	let cancelTarget = $state<Booking | null>(null);
	let cancelDialogOpen = $state(false);
	let cancelLoading = $state(false);
	let cancelError = $state('');

	function openCancelDialog(booking: Booking) {
		cancelTarget = booking;
		cancelError = '';
		cancelDialogOpen = true;
	}

	async function cancelBooking() {
		if (!cancelTarget) return;
		cancelLoading = true;
		cancelError = '';
		try {
			const res = await fetch(`/api/partners/bookings/${cancelTarget.id}/cancel`, { method: 'POST' });
			if (res.ok) {
				cancelDialogOpen = false;
				await invalidateAll();
			} else if (res.status === 403) {
				cancelError = "Cette réservation n'est pas la vôtre.";
			} else if (res.status === 400) {
				cancelError = 'Cette réservation ne peut plus être annulée.';
			} else {
				cancelError = "Impossible d'annuler la réservation. Veuillez réessayer.";
			}
		} catch {
			cancelError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			cancelLoading = false;
		}
	}

	const filtered = $derived(
		activeTab === 'all'
			? data.bookings
			: activeTab === 'upcoming'
				? data.bookings.filter((b) => b.status === 'confirmed')
				: data.bookings.filter((b) => b.status === activeTab),
	);

	const upcomingCount = $derived(data.bookings.filter((b) => b.status === 'confirmed').length);
	const completedCount = $derived(data.bookings.filter((b) => b.status === 'completed').length);
	const completedEarnings = $derived(
		data.bookings
			.filter((b) => b.status === 'completed')
			.reduce((s, b) => s + b.amountInCents, 0),
	);

	const tabs: [Tab, string][] = [
		['upcoming', 'À venir'],
		['completed', 'Historique'],
		['no_show', 'Absents'],
		['cancelled', 'Annulés'],
		['all', 'Tout'],
	];

	function switchTab(tab: Tab) {
		activeTab = tab;
		const url = new URL(window.location.href);
		if (tab === 'all') {
			url.searchParams.delete('status');
		} else {
			url.searchParams.set('status', tab);
		}
		window.history.replaceState({}, '', url.toString());
	}

	async function completeBooking(bookingId: string) {
		actionLoading = { ...actionLoading, [bookingId + '_complete']: true };
		actionError = '';
		try {
			const res = await fetch(`/api/bookings/${bookingId}/complete`, { method: 'POST' });
			if (res.ok) {
				await invalidateAll();
			} else {
				actionError = 'Impossible de terminer la réservation. Veuillez réessayer.';
			}
		} catch {
			actionError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			actionLoading = { ...actionLoading, [bookingId + '_complete']: false };
		}
	}

	async function markNoShow(bookingId: string) {
		actionLoading = { ...actionLoading, [bookingId + '_noshow']: true };
		actionError = '';
		try {
			const res = await fetch(`/api/bookings/${bookingId}/no-show`, { method: 'POST' });
			if (res.ok) {
				await invalidateAll();
			} else {
				actionError = 'Impossible de marquer comme absent. Veuillez réessayer.';
			}
		} catch {
			actionError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			actionLoading = { ...actionLoading, [bookingId + '_noshow']: false };
		}
	}
	/** "Reste payé · N €": a started session with money due at the practice */
	function canRecordOnSite(booking: Booking): boolean {
		return booking.onSite.recordable && new Date(booking.startTime) <= new Date();
	}

	/** Records ("Reste payé") or undoes the on-site payment */
	async function runOnSitePayment(bookingId: string, action: 'record' | 'undo') {
		const key = bookingId + '_onsite';
		actionLoading = { ...actionLoading, [key]: true };
		actionError = '';
		try {
			const message = await changeOnSitePayment(bookingId, action);
			if (message) actionError = message;
			await invalidateAll();
		} finally {
			actionLoading = { ...actionLoading, [key]: false };
		}
	}

	async function reopenBooking(bookingId: string) {
		actionLoading = { ...actionLoading, [bookingId + '_reopen']: true };
		actionError = '';
		try {
			const res = await fetch(`/api/bookings/${bookingId}/reopen`, { method: 'POST' });
			if (res.ok) {
				await invalidateAll();
			} else {
				actionError = 'Impossible d’annuler la clôture. Veuillez réessayer.';
			}
		} catch {
			actionError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			actionLoading = { ...actionLoading, [bookingId + '_reopen']: false };
		}
	}
</script>

<svelte:head>
	<title>Réservations | Staff</title>
</svelte:head>

<!-- An administrator without a practitioner profile: nothing to show yet -->
{#if data.profileMissing}
	<div class="container mx-auto px-4 py-8 lg:py-12">
		<div class="mb-8 flex items-center justify-between gap-4">
			<h1 class="text-3xl lg:text-4xl font-bold text-foreground">Réservations</h1>
			<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
			<div class="lg:hidden">
				<PageActionsMenu />
			</div>
		</div>
		<PractitionerProfileMissing />
	</div>
{:else}
<div class="container mx-auto px-4 py-8 lg:py-12">
	<div class="mb-8">
		<div class="mb-2 flex items-center justify-between gap-4">
			<h1 class="text-3xl lg:text-4xl font-bold text-foreground">Réservations</h1>
			<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
			<div class="lg:hidden">
				<PageActionsMenu />
			</div>
		</div>
		<p class="text-muted-foreground">Suivi de vos séances et clients</p>
	</div>

	<!-- Summary -->
	<div class="grid grid-cols-3 gap-2 sm:gap-4 mb-8">
		<div class="bg-background min-w-0 rounded-2xl border border-border p-3 sm:p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2 truncate">À venir</p>
			<p class="text-lg sm:text-2xl font-bold text-foreground tabular-nums whitespace-nowrap">{upcomingCount}</p>
		</div>
		<div class="bg-background min-w-0 rounded-2xl border border-border p-3 sm:p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2 truncate">Terminées</p>
			<p class="text-lg sm:text-2xl font-bold text-foreground tabular-nums whitespace-nowrap">{completedCount}</p>
		</div>
		<div class="bg-background min-w-0 rounded-2xl border border-border p-3 sm:p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2 truncate">Revenus</p>
			<p class="text-lg sm:text-2xl font-bold text-foreground tabular-nums whitespace-nowrap">{formatCents(completedEarnings)}</p>
		</div>
	</div>

	<!-- Tabs -->
	<!-- One row at every width: an underlined tab bar that scrolls sideways rather than wrapping -->
	<div
		role="tablist"
		class="mb-6 -mx-4 px-4 sm:mx-0 sm:px-0 flex gap-5 sm:gap-6 overflow-x-auto border-b border-border [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
	>
		{#each tabs as [val, label]}
			<button
				role="tab"
				aria-selected={activeTab === val}
				class="-mb-px shrink-0 whitespace-nowrap border-b-2 py-2.5 text-sm font-medium transition-colors {activeTab === val
					? 'border-foreground text-foreground'
					: 'border-transparent text-muted-foreground hover:text-foreground'}"
				onclick={() => switchTab(val)}
			>
				{label}
			</button>
		{/each}
	</div>

	<!-- Action error banner -->
	{#if actionError}
		<div class="flex items-center gap-2 mb-4 px-4 py-3 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm">
			<AlertCircle size={16} class="flex-shrink-0" />
			{actionError}
		</div>
	{/if}

	<!-- Bookings List -->
	<div class="space-y-3">
		{#each filtered as booking (booking.id)}
			<div class="bg-background rounded-xl border border-border p-4 sm:p-5 hover:shadow-mini transition-shadow">
				<div class="flex flex-col sm:flex-row sm:items-start gap-3 sm:gap-4">
					<!-- Main Info -->
					<div class="flex-1 min-w-0">
						<div class="flex items-start justify-between gap-3">
							<a href="/staff/clients/{booking.id}" class="block font-semibold text-foreground truncate hover:underline">{booking.clientName}</a>
							<!-- On the phone the amount sits beside the name; from sm up it heads the actions column -->
							<p class="font-semibold text-foreground tabular-nums sm:hidden">{formatCents(booking.amountInCents)}</p>
						</div>
						<p class="text-sm text-muted-foreground mb-2">{booking.productName}</p>

						<div class="flex flex-wrap items-center gap-x-3 gap-y-2 text-sm text-muted-foreground">
							<span class="flex items-center gap-1.5">
								<Clock size={13} />
								{formatDateTime(booking.startTime)} – {formatTime(booking.endTime)}
							</span>
							<span class="px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(booking.status)}">
								{statusLabel(booking.status)}
							</span>
							<!-- Paying is what confirms a booking, so "paid" is the normal case: only a refund, a dispute or a failure is worth a pill -->
							{#if paymentNeedsAttention(booking.paymentStatus)}
								<span class="px-2 py-0.5 rounded-full text-xs font-medium border {paymentStatusBadge(booking.paymentStatus)}">
									{paymentStatusLabel(booking.paymentStatus)}
								</span>
							{:else if booking.status !== 'cancelled' && booking.onSite.label}
								<!-- What was paid, and what is left to pay at the practice -->
								<span class="text-xs">{booking.onSite.label}</span>
							{/if}
						</div>
					</div>

					<!-- Amount + Actions -->
					<div class="flex flex-col items-stretch sm:items-end gap-2 flex-shrink-0">
						<p class="hidden sm:block font-semibold text-foreground tabular-nums">{formatCents(booking.amountInCents)}</p>
						{#if isCancellable(booking)}
							<button
								class="inline-flex flex-1 sm:flex-none items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-red-50 text-red-700 border border-red-200 rounded-md hover:bg-red-100 transition-colors"
								onclick={() => openCancelDialog(booking)}
							>
								<Ban size={13} />
								Annuler la réservation
							</button>
						{/if}
						{#if isActionable(booking)}
							<div class="flex gap-2">
								<button
									class="inline-flex flex-1 sm:flex-none items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-foreground text-background rounded-md hover:bg-foreground/90 transition-colors disabled:opacity-50"
									title="Marquer comme terminé"
									onclick={() => completeBooking(booking.id)}
									disabled={actionLoading[booking.id + '_complete']}
								>
									<CheckCircle size={13} />
									{actionLoading[booking.id + '_complete'] ? '…' : 'Terminer'}
								</button>
								<button
									class="inline-flex flex-1 sm:flex-none items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium border border-border text-foreground rounded-md hover:bg-muted transition-colors disabled:opacity-50"
									title="Marquer absent"
									onclick={() => markNoShow(booking.id)}
									disabled={actionLoading[booking.id + '_noshow']}
								>
									<XCircle size={13} />
									{actionLoading[booking.id + '_noshow'] ? '…' : 'Absent'}
								</button>
							</div>
						{/if}
						{#if canRecordOnSite(booking)}
							<button
								class="inline-flex h-11 flex-1 sm:flex-none items-center justify-center gap-1.5 px-4 text-sm font-medium bg-foreground text-background rounded-md hover:bg-foreground/90 transition-colors disabled:opacity-50"
								onclick={() => runOnSitePayment(booking.id, 'record')}
								disabled={actionLoading[booking.id + '_onsite']}
							>
								{actionLoading[booking.id + '_onsite'] ? '…' : `Reste payé · ${booking.onSite.dueLabel}`}
							</button>
						{:else if booking.onSite.recorded && booking.onSite.undoable}
							<button
								class="inline-flex h-11 flex-1 sm:flex-none items-center justify-center gap-1.5 px-4 text-sm font-medium text-muted-foreground bg-background border border-border rounded-md hover:bg-muted transition-colors disabled:opacity-50"
								aria-label="Annuler le paiement sur place"
								title="Annuler le paiement sur place"
								onclick={() => runOnSitePayment(booking.id, 'undo')}
								disabled={actionLoading[booking.id + '_onsite']}
							>
								<RotateCcw size={13} />
								{actionLoading[booking.id + '_onsite'] ? '…' : 'Annuler'}
							</button>
						{/if}
						{#if isReopenable(booking)}
							<button
								class="inline-flex flex-1 sm:flex-none items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium text-muted-foreground bg-background border border-border rounded-md hover:bg-muted transition-colors disabled:opacity-50"
								title="Remettre la séance en confirmée"
								onclick={() => reopenBooking(booking.id)}
								disabled={actionLoading[booking.id + '_reopen']}
							>
								<RotateCcw size={13} />
								{actionLoading[booking.id + '_reopen'] ? '…' : 'Annuler la clôture'}
							</button>
						{/if}
					</div>
				</div>
			</div>
		{:else}
			<div class="text-center py-12 text-muted-foreground bg-background rounded-xl border border-border">
				Aucune réservation trouvée
			</div>
		{/each}
	</div>
</div>
{/if}

<Dialog.Root bind:open={cancelDialogOpen}>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
		/>
		<Dialog.Content
			class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-6 sm:p-8 sm:max-w-[440px]"
		>
			<Dialog.Title class="w-full text-xl font-semibold tracking-tight">Annuler la réservation</Dialog.Title>
			{#if cancelTarget}
				<Dialog.Description class="text-muted-foreground mt-2 text-sm">
					{cancelTarget.clientName}, {cancelTarget.productName}, {formatDateTime(cancelTarget.startTime)}.
				</Dialog.Description>
				<p class="mt-4 text-sm text-foreground">
					{#if cancelTarget.practitionerRefundCents > 0}
						Le client sera intégralement remboursé ({formatCents(cancelTarget.practitionerRefundCents)})
						et prévenu par email.
					{:else}
						Aucun paiement en ligne à rembourser. Le client sera prévenu par email.
					{/if}
					Le créneau sera de nouveau proposé à la réservation.
				</p>
			{/if}
			{#if cancelError}
				<p class="mt-4 text-sm text-destructive">{cancelError}</p>
			{/if}
			<div class="mt-8 flex w-full flex-col-reverse sm:flex-row sm:justify-end gap-3">
				<button
					type="button"
					onclick={() => (cancelDialogOpen = false)}
					disabled={cancelLoading}
					class="h-input rounded-input border border-border-input hover:bg-dark-04 inline-flex items-center justify-center px-6 text-sm font-medium active:scale-[0.98] cursor-pointer transition-all disabled:opacity-50"
				>
					Garder la réservation
				</button>
				<button
					type="button"
					onclick={cancelBooking}
					disabled={cancelLoading}
					class="h-input rounded-input bg-destructive text-white shadow-mini hover:bg-destructive/90 inline-flex items-center justify-center px-6 text-sm font-semibold active:scale-[0.98] cursor-pointer transition-all disabled:opacity-50"
				>
					{cancelLoading ? 'Annulation…' : "Confirmer l'annulation"}
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
