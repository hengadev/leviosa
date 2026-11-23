// The app's service worker, bundled by @vite-pwa/sveltekit (strategies:
// 'buildSW' in vite.config.ts), which injects the precache manifest into
// self.__WB_MANIFEST. It keeps the generated Workbox worker's precache,
// runtime caches and autoUpdate, and adds the practitioner notifications
// (pwa/push.js, ADR-0020), which a generated worker cannot import. Kept
// outside src/ so the app's DOM-typed check does not read it.
//
// The generated worker bound its navigation fallback to '/offline.html',
// which the SvelteKit manifest transform precaches as '/offline': it threw
// "non-precached-url" on startup, so no service worker ever installed. Bound
// as it was meant, it would also have answered every page with the offline
// page. Here pages come from the network and the offline page answers only
// when the network fails.
import { clientsClaim } from '@vite-pwa/workbox-swkit/core';
import { matchPrecache, precacheAndRoute } from '@vite-pwa/workbox-swkit/precaching';
import { NavigationRoute, registerRoute } from '@vite-pwa/workbox-swkit/routing';
import { CacheFirst, NetworkFirst, NetworkOnly } from '@vite-pwa/workbox-swkit/strategies';
import { ExpirationPlugin } from '@vite-pwa/workbox-swkit/expiration';
import { CacheableResponsePlugin } from '@vite-pwa/workbox-swkit/cacheable-response';
import './push.js';

// registerType 'autoUpdate': a new worker takes over at once.
self.skipWaiting();
clientsClaim();

precacheAndRoute(self.__WB_MANIFEST, { cleanURLs: true, parallel: { enabled: false, concurrency: 5 } });

// Always fetch HTML pages from the network so server-side redirects
// (e.g. maintenance mode) are never bypassed by the service worker cache;
// offline, the precached offline page (static/offline.html) answers.
registerRoute(
	new NavigationRoute(
		new NetworkOnly({
			plugins: [{ handlerDidError: async () => (await matchPrecache('/offline')) ?? Response.error() }]
		}),
		{ denylist: [/^\/api\//, /\/oauth\//] }
	)
);
registerRoute(({ request }) => request.mode === 'navigate', new NetworkOnly(), 'GET');

registerRoute(
	/^https:\/\/fonts\.(googleapis|gstatic)\.com/i,
	new CacheFirst({
		cacheName: 'google-fonts-cache',
		plugins: [
			new ExpirationPlugin({ maxEntries: 50, maxAgeSeconds: 60 * 60 * 24 * 365 }),
			new CacheableResponsePlugin({ statuses: [0, 200] })
		]
	}),
	'GET'
);

registerRoute(
	/^https:\/\/api\./i,
	new NetworkFirst({
		cacheName: 'api-cache',
		networkTimeoutSeconds: 10,
		plugins: [new ExpirationPlugin({ maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 })]
	}),
	'GET'
);

registerRoute(
	/\.(?:png|jpg|jpeg|svg|gif|webp)$/i,
	new CacheFirst({
		cacheName: 'image-cache',
		plugins: [new ExpirationPlugin({ maxEntries: 200, maxAgeSeconds: 60 * 60 * 24 * 30 })]
	}),
	'GET'
);
