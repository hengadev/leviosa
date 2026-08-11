<script lang="ts">
	import type { PageProps } from './$types';
	import { invalidateAll } from '$app/navigation';
	import {
		Clock,
		MapPin,
		FileText,
		Save,
		X,
		Pencil,
		AlertCircle,
		XCircle,
		ArrowLeft,
		MessageSquare,
		CalendarClock,
		Undo2,
		Receipt,
		ExternalLink,
	} from '@lucide/svelte';

	let { data }: PageProps = $props();
	let booking = $derived(data.booking);

	const now = new Date();
	let cancelling = $state(false);
	let cancelError = $state('');

	let editingNotes = $state(false);
	let notesValue = $state(booking.client_notes ?? '');
	let notesOverride = $state(booking.client_notes ?? '');
	let savingNotes = $state(false);
	let notesError = $state('');

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric',
		});
	}

	function formatTime(iso: string): string {
		return new Date(iso).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
	}

	function formatCents(cents: number): string {
		return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100);
	}

	function formatDateTime(iso: string): string {
		return new Date(iso).toLocaleString('fr-FR', {
			day: 'numeric',
			month: 'long',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	// Plain-language cancellation-deadline copy, derived from the booking's
	// cancellation_deadline (product's cancellation window applied to the
	// slot). Shown ahead of the cancel action so the client understands the
	// consequence before acting, not after (issue 011).
	let cancellationNotice = $derived.by(() => {
		if (!booking.cancellation_deadline) return null;
		const deadline = new Date(booking.cancellation_deadline);
		const feeApplies = now >= deadline;
		return {
			feeApplies,
			text: feeApplies
				? `Des frais d'annulation s'appliquent depuis le ${formatDateTime(booking.cancellation_deadline)}.`
				: `Annulation gratuite jusqu'au ${formatDateTime(booking.cancellation_deadline)}.`,
		};
	});

	// Plain-language refund status. Written to read a generic amount/date pair
	// rather than assuming "full refund or nothing", so it renders correctly
	// once a graduated refund policy populates a partial amount here too.
	let refundNotice = $derived.by(() => {
		if (!booking.refund_amount_cents) return null;
		const amount = formatCents(booking.refund_amount_cents);
		return booking.refunded_at
			? `Remboursé de ${amount} le ${formatDate(booking.refunded_at)}.`
			: `Remboursé de ${amount}.`;
	});

	function statusBadge(status: string): string {
		switch (status) {
			case 'confirmed': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400';
			case 'completed': return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400';
			case 'cancelled': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400';
			case 'no_show': return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400';
			default: return 'bg-muted text-muted-foreground';
		}
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'confirmed': return 'À venir';
			case 'completed': return 'Terminé';
			case 'cancelled': return 'Annulé';
			case 'no_show': return 'Absent';
			default: return status;
		}
	}

	function paymentStatusLabel(status: string): string {
		switch (status) {
			case 'paid': return 'Payé';
			case 'pending': return 'En attente';
			case 'failed': return 'Échoué';
			case 'refunded': return 'Remboursé';
			default: return status;
		}
	}

	function isCancellable(): boolean {
		return booking.status === 'confirmed' && new Date(booking.slot_start_time) > now;
	}

	async function cancelBooking() {
		cancelling = true;
		cancelError = '';
		try {
			const res = await fetch(`/api/bookings/${booking.id}/cancel`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ reason: 'Annulé par le client' }),
			});
			if (res.ok) {
				await invalidateAll();
			} else {
				cancelError = 'Impossible d\'annuler. Veuillez réessayer.';
			}
		} catch {
			cancelError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			cancelling = false;
		}
	}

	function startEditNotes() {
		editingNotes = true;
		notesValue = notesOverride;
		notesError = '';
	}

	function cancelEditNotes() {
		editingNotes = false;
		notesValue = notesOverride;
		notesError = '';
	}

	async function saveNotes() {
		savingNotes = true;
		notesError = '';
		try {
			const res = await fetch(`/api/bookings/${booking.id}/notes`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ client_notes: notesValue }),
			});
			if (res.ok) {
				notesOverride = notesValue;
				editingNotes = false;
			} else {
				notesError = 'Impossible de sauvegarder. Veuillez réessayer.';
			}
		} catch {
			notesError = 'Erreur réseau. Veuillez réessayer.';
		} finally {
			savingNotes = false;
		}
	}
