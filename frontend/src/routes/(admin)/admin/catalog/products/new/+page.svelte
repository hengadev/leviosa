<script lang="ts">
	// The product creation form, in tabs like germinal's event editor: every
	// panel stays mounted (only hidden) so what is typed survives switching,
	// the open tab lives in the address hash, and a tab still missing a
	// required value is marked. What is typed is kept as a draft in this
	// browser (not on the backend: creating a product creates it in Stripe)
	// and restored on the next visit; « Dupliquer » on a product card opens
	// this form filled from that product (?from=<id>), with a draft of its own.
	import { enhance, type SubmitFunction } from "$app/forms";
	import { goto, replaceState } from "$app/navigation";
	import { onMount, tick } from "svelte";
	import { ArrowLeft, ArrowRight, Package, Clock, MapPin, Calendar, FileText, Tag, Euro, X, ImageIcon, Plus, Trash2, LoaderCircle } from "@lucide/svelte";
	import type { PageData } from "./$types";
	import type { ProductFormValues } from "./+page.server";
	import DepositField from "../../DepositField.svelte";
	import { getToastContext } from "#lib/components/toast/index.ts";

	let { data }: { data: PageData } = $props();

	const toast = getToastContext();

	type Status = "published" | "draft" | "archived";
	type Values = ProductFormValues & { status: Status };

	const blank: Values = {
		name: "",
		description: "",
		categoryId: "",
		duration: 60,
		price: "",
		availability: "in-person",
		cancellationHours: 24,
		cancellationTiers: [],
		depositType: "none",
		depositValue: null,
		status: "draft",
	};
	// What the form opens with: a copy of a product, or a blank form.
	const initial: Values = data.duplicate ? { ...data.duplicate, status: "draft" } : blank;

	// Form state
	let name = $state(initial.name);
	let description = $state(initial.description);
	let categoryId = $state(initial.categoryId);
	let duration = $state(initial.duration);
	let price = $state<string | number | null>(initial.price);
	let status = $state<Status>(initial.status);
	let availability = $state(initial.availability);
	let cancellationHours = $state(initial.cancellationHours);
	let cancellationTiers = $state(initial.cancellationTiers.map((t) => ({ ...t })));
	let depositType = $state(initial.depositType);
	let depositValue = $state(initial.depositValue);

	function values(): Values {
		return {
			name,
			description,
			categoryId,
			duration,
			price: price == null ? "" : String(price),
			availability,
			cancellationHours,
			cancellationTiers: cancellationTiers.map((t) => ({ ...t })),
			depositType,
			depositValue,
			status,
		};
	}

	function apply(v: Values) {
		name = v.name;
		description = v.description;
		categoryId = v.categoryId;
		duration = v.duration;
		price = v.price;
		availability = v.availability;
		cancellationHours = v.cancellationHours;
		cancellationTiers = v.cancellationTiers.map((t) => ({ ...t }));
		depositType = v.depositType;
		depositValue = v.depositValue;
		status = v.status;
	}

	function addCancellationTier() {
		cancellationTiers = [...cancellationTiers, { minHoursBefore: 0, refundPercentage: 0 }];
	}

	function removeCancellationTier(index: number) {
		cancellationTiers = cancellationTiers.filter((_, i) => i !== index);
	}

	// ---- Draft, kept in this browser ----

	const draftKey = `leviosa:admin:product-draft${data.from ? `:${data.from}` : ""}`;
	let draftSavedAt = $state<Date | null>(null);
	let draftReady = $state(false);

	function readDraft(): { values: Values; savedAt: string } | null {
		try {
			const raw = localStorage.getItem(draftKey);
			return raw ? JSON.parse(raw) : null;
		} catch {
			return null;
		}
	}

	function clearDraft() {
		try {
			localStorage.removeItem(draftKey);
		} catch {
			// Storage unavailable: there is no draft to clear.
		}
		draftSavedAt = null;
	}

	onMount(() => {
		const draft = readDraft();
		if (draft?.values) {
			apply({ ...initial, ...draft.values });
			draftSavedAt = new Date(draft.savedAt);
			toast.info("Brouillon restauré", "Vous reprenez là où vous vous étiez arrêté.");
		}
		draftReady = true;
	});

	// Every change is saved, half a second after the last keystroke. A form
	// back to what it opened with keeps no draft.
	$effect(() => {
		const snapshot = JSON.stringify(values());
		if (!draftReady) return;
		const timer = setTimeout(() => {
			try {
				if (snapshot === JSON.stringify(initial)) {
					localStorage.removeItem(draftKey);
					draftSavedAt = null;
					return;
				}
				const savedAt = new Date();
				localStorage.setItem(draftKey, JSON.stringify({ values: JSON.parse(snapshot), savedAt: savedAt.toISOString() }));
				draftSavedAt = savedAt;
			} catch {
				// Storage full or blocked: the form still works, without a draft.
			}
		}, 500);
		return () => clearTimeout(timer);
	});

	function discardDraft() {
		clearDraft();
		apply(initial);
		removeImage();
		selectTab("general");
	}

	const savedAtLabel = $derived(
		draftSavedAt ? draftSavedAt.toLocaleTimeString("fr-FR", { hour: "2-digit", minute: "2-digit" }) : null,
	);

	// ---- Image ----

	let imagePreview = $state<string | null>(null);
	let imageInput = $state<HTMLInputElement>();

	function handleImageSelect(event: Event) {
		const file = (event.target as HTMLInputElement).files?.[0];
		imagePreview = file ? URL.createObjectURL(file) : null;
	}

	// A pasted image goes into the file input, which is what the form sends.
	function handleImagePaste(event: ClipboardEvent) {
		for (const item of event.clipboardData?.items ?? []) {
			const file = item.type.startsWith("image/") ? item.getAsFile() : null;
			if (file && imageInput) {
				const transfer = new DataTransfer();
				transfer.items.add(file);
				imageInput.files = transfer.files;
				imagePreview = URL.createObjectURL(file);
				return;
			}
		}
	}

	function removeImage() {
		imagePreview = null;
		if (imageInput) imageInput.value = "";
	}

	// ---- Tabs ----

	const tabs = [
		{ id: "general", title: "Général" },
		{ id: "tarif", title: "Tarif & acompte" },
		{ id: "annulation", title: "Annulation" },
		{ id: "publication", title: "Image & publication" },
	] as const;
	type TabId = (typeof tabs)[number]["id"];

	let activeTab = $state<TabId>("general");
	let tabEls: HTMLButtonElement[] = $state([]);

	function tabFromHash(hash: string): TabId {
		return tabs.find((t) => `#${t.id}` === hash)?.id ?? "general";
	}

	onMount(() => {
		activeTab = tabFromHash(location.hash);
		const follow = () => (activeTab = tabFromHash(location.hash));
		window.addEventListener("hashchange", follow);
		return () => window.removeEventListener("hashchange", follow);
	});

	function selectTab(id: TabId) {
		activeTab = id;
		replaceState(`${location.pathname}${location.search}#${id}`, {});
	}

	const activeIndex = $derived(tabs.findIndex((t) => t.id === activeTab));

	// Arrow keys / Home / End move between tabs (WAI-ARIA tabs pattern).
	function handleTabKeydown(event: KeyboardEvent) {
		const last = tabs.length - 1;
		const target =
			event.key === "ArrowRight" ? (activeIndex === last ? 0 : activeIndex + 1)
			: event.key === "ArrowLeft" ? (activeIndex === 0 ? last : activeIndex - 1)
			: event.key === "Home" ? 0
			: event.key === "End" ? last
			: null;
		if (target === null) return;
		event.preventDefault();
		selectTab(tabs[target].id);
		tabEls[target]?.focus();
	}

	// What each tab still needs before the product can be created.
	const missing = $derived<Record<TabId, number>>({
		general: [name.trim(), description.trim(), categoryId, duration >= 20].filter((ok) => !ok).length,
		tarif:
			(price === "" || price == null ? 1 : 0) +
			(depositType !== "none" && (depositValue == null || Number.isNaN(depositValue)) ? 1 : 0),
		annulation: 0,
		publication: 0,
	});

	const pointsLabel = (n: number) => (n === 1 ? "1 champ à compléter" : `${n} champs à compléter`);

	// ---- Submit ----

	let submitting = $state(false);

	// The form validates itself: a required field on a hidden tab would
	// otherwise block the submission without showing anything. The first
	// invalid field's tab opens and the browser shows its message.
	async function revealInvalid(form: HTMLFormElement) {
		const field = form.querySelector<HTMLInputElement>(":invalid:not(fieldset)");
		if (!field) return;
		const panel = field.closest<HTMLElement>("[role=tabpanel]");
		const id = panel?.id.replace("panel-", "") as TabId | undefined;
		if (id) selectTab(id);
		await tick();
		field.reportValidity();
	}

	const createProductEnhance: SubmitFunction = ({ formElement, cancel }) => {
		if (!formElement.checkValidity()) {
			cancel();
			revealInvalid(formElement);
			return;
		}
		submitting = true;
		return async ({ result }) => {
			submitting = false;
			if (result.type === "success") {
				clearDraft();
				toast.success("Succès", "Produit créé avec succès");
				goto("/admin/catalog");
			} else if (result.type === "failure") {
				const failure = result.data as { error?: string; created?: boolean } | undefined;
				toast.error("Erreur", failure?.error ?? "Une erreur est survenue");
				// The product exists (as a draft): the catalog shows it.
				if (failure?.created) {
					clearDraft();
					goto("/admin/catalog");
				}
			} else if (result.type === "error") {
				toast.error("Erreur", "Une erreur est survenue. Veuillez réessayer.");
			}
		};
	};

	const categoriesWithoutDefault = $derived(data.categories.filter((c) => c.id !== "default"));

	const inputClass =
		"w-full pl-10 pr-4 py-3 border border-border-input rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent";
	const selectClass = `${inputClass} appearance-none`;
	const labelClass = "block text-sm font-medium text-foreground-alt mb-2";
	const iconClass = "absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground";
