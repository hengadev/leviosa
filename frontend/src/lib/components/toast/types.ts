import { toastStore } from './store.svelte';

export type ToastType = 'success' | 'error' | 'info' | 'warning';

export type Toast = {
	id: number;
	type: ToastType;
	title: string;
	message?: string;
	duration?: number;
};

export type ToastContext = {
	toasts: Array<Toast>;
	success: (title: string, message?: string, duration?: number) => void;
	error: (title: string, message?: string, duration?: number) => void;
	info: (title: string, message?: string, duration?: number) => void;
	warning: (title: string, message?: string, duration?: number) => void;
};

/** The app's toasts: success, error, info and warning messages. */
export function getToastContext(): ToastContext {
	return toastStore;
}
