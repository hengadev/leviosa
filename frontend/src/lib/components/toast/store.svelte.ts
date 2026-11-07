import { browser } from '$app/environment';
import type { Toast, ToastContext, ToastType } from './types';

// One list of toasts for the whole app, read by the container the root
// layout renders. A module, not a context: the container is a sibling of
// the pages, so a context it set would never reach them.

let toasts = $state<Toast[]>([]);
let nextId = 0;

function removeToast(id: number) {
	toasts = toasts.filter((t) => t.id !== id);
}

function addToast(type: ToastType, title: string, message?: string, duration = 4000) {
	// Toasts answer a user's action in the browser; on the server the list
	// would be shared between requests.
	if (!browser) return;
	const id = nextId++;
	toasts.push({ id, type, title, message, duration });
	if (duration > 0) setTimeout(() => removeToast(id), duration);
}

export const toastStore: ToastContext & { remove: (id: number) => void } = {
	get toasts() {
		return toasts;
	},
	success: (title, message, duration) => addToast('success', title, message, duration),
	error: (title, message, duration) => addToast('error', title, message, duration),
	info: (title, message, duration) => addToast('info', title, message, duration),
	warning: (title, message, duration) => addToast('warning', title, message, duration),
	remove: removeToast,
};