</script>

<svelte:head>
	<title>Nouveau Produit | Admin</title>
</svelte:head>

{#snippet nextButton(next: TabId)}
	<div class="flex justify-end pt-2">
		<button
			type="button"
			onclick={() => selectTab(next)}
			class="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-foreground-alt hover:text-foreground hover:bg-muted rounded-lg transition-colors"
		>
			Suivant : {tabs.find((t) => t.id === next)?.title}
			<ArrowRight size={16} />
		</button>
	</div>
{/snippet}

<div class="px-4 py-8 lg:py-12">
	<div class="mb-8">
		<a
			href="/admin/catalog"
			class="inline-flex items-center gap-2 text-foreground-alt hover:text-foreground transition-colors mb-4"
		>
			<ArrowLeft size={20} />
			<span>Retour au Catalogue</span>
		</a>
		<h1 class="text-3xl lg:text-4xl font-bold mb-2">
			{data.duplicate ? "Dupliquer un produit" : "Créer un Nouveau Produit"}
		</h1>
		<p class="text-muted-foreground">
			{data.duplicate
				? "Le formulaire reprend le produit d'origine, sauf son image. Le nouveau produit est créé à part."
				: "Remplissez les détails pour créer un nouveau produit"}
		</p>
	</div>

	<div class="bg-background rounded-lg border border-border-card">
		<form method="POST" enctype="multipart/form-data" novalidate use:enhance={createProductEnhance}>
			<div class="overflow-x-auto px-4 lg:px-6 pt-4">
				<div
					role="tablist"
					aria-label="Sections du produit"
					class="inline-flex items-center w-fit gap-1 text-sm font-semibold border-b border-border-card"
				>
					{#each tabs as tab, i (tab.id)}
						<button
							bind:this={tabEls[i]}
							type="button"
							role="tab"
							id="tab-{tab.id}"
							aria-selected={tab.id === activeTab}
							aria-controls="panel-{tab.id}"
							tabindex={tab.id === activeTab ? 0 : -1}
							onclick={() => selectTab(tab.id)}
							onkeydown={handleTabKeydown}
							class="inline-flex items-center whitespace-nowrap px-3 md:px-4 py-2 border-b-2 mb-[-1px] transition-colors cursor-pointer {tab.id === activeTab
								? 'border-b-foreground text-foreground'
								: 'border-transparent text-muted-foreground hover:text-foreground-alt'}"
						>
							{tab.title}
							{#if missing[tab.id] > 0}
								<!-- A hollow ring, named for screen readers: still to fill in. -->
								<span
									aria-hidden="true"
									class="ml-1.5 inline-block size-2 rounded-full border-[1.5px] border-current opacity-70"
									title={pointsLabel(missing[tab.id])}
								></span>
								<span class="sr-only">({pointsLabel(missing[tab.id])})</span>
							{/if}
						</button>
					{/each}
				</div>
			</div>

			<!-- Général -->
			<div role="tabpanel" id="panel-general" aria-labelledby="tab-general" hidden={activeTab !== "general"} class="p-6 lg:p-8 space-y-6">
				<div>
					<label for="name" class={labelClass}>Nom du Produit <span class="text-red-500">*</span></label>
					<div class="relative">
						<Package size={18} class={iconClass} />
						<input id="name" name="name" type="text" bind:value={name} required placeholder="Séance d'ostéopathie" class={inputClass} />
					</div>
				</div>

				<div>
					<label for="description" class={labelClass}>Description <span class="text-red-500">*</span></label>
					<div class="relative">
						<FileText size={18} class="absolute left-3 top-3 text-muted-foreground" />
						<textarea
							id="description"
							name="description"
							bind:value={description}
							required
							rows="4"
							placeholder="Décrivez ce produit..."
							class="{inputClass} resize-none"
						></textarea>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<label for="categoryId" class={labelClass}>Catégorie <span class="text-red-500">*</span></label>
						<div class="relative">
							<Tag size={18} class={iconClass} />
							<select id="categoryId" name="categoryId" bind:value={categoryId} required class={selectClass}>
								<option value="">Sélectionner une catégorie</option>
								{#each categoriesWithoutDefault as cat (cat.id)}
									<option value={cat.id}>{cat.name}</option>
								{/each}
							</select>
						</div>
						{#if categoriesWithoutDefault.length === 0}
							<p class="mt-1 text-xs text-muted-foreground">
								Aucune catégorie : créez-en une depuis le catalogue d'abord.
							</p>
						{/if}
					</div>

					<div>
						<label for="duration" class={labelClass}>Durée (minutes) <span class="text-red-500">*</span></label>
						<div class="relative">
							<Clock size={18} class={iconClass} />
							<input id="duration" name="duration" type="number" bind:value={duration} required min="20" step="10" placeholder="60" class={inputClass} />
						</div>
						<p class="mt-1 text-xs text-muted-foreground">
							Minimum 20 minutes, par incréments de 10. Comprend le temps entre deux séances.
						</p>
					</div>
				</div>

				<div class="md:w-1/2 md:pr-3">
					<label for="availability" class={labelClass}>Disponibilité <span class="text-red-500">*</span></label>
					<div class="relative">
						<MapPin size={18} class={iconClass} />
						<select id="availability" name="availability" bind:value={availability} required class={selectClass}>
							<option value="online">En ligne</option>
							<option value="in-person">En présentiel</option>
							<option value="hybrid">Hybride</option>
						</select>
					</div>
				</div>

				{@render nextButton("tarif")}
			</div>

			<!-- Tarif & acompte -->
			<div role="tabpanel" id="panel-tarif" aria-labelledby="tab-tarif" hidden={activeTab !== "tarif"} class="p-6 lg:p-8 space-y-6">
				<div class="md:w-1/2 md:pr-3">
					<label for="price" class={labelClass}>Prix (€) <span class="text-red-500">*</span></label>
					<div class="relative">
						<Euro size={18} class={iconClass} />
						<input id="price" name="price" type="number" bind:value={price} required min="0" step="0.01" placeholder="70.00" class={inputClass} />
					</div>
					<p class="mt-1 text-xs text-muted-foreground">Le prix d'une séance, payé une fois.</p>
				</div>

				<DepositField priceEuros={price == null ? "" : String(price)} bind:depositType bind:depositValue />

				{@render nextButton("annulation")}
			</div>

			<!-- Annulation -->
			<div role="tabpanel" id="panel-annulation" aria-labelledby="tab-annulation" hidden={activeTab !== "annulation"} class="p-6 lg:p-8 space-y-6">
				<div class="md:w-1/2 md:pr-3">
					<label for="cancellationHours" class={labelClass}>Délai d'annulation (heures)</label>
					<div class="relative">
						<Calendar size={18} class={iconClass} />
						<input id="cancellationHours" name="cancellationHours" type="number" bind:value={cancellationHours} min="0" placeholder="24" class={inputClass} />
					</div>
					<p class="mt-1 text-xs text-muted-foreground">Délai d'annulation avant la séance.</p>
				</div>

				<div class="pt-4 border-t border-border-card">
					<div class="flex items-center justify-between mb-2">
						<span class="block text-sm font-medium text-foreground-alt">Paliers de frais d'annulation</span>
						<button
							type="button"
							onclick={addCancellationTier}
							class="inline-flex items-center gap-1 text-sm text-foreground hover:opacity-80 transition-opacity"
						>
							<Plus size={16} />
							Ajouter un palier
						</button>
					</div>
					<p class="text-xs text-muted-foreground mb-3">
						Remplace le blocage strict d'annulation (ADR-0006) par des frais gradués selon le délai avant le rendez-vous (ADR-0014). Laissez vide pour ne pas utiliser les frais gradués. Le dernier palier doit couvrir 0h pour être valide.
					</p>

					{#if cancellationTiers.length === 0}
						<p class="text-sm text-muted-foreground italic mb-2">Aucun palier configuré.</p>
					{:else}
						<div class="space-y-2 mb-2">
							{#each cancellationTiers as tier, i}
								<div class="flex items-center gap-3 bg-muted/50 rounded-lg p-3">
									<div class="flex-1">
										<label for={`tier-hours-${i}`} class="block text-xs text-muted-foreground mb-1">
											Heures minimum avant le rendez-vous
										</label>
										<input
											id={`tier-hours-${i}`}
											type="number"
											min="0"
											bind:value={tier.minHoursBefore}
											class="w-full px-3 py-2 text-sm border border-border-input rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent"
										/>
									</div>
									<div class="flex-1">
										<label for={`tier-refund-${i}`} class="block text-xs text-muted-foreground mb-1">
											Pourcentage remboursé
										</label>
										<input
											id={`tier-refund-${i}`}
											type="number"
											min="0"
											max="100"
											bind:value={tier.refundPercentage}
											class="w-full px-3 py-2 text-sm border border-border-input rounded-lg bg-background focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent"
										/>
									</div>
									<button
										type="button"
										onclick={() => removeCancellationTier(i)}
										class="mt-5 p-2 text-muted-foreground hover:text-red-600 transition-colors"
										aria-label="Supprimer ce palier"
									>
										<Trash2 size={16} />
									</button>
								</div>
							{/each}
						</div>
					{/if}
					<input type="hidden" name="cancellationTiers" value={JSON.stringify(cancellationTiers)} />
				</div>

				{@render nextButton("publication")}
			</div>

			<!-- Image & publication -->
			<div role="tabpanel" id="panel-publication" aria-labelledby="tab-publication" hidden={activeTab !== "publication"} class="p-6 lg:p-8 space-y-6">
				<div>
					<span class={labelClass}>Image du Produit</span>
					<p class="text-xs text-muted-foreground mb-3">
						Facultative. JPEG, PNG ou WebP, 10 Mo maximum. L'image n'est pas gardée dans le brouillon.
					</p>

					<!-- The file input stays in the page under the preview: it is what the form sends. -->
					<div
						class="relative w-full aspect-video max-h-96 rounded-lg overflow-hidden {imagePreview
							? 'bg-muted border border-border-card'
							: 'bg-muted border-2 border-dashed border-border-card hover:border-foreground/30 transition-colors'}"
						onpaste={handleImagePaste}
					>
						<input
							bind:this={imageInput}
							type="file"
							name="imageFile"
							accept="image/*"
							onchange={handleImageSelect}
							class="absolute inset-0 w-full h-full opacity-0 cursor-pointer {imagePreview ? 'pointer-events-none' : ''}"
							aria-label="Télécharger une image"
						/>
						{#if imagePreview}
							<img src={imagePreview} alt="Aperçu de l'image" class="w-full h-full object-contain" />
							<button
								type="button"
								onclick={removeImage}
								class="absolute top-3 right-3 p-2 bg-black/60 hover:bg-red-600 text-white rounded-lg transition-colors"
								aria-label="Retirer l'image"
							>
								<X size={16} />
							</button>
						{:else}
							<div class="absolute inset-0 flex flex-col items-center justify-center gap-3 pointer-events-none">
								<div class="w-16 h-16 rounded-full bg-foreground/10 flex items-center justify-center">
									<ImageIcon size={32} class="text-foreground-alt" />
								</div>
								<div class="text-center">
									<p class="text-sm font-medium text-foreground">Cliquez pour télécharger ou collez une image</p>
									<p class="text-xs text-muted-foreground mt-1">ou glissez-déposez un fichier ici</p>
								</div>
							</div>
						{/if}
					</div>
				</div>

				<div class="md:w-1/2 md:pr-3">
					<label for="status" class={labelClass}>Statut <span class="text-red-500">*</span></label>
					<div class="relative">
						<Tag size={18} class={iconClass} />
						<select id="status" name="status" bind:value={status} required class={selectClass}>
							<option value="draft">Brouillon</option>
							<option value="published">Publié</option>
							<option value="archived">Archivé</option>
						</select>
					</div>
				</div>
			</div>

			<!-- Draft state and actions, under every tab -->
			<div class="flex flex-col-reverse sm:flex-row sm:items-center gap-3 px-6 lg:px-8 py-4 border-t border-border-card">
				<p class="flex-1 text-xs text-muted-foreground" aria-live="polite">
					{#if savedAtLabel}
						Brouillon enregistré sur cet appareil à {savedAtLabel} ·
						<button type="button" onclick={discardDraft} class="underline underline-offset-2 hover:text-foreground">
							Repartir de zéro
						</button>
					{:else}
						Ce que vous saisissez est gardé sur cet appareil jusqu'à la création.
					{/if}
				</p>
				<a
					href="/admin/catalog"
					class="px-6 py-3 bg-background border border-border-input text-foreground-alt rounded-lg hover:bg-muted transition-colors font-medium text-center"
				>
					Annuler
				</a>
				<button
					type="submit"
					disabled={submitting}
					class="inline-flex items-center justify-center gap-2 px-6 py-3 bg-foreground text-background rounded-lg hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-foreground focus:ring-offset-2 transition-colors font-medium disabled:opacity-60"
				>
					{#if submitting}
						<LoaderCircle size={18} class="animate-spin" />
						Création…
					{:else}
						Créer le Produit
					{/if}
				</button>
			</div>
		</form>
	</div>
</div>
