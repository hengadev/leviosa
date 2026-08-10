import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';
import { mockProducts, mockPartners } from '$lib/data/mockData';

// Public static routes that are always reachable without required query params or
// dynamic data. Deliberately excludes `/book/confirmation` (requires a booking_id
// param, 400s without it) and `/coming-soon` (carries its own `noindex` meta tag).
const STATIC_ROUTES = ['/', '/about', '/services', '/team', '/book', '/bookings', '/legal/privacy', '/legal/terms'];

export const GET: RequestHandler = async ({ url, fetch }) => {
    let productIds: string[] = [];
    let partnerIds: string[] = [];

    if (env.USE_MOCK_DATA === 'true') {
        productIds = mockProducts.map((p) => p.id);
        partnerIds = mockPartners.map((p) => p.id);
    } else {
        // Reuse the same public endpoints the services/team listing pages already
        // load from, so the sitemap can't drift from what's actually published.
        try {
            const productsRes = await fetch(`${env.API_URL}/products`);
            if (productsRes.ok) {
                const products = (await productsRes.json()) as Array<{ id: string }>;
                productIds = products.map((p) => p.id);
            }
        } catch {
            // Sitemap generation is best-effort — fall back to static routes only.
        }

        try {
            const partnersRes = await fetch(`${env.API_URL}/partners`);
            if (partnersRes.ok) {
                const partners = (await partnersRes.json()) as Array<{ id: string }>;
                partnerIds = partners.map((p) => p.id);
            }
        } catch {
            // Sitemap generation is best-effort — fall back to static routes only.
        }
    }

    const paths = [
        ...STATIC_ROUTES,
        ...productIds.map((id) => `/services/${id}`),
        ...partnerIds.map((id) => `/team/${id}`)
    ];

    const urlEntries = paths
        .map((path) => `  <url>\n    <loc>${url.origin}${path}</loc>\n  </url>`)
        .join('\n');

    const body = `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urlEntries}\n</urlset>\n`;

    return new Response(body, {
        headers: {
            'Content-Type': 'application/xml',
            'Cache-Control': 'max-age=0, s-maxage=3600'
        }
    });
};
