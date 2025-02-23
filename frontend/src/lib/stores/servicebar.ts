import { browser } from '$app/environment';
import { writable, get } from 'svelte/store';

import type { ServiceState } from '$lib/types';

function createNavbarStore(key: string, initValue: ServiceState) {
	const store = writable<ServiceState>(initValue);
	if (!browser) return store;

	const storedValueStr = localStorage.getItem(key);
	if (storedValueStr != null) store.set(JSON.parse(storedValueStr));

	store.subscribe((val) => {
		if ([null, undefined].includes(val)) localStorage.removeItem(key);
		else localStorage.setItem(key, JSON.stringify(val));
	});

	window.addEventListener('storage', () => {
		const storedValueStr = localStorage.getItem(key);
		if (storedValueStr == null) return;

		const localValue: ServiceState = JSON.parse(storedValueStr);
		if (localValue !== get(store)) store.set(localValue);
	});
	return store;
}

export const servicestate = createNavbarStore('serviceState', 'A propos');
