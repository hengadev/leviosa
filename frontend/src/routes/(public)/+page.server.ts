import type { PageServerLoad } from "./$types";
import { env } from "$env/dynamic/private";
import { mockCategories, mockProducts, mockPrices, mockPartners, mockUsers } from "$lib/data/mockData";

// Homepage highlight sections (Services, Team) reuse the same catalog/partner
// data already powering the dedicated (public)/services and (public)/team
// listing pages — just trimmed down to the single flagship service and the
// first few team members shown on the homepage.
export const load: PageServerLoad = async ({ fetch }) => {
    if (env.USE_MOCK_DATA === "true") {
        const featuredProduct = mockProducts.find((p) => p.status === "published") ?? null;
        const featuredCategory = featuredProduct
            ? mockCategories.find((c) => c.id === featuredProduct.category) ?? null
            : null;
        const featuredPrice =
            (featuredProduct &&
                mockPrices.find(
                    (p) =>
                        p.productId === featuredProduct.id &&
                        p.isActive &&
                        p.interval === "one_time",
                )?.amount) ??
            null;

        const teamMembers = mockPartners.slice(0, 3).map((partner) => {
            const user = mockUsers.find((u) => u.id === partner.user_id);
            return {
                id: partner.id,
                firstname: user?.first_name ?? "—",
                lastname: user?.last_name ?? "—",
                occupation: partner.occupation ?? "",
                quote: partner.quote ?? "",
                tags: partner.tags ?? [],
                picture: partner.picture,
            };
        });

        return {
            featuredService: featuredProduct
                ? {
                      id: featuredProduct.id,
                      name: featuredProduct.name,
                      description: featuredProduct.description,
                      duration: featuredProduct.duration,
                      category: featuredCategory?.name ?? "",
                  }
                : null,
            featuredPrice,
            teamMembers,
        };
    }

    // Fetch published products, use the first as the homepage's flagship service
    const productsRes = await fetch(`${env.API_URL}/products`);
    const products = productsRes.ok ? await productsRes.json() : [];
    const featuredProduct = products[0] ?? null;

    let featuredPrice: number | null = null;
    if (featuredProduct) {
        try {
            const pricesRes = await fetch(`${env.API_URL}/products/${featuredProduct.id}/prices`);
            if (pricesRes.ok) {
                const prices = await pricesRes.json();
                const price =
                    prices.find(
                        (p: { interval: string }) => p.interval === "one_time",
                    ) ?? prices[0];
                if (price) featuredPrice = price.amount;
            }
        } catch {
            // Price fetch failure is non-blocking
        }
    }

    // Fetch partners, use the first three as the homepage's team highlight
    const partnersRes = await fetch(`${env.API_URL}/partners`);
    const partnersData = partnersRes.ok ? await partnersRes.json() : [];
    const teamMembers = partnersData.slice(0, 3).map((partner: any) => ({
        id: partner.id,
        firstname: partner.first_name ?? "—",
        lastname: partner.last_name ?? "—",
        occupation: partner.occupation ?? "",
        quote: partner.quote ?? "",
        tags: partner.tags ?? [],
        picture: partner.picture,
    }));

    return {
        featuredService: featuredProduct
            ? {
                  id: featuredProduct.id,
                  name: featuredProduct.name,
                  description: featuredProduct.description,
                  duration: featuredProduct.duration,
                  category: featuredProduct.category?.name ?? "",
              }
            : null,
        featuredPrice,
        teamMembers,
    };
};
