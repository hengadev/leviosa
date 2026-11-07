<script lang="ts">
	import { Clock, Calendar, MapPin, SquarePen, Trash2, Archive, ArchiveRestore, Copy, ImageIcon } from "@lucide/svelte";
	import { untrack } from "svelte";
	import { enhance } from "$app/forms";
	import type { SubmitFunction } from "@sveltejs/kit";
	import { getToastContext } from "$lib/components/toast";
	import { type SuperValidated } from "sveltekit-superforms";
	import { superForm } from "sveltekit-superforms";
	import AlertDialog from "./AlertDialog.svelte";
	import { type CardType, type Category } from "$lib/data/mockData";
	import type { DeleteProduct } from "./schemas";

	type Props = {
		card: CardType;
		statuses: Set<string>;
		categories: Category[];
		availabilities: Set<string>;
		deleteProductForm: SuperValidated<DeleteProduct>;
	};

	let {
		card,
		statuses,
		categories,
		availabilities,
		deleteProductForm,
	}: Props = $props();

	// Derived: after an action the page reloads its data and the card follows.
	let { id, name, price, description, duration, image, published } = $derived(card);

	const toast = getToastContext();

	const statusLabel: Record<string, string> = {
		published: "Publié",
		draft: "Brouillon",
		archived: "Archivé",
	};

	// Archived: kept for the bookings that refer to it, no longer offered.
	// Restoring brings it back as a draft, to publish again on purpose.
	const archived = $derived(published === "archived");
	const statusEnhance: SubmitFunction = () => {
		return async ({ result, update }) => {
			if (result.type === "success") {
				toast.success("Succès", archived ? "Produit restauré en brouillon" : "Produit archivé");
				await update();
			} else if (result.type === "failure") {
				toast.error("Erreur", (result.data as { productError?: string })?.productError ?? "Une erreur est survenue");
			} else if (result.type === "error") {
				toast.error("Erreur", "Une erreur est survenue. Veuillez réessayer.");
			}
		};
	};

	// One delete form per card, each with its own id.
	// The list is keyed by product: a card's product never changes.
	const { enhance: deleteEnhance } = superForm(deleteProductForm, {
		id: `delete-product-${untrack(() => card.id)}`,
		resetForm: true,
	});

	function formatMinutes(totalMinutes: number) {
		const hours = Math.floor(totalMinutes / 60);
		const minutes = totalMinutes % 60;

		let result = "";
		if (hours > 0) {
			result += `${hours}h`;
		}
		if (minutes > 0 || hours === 0) {
			result += `${minutes}min`;
		}

		return result;
	}

	const getStatusBadge = (status: string) => {
		switch (status) {
			case "published":
				return "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400";
			case "draft":
				return "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400";
			case "archived":
				return "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400";
			default:
				return "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400";
		}
	};

	const getAvailabilityBadge = (availability: string) => {
		switch (availability) {
			case "online":
				return "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400";
			case "in-person":
				return "bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400";
			case "hybrid":
				return "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400";
			default:
				return "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400";
		}
	};
</script>

<div
	{id}
	class="bg-card border border-border-card rounded-lg overflow-hidden hover:shadow-card transition-shadow"
>
	<div class="aspect-video bg-muted relative">
		{#if image}
			<img src={image} alt={name} class="w-full h-full object-cover" />
		{:else}
			<div class="w-full h-full flex items-center justify-center text-muted-foreground" aria-hidden="true">
				<ImageIcon size={32} strokeWidth={1.5} />
			</div>
		{/if}
		<div class="absolute top-3 right-3 flex gap-2">
			<span
				class="inline-flex items-center px-2.5 py-0.5 text-xs font-medium rounded-full {getStatusBadge(
					published
				)}"
			>
				{statusLabel[published] ?? published}
			</span>
		</div>
	</div>

	<div class="p-5">
		<div class="flex items-start justify-between mb-3">
			<h3 class="font-semibold text-foreground line-clamp-1">{name}</h3>
			<span class="text-lg font-bold text-primary">{price}€</span>
		</div>

		<p class="text-sm text-foreground-alt line-clamp-2 mb-3">{description}</p>

		<div class="flex flex-wrap gap-2 mb-4">
			<span
				class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-md bg-muted text-foreground-alt"
			>
				<Clock size={12} />
				{formatMinutes(duration)}
			</span>
			<span
				class="inline-flex items-center gap-1 px-2 py-1 text-xs font-medium rounded-md bg-muted text-foreground-alt"
			>
				<MapPin size={12} />
				{card.availability}
			</span>
		</div>

		<div class="flex items-center justify-between pt-3 border-t border-border-card">
			<a
				href="/admin/catalog/products/{card.id}"
				class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-foreground-alt hover:text-foreground hover:bg-muted rounded-lg transition-colors"
			>
				<SquarePen size={14} />
				<span>Modifier</span>
			</a>

			<div class="flex gap-1">
				<!-- Opens the creation form filled from this product. -->
				<a
					href="/admin/catalog/products/new?from={card.id}"
					class="p-1.5 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors"
					aria-label="Dupliquer"
					title="Dupliquer"
				>
					<Copy size={14} />
				</a>
				<form method="POST" action="?/setProductStatus" use:enhance={statusEnhance}>
					<input type="hidden" name="id" value={id} />
					<input type="hidden" name="status" value={archived ? "draft" : "archived"} />
					<button
						type="submit"
						class="p-1.5 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors cursor-pointer"
						aria-label={archived ? "Restaurer en brouillon" : "Archiver"}
						title={archived ? "Restaurer en brouillon" : "Archiver : n'est plus proposé, reste lié aux réservations"}
					>
						{#if archived}
							<ArchiveRestore size={14} />
						{:else}
							<Archive size={14} />
						{/if}
					</button>
				</form>
				<AlertDialog
					label="Supprimer"
					message={`Le produit « ${name} » et toutes ses données seront supprimés.`}
					triggerClass="p-1.5 text-muted-foreground hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950 rounded-lg transition-colors cursor-pointer"
				>
					{#snippet trigger()}
						<Trash2 size={14} />
					{/snippet}
					<form method="POST" action="?/deleteProduct" use:deleteEnhance>
						<input type="hidden" name="id" value={id} />
						<button
							type="submit"
							class="px-4 py-2 rounded-lg bg-red-600 text-white text-sm font-medium hover:bg-red-700 transition-colors cursor-pointer"
						>
							Supprimer
						</button>
					</form>
				</AlertDialog>
			</div>
		</div>
	</div>
</div>
