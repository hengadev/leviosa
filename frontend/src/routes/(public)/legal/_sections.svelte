<script lang="ts">
    import type { LegalSection } from "./legal";

    let { title, updated, intro, sections }: { title: string; updated: string; intro?: string; sections: LegalSection[] } =
        $props();

    const lastUpdated = $derived(
        new Intl.DateTimeFormat("fr-FR", { day: "numeric", month: "long", year: "numeric", timeZone: "UTC" }).format(
            new Date(updated),
        ),
    );
</script>

<h1 class="font-display text-3xl md:text-4xl">{title}</h1>
<p class="mt-3 mb-12 text-sm text-muted-foreground">Dernière mise à jour : {lastUpdated}</p>

<div class="grid gap-10 leading-relaxed text-foreground-alt">
    {#if intro}
        <p>{intro}</p>
    {/if}
    {#each sections as section, i (section.title)}
        <section class="grid gap-3">
            <h2 class="text-xl font-semibold text-foreground">{i + 1}. {section.title}</h2>
            {#each section.paragraphs ?? [] as paragraph (paragraph)}
                <p>{paragraph}</p>
            {/each}
            {#if section.items}
                <ul class="grid list-disc gap-2 pl-6">
                    {#each section.items as item (item)}
                        <li>{item}</li>
                    {/each}
                </ul>
            {/if}
            {#each section.after ?? [] as paragraph (paragraph)}
                <p>{paragraph}</p>
            {/each}
        </section>
    {/each}
</div>