</script>

<svelte:head>
	<title>Réservation {booking.id.slice(0, 8)}… | Leviosa</title>
</svelte:head>

<div class="space-y-6">
	<!-- Back link -->
	<a
		href="/client/bookings"
		class="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
	>
		<ArrowLeft size={16} />
		Retour aux réservations
	</a>

	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<h1 class="text-2xl font-bold text-foreground">Détail de la réservation</h1>
			<p class="text-sm text-muted-foreground mt-0.5">Réf. {booking.id}</p>
		</div>
		<div class="flex items-center gap-2">
			<span class="px-2.5 py-0.5 rounded-full text-xs font-medium {statusBadge(booking.status)}">
				{statusLabel(booking.status)}
			</span>
			<span class="px-2 py-0.5 rounded-full text-xs border {booking.payment_status === 'paid'
				? 'bg-green-50 text-green-600 border-green-200 dark:bg-green-900/20 dark:text-green-400 dark:border-green-800'
				: 'bg-muted text-muted-foreground border-border'}">
				{paymentStatusLabel(booking.payment_status)}
			</span>
		</div>
	</div>

	<!-- Detail card -->
	<div class="bg-card rounded-lg border border-border p-5 sm:p-6 space-y-5">
		<!-- Date/Time -->
		<div class="flex items-start gap-3">
			<Clock size={18} class="text-muted-foreground mt-0.5 flex-shrink-0" />
			<div>
				<p class="text-sm font-medium text-muted-foreground">Date et heure</p>
				<p class="text-foreground">{formatDate(booking.slot_start_time)}</p>
				<p class="text-sm text-muted-foreground">{formatTime(booking.slot_start_time)} – {formatTime(booking.slot_end_time)}</p>
			</div>
		</div>

		<!-- Price -->
		{#if booking.total_price_cents}
			<div class="flex items-start gap-3">
				<div class="w-[18px] mt-0.5 flex-shrink-0 text-center text-muted-foreground text-sm font-bold">€</div>
				<div>
					<p class="text-sm font-medium text-muted-foreground">Prix</p>
					<p class="text-foreground font-medium">{formatCents(booking.total_price_cents)}</p>
				</div>
			</div>
		{/if}

		<!-- Receipt -->
		{#if booking.payment_status === 'paid' && booking.receipt_url}
			<div class="flex items-start gap-3">
				<Receipt size={18} class="text-muted-foreground mt-0.5 flex-shrink-0" />
				<div>
					<p class="text-sm font-medium text-muted-foreground">Reçu</p>
					<a
						href={booking.receipt_url}
						target="_blank"
						rel="noopener noreferrer"
						class="inline-flex items-center gap-1 text-sm font-medium text-foreground hover:underline"
					>
						Voir le reçu
						<ExternalLink size={13} />
					</a>
				</div>
			</div>
		{/if}

		<!-- Cancellation info -->
		{#if booking.status === 'cancelled' && booking.cancelled_at}
			<div class="flex items-start gap-3">
				<AlertCircle size={18} class="text-red-500 mt-0.5 flex-shrink-0" />
				<div>
					<p class="text-sm font-medium text-muted-foreground">Annulée le {new Date(booking.cancelled_at).toLocaleDateString('fr-FR')}</p>
					{#if booking.cancellation_reason}
						<p class="text-sm text-muted-foreground">{booking.cancellation_reason}</p>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Refund status -->
		{#if refundNotice}
			<div class="flex items-start gap-3">
				<Undo2 size={18} class="text-sky-600 dark:text-sky-400 mt-0.5 flex-shrink-0" />
				<div>
					<p class="text-sm font-medium text-muted-foreground">Remboursement</p>
					<p class="text-foreground text-sm">{refundNotice}</p>
				</div>
			</div>
		{/if}

		<!-- Completed info -->
		{#if booking.status === 'completed' && booking.completed_at}
			<div class="flex items-start gap-3">
				<div class="w-[18px] mt-0.5 flex-shrink-0 text-center text-green-600 text-sm">✓</div>
				<div>
					<p class="text-sm font-medium text-muted-foreground">Terminée le {new Date(booking.completed_at).toLocaleDateString('fr-FR')}</p>
				</div>
			</div>
		{/if}

		<!-- Partner notes (read-only) -->
		{#if booking.partner_notes}
			<div class="flex items-start gap-3">
				<FileText size={18} class="text-muted-foreground mt-0.5 flex-shrink-0" />
				<div>
					<p class="text-sm font-medium text-muted-foreground">Notes du praticien</p>
					<p class="text-foreground text-sm">{booking.partner_notes}</p>
				</div>
			</div>
		{/if}

		<!-- Client notes (editable) -->
		<div class="flex items-start gap-3">
			<FileText size={18} class="text-muted-foreground mt-0.5 flex-shrink-0" />
			<div class="flex-1">
				<p class="text-sm font-medium text-muted-foreground mb-2">Vos notes</p>
				{#if editingNotes}
					<div class="flex flex-col gap-2">
						<textarea
							class="w-full rounded-md border border-border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-none"
							rows="3"
							placeholder="Ajoutez une note pour le praticien…"
							bind:value={notesValue}
						></textarea>
						{#if notesError}
							<p class="text-xs text-red-600">{notesError}</p>
						{/if}
						<div class="flex gap-2">
							<button
								class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-foreground text-background rounded-md hover:bg-foreground/90 transition-colors disabled:opacity-50"
								onclick={saveNotes}
								disabled={savingNotes}
							>
								<Save size={13} />
								{savingNotes ? 'Enregistrement…' : 'Enregistrer'}
							</button>
							<button
								class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium bg-muted text-muted-foreground rounded-md hover:bg-muted/80 transition-colors"
								onclick={cancelEditNotes}
							>
								<X size={13} />
								Annuler
							</button>
						</div>
					</div>
				{:else}
					<div class="flex items-start gap-2">
						{#if notesOverride}
							<p class="text-foreground text-sm flex-1">{notesOverride}</p>
						{:else}
							<p class="text-muted-foreground text-sm flex-1 italic">Aucune note</p>
						{/if}
						<button
							class="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
							onclick={startEditNotes}
						>
							<Pencil size={13} />
							<span class="text-xs">{notesOverride ? 'Modifier' : 'Ajouter'}</span>
						</button>
					</div>
				{/if}
			</div>
		</div>

		<!-- Messaging shortcut (placeholder for Issue 008) -->
		{#if booking.status === 'confirmed' || booking.status === 'completed'}
			<div class="flex items-start gap-3">
				<MessageSquare size={18} class="text-muted-foreground mt-0.5 flex-shrink-0" />
				<div>
					<a
						href="/client/messages?partner={booking.partner_id}"
						class="text-sm font-medium text-foreground hover:underline"
					>
						Contacter le praticien
					</a>
				</div>
			</div>
		{/if}
	</div>

	<!-- Cancel action -->
	{#if isCancellable()}
		{#if cancellationNotice}
			<div
				class="flex items-center gap-2 px-4 py-3 rounded-lg text-sm border {cancellationNotice.feeApplies
					? 'bg-amber-50 dark:bg-amber-950/40 border-amber-200 dark:border-amber-800 text-amber-800 dark:text-amber-400'
					: 'bg-muted border-border text-muted-foreground'}"
			>
				<CalendarClock size={16} class="flex-shrink-0" />
				{cancellationNotice.text}
			</div>
		{/if}
		{#if cancelError}
			<div class="flex items-center gap-2 px-4 py-3 rounded-lg bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm">
				<AlertCircle size={16} class="flex-shrink-0" />
				{cancelError}
			</div>
		{/if}
		<div class="pt-2">
			<button
				class="inline-flex items-center gap-1.5 px-4 py-2 text-sm font-medium text-red-600 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100 transition-colors disabled:opacity-50"
				onclick={cancelBooking}
				disabled={cancelling}
			>
				<XCircle size={16} />
				{cancelling ? 'Annulation en cours…' : 'Annuler cette réservation'}
			</button>
		</div>
	{/if}
</div>
