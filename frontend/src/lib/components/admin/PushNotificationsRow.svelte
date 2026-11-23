<script lang="ts">
	// The "Notifications" row of Cabinet: switches practitioner notifications
	// (Web Push, ADR-0020) on or off for this device. On iPhone they work only
	// in the app added to the home screen (iOS 16.4+), and the permission is
	// asked from the tap on the switch.
	import { onMount } from "svelte";
	import { Bell } from "@lucide/svelte";

	type State = "loading" | "on" | "off" | "busy" | "unavailable" | "install" | "denied" | "unsupported";

	let status = $state<State>("loading");
	let errorMessage = $state("");
	let publicKey = "";

	const hints: Record<State, string> = {
		loading: "Vérification…",
		on: "Activées sur cet appareil : nouvelles réservations.",
		off: "Désactivées sur cet appareil.",
		busy: "Un instant…",
		unavailable: "Les notifications ne sont pas disponibles pour le moment.",
		install: "Ajoutez d'abord l'app à l'écran d'accueil (Partager → Sur l'écran d'accueil)",
		denied:
			"Notifications refusées sur cet appareil. Pour les autoriser : Réglages → Notifications → Leviosa → Autoriser les notifications, puis revenez ici.",
		unsupported: "Ce navigateur ne permet pas les notifications.",
	};

	const checked = $derived(status === "on");
	const actionable = $derived(status === "on" || status === "off");

	function isIOS(): boolean {
		const ua = navigator.userAgent;
		// iPadOS reports itself as a Mac with a touch screen.
		return /iPad|iPhone|iPod/.test(ua) || (/Macintosh/.test(ua) && navigator.maxTouchPoints > 1);
	}

	function isStandalone(): boolean {
		return (
			window.matchMedia("(display-mode: standalone)").matches ||
			(navigator as Navigator & { standalone?: boolean }).standalone === true
		);
	}

	function pushSupported(): boolean {
		return "serviceWorker" in navigator && "PushManager" in window && "Notification" in window;
	}

	// The VAPID public key is base64url; PushManager wants its bytes.
	function keyBytes(base64url: string): Uint8Array<ArrayBuffer> {
		const padded = (base64url + "=".repeat((4 - (base64url.length % 4)) % 4)).replace(/-/g, "+").replace(/_/g, "/");
		const raw = atob(padded);
		const bytes = new Uint8Array(new ArrayBuffer(raw.length));
		for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
		return bytes;
	}

	// Nothing else in the app registers the service worker (the PWA plugin's
	// registration script is never injected), so this row does it: without
	// it there is no registration and the switch stays disabled.
	async function registration(): Promise<ServiceWorkerRegistration | undefined> {
		const existing = await navigator.serviceWorker.getRegistration();
		if (existing) return existing;
		try {
			await navigator.serviceWorker.register("/sw.js", { scope: "/" });
			// pushManager.subscribe needs an active worker.
			return await Promise.race([
				navigator.serviceWorker.ready,
				new Promise<undefined>((resolve) => setTimeout(() => resolve(undefined), 10000)),
			]);
		} catch {
			return undefined;
		}
	}

	async function store(subscription: PushSubscription): Promise<boolean> {
		const res = await fetch("/api/notifications/push/subscription", {
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(subscription.toJSON()),
		});
		return res.ok;
	}

	onMount(async () => {
		try {
			const res = await fetch("/api/notifications/push");
			const info = res.ok ? ((await res.json()) as { enabled?: boolean; public_key?: string }) : {};
			if (!info.enabled || !info.public_key) {
				status = "unavailable";
				return;
			}
			publicKey = info.public_key;
		} catch {
			status = "unavailable";
			return;
		}

		if (isIOS() && !isStandalone()) {
			status = "install";
			return;
		}
		if (!pushSupported()) {
			status = "unsupported";
			return;
		}
		if (Notification.permission === "denied") {
			status = "denied";
			return;
		}

		const reg = await registration();
		if (!reg) {
			status = "unsupported";
			return;
		}
		const subscription = await reg.pushManager.getSubscription();
		if (subscription && Notification.permission === "granted") {
			status = "on";
			// Keep the server in step with this device (best effort).
			store(subscription).catch(() => {});
		} else {
			status = "off";
		}
	});

	async function switchOn() {
		errorMessage = "";
		// Asked first, straight from the tap: iOS refuses it otherwise.
		const permission = await Notification.requestPermission();
		if (permission === "denied") {
			status = "denied";
			return;
		}
		if (permission !== "granted") {
			status = "off";
			return;
		}

		status = "busy";
		let subscription: PushSubscription | undefined;
		try {
			const reg = await registration();
			if (!reg) {
				status = "unsupported";
				return;
			}
			subscription = await reg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: keyBytes(publicKey),
			});
			if (!(await store(subscription))) throw new Error("store failed");
			status = "on";
		} catch {
			await subscription?.unsubscribe().catch(() => {});
			errorMessage = "Impossible d'activer les notifications. Réessayez.";
			status = "off";
		}
	}

	async function switchOff() {
		errorMessage = "";
		status = "busy";
		try {
			const reg = await registration();
			const subscription = await reg?.pushManager.getSubscription();
			if (subscription) {
				const res = await fetch("/api/notifications/push/subscription", {
					method: "DELETE",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({ endpoint: subscription.endpoint }),
				});
				if (!res.ok) throw new Error("delete failed");
				await subscription.unsubscribe();
			}
			status = "off";
		} catch {
			errorMessage = "Impossible de désactiver les notifications. Réessayez.";
			status = "on";
		}
	}

	function toggle() {
		if (status === "off") switchOn();
		else if (status === "on") switchOff();
	}
</script>

<button
	type="button"
	role="switch"
	aria-checked={checked}
	aria-describedby="push-notifications-hint"
	disabled={!actionable}
	onclick={toggle}
	class="flex min-h-11 w-full items-center gap-3 px-4 py-3 text-left transition-colors enabled:cursor-pointer enabled:hover:bg-muted disabled:cursor-default"
>
	<Bell size={20} strokeWidth={1.5} class="shrink-0 text-muted-foreground" />
	<span class="min-w-0 flex-1">
		<span class="block font-medium text-foreground">Notifications</span>
		<span id="push-notifications-hint" class="block text-xs text-muted-foreground" aria-live="polite">
			{hints[status]}
		</span>
		{#if errorMessage}
			<span class="mt-1 block text-xs text-red-600" role="alert">{errorMessage}</span>
		{/if}
	</span>
	<span
		aria-hidden="true"
		class="relative inline-flex h-7 w-12 shrink-0 items-center rounded-full border transition-colors {checked
			? 'border-transparent bg-foreground'
			: 'border-border-card bg-muted'} {actionable ? '' : 'opacity-50'}"
	>
		<span
			class="inline-block size-5 rounded-full bg-background shadow transition-transform {checked
				? 'translate-x-6'
				: 'translate-x-1'}"
		></span>
	</span>
</button>
