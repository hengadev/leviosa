import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

export const GET: RequestHandler = ({ url }) => {
    // Pre-launch: keep crawlers out entirely and don't advertise a sitemap
    // there's no point indexing yet (mirrors the coming-soon page's own
    // `noindex` meta tag).
    if (env.COMING_SOON === 'true') {
        return new Response('User-agent: *\nDisallow: /\n', {
            headers: { 'Content-Type': 'text/plain' }
        });
    }

    const body = [
        'User-agent: *',
        'Allow: /',
        'Disallow: /admin',
        'Disallow: /staff',
        'Disallow: /client',
        'Disallow: /auth',
        'Disallow: /api',
        'Disallow: /book/confirmation',
        'Disallow: /coming-soon',
        '',
        `Sitemap: ${url.origin}/sitemap.xml`,
        ''
    ].join('\n');

    return new Response(body, {
        headers: {
            'Content-Type': 'text/plain',
            'Cache-Control': 'max-age=0, s-maxage=3600'
        }
    });
};
