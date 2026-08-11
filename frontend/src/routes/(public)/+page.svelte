<script lang="ts">
    import Hero from "./_hero.svelte";
    import How from "./_how.svelte";
    import Stats from "./_stats.svelte";
    import Services from "./_services.svelte";
    import Why from "./_why.svelte";
    import Team from "./_team.svelte";
    import Testimonies from "./_testimonies.svelte";
    import Faq from "./_faq.svelte";
    import Ready from "./_ready.svelte";
    import type { PageProps } from "./$types";
    import { page } from "$app/state";
    import { jsonLdScriptTag } from "$lib/utils/seo";

    let { data }: PageProps = $props();
    let { featuredService, featuredPrice, teamMembers } = data;

    const description =
        "Réservez instantanément une consultation avec un praticien vérifié : ostéopathie, psychologie, coaching. Créneaux disponibles, réservation en quelques secondes.";

    const localBusinessJsonLd = jsonLdScriptTag({
        "@context": "https://schema.org",
        "@type": "LocalBusiness",
        name: "Leviosa Care",
        url: page.url.origin,
        description,
        image: `${page.url.origin}/pwa-512x512.png`,
        email: "contact@leviosa.care",
        sameAs: ["https://www.instagram.com/leviosa_care"],
    });
</script>

<svelte:head>
    <title>Leviosa Care | Réservez un soin bien-être en ligne</title>
    <meta name="description" content={description} />
    {@html localBusinessJsonLd}
</svelte:head>

<main class="relative overflow-hidden">
    <Hero />
    <How />
    <Services service={featuredService} price={featuredPrice} />
    <Stats />
    <Why />
    <Team members={teamMembers} />
    <Testimonies />
    <Faq />
    <Ready />
</main>
