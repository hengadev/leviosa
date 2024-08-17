import { browser } from '$app/environment';
import { writable, get } from 'svelte/store';

export type navState = 'events' | 'photos' | 'profile' | 'home' | 'messagerie';

// the fetch from the local storage is quite slow, but it works for now so...

function createNavbarStore(key: string, initValue: navState) {
    const store = writable<navState>(initValue);
    if (!browser) return store;

    const storedValueStr = localStorage.getItem(key);
    if (storedValueStr != null) store.set(JSON.parse(storedValueStr));

    store.subscribe((val) => {
        if ([null, undefined].includes(val)) {
            localStorage.removeItem(key);
        } else {
            localStorage.setItem(key, JSON.stringify(val));
        }
    });

    window.addEventListener('storage', () => {
        const storedValueStr = localStorage.getItem(key);
        if (storedValueStr == null) return;

        const localValue: navState = JSON.parse(storedValueStr);
        if (localValue !== get(store)) store.set(localValue);
    });
    return store;
}

export const navstate = createNavbarStore('navState', 'home');
