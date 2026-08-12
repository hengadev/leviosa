// Thin wrapper around the self-hosted, cookieless Plausible tracking script.
//
// The script tag itself is only rendered on public routes — see
// src/routes/(public)/+layout.svelte — so this module is a safe no-op
// everywhere else (client/staff/admin dashboards never load the script,
// window.plausible stays undefined there). It's also a safe no-op in any
// environment where PUBLIC_PLAUSIBLE_SRC isn't set (e.g. local dev), or
// where the script failed to load (ad blockers, offline).
//
// See docs/issues/008-cookieless-analytics.md — no cookies are set by this
// mechanism; Plausible identifies unique visitors from a daily-rotating,
// non-cross-site hash instead.

declare global {
    interface Window {
        plausible?: (
            eventName: string,
            options?: { props?: Record<string, string | number | boolean> }
        ) => void;
    }
}

/**
 * Fire a custom Plausible event (e.g. a booking-funnel step). No-ops if the
 * tracking script hasn't loaded — callers never need to guard this themselves.
 */
export function trackEvent(
    eventName: string,
    props?: Record<string, string | number | boolean>
): void {
    if (typeof window === "undefined" || typeof window.plausible !== "function") return;
    window.plausible(eventName, props ? { props } : undefined);
}
