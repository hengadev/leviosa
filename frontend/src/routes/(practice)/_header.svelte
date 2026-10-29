<script lang="ts">
    import { dictionary, type Locale } from "$lib/i18n/practice";

    // Header of the practice page: the wordmark and the FR/EN switch. The
    // switch navigates between the page's two locales (`/` is French, `/en`
    // English) rather than swapping text in place, so each locale has one
    // address, one <html lang> and one shareable URL.
    let { locale }: { locale: Locale } = $props();

    const t = $derived(dictionary[locale]);
    const hrefFor = (option: Locale) => (option === "fr" ? "/" : "/en");
</script>

<header class="sticky top-0 z-40 border-b border-neutral-800 bg-neutral-950/90 text-white backdrop-blur">
    <div class="mx-auto flex w-full max-w-5xl items-center gap-3 px-4 py-3.5">
        <a href="#top" class="text-lg font-bold tracking-tight focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none">
            leviosa<span class="text-neutral-400">.care</span>
        </a>
        <nav class="ml-auto flex gap-1" aria-label={t.languageSwitch}>
            {#each ["fr", "en"] as const as option (option)}
                <a
                    href={hrefFor(option)}
                    lang={option}
                    aria-current={locale === option ? "page" : undefined}
                    class="min-w-9 rounded-md px-2 py-1 text-sm font-semibold uppercase transition-colors focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none {locale === option
                        ? 'bg-white text-neutral-950'
                        : 'text-neutral-400 hover:text-white'}"
                >
                    {option}
                </a>
            {/each}
        </nav>
    </div>
</header>
