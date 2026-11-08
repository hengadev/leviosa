import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { withPwa } from '@vite-pwa/sveltekit';

// The build's environment (no @types/node in this project).
const buildEnv: Record<string, string | undefined> =
	(globalThis as { process?: { env: Record<string, string | undefined> } }).process?.env ?? {};

export default defineConfig({
	plugins: [
		tailwindcss(),
		// withPwa runs the sveltekit() plugin itself, with the SvelteKit config
		// (formerly svelte.config.js) as its first argument.
		...withPwa(
			{
				preprocess: vitePreprocess(),
				adapter: adapter(),
				// adapter-node 6 dropped the ORIGIN environment variable. Behind
				// Caddy the origin is derived per request (PROTOCOL_HEADER,
				// HOST_HEADER), so CI images leave this unset; only the local
				// stack without a proxy (local.compose.yaml) pins it at build time.
				paths: { origin: buildEnv.KIT_PATHS_ORIGIN || undefined },
				alias: {
					'@/*': '#lib',
					'@assets/*': 'src/assets'
				}
			},
			{
				registerType: 'autoUpdate',
				includeManifestIcons: false,
				manifest: false,
				generateSW: {
					// The plugin also globs prerendered/**, and the app prerenders
					// nothing: an empty pattern must not fail the build.
					globStrict: false,
					globPatterns: ['**/*.{js,css,ico,png,svg,webp,woff,woff2,html}'],
					navigateFallback: '/offline.html',
					navigateFallbackDenylist: [/^\/api\//, /\/oauth\//],
					runtimeCaching: [
						{
							// Always fetch HTML pages from the network so server-side redirects
							// (e.g. maintenance mode) are never bypassed by the service worker cache.
							urlPattern: ({ request }) => request.mode === 'navigate',
							handler: 'NetworkOnly'
						},
						{
							urlPattern: /^https:\/\/fonts\.(googleapis|gstatic)\.com/i,
							handler: 'CacheFirst',
							options: {
								cacheName: 'google-fonts-cache',
								expiration: {
									maxEntries: 50,
									maxAgeSeconds: 60 * 60 * 24 * 365
								},
								cacheableResponse: {
									statuses: [0, 200]
								}
							}
						},
						{
							urlPattern: /^https:\/\/api\./i,
							handler: 'NetworkFirst',
							options: {
								cacheName: 'api-cache',
								expiration: {
									maxEntries: 100,
									maxAgeSeconds: 60 * 60 * 24
								},
								networkTimeoutSeconds: 10
							}
						},
						{
							urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/i,
							handler: 'CacheFirst',
							options: {
								cacheName: 'image-cache',
								expiration: {
									maxEntries: 200,
									maxAgeSeconds: 60 * 60 * 24 * 30
								}
							}
						}
					]
				},
				devOptions: {
					enabled: false
				}
			}
		)
	]
});
