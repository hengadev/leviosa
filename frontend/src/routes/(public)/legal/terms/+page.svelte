<script lang="ts">
    import Sections from "../_sections.svelte";
    import { COMPANY_NAME, CONTACT_EMAIL, type LegalSection } from "../legal";

    // The cancellation section states the current policy: no cancellation
    // tier, so every cancellation by the client is refunded in full. If the
    // product gets tiers (ADR-0014), this section must list them.
    const sections: LegalSection[] = [
        {
            title: "Objet",
            paragraphs: [
                `Les présentes conditions générales de vente (« CGV ») régissent la réservation et le paiement en ligne des séances proposées sur le site ${COMPANY_NAME}. En réservant une séance, vous acceptez ces conditions sans réserve.`,
            ],
        },
        {
            title: "Prix",
            paragraphs: [
                "Le prix de la séance, sa durée et le montant de l'acompte sont indiqués sur le site au moment de la réservation, en euros, toutes taxes comprises. Le prix applicable est celui affiché lors de la réservation.",
            ],
        },
        {
            title: "Réservation",
            paragraphs: [
                "Vous choisissez un créneau parmi ceux proposés, puis renseignez vos coordonnées. Le créneau vous est réservé pendant le délai de paiement indiqué à l'écran ; passé ce délai sans paiement, il est libéré.",
                "La réservation est confirmée une fois le paiement en ligne accepté. Un e-mail de confirmation vous est alors envoyé ; son lien vous permet de consulter ou d'annuler votre réservation.",
            ],
        },
        {
            title: "Paiement et acompte",
            paragraphs: [
                "Le paiement en ligne se fait par carte bancaire, de manière sécurisée, par l'intermédiaire de Stripe. Vos données bancaires ne transitent pas par le cabinet.",
                "Lorsque le site l'indique, seul un acompte est payé en ligne au moment de la réservation. Cette somme est un acompte au sens de l'article L214-1 du Code de la consommation, et non des arrhes : elle constitue un premier versement sur le prix de la séance. Le solde est réglé au cabinet, par carte bancaire, le jour de la séance.",
            ],
        },
        {
            title: "Annulation et remboursement",
            items: [
                "Annulation par le client : vous pouvez annuler votre réservation jusqu'au début de la séance, depuis le lien reçu par e-mail. Les sommes payées en ligne vous sont alors intégralement remboursées.",
                "Annulation par le praticien : si le praticien annule une séance, les sommes payées en ligne vous sont intégralement remboursées et vous en êtes informé(e) par e-mail.",
                "Absence : en cas d'absence à la séance sans annulation préalable, les sommes payées en ligne ne sont pas remboursées.",
                "Retard : en cas de retard, la séance se termine à l'heure prévue.",
            ],
            after: [
                "Les remboursements sont effectués sur le moyen de paiement utilisé, sous 5 à 10 jours ouvrés selon votre banque.",
            ],
        },
        {
            title: "Droit de rétractation",
            paragraphs: [
                "Vous disposez d'un délai de 14 jours à compter de la réservation pour vous rétracter, en annulant votre réservation depuis le lien reçu par e-mail : les sommes payées en ligne vous sont intégralement remboursées. Si la séance a lieu avant la fin de ce délai, vous demandez expressément, en réservant, qu'elle soit exécutée à la date choisie.",
            ],
        },
        {
            title: "Responsabilité",
            paragraphs: [
                `La responsabilité de ${COMPANY_NAME} au titre de la réservation est limitée au montant payé pour la séance concernée.`,
            ],
        },
        {
            title: "Réclamations et droit applicable",
            paragraphs: [
                `Pour toute question ou réclamation, écrivez à ${CONTACT_EMAIL}.`,
                "Les présentes conditions sont soumises au droit français. En cas de litige, une solution amiable sera recherchée avant toute action ; à défaut, les tribunaux compétents sont ceux désignés par les règles de droit commun.",
            ],
        },
    ];
</script>

<svelte:head>
    <title>Conditions générales de vente | Leviosa Care</title>
    <meta name="description" content="Conditions générales de vente des séances réservées sur Leviosa Care : prix, acompte, paiement, annulation et remboursement." />
</svelte:head>

<Sections title="Conditions générales de vente" updated="2026-11-07" {sections} />
