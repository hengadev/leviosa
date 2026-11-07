import type { RequestHandler } from './$types';

// The public site is the practice page (issue 009) plus the pages that stay:
// the booking lookup and the legal pages. The old multi-practitioner routes
// (services, team, about, contact, book) redirect to `/` and are not listed.
// `/coming-soon` carries its own `noindex` meta tag.
const STATIC_ROUTES = ['/', '/en', '/bookings', '/legal/privacy', '/legal/terms', '/legal/mentions-legales'];

export const GET: RequestHandler = async ({ url }) => {
    const paths = [...STATIC_ROUTES];

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
