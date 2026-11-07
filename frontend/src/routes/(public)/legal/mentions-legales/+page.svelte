<script lang="ts">
    import Sections from "../_sections.svelte";
    import { COMPANY_NAME, CONTACT_EMAIL, HOST, type LegalSection } from "../legal";

    const sections: LegalSection[] = [
        {
            title: "Éditeur du site",
            paragraphs: [
                `Le présent site est édité par ${COMPANY_NAME}.`,
                `E-mail : ${CONTACT_EMAIL}`,
                `Directeur de la publication : ${COMPANY_NAME}`,
            ],
        },
        {
            title: "Hébergement",
            paragraphs: [`Le présent site est hébergé par ${HOST.name}, ${HOST.address}.`],
        },
        {
            title: "Propriété intellectuelle",
            paragraphs: [
                `L'ensemble des contenus du présent site (textes, images, vidéos, logos) est la propriété de ${COMPANY_NAME} et est protégé par le droit d'auteur. Toute reproduction, représentation, modification ou adaptation, totale ou partielle, sans autorisation écrite préalable de ${COMPANY_NAME}, est interdite.`,
            ],
        },
        {
            title: "Données personnelles",
            paragraphs: [
                "La collecte et le traitement de vos données personnelles sont décrits dans la politique de confidentialité, accessible en bas de chaque page.",
            ],
        },
    ];
</script>

<svelte:head>
    <title>Mentions légales | Leviosa Care</title>
    <meta name="description" content="Mentions légales du site Leviosa Care : éditeur, hébergeur et propriété intellectuelle." />
</svelte:head>

<Sections title="Mentions légales" updated="2026-11-07" {sections} />
