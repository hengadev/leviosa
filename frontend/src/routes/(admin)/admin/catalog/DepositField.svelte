<script lang="ts">
	// Deposit (acompte) editor for the admin product forms — issue 013.
	// Submits `depositType` (none | fixed | percent) and `depositValue`
	// (euros for a fixed amount, whole percentage points otherwise); the
	// form action turns them into the API payload (lib/deposit.ts).
	import { amountDueOnlineCents, type Deposit, type DepositType } from "$lib/deposit";

	let { deposit = null, priceEuros = "" }: { deposit?: Deposit | null; priceEuros?: string } = $props();

	let depositType = $state<"none" | DepositType>(deposit?.type ?? "none");
	let depositValue = $state<number | null>(
		deposit ? (deposit.type === "fixed" ? deposit.value / 100 : deposit.value) : null,
	);

	const priceCents = $derived(Math.round(Number(String(priceEuros).replace(",", ".")) * 100));

	const formatEuros = (cents: number) =>
		new Intl.NumberFormat("fr-FR", { style: "currency", currency: "EUR" }).format(cents / 100);

	const preview = $derived.by(() => {
		if (!Number.isFinite(priceCents) || priceCents <= 0) return null;
		if (depositType === "none") return `Le prix complet (${formatEuros(priceCents)}) est payé en ligne.`;
		if (depositValue == null || !Number.isFinite(depositValue)) return null;
		const d: Deposit =
			depositType === "fixed"
				? { type: "fixed", value: Math.round(depositValue * 100) }
				: { type: "percent", value: depositValue };
		if (d.type === "fixed" && d.value > priceCents) return "L'acompte ne peut pas dépasser le prix.";
		const online = amountDueOnlineCents(d, priceCents);
		return `${formatEuros(online)} payés en ligne, ${formatEuros(priceCents - online)} réglés au cabinet.`;
	});
</script>

<div class="pt-4 border-t border-border-card">
	<span class="block text-sm font-medium text-foreground-alt mb-2">Acompte</span>
	<p class="text-xs text-muted-foreground mb-3">
		Partie du prix payée en ligne à la réservation ; le solde est réglé au cabinet. Sans acompte, le prix complet est payé en ligne.
	</p>
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
		<div>
			<label for="depositType" class="block text-xs text-muted-foreground mb-1">Type d'acompte</label>
			<select
				id="depositType"
				name="depositType"
				bind:value={depositType}
				class="w-full px-3 py-2 text-sm border border-border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent bg-background"
			>
				<option value="none">Aucun acompte</option>
				<option value="fixed">Montant fixe (€)</option>
				<option value="percent">Pourcentage du prix (%)</option>
			</select>
		</div>
		{#if depositType !== "none"}
			<div>
				<label for="depositValue" class="block text-xs text-muted-foreground mb-1">
					{depositType === "fixed" ? "Montant (€)" : "Pourcentage (%)"}
				</label>
				<input
					id="depositValue"
					name="depositValue"
					type="number"
					required
					min={depositType === "fixed" ? 0 : 1}
					max={depositType === "percent" ? 100 : undefined}
					step={depositType === "fixed" ? 0.01 : 1}
					bind:value={depositValue}
					class="w-full px-3 py-2 text-sm border border-border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent"
				/>
			</div>
		{/if}
	</div>
	{#if preview}
		<p class="mt-2 text-xs text-muted-foreground">{preview}</p>
	{/if}
</div>
