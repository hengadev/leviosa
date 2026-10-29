<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	import type { PublicBooking } from './+page.server';
	import { cap, dictionary, fmt, type Locale } from '$lib/i18n/practice';
	import { Mail, Clock, CheckCircle, XCircle, AlertCircle, Search, Phone as PhoneIcon, Loader2 } from '@lucide/svelte';

	let { data, form }: PageProps = $props();

	// Contact method toggle for the manual lookup form
	let contactMethod: 'email' | 'phone' = $state('email');

	// Cancel confirmation state
	let showCancelConfirm = $state(false);
	let cancelling = $state(false);

	// Booking detail (from either token path or manual form action). After a
	// cancellation the page data is reloaded; the action's own result is used
	// meanwhile.
	let looked = $derived<PublicBooking | null>(
		data.booking ?? (form?.action === 'lookup' && form.success ? (form.booking ?? null) : null)
	);
	let justCancelled = $derived(form?.action === 'cancel' && form.success === true);
	let booking = $derived<PublicBooking | null>(
		justCancelled && form?.action === 'cancel' && form.booking ? { ...looked, ...form.booking } : looked
	);
	let displayError = $derived(data.lookupError ?? (form?.action === 'lookup' && !form.success ? form.error : null));
	let cancelFailure = $derived(form?.action === 'cancel' && !form.success ? form.failure : null);

	// The booking page speaks the language the client booked in.
	let locale = $derived<Locale>(booking?.locale === 'en' ? 'en' : 'fr');
	let t = $derived(dictionary[locale].booking);

	// Cancel requires a booking token — only available on the token URL path
	let canCancel = $derived(
		!!data.token &&
			!!booking &&
			booking.status === 'confirmed' &&
			!!booking.refund_preview &&
			new Date(booking.slot_start_time) > new Date()
	);

	let paid = $derived(booking?.amount_paid_online_cents ?? 0);
	let balance = $derived(booking ? Math.max(booking.total_price_cents - paid, 0) : 0);
	let preview = $derived(booking?.refund_preview ?? null);
	let refunded = $derived(booking?.refunded_amount_cents ?? null);

	const price = (cents: number) => fmt.price(cents, locale);

	function previewText(): string {
		if (!preview) return '';
		if (paid <= 0) return t.previewNothingPaid;
		if (preview.amount_cents <= 0) return t.previewNone(price(paid));
		if (preview.amount_cents >= paid) return t.previewFull(price(preview.amount_cents));
		return t.previewPartial(price(preview.amount_cents), preview.percentage, price(paid));
	}

	function dateLabel(iso: string): string {
		const date = new Date(iso);
		const year = new Intl.DateTimeFormat(locale === 'fr' ? 'fr-FR' : 'en-GB', {
			year: 'numeric',
			timeZone: 'Europe/Paris'
		}).format(date);
		return `${cap(fmt.long(date, locale))} ${year}`;
	}

	function statusColor(status: string): string {
		switch (status) {
			case 'confirmed':
				return 'text-sky-900 bg-sky-100';
			case 'completed':
				return 'text-green-800 bg-green-100';
			case 'cancelled':
			case 'expired':
				return 'text-red-800 bg-red-100';
			default:
				return 'text-neutral-800 bg-neutral-100';
		}
	}
</script>

