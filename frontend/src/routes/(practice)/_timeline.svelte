<script lang="ts">
    import { onMount } from "svelte";
    import { infoBlocks, type Locale } from "$lib/i18n/practice";

    // "A session, minute by minute": the hour as a proportional bar, then the
    // steps under it sharing the same column scale. Static placeholder copy
    // from the practice dictionary (issue 021 replaces it with the
    // practitioner's real wording).
    let { locale }: { locale: Locale } = $props();
    const info = $derived(infoBlocks[locale]);
    /** The longest step (the treatment) is the one drawn in ink. */
    const main = $derived(info.steps.reduce((best, step, index, all) => (step.to - step.from > all[best].to - all[best].from ? index : best), 0));

    // The bar is drawn full by default. If it starts below the fold, it is
    // emptied and fills segment by segment as it scrolls into view: the
    // section's one moving piece.
    let bar: HTMLElement;
    let drawn = $state(true);
    onMount(() => {
        if (matchMedia("(prefers-reduced-motion: reduce)").matches) return;
        if (bar.getBoundingClientRect().top < innerHeight) return;
        drawn = false;
        const observer = new IntersectionObserver(
            ([entry]) => {
                if (!entry.isIntersecting) return;
                drawn = true;
                observer.disconnect();
            },
            { threshold: 0.6 },
        );
        observer.observe(bar);
        return () => observer.disconnect();
    });

    /** Columns sized by minutes, so the bar and the text under it share one scale. */
    const columns = $derived(info.steps.map((step) => `${step.to - step.from}fr`).join(" "));
</script>

<div class="space-y-8">
    <!-- The hour as a bar: each segment's width is its share of the session. -->
    <!-- Phones: a flex row; from sm, the same columns and gap as the steps below, so each segment
         sits exactly over its text. -->
    <div bind:this={bar} class="flex h-3 gap-1 sm:grid sm:grid-cols-[var(--cols)] sm:gap-4" style="--cols: {columns}" aria-hidden="true">
        {#each info.steps as step, index (step.title)}
            <span
                class="h-full origin-left rounded-full transition-transform duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] {index === main
                    ? 'bg-neutral-950'
                    : 'bg-neutral-300'} {drawn ? 'scale-x-100' : 'scale-x-0'}"
                style="flex: {step.to - step.from} 1 0%; transition-delay: {drawn ? index * 160 : 0}ms"
            ></span>
        {/each}
    </div>

    <ol class="grid gap-6 sm:grid-cols-[var(--cols)] sm:gap-4" style="--cols: {columns}">
        {#each info.steps as step, index (step.title)}
            <li class="grid grid-cols-[5.5rem_1fr] gap-4 sm:block">
                <p class="text-sm font-semibold text-neutral-600 tabular-nums">
                    {step.from}–{step.to}&nbsp;<span class="text-xs">min</span>
                </p>
                <div class="sm:mt-2">
                    <h4 class="text-lg font-bold tracking-tight {index === main ? '' : 'text-neutral-800'}">{step.title}</h4>
                    <p class="mt-1 text-sm leading-relaxed text-pretty text-neutral-600">{step.body}</p>
                </div>
            </li>
        {/each}
    </ol>
</div>
