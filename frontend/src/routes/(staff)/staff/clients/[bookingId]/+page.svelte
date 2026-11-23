<script lang="ts">
	import PageActionsMenu from '#lib/components/admin/PageActionsMenu.svelte';
	import SmsButton from '#lib/components/practice/SmsButton.svelte';
	import { Phone, Mail, ChevronLeft } from '@lucide/svelte';
	import type { PageProps } from './$types';
	import type { ClientSession } from './+page.server';

	let { data }: PageProps = $props();

	function formatDateTime(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			timeZone: 'Europe/Paris',
			weekday: 'short',
			day: 'numeric',
			month: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatTime(iso: string): string {
		return new Date(iso).toLocaleTimeString('fr-FR', {
			timeZone: 'Europe/Paris',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'confirmed':
				return 'À venir';
			case 'completed':
				return 'Terminée';
			case 'no_show':
				return 'Absent';
			case 'cancelled':
				return 'Annulée';
			default:
				return status;
		}
	}

	function paymentStatusLabel(status: string): string {
		switch (status) {
			case 'paid':
				return 'Payé';
			case 'pending':
				return 'À payer';
			case 'failed':
				return 'Échec';
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

	/** Digits and a leading + only, for tel: and sms: links. */
	function dialable(phone: string): string {
		const trimmed = phone.trim();
		return (trimmed.startsWith('+') ? '+' : '') + trimmed.replace(/\D/g, '');
	}

	/** "+33612345678" → "06 12 34 56 78"; other numbers as stored. */
	function displayPhone(phone: string): string {
		const digits = dialable(phone);
		if (/^\+33\d{9}$/.test(digits)) {
			return ('0' + digits.slice(3)).replace(/(\d{2})(?=\d)/g, '$1 ');
		}
		return phone.trim();
	}

	function firstName(name: string): string {
		return name.split(' ')[0] || name;
	}
</script>

<svelte:head>
	<title>Fiche client | Staff</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
	<div class="mb-8">
		<a
			href="/staff/agenda/reservations"
			class="mb-4 inline-flex min-h-11 items-center gap-1 text-sm font-medium text-muted-foreground hover:text-foreground"
		>
			<ChevronLeft size={16} />
			Réservations
		</a>
		<div class="flex items-start justify-between gap-4">
			<h1 class="text-3xl font-bold break-words text-foreground lg:text-4xl">
				{data.client?.name ?? 'Client'}
			</h1>
			<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
			<div class="lg:hidden">
				<PageActionsMenu />
			</div>
		</div>
	</div>

	{#if data.client}
		<!-- Phone-first contact: Appeler and SMS before the email -->
		<section class="mb-8 rounded-2xl border border-border-card p-4 sm:p-5">
			{#if data.client.phone}
				<div class="grid grid-cols-[1fr_2.75rem] gap-2">
					<a
						href="tel:{dialable(data.client.phone)}"
						class="flex h-11 items-center justify-center gap-2 rounded-xl border border-border-card font-medium hover:bg-muted"
					>
						<Phone size={18} />
						Appeler {displayPhone(data.client.phone)}
					</a>
					<SmsButton phone={dialable(data.client.phone)} firstName={firstName(data.client.name)} />
				</div>
			{/if}
			{#if data.client.email}
				<a
					href="mailto:{data.client.email}"
					class="mt-3 flex min-h-11 items-center gap-2 break-all text-sm text-muted-foreground hover:text-foreground"
				>
					<Mail size={16} class="shrink-0" />
					{data.client.email}
				</a>
			{/if}
		</section>

		<section class="mb-8">
			<h2 class="mb-2 px-1 text-sm font-semibold">À venir</h2>
			<ul class="divide-y divide-border-card rounded-2xl border border-border-card">
				{#each data.client.upcoming as session (session.id)}
					<li class="flex items-center gap-3 px-4 py-3">
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-medium">{session.productName}</p>
							<p class="text-sm text-muted-foreground">
								{formatDateTime(session.startTime)} – {formatTime(session.endTime)}
							</p>
							<p class="text-xs text-muted-foreground">
								{statusLabel(session.status)} · {paymentStatusLabel(session.paymentStatus)}
							</p>
						</div>
					</li>
				{:else}
					<li class="px-4 py-3 text-sm text-muted-foreground">Aucune séance à venir</li>
				{/each}
			</ul>
		</section>

		<section>
			<h2 class="mb-2 px-1 text-sm font-semibold">Passées</h2>
			<ul class="divide-y divide-border-card rounded-2xl border border-border-card">
				{#each data.client.past as session (session.id)}
					<li class="flex items-center gap-3 px-4 py-3">
						<div class="min-w-0 flex-1">
							<p class="truncate text-sm font-medium {session.status === 'cancelled' ? 'line-through opacity-70' : ''}">
								{session.productName}
							</p>
							<p class="text-sm text-muted-foreground">
								{formatDateTime(session.startTime)} – {formatTime(session.endTime)}
							</p>
							<p class="text-xs text-muted-foreground">
								{statusLabel(session.status)} · {paymentStatusLabel(session.paymentStatus)}
							</p>
						</div>
					</li>
				{:else}
					<li class="px-4 py-3 text-sm text-muted-foreground">Aucune séance passée</li>
				{/each}
			</ul>
		</section>
	{:else}
		<section class="rounded-2xl border border-border-card p-6 text-center">
			<p class="font-medium text-foreground">Client introuvable</p>
			<p class="mt-1 text-sm text-muted-foreground">Cette réservation n'est pas la vôtre ou n'existe plus.</p>
		</section>
	{/if}
</div>
