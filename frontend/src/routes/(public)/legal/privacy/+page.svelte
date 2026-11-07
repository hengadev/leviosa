<script lang="ts">
    import Sections from "../_sections.svelte";
    import { COMPANY_NAME, CONTACT_EMAIL, type LegalSection } from "../legal";

    const sections: LegalSection[] = [
        {
            title: "Responsable du traitement",
            paragraphs: [`Le responsable du traitement est ${COMPANY_NAME}, joignable à ${CONTACT_EMAIL}.`],
        },
        {
            title: "Données collectées",
            paragraphs: ["Nous collectons les données suivantes :"],
            items: [
                "Réservation : prénom, nom, adresse e-mail, numéro de téléphone, créneau choisi — collectés lorsque vous réservez une séance, pour vous envoyer la confirmation, le rappel et vous joindre en cas d'imprévu.",
                "Paiement : le montant payé, remboursé, et la référence du paiement. Vos données bancaires sont saisies directement auprès de Stripe : elles ne sont ni transmises au cabinet ni conservées par nous.",
                "Compte : si vous créez un compte, les informations demandées à l'inscription (identité et coordonnées).",
                "Données techniques : adresse IP et type de navigateur, collectés automatiquement pour la sécurité du site et la prévention de la fraude.",
            ],
            after: [
                "Le site ne collecte aucune donnée de santé : aucun motif de consultation, symptôme ni antécédent ne vous est demandé. Les notes cliniques de votre praticien sont tenues dans ses propres outils, en dehors de ce site.",
            ],
        },
        {
            title: "Base légale",
            items: [
                "Exécution du contrat : gérer votre réservation, son paiement, son annulation et son remboursement.",
                "Intérêt légitime : assurer la sécurité du site et prévenir la fraude.",
                "Obligation légale : conserver les pièces comptables.",
            ],
        },
        {
            title: "Durées de conservation",
            items: [
                "Réservations : le temps de la relation, puis au plus 3 ans après votre dernière séance.",
                "Pièces comptables (montants, dates, références de paiement) : 10 ans, comme l'exige la loi.",
                "Journaux techniques du serveur : 30 jours.",
                "Sauvegardes : copies chiffrées, stockées dans l'Union européenne et supprimées après 90 jours.",
            ],
        },
        {
            title: "Destinataires et sous-traitants",
            paragraphs: [
                `Vos données ne sont accessibles qu'à ${COMPANY_NAME}, à votre praticien pour les séances que vous réservez, et aux prestataires suivants, qui les traitent pour notre compte :`,
            ],
            items: [
                "Stripe (paiements) — Stripe Payments Europe, Irlande.",
                "Amazon Web Services (chiffrement des données et stockage de fichiers) — Francfort, Allemagne.",
                "Brevo et Zoho (envoi des e-mails).",
                "Twilio (SMS) — traitement en Irlande.",
                "Hetzner (hébergement du site et de la base de données) — Allemagne.",
                "Cloudflare (nom de domaine et protection du réseau).",
            ],
            after: [
                "Certains de ces prestataires appartiennent à des groupes établis aux États-Unis. Tout transfert de données hors de l'Union européenne est encadré par le Data Privacy Framework UE–États-Unis ou par les clauses contractuelles types de la Commission européenne.",
            ],
        },
        {
            title: "Vos droits",
            paragraphs: ["Vous disposez des droits suivants sur vos données :"],
            items: [
                "accès à vos données ;",
                "rectification des données inexactes ;",
                "effacement de vos données ;",
                "portabilité de vos données ;",
                "opposition au traitement fondé sur l'intérêt légitime ;",
                "limitation du traitement.",
            ],
            after: [
                `Pour exercer l'un de ces droits, écrivez à ${CONTACT_EMAIL}.`,
                "Si vous estimez que vos droits ne sont pas respectés, vous pouvez introduire une réclamation auprès de la CNIL (www.cnil.fr).",
            ],
        },
        {
            title: "Cookies",
            paragraphs: [
                "Le site n'utilise que des cookies strictement nécessaires à son fonctionnement : le cookie de session qui vous garde connecté(e) à votre compte, et, au moment du paiement, ceux que Stripe dépose pour sécuriser la transaction. Ils ne sont pas soumis à consentement.",
                "Le site n'utilise aucun cookie de mesure d'audience ni de publicité.",
            ],
        },
    ];
</script>

<svelte:head>
    <title>Politique de confidentialité | Leviosa Care</title>
    <meta name="description" content="Politique de confidentialité de Leviosa Care : comment nous collectons, utilisons et protégeons vos données personnelles." />
</svelte:head>

<Sections
    title="Politique de confidentialité"
    updated="2026-11-07"
    intro={`${COMPANY_NAME} s'engage à protéger vos données personnelles. Cette politique explique quelles données nous collectons, pourquoi, combien de temps nous les conservons et quels sont vos droits, conformément au Règlement général sur la protection des données (RGPD).`}
    {sections}
/>
