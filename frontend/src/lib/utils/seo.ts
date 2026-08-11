/**
 * Small shared helpers for per-page SEO metadata: trimming long copy down
 * to a reasonable meta-description length, and safely serializing JSON-LD
 * structured data for embedding inside a <script type="application/ld+json">
 * tag in <svelte:head>.
 */

/** Truncate text to a meta-description-friendly length, breaking on a word boundary. */
export function truncateDescription(text: string, maxLength = 155): string {
    const clean = text.trim().replace(/\s+/g, " ");
    if (clean.length <= maxLength) return clean;
    const cut = clean.slice(0, maxLength);
    const lastSpace = cut.lastIndexOf(" ");
    return `${cut.slice(0, lastSpace > 0 ? lastSpace : maxLength)}…`;
}

/**
 * Serialize a JSON-LD object for safe inline embedding. Escapes "</" so a
 * value containing "</script>" can't prematurely close the tag.
 */
export function serializeJsonLd(data: Record<string, unknown>): string {
    return JSON.stringify(data).replace(/</g, "\\u003c");
}

/**
 * Build a full `<script type="application/ld+json">` tag as an HTML string,
 * for use with Svelte's `{@html ...}` inside <svelte:head>. Svelte does not
 * evaluate `{expression}` mustaches placed as text inside a nested <script>
 * element (script content is parsed as raw text, per the HTML spec), so the
 * tag has to be injected as a pre-built HTML string instead.
 */
export function jsonLdScriptTag(data: Record<string, unknown>): string {
    return `<script type="application/ld+json">${serializeJsonLd(data)}<\/script>`;
}
