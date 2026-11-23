// Practitioner notifications (Web Push, ADR-0020), bundled into the service
// worker (pwa/sw.js).
// The backend sends { title, body, url, tag }: never a full name, a phone,
// an email, an amount or the door code. A tap opens or focuses `url`, a path
// on this origin.

const DEFAULT_URL = '/admin';

function samePath(url) {
	// Only paths on this origin: a payload can never send the app elsewhere.
	return typeof url === 'string' && url.startsWith('/') && !url.startsWith('//') ? url : DEFAULT_URL;
}

self.addEventListener('push', (event) => {
	let data = {};
	try {
		data = event.data ? event.data.json() : {};
	} catch {
		data = { body: event.data ? event.data.text() : '' };
	}
	const title = data.title || 'Leviosa';
	event.waitUntil(
		self.registration.showNotification(title, {
			body: data.body || '',
			tag: data.tag || undefined,
			icon: '/pwa-admin-192x192.png',
			badge: '/pwa-admin-192x192.png',
			data: { url: samePath(data.url) }
		})
	);
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const target = new URL(samePath(event.notification.data && event.notification.data.url), self.location.origin).href;

	event.waitUntil(
		(async () => {
			const windows = await self.clients.matchAll({ type: 'window', includeUncontrolled: true });
			const open = windows.find((client) => new URL(client.url).origin === self.location.origin);
			if (open) {
				// Focus is a courtesy: never let it stop the page from opening.
				await open.focus().catch(() => {});
				if (open.url === target) return;
				if ('navigate' in open) {
					try {
						await open.navigate(target);
						return;
					} catch {
						// An uncontrolled window cannot be navigated: open the page.
					}
				}
			}
			await self.clients.openWindow(target);
		})()
	);
});
