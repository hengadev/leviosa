<script lang="ts">
	import PageActionsMenu from "#lib/components/admin/PageActionsMenu.svelte";
	import type { PageProps } from './$types';
	import {
		DollarSign,
		ArrowDown,
		TrendingUp,
		CreditCard,
		Calendar,
		Download,
		AlertTriangle,
		ChevronLeft,
		ChevronRight,
		Users,
		Hourglass
	} from '@lucide/svelte';
	import { goto } from '$app/navigation';

	let { data }: PageProps = $props();

	function formatCents(cents: number): string {
		return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(cents / 100);
	}

	function formatDateTime(isoString: string): string {
		const date = new Date(isoString);
		return date.toLocaleDateString('fr-FR', {
			day: '2-digit',
			month: '2-digit',
			year: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getTypeBadge(type: string) {
		return type === 'payment'
			? 'bg-muted text-foreground'
			: 'bg-muted text-muted-foreground';
	}

	function getTypeLabel(type: string): string {
		return type === 'payment' ? 'Paiement' : 'Remboursement';
	}

	function getBookingStatusBadge(status: string) {
		switch (status) {
			case 'completed':
				return 'bg-muted text-foreground';
			case 'cancelled':
				return 'bg-muted text-muted-foreground';
			case 'confirmed':
				return 'bg-muted text-foreground';
			case 'no_show':
				return 'bg-muted text-foreground';
			default:
				return 'bg-muted text-foreground';
		}
	}

	function getBookingStatusLabel(status: string): string {
		switch (status) {
			case 'completed': return 'Complété';
			case 'cancelled': return 'Annulé';
			case 'confirmed': return 'Confirmé';
			case 'no_show': return 'Absence';
			default: return status;
		}
	}

	function getPaymentStatusBadge(status: string) {
		switch (status) {
			case 'paid':
				return 'bg-muted text-foreground';
			case 'refunded':
				return 'bg-muted text-foreground';
			case 'disputed':
				return 'border border-foreground text-foreground';
			default:
				return 'bg-muted text-foreground';
		}
	}

	function getPaymentStatusLabel(status: string): string {
		switch (status) {
			case 'paid': return 'Payé';
			case 'refunded': return 'Remboursé';
			case 'disputed': return 'Contesté';
			default: return status;
		}
	}

	function getPaymentModeBadge(mode: string) {
		return mode === 'on_site'
			? 'border border-foreground text-foreground'
			: 'bg-muted text-foreground';
	}

	function getPaymentModeLabel(mode: string): string {
		return mode === 'on_site' ? 'Sur place' : 'En ligne';
	}

	// --- Month selector (French) ---

	// Selected month derived from the 'from' date, "YYYY-MM"
	let selectedMonth: string = $derived(data.from.substring(0, 7));

	function capitalize(s: string): string {
		return s.charAt(0).toUpperCase() + s.slice(1);
	}

	let monthLabel: string = $derived.by(() => {
		const [year, month] = selectedMonth.split('-').map(Number);
		const label = new Date(year, month - 1, 1).toLocaleDateString('fr-FR', {
			month: 'long',
			year: 'numeric'
		});
		return capitalize(label);
	});

	function gotoMonth(yearMonth: string) {
		const [year, month] = yearMonth.split('-').map(Number);
		const from = `${yearMonth}-01`;
		// Last day of selected month
		const lastDay = new Date(year, month, 0).getDate();
		const to = `${yearMonth}-${String(lastDay).padStart(2, '0')}`;
		goto(`/admin/compta?from=${from}&to=${to}`);
	}

	function shiftMonth(delta: number) {
		const [year, month] = selectedMonth.split('-').map(Number);
		const d = new Date(year, month - 1 + delta, 1);
		const yearMonth = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
		gotoMonth(yearMonth);
	}
</script>

<svelte:head>
	<title>Comptabilité | Admin</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
	<div class="mb-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
		<div>
			<div class="mb-1 flex items-center justify-between gap-4">
				<h1 class="text-3xl lg:text-4xl font-bold text-foreground">Comptabilité</h1>
				<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
				<div class="lg:hidden">
					<PageActionsMenu />
				</div>
			</div>
			<p class="text-muted-foreground">
				Suivi des revenus et des transactions
			</p>
		</div>
		<div class="flex items-center gap-3">
			<!-- Month selector (French) -->
			<div class="flex items-center gap-1" role="group" aria-label="Choisir le mois">
				<button
					type="button"
					onclick={() => shiftMonth(-1)}
					aria-label="Mois précédent"
					class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-border bg-card text-foreground hover:bg-muted/50 transition-colors"
				>
					<ChevronLeft size={20} />
				</button>
				<span class="min-w-32 text-center text-sm font-semibold text-foreground tabular-nums">
					{monthLabel}
				</span>
				<button
					type="button"
					onclick={() => shiftMonth(1)}
					aria-label="Mois suivant"
					class="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg border border-border bg-card text-foreground hover:bg-muted/50 transition-colors"
				>
					<ChevronRight size={20} />
				</button>
			</div>
			<!-- CSV export -->
			<a
				href={`/admin/compta/export?from=${data.from}&to=${data.to}`}
				download
				class="flex h-11 items-center gap-2 bg-card border border-border rounded-lg px-4 text-sm text-foreground hover:bg-muted/50 transition-colors"
			>
				<Download size={16} />
				<span><span class="hidden sm:inline">Exporter&nbsp;</span>CSV</span>
			</a>
		</div>
	</div>

	<!-- Last 7 days -->
	{#if data.last7Days}
	<h2 class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Activité</h2>
	<div class="mb-6 grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-6">
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:mb-3">
				<TrendingUp size={16} class="text-muted-foreground" />
				<span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Revenus (7 j)</span>
			</div>
			<div class="text-base lg:text-2xl font-bold text-foreground">{formatCents(data.last7Days.revenueThisWeek)}</div>
		</div>
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:mb-3">
				<Calendar size={16} class="text-muted-foreground" />
				<span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Réservations (7 j)</span>
			</div>
			<div class="text-base lg:text-2xl font-bold text-foreground">{data.last7Days.bookingsThisWeek}</div>
		</div>
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:mb-3">
				<Users size={16} class="text-muted-foreground" />
				<span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">À venir</span>
			</div>
			<div class="text-base lg:text-2xl font-bold text-foreground">{data.last7Days.upcomingBookingsCount}</div>
		</div>
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:mb-3">
				<Hourglass size={16} class="text-muted-foreground" />
				<span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">En attente</span>
			</div>
			<div class="text-base lg:text-2xl font-bold text-foreground">{data.last7Days.pendingBookingsCount}</div>
		</div>
	</div>
	{/if}

	<!-- Summary KPIs: compact 2×2 on a phone, net revenue most prominent -->
	<h2 class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{monthLabel}</h2>
	<div class="grid grid-cols-2 lg:grid-cols-5 gap-3 lg:gap-6 mb-6 lg:mb-8">
		<!-- Gross Revenue -->
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:items-start lg:justify-between lg:mb-4">
				<div class="lg:p-3 lg:bg-muted lg:rounded-lg">
					<DollarSign size={18} class="text-foreground lg:hidden" />
					<DollarSign size={24} class="hidden lg:block text-foreground" />
				</div>
			</div>
			<div class="text-base lg:text-3xl font-bold text-foreground lg:mb-1">
				{formatCents(data.summary.grossRevenueInCents)}
			</div>
			<p class="text-xs lg:text-sm text-muted-foreground">Chiffre d'affaires brut</p>
		</div>

		<!-- Refunds -->
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:items-start lg:justify-between lg:mb-4">
				<div class="lg:p-3 lg:bg-muted lg:rounded-lg">
					<ArrowDown size={18} class="text-foreground lg:hidden" />
					<ArrowDown size={24} class="hidden lg:block text-foreground" />
				</div>
			</div>
			<div class="text-base lg:text-3xl font-bold text-foreground lg:mb-1">
				-{formatCents(data.summary.refundsInCents)}
			</div>
			<p class="text-xs lg:text-sm text-muted-foreground">Remboursements</p>
		</div>

		<!-- Disputed (open Stripe chargebacks — no automatic refund/cancellation, manual response required) -->
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:items-start lg:justify-between lg:mb-4">
				<div class="lg:p-3 lg:bg-muted lg:rounded-lg">
					<AlertTriangle size={18} class="text-foreground lg:hidden" />
					<AlertTriangle size={24} class="hidden lg:block text-foreground" />
				</div>
			</div>
			<div class="text-base lg:text-3xl font-bold text-foreground lg:mb-1">
				{formatCents(data.summary.disputedInCents)}
			</div>
			<p class="text-xs lg:text-sm text-muted-foreground">En litige</p>
		</div>

		<!-- On-site (paid at the practice, recorded by the practitioner) -->
		<div class="bg-card rounded-lg border border-border p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:items-start lg:justify-between lg:mb-4">
				<div class="lg:p-3 lg:bg-muted lg:rounded-lg">
					<CreditCard size={18} class="text-foreground lg:hidden" />
					<CreditCard size={24} class="hidden lg:block text-foreground" />
				</div>
			</div>
			<div class="text-base lg:text-3xl font-bold text-foreground lg:mb-1">
				{formatCents(data.summary.onSiteInCents)}
			</div>
			<p class="text-xs lg:text-sm text-muted-foreground">Sur place</p>
		</div>

		<!-- Net Revenue -->
		<div class="bg-card rounded-lg border-2 border-foreground p-3 lg:p-6">
			<div class="flex items-center gap-2 mb-1 lg:items-start lg:justify-between lg:mb-4">
				<div class="lg:p-3 lg:bg-muted lg:rounded-lg">
					<TrendingUp size={18} class="text-foreground lg:hidden" />
					<TrendingUp size={24} class="hidden lg:block text-foreground" />
				</div>
			</div>
			<div class="text-lg lg:text-3xl font-bold text-foreground lg:mb-1">
				{formatCents(data.summary.netRevenueInCents)}
			</div>
			<p class="text-xs lg:text-sm text-muted-foreground">Chiffre d'affaires net</p>
		</div>
	</div>

	<!-- Transactions: cards on a phone, table from lg up -->
	<div class="bg-card rounded-lg border border-border mb-8">
		<div class="p-4 lg:p-6 border-b border-border">
			<h2 class="text-lg lg:text-xl font-semibold text-foreground">Transactions</h2>
		</div>
		{#if data.transactions.length === 0}
			<div class="p-12 text-center">
				<CreditCard size={48} class="mx-auto mb-4 text-muted-foreground/50" />
				<p class="text-muted-foreground">Aucune transaction pour cette période.</p>
			</div>
		{:else}
			<!-- Mobile: cards -->
			<ul class="divide-y divide-border lg:hidden">
				{#each data.transactions as transaction (transaction.key)}
					<li class="p-4">
						<div class="flex items-start justify-between gap-3">
							<div class="min-w-0">
								<p class="truncate text-sm font-medium text-foreground">{transaction.description}</p>
								<p class="mt-0.5 text-sm text-muted-foreground truncate">
									{transaction.clientName}{transaction.partnerName ? ` · ${transaction.partnerName}` : ''}
								</p>
								<p class="mt-0.5 text-xs text-muted-foreground">{formatDateTime(transaction.date)}</p>
							</div>
							<div class="shrink-0 text-right">
								<p class="text-sm font-semibold {transaction.type === 'payment' ? 'text-foreground' : 'text-muted-foreground'}">
									{transaction.type === 'refund' ? '-' : ''}{formatCents(transaction.amountInCents)}
								</p>
								<span class="mt-1 inline-block px-3 py-1 rounded-full text-xs font-medium {getPaymentStatusBadge(transaction.paymentStatus)}">
									{getPaymentStatusLabel(transaction.paymentStatus)}
								</span>
							</div>
						</div>
						<div class="mt-2 flex flex-wrap gap-2">
							<span class="px-3 py-1 rounded-full text-xs font-medium {getTypeBadge(transaction.type)}">
								{getTypeLabel(transaction.type)}
							</span>
							<span class="px-3 py-1 rounded-full text-xs font-medium {getPaymentModeBadge(transaction.paymentMode)}">
								{getPaymentModeLabel(transaction.paymentMode)}
							</span>
							<span class="px-3 py-1 rounded-full text-xs font-medium {getBookingStatusBadge(transaction.bookingStatus)}">
								{getBookingStatusLabel(transaction.bookingStatus)}
							</span>
						</div>
					</li>
				{/each}
			</ul>
			<!-- Desktop: table -->
			<div class="hidden lg:block overflow-x-auto">
				<table class="w-full">
					<thead>
						<tr class="bg-muted/50 border-b border-border">
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Date</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Produit</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Client</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Praticien</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Type</th>
						<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Mode</th>
							<th class="text-right py-4 px-6 text-sm font-medium text-muted-foreground">Montant</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Paiement</th>
							<th class="text-left py-4 px-6 text-sm font-medium text-muted-foreground">Réservation</th>
						</tr>
					</thead>
					<tbody>
						{#each data.transactions as transaction (transaction.key)}
							<tr class="border-b border-border hover:bg-muted/30 transition-colors">
								<td class="py-4 px-6 text-sm text-muted-foreground">{formatDateTime(transaction.date)}</td>
								<td class="py-4 px-6">
									<span class="text-sm font-medium text-foreground">{transaction.description}</span>
								</td>
								<td class="py-4 px-6 text-sm text-foreground">{transaction.clientName}</td>
								<td class="py-4 px-6 text-sm text-foreground">{transaction.partnerName}</td>
								<td class="py-4 px-6">
									<span class="px-3 py-1 rounded-full text-xs font-medium {getTypeBadge(transaction.type)}">
										{getTypeLabel(transaction.type)}
									</span>
								</td>
								<td class="py-4 px-6">
									<span class="px-3 py-1 rounded-full text-xs font-medium {getPaymentModeBadge(transaction.paymentMode)}">
										{getPaymentModeLabel(transaction.paymentMode)}
									</span>
								</td>
								<td class="py-4 px-6 text-right">
									<span class="text-sm font-medium {transaction.type === 'payment'
										? 'text-foreground'
										: 'text-muted-foreground'}">
										{transaction.type === 'refund' ? '-' : ''}{formatCents(transaction.amountInCents)}
									</span>
								</td>
								<td class="py-4 px-6">
									<span class="px-3 py-1 rounded-full text-xs font-medium {getPaymentStatusBadge(transaction.paymentStatus)}">
										{getPaymentStatusLabel(transaction.paymentStatus)}
									</span>
								</td>
								<td class="py-4 px-6">
									<span class="px-3 py-1 rounded-full text-xs font-medium {getBookingStatusBadge(transaction.bookingStatus)}">
										{getBookingStatusLabel(transaction.bookingStatus)}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</div>
