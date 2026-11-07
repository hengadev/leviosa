import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { depositFromForm, parseDeposit, type Deposit } from "$lib/deposit";
import { mockAdminCategories, type Category, type CancellationTier } from "$lib/data/mockData";

interface BackendCategory {
	id: string;
	name: string;
	description: string;
	status: string;
}

// GET /admin/categories returns CategoryWithImage: { category, image? }
interface BackendCategoryWithImage {
	category: BackendCategory;
}

function mapBackendCategoryToFrontend(cat: BackendCategory): Category {
	return {
		id: cat.id,
		name: cat.name,
		description: cat.description,
		status: cat.status as 'published' | 'draft' | 'archived',
	};
}

// Body of POST /admin/products (backend CreateProductWithPriceRequest): the
// product and its first price are created together. The backend rejects
// unknown fields, so the shape must match exactly.
interface CreateProductRequest {
	product: {
		name: string;
		description: string;
		category: string;
		duration: number;
		availability: 'online' | 'in-person' | 'hybrid';
		bufferTime: number;
		cancellationHours: number;
		cancellationTiers?: CancellationTier[];
		deposit?: Deposit;
	};
	price: {
		amount: number; // cents
		currency: string;
		interval: 'one_time';
	};
}

interface CreateProductResponse {
	product_id: string;
	price_id: string;
}

/** The form's values: a blank form, or a product being duplicated. */
export type ProductFormValues = {
	name: string;
	description: string;
	categoryId: string;
	duration: number;
	price: string;
	availability: 'online' | 'in-person' | 'hybrid';
	cancellationHours: number;
	cancellationTiers: CancellationTier[];
	depositType: 'none' | 'fixed' | 'percent';
	depositValue: number | null;
};

// GET /admin/products/{id}: the product, its prices and its image.
interface BackendProductAggregator {
	product: {
		name: string;
		description: string;
		category: { id: string };
		duration: number;
		availability: 'online' | 'in-person' | 'hybrid';
		cancellationHours: number;
		cancellationTiers?: CancellationTier[];
		deposit?: unknown;
	};
	prices?: { amount: number; interval: string; isActive: boolean }[];
}

/** A product's values for a copy of it: everything but its image and status. */
function duplicateValues(agg: BackendProductAggregator): ProductFormValues {
	const p = agg.product;
	const price = (agg.prices ?? []).find((pr) => pr.isActive && pr.interval === 'one_time');
	const deposit = parseDeposit(p.deposit);
	return {
		name: `${p.name} (copie)`,
		description: p.description,
		categoryId: p.category?.id ?? '',
		duration: p.duration,
		price: price ? (price.amount / 100).toFixed(2) : '',
		availability: p.availability,
		cancellationHours: p.cancellationHours,
		cancellationTiers: (p.cancellationTiers ?? []).map((t) => ({ ...t })),
		depositType: deposit?.type ?? 'none',
		depositValue: deposit ? (deposit.type === 'fixed' ? deposit.value / 100 : deposit.value) : null,
	};
}

export const load: PageServerLoad = async ({ fetch, url }) => {
	let categories: Category[];
	// ?from=<id>: the « Dupliquer » action of a product card.
	const from = url.searchParams.get('from');
	let duplicate: ProductFormValues | null = null;

	if (env.USE_MOCK_DATA === 'true') {
		categories = mockAdminCategories;
	} else {
		try {
			const res = await fetch(`${env.API_URL}/admin/categories`);
			if (!res.ok) throw new Error(`Failed to fetch categories: ${res.status} ${res.statusText}`);
			const backendCategories: BackendCategoryWithImage[] = await res.json();
			categories = backendCategories.map((item) => mapBackendCategoryToFrontend(item.category));
		} catch (err) {
			console.error('Error loading categories:', err);
			categories = [];
		}

		if (from) {
			const res = await fetch(`${env.API_URL}/admin/products/${encodeURIComponent(from)}`).catch(() => null);
			if (res?.ok) duplicate = duplicateValues(await res.json());
			else console.error('Error loading the product to duplicate:', res?.status);
		}
	}

	return {
		categories: [{ id: "default", name: "Toutes les catégories" }, ...categories],
		duplicate,
		// The product the copy comes from: its draft is kept apart from a blank one.
		from: duplicate ? from : null,
	};
};

