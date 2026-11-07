<script lang="ts">
	import type { PageProps } from './$types';
	import type { Booking } from './+page.server';
	import { invalidateAll } from '$app/navigation';
	import { Dialog } from 'bits-ui';
	import { Clock, MapPin, Mail, Phone, CheckCircle, XCircle, AlertCircle, Ban } from '@lucide/svelte';

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

	function statusBadge(status: string): string {
		switch (status) {
			case 'confirmed':
				return 'bg-blue-100 text-blue-700';
			case 'completed':
				return 'bg-green-100 text-green-700';
			case 'no_show':
				return 'bg-orange-100 text-orange-700';
			case 'cancelled':
				return 'bg-red-100 text-red-700';
			default:
				return 'bg-gray-100 text-gray-700';
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

	function paymentStatusBadge(status: string): string {
		switch (status) {
			case 'paid':
				return 'bg-green-50 text-green-600 border-green-200';
			case 'pending':
				return 'bg-yellow-50 text-yellow-600 border-yellow-200';
			case 'failed':
				return 'bg-red-50 text-red-600 border-red-200';
			case 'refunded':
				return 'bg-purple-50 text-purple-600 border-purple-200';
			default:
				return 'bg-gray-50 text-gray-600 border-gray-200';
		}
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
			default:
				return status;
		}
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
</script>

<svelte:head>
	<title>Réservations | Staff</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
	<div class="mb-8">
		<h1 class="text-3xl lg:text-4xl font-bold mb-2 text-foreground">Réservations</h1>
		<p class="text-muted-foreground">Suivi de vos séances et clients</p>
	</div>

	<!-- Summary -->
	<div class="grid grid-cols-3 gap-4 mb-8">
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">À venir</p>
			<p class="text-2xl font-bold text-blue-600 dark:text-blue-400 tabular-nums">{upcomingCount}</p>
		</div>
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Terminées</p>
			<p class="text-2xl font-bold text-green-600 dark:text-green-400 tabular-nums">{completedCount}</p>
		</div>
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Revenus</p>
			<p class="text-2xl font-bold text-foreground tabular-nums">{formatCents(completedEarnings)}</p>
		</div>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1 mb-6 bg-muted p-1 rounded-lg w-fit flex-wrap">
		{#each tabs as [val, label]}
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-colors {activeTab === val
					? 'bg-card shadow-mini text-foreground'
					: 'text-muted-foreground hover:text-foreground'}"
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
				<div class="flex flex-col sm:flex-row sm:items-start gap-4">
					<!-- Client Avatar -->
					<div
						class="w-10 h-10 rounded-full bg-muted flex items-center justify-center flex-shrink-0"
					>
						<span class="text-sm font-semibold text-foreground">{booking.clientInitials}</span>
					</div>

					<!-- Main Info -->
					<div class="flex-1 min-w-0">
						<div class="flex flex-wrap items-center gap-2 mb-1">
							<p class="font-semibold text-foreground">{booking.clientName}</p>
							<span class="px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(booking.status)}">
								{statusLabel(booking.status)}
							</span>
							<span class="px-2 py-0.5 rounded-full text-xs font-medium border {paymentStatusBadge(booking.paymentStatus)}">
								{paymentStatusLabel(booking.paymentStatus)}
							</span>
						</div>
						<p class="text-sm text-muted-foreground mb-2">{booking.productName}</p>

						<div class="flex flex-wrap gap-4 text-sm text-muted-foreground">
							<span class="flex items-center gap-1.5">
								<Clock size={13} />
								{formatDateTime(booking.startTime)} – {formatTime(booking.endTime)}
							</span>
							<span class="flex items-center gap-1.5">
								<MapPin size={13} />
								{booking.roomName}
							</span>
						</div>

						<!-- Client contact details (no notes: Leviosa stores no health data, ADR-0019) -->
						{#if booking.clientEmail || booking.clientPhone}
							<div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
								{#if booking.clientEmail}
									<a
										href="mailto:{booking.clientEmail}"
										class="flex items-center gap-1.5 hover:text-foreground transition-colors"
									>
										<Mail size={13} />
										{booking.clientEmail}
									</a>
								{/if}
								{#if booking.clientPhone}
									<a
										href="tel:{booking.clientPhone}"
										class="flex items-center gap-1.5 hover:text-foreground transition-colors"
									>
										<Phone size={13} />
										{booking.clientPhone}
									</a>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Amount + Actions -->
					<div class="flex flex-col items-end gap-2 flex-shrink-0">
						<p class="font-semibold text-foreground">{formatCents(booking.amountInCents)}</p>
						{#if isCancellable(booking)}
							<button
								class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-red-50 text-red-700 border border-red-200 rounded-md hover:bg-red-100 transition-colors"
								onclick={() => openCancelDialog(booking)}
							>
								<Ban size={13} />
								Annuler la réservation
							</button>
						{/if}
						{#if isActionable(booking)}
							<div class="flex gap-2">
								<button
									class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors disabled:opacity-50"
									title="Marquer comme terminé"
									onclick={() => completeBooking(booking.id)}
									disabled={actionLoading[booking.id + '_complete']}
								>
									<CheckCircle size={13} />
									{actionLoading[booking.id + '_complete'] ? '…' : 'Terminer'}
								</button>
								<button
									class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-orange-100 text-orange-700 rounded-md hover:bg-orange-200 transition-colors disabled:opacity-50"
									title="Marquer absent"
									onclick={() => markNoShow(booking.id)}
									disabled={actionLoading[booking.id + '_noshow']}
								>
									<XCircle size={13} />
									{actionLoading[booking.id + '_noshow'] ? '…' : 'Absent'}
								</button>
							</div>
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
