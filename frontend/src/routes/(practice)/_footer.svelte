<script lang="ts">
    import { dictionary, type Locale } from "$lib/i18n/practice";
    import type { PracticePageData } from "./practice";

    let { locale, data }: { locale: Locale; data: PracticePageData } = $props();

    const t = $derived(dictionary[locale]);
</script>

<footer class="bg-neutral-950 px-4 py-10 text-sm text-neutral-400">
    <div class="mx-auto flex max-w-5xl flex-wrap items-center gap-x-6 gap-y-3">
        <a href="#top" class="font-bold text-white focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none">
            leviosa<span class="text-neutral-400">.care</span>
        </a>
        {#if data.address}
            <span>{data.address}</span>
        {/if}
        <a href="mailto:{data.email}" class="hover:text-white focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none">
            {data.email}
        </a>
        <span class="flex gap-4 sm:ml-auto">
            <!-- The mentions légales page arrives with issue 001; it will join these links. -->
            <a href="/legal/privacy" class="hover:text-white focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none">{t.privacy}</a>
            <a href="/legal/terms" class="hover:text-white focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none">{t.terms}</a>
        </span>
    </div>
</footer>