export const actions: Actions = {
	default: async ({ request, fetch }) => {
		const formData = await request.formData();

		const name = formData.get('name') as string;
		const description = formData.get('description') as string;
		const categoryId = formData.get('categoryId') as string;
		const duration = formData.get('duration') as string;
		const price = formData.get('price') as string;
		const status = formData.get('status') as 'published' | 'draft' | 'archived';
		const availability = formData.get('availability') as 'online' | 'in-person' | 'hybrid';
		const cancellationHours = formData.get('cancellationHours') as string;
		const cancellationTiersStr = formData.get('cancellationTiers') as string | null;

		// A session is paid once: its price is a one-time price, in cents.
		const requestBody: CreateProductRequest = {
			product: {
				name,
				description,
				category: categoryId,
				duration: parseInt(duration, 10),
				availability,
				// No buffer: the session's duration includes the practitioner's own break.
				bufferTime: 0,
				cancellationHours: parseInt(cancellationHours, 10),
			},
			price: {
				amount: Math.round(parseFloat(price) * 100),
				currency: 'eur',
				interval: 'one_time',
			},
		};

		const deposit = depositFromForm(formData.get('depositType'), formData.get('depositValue'));
		if (!deposit.ok) return fail(400, { error: deposit.error });
		if (deposit.deposit) requestBody.product.deposit = deposit.deposit;

		if (cancellationTiersStr) {
			try {
				const parsed: CancellationTier[] = JSON.parse(cancellationTiersStr);
				if (parsed.length > 0) {
					requestBody.product.cancellationTiers = parsed.map((t) => ({
						minHoursBefore: Number(t.minHoursBefore),
						refundPercentage: Number(t.refundPercentage),
					}));
				}
			} catch (e) {
				console.error('Invalid cancellationTiers payload:', e);
				return fail(400, { error: 'Paliers de remboursement invalides.' });
			}
		}

		// Publishing needs an image (the backend refuses it otherwise): say
		// so before creating anything.
		const imageFile = formData.get('imageFile');
		const hasImage = imageFile instanceof File && imageFile.size > 0;
		if (status === 'published' && !hasImage) {
			return fail(400, { error: 'Ajoutez une image pour publier le produit, ou créez-le en brouillon.' });
		}

		try {
			const res = await fetch(`${env.API_URL}/admin/products`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestBody),
			});

			if (!res.ok) {
				if (res.status === 401) return fail(401, { error: 'Session expirée. Veuillez vous reconnecter.' });
				if (res.status === 403) return fail(403, { error: "Vous n'avez pas les droits pour effectuer cette action." });
				if (res.status === 409) return fail(409, { error: 'Cette ressource est déjà utilisée.' });
				if (res.status >= 500) return fail(500, { error: 'Une erreur serveur est survenue. Veuillez réessayer.' });
				const errorData = await res.json().catch(() => null);
				console.error('Product creation refused:', res.status, errorData);
				return fail(400, { error: 'Le produit a été refusé : vérifiez les champs saisis.' });
			}

			const created: CreateProductResponse = await res.json();

			if (hasImage) {
				const imageFormData = new FormData();
				imageFormData.append('image', imageFile);
				imageFormData.append('parent_id', created.product_id);
				imageFormData.append('parent_type', 'product');
				imageFormData.append('title', name);
				imageFormData.append('is_active', 'true');

				const imageRes = await fetch(`${env.API_URL}/admin/images`, {
					method: 'POST',
					body: imageFormData,
				});
				if (!imageRes.ok) {
					console.error('Image upload failed after product creation:', imageRes.status, await imageRes.text());
					return fail(400, {
						created: true,
						error: "Produit créé en brouillon, mais l'image n'a pas pu être envoyée. Ajoutez-la depuis la fiche du produit.",
					});
				}
			}

			// The backend creates every product as a draft; apply the chosen
			// status afterwards (publishing needs the image uploaded above).
			if (status && status !== 'draft') {
				const statusRes = await fetch(`${env.API_URL}/admin/products/${created.product_id}`, {
					method: 'PATCH',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ publishedStatus: status }),
				});
				if (!statusRes.ok) {
					console.error('Product status update failed:', statusRes.status, await statusRes.text());
					return fail(400, { created: true, error: "Produit créé en brouillon : le statut choisi n'a pas pu être appliqué." });
				}
			}

			return { success: true };
		} catch (e) {
			console.error('Product creation failed:', e);
			return fail(500, { error: 'Une erreur est survenue lors de la création du produit.' });
		}
	},
};
