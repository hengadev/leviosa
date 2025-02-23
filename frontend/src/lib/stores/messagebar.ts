import { browser } from '$app/environment';
import { writable, get } from 'svelte/store';

import type { MessageState } from '$lib/types';

function createEventbarStore(key: string, initValue: MessageState) {
	const store = writable<MessageState>(initValue);
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

		const localValue: MessageState = JSON.parse(storedValueStr);
		if (localValue !== get(store)) store.set(localValue);
	});
	return store;
}

export const messagestate = createEventbarStore('messageState', 'Conversations');