<svelte:head>
	<title>{booking ? t.pageTitle : 'Mes réservations'} | Leviosa Care</title>
	<meta name="description" content="Retrouvez et gérez votre réservation Leviosa Care : détails, statut et annulation." />
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="min-h-screen bg-neutral-50 px-4 py-24 text-neutral-950 md:py-32 lg:px-8">
	<div class="mx-auto max-w-xl" lang={booking ? locale : 'fr'}>
		{#if booking}
			<!-- ═══ Booking detail view ═══ -->
			<h1 class="mb-6 text-3xl font-bold tracking-tight md:text-4xl">{t.heading}</h1>

			<div class="rounded-2xl border border-neutral-200 bg-white p-5 shadow-[0_1px_2px_rgb(0_0_0/0.04)] sm:p-8">
				<span
					class="mb-5 inline-flex items-center gap-1.5 rounded-full px-3 py-1 text-sm font-medium {statusColor(booking.status)}"
				>
					{#if booking.status === 'confirmed' || booking.status === 'completed'}
						<CheckCircle size={16} aria-hidden="true" />
					{:else if booking.status === 'cancelled' || booking.status === 'expired'}
						<XCircle size={16} aria-hidden="true" />
					{:else}
						<Clock size={16} aria-hidden="true" />
					{/if}
					{t.status[booking.status] ?? booking.status}
				</span>

				<dl class="divide-y divide-neutral-200">
					{#if booking.product_name}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.service}</dt>
							<dd class="font-semibold sm:text-right">{booking.product_name}</dd>
						</div>
					{/if}
					{#if booking.partner_name}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.practitioner}</dt>
							<dd class="font-semibold sm:text-right">{booking.partner_name}</dd>
						</div>
					{/if}
					<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
						<dt class="text-neutral-600">{t.date}</dt>
						<dd class="font-semibold sm:text-right">{dateLabel(booking.slot_start_time)}</dd>
					</div>
					<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
						<dt class="text-neutral-600">{t.time}</dt>
						<dd class="font-semibold sm:text-right">
							{fmt.time(new Date(booking.slot_start_time), locale)} – {fmt.time(new Date(booking.slot_end_time), locale)}
							<span class="font-normal text-neutral-600">({t.timeZoneNote})</span>
						</dd>
					</div>
					{#if booking.total_price_cents}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.totalPrice}</dt>
							<dd class="font-semibold sm:text-right">{price(booking.total_price_cents)}</dd>
						</div>
					{/if}
					{#if paid > 0}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.paidOnline}</dt>
							<dd class="font-semibold sm:text-right">{price(paid)}</dd>
						</div>
					{/if}
					{#if balance > 0 && booking.status === 'confirmed'}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.balanceDue}</dt>
							<dd class="font-semibold sm:text-right">{price(balance)}</dd>
						</div>
					{/if}
					{#if booking.status === 'cancelled' && refunded !== null}
						<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
							<dt class="text-neutral-600">{t.refunded}</dt>
							<dd class="font-semibold sm:text-right">{price(refunded)}</dd>
						</div>
					{/if}
					<div class="flex flex-col gap-0.5 py-3 sm:flex-row sm:justify-between sm:gap-4">
						<dt class="text-neutral-600">{t.reference}</dt>
						<dd class="font-mono text-sm break-all text-neutral-600 sm:text-right">{booking.id}</dd>
					</div>
				</dl>
			</div>

			{#if justCancelled}
				<!-- ═══ Cancellation result ═══ -->
				<div class="mt-6 rounded-2xl border border-neutral-200 bg-white p-5 sm:p-6" role="status">
					<h2 class="text-lg font-semibold">{t.cancelledHeading}</h2>
					<p class="mt-2 text-pretty text-neutral-600">
						{refunded && refunded > 0 ? t.cancelledRefund(price(refunded)) : t.cancelledNoRefund}
					</p>
					<p class="mt-2 text-neutral-600">{t.cancelledEmail}</p>
				</div>
			{/if}

			{#if cancelFailure}
				<div
					class="mt-6 flex items-start gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
					role="alert"
				>
					<AlertCircle size={16} class="mt-0.5 shrink-0" aria-hidden="true" />
					{cancelFailure === 'not_cancellable' ? t.notCancellable : t.cancelFailed}
				</div>
			{/if}

			{#if canCancel && preview}
				<!-- ═══ Refund preview and cancellation ═══ -->
				<section class="mt-6 rounded-2xl border border-neutral-200 bg-white p-5 sm:p-6" aria-labelledby="refund-preview">
					<h2 id="refund-preview" class="text-lg font-semibold">{t.previewHeading}</h2>
					<p class="mt-2 text-pretty">{previewText()}</p>
					{#if preview.min_hours_before !== undefined && paid > 0}
						<p class="mt-1 text-sm text-neutral-600">{t.previewTier(preview.min_hours_before)}</p>
					{/if}

					{#if showCancelConfirm}
						<div class="mt-5 border-t border-neutral-200 pt-5">
							<p class="font-semibold">{t.confirmHeading}</p>
							<p class="mt-1 text-pretty text-neutral-600">
								{preview.amount_cents > 0 ? t.confirmRefund(price(preview.amount_cents)) : t.confirmNoRefund}
							</p>

							<form
								method="POST"
								action="?/cancel"
								class="mt-4 flex flex-col gap-3 sm:flex-row"
								use:enhance={() => {
									cancelling = true;
									return async ({ update }) => {
										await update({ reset: false });
										cancelling = false;
										showCancelConfirm = false;
									};
								}}
							>
								<input type="hidden" name="booking_id" value={booking.id} />
								<input type="hidden" name="token" value={data.token ?? ''} />

								<button
									type="submit"
									disabled={cancelling}
									class="inline-flex min-h-11 flex-1 items-center justify-center gap-2 rounded-xl bg-red-700 px-5 py-2.5 font-medium text-white transition-colors hover:bg-red-800 focus-visible:ring-2 focus-visible:ring-red-700 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-60"
								>
									{#if cancelling}
										<Loader2 size={18} class="animate-spin" aria-hidden="true" />
										{t.cancelling}
									{:else}
										{t.confirm}
									{/if}
								</button>
								<button
									type="button"
									disabled={cancelling}
									class="inline-flex min-h-11 items-center justify-center rounded-xl border border-neutral-300 bg-white px-5 py-2.5 font-medium transition-colors hover:bg-neutral-100 focus-visible:ring-2 focus-visible:ring-neutral-950 focus-visible:ring-offset-2 focus-visible:outline-none disabled:opacity-60"
									onclick={() => {
										showCancelConfirm = false;
									}}
								>
									{t.keep}
								</button>
							</form>
						</div>
					{:else}
						<button
							type="button"
							class="mt-5 inline-flex min-h-11 w-full items-center justify-center rounded-xl border border-red-200 bg-red-50 px-5 py-2.5 font-medium text-red-800 transition-colors hover:bg-red-100 focus-visible:ring-2 focus-visible:ring-red-700 focus-visible:ring-offset-2 focus-visible:outline-none sm:w-auto"
							onclick={() => {
								showCancelConfirm = true;
							}}
						>
							{t.cancel}
						</button>
					{/if}
				</section>
			{/if}

			<div class="mt-8 text-center">
				<a href={locale === 'en' ? '/en' : '/'} class="text-neutral-600 underline hover:text-neutral-950">
					{booking.status === 'cancelled' ? t.bookAgain : t.backToPractice}
				</a>
			</div>
		{:else}
			<!-- ═══ Manual lookup form ═══ -->
			<div class="mb-8 text-center">
				<h1 class="mb-2 text-3xl font-bold tracking-tight md:text-4xl">Retrouver ma réservation</h1>
				<p class="text-neutral-600">Entrez votre référence et vos coordonnées pour accéder à votre réservation</p>
			</div>

			{#if displayError}
				<div
					class="mb-6 flex items-center gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800"
					role="alert"
				>
					<AlertCircle size={16} class="shrink-0" aria-hidden="true" />
					{displayError}
				</div>
			{/if}

			<form
				method="POST"
				action="?/lookup"
				class="space-y-5 rounded-2xl border border-neutral-200 bg-white p-5 sm:p-8"
				use:enhance
			>
				<!-- Booking reference -->
				<div>
					<label for="ref" class="mb-1.5 block text-sm font-medium text-neutral-700">Référence de réservation</label>
					<input
						id="ref"
						name="ref"
						type="text"
						placeholder="ex. a1b2c3d4-e5f6-7890-abcd-ef1234567890"
						class="w-full rounded-xl border border-neutral-300 px-4 py-3 font-mono text-sm placeholder:text-neutral-500 focus:border-transparent focus:ring-2 focus:ring-neutral-950 focus:outline-none"
						required
					/>
				</div>

				<!-- Contact method toggle -->
				<div>
					<span class="mb-2 block text-sm font-medium text-neutral-700">Vérification</span>
					<div class="mb-3 flex rounded-xl border border-neutral-300 bg-neutral-50 p-1">
						<button
							type="button"
							class="flex flex-1 items-center justify-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-all {contactMethod ===
							'email'
								? 'bg-white text-neutral-950 shadow-sm'
								: 'text-neutral-600 hover:text-neutral-950'}"
							onclick={() => {
								contactMethod = 'email';
							}}
						>
							<Mail size={14} aria-hidden="true" />
							Email
						</button>
						<button
							type="button"
							class="flex flex-1 items-center justify-center gap-1.5 rounded-lg px-4 py-2 text-sm font-medium transition-all {contactMethod ===
							'phone'
								? 'bg-white text-neutral-950 shadow-sm'
								: 'text-neutral-600 hover:text-neutral-950'}"
							onclick={() => {
								contactMethod = 'phone';
							}}
						>
							<PhoneIcon size={14} aria-hidden="true" />
							Téléphone
						</button>
					</div>

					{#if contactMethod === 'email'}
						<input
							name="email"
							type="email"
							aria-label="Email"
							placeholder="votre@email.com"
							class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm placeholder:text-neutral-500 focus:border-transparent focus:ring-2 focus:ring-neutral-950 focus:outline-none"
						/>
					{:else}
						<input
							name="phone"
							type="tel"
							aria-label="Téléphone"
							placeholder="+33 6 12 34 56 78"
							class="w-full rounded-xl border border-neutral-300 px-4 py-3 text-sm placeholder:text-neutral-500 focus:border-transparent focus:ring-2 focus:ring-neutral-950 focus:outline-none"
						/>
					{/if}
				</div>

				<button
					type="submit"
					class="flex min-h-11 w-full items-center justify-center gap-2 rounded-xl bg-neutral-950 py-3 font-medium text-white transition-colors hover:bg-neutral-950/90 focus-visible:ring-2 focus-visible:ring-neutral-950 focus-visible:ring-offset-2 focus-visible:outline-none"
				>
					<Search size={18} aria-hidden="true" />
					Rechercher
				</button>
			</form>

			<p class="mt-6 text-center text-sm text-neutral-600">
				Un lien direct vous a été envoyé par e-mail lors de votre réservation.
			</p>
		{/if}
	</div>
</div>
