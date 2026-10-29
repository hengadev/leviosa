// FR/EN dictionary for the public practice page only (issue 009). The admin
// and partner areas stay French-only by decision of the single-practitioner
// PRD; if the app ever grows real i18n, this module is the seam to lift.
//
// Static page copy (timeline steps, "good to know" bodies) is PLACEHOLDER
// copy ported from the chosen prototype (variant D): the practitioner's real
// wording replaces it before launch (issue 021).

export type Locale = "fr" | "en";

export const locales: Locale[] = ["fr", "en"];

/** Locale for Intl formatting: FR conventions for fr, GB for en (€ still EUR). */
const tag = (locale: Locale) => (locale === "fr" ? "fr-FR" : "en-GB");

/** Slots are computed and shown in the practice's time zone, wherever the visitor is. */
export const PRACTICE_TIME_ZONE = "Europe/Paris";

const timeZone = PRACTICE_TIME_ZONE;

export const fmt = {
    price: (cents: number, locale: Locale) =>
        new Intl.NumberFormat(tag(locale), {
            style: "currency",
            currency: "EUR",
            maximumFractionDigits: cents % 100 === 0 ? 0 : 2,
        }).format(cents / 100),
    time: (date: Date, locale: Locale) =>
        new Intl.DateTimeFormat(tag(locale), { hour: "2-digit", minute: "2-digit", timeZone }).format(date),
    weekdayShort: (date: Date, locale: Locale) =>
        new Intl.DateTimeFormat(tag(locale), { weekday: "short", timeZone }).format(date),
    dayMonth: (date: Date, locale: Locale) =>
        new Intl.DateTimeFormat(tag(locale), { day: "numeric", month: "short", timeZone }).format(date),
    long: (date: Date, locale: Locale) =>
        new Intl.DateTimeFormat(tag(locale), { weekday: "long", day: "numeric", month: "long", timeZone }).format(date),
    month: (date: Date, locale: Locale) =>
        new Intl.DateTimeFormat(tag(locale), { month: "long", year: "numeric", timeZone }).format(date),
};

/** French weekdays and months are lowercase; a label starting a line wants a capital. */
export const cap = (text: string) => text.charAt(0).toUpperCase() + text.slice(1);

export const strings = {
    fr: {
        // header
        languageSwitch: "Langue",
        // hero
        videoMissing: "Vidéo du cabinet à fournir",
        bookSession: "Réserver une séance",
        skipToBooking: "Aller à la réservation",
        duration: "Durée",
        price: "Tarif",
        address: "Adresse",
        payment: "Paiement",
        paymentValue: "En ligne, par carte",
        minutes: (n: number) => `${n} min`,
        // booking card
        intro: (h: number) =>
            `Choisissez un créneau parmi les disponibilités du cabinet, puis réglez la séance en ligne. Annulation gratuite jusqu'à ${h} h avant.`,
        pickDay: "Choisissez un jour pour voir les horaires.",
        noSlots: "Aucun créneau ce jour-là.",
        slotsLeft: (n: number) => (n === 1 ? "1 créneau" : `${n} créneaux`),
        prevMonth: "Mois précédent",
        nextMonth: "Mois suivant",
        timeLegend: "Heure",
        timeZoneNote: "Heures de Paris",
        noSlotsYet: "Aucun créneau n'est ouvert à la réservation pour le moment.",
        slotsUnavailable: "Les créneaux ne peuvent pas être affichés pour le moment. Réessayez un peu plus tard.",
        // details step (stub until the booking form, issue 014)
        detailsHeading: "Vos coordonnées",
        changeSlot: "Changer de créneau",
        detailsSoon: "La saisie de vos coordonnées et le paiement en ligne arrivent bientôt.",
        // footer
        legal: "Mentions légales",
        privacy: "Confidentialité",
        terms: "CGV",
        backToTop: "Revenir en haut de page",
        weekdays: ["lun.", "mar.", "mer.", "jeu.", "ven.", "sam.", "dim."],
    },
    en: {
        languageSwitch: "Language",
        videoMissing: "Practice video to be provided",
        bookSession: "Book a session",
        skipToBooking: "Go to booking",
        duration: "Duration",
        price: "Price",
        address: "Address",
        payment: "Payment",
        paymentValue: "Online, by card",
        minutes: (n: number) => `${n} min`,
        intro: (h: number) =>
            `Pick one of the practice's open slots, then pay for the session online. Free cancellation up to ${h} h before.`,
        pickDay: "Pick a day to see the times.",
        noSlots: "No slots that day.",
        slotsLeft: (n: number) => (n === 1 ? "1 slot" : `${n} slots`),
        prevMonth: "Previous month",
        nextMonth: "Next month",
        timeLegend: "Time",
        timeZoneNote: "Paris time",
        noSlotsYet: "No slots are open for booking right now.",
        slotsUnavailable: "The slots can't be shown right now. Please try again a little later.",
        detailsHeading: "Your details",
        changeSlot: "Change slot",
        detailsSoon: "Entering your details and paying online is coming soon.",
        legal: "Legal notice",
        privacy: "Privacy",
        terms: "Terms",
        backToTop: "Back to top",
        weekdays: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
    },
};

/** The typed dictionary: the French shape is the contract every locale fulfils. */
export type Strings = (typeof strings)["fr"];

export const dictionary: Record<Locale, Strings> = strings;

/** The session timeline and practical points. PLACEHOLDER copy from the prototype. */
export const infoBlocks = {
    fr: {
        heading: "Une séance, minute par minute",
        intro: "Ce qui se passe pendant l'heure, du premier échange aux conseils pour la suite.",
        steps: [
            { from: 0, to: 10, title: "Échange", body: "Vos douleurs, votre activité, vos antécédents." },
            { from: 10, to: 20, title: "Bilan", body: "Tests de mobilité pour trouver d'où vient la tension." },
            { from: 20, to: 50, title: "Traitement", body: "Techniques manuelles adaptées à ce que le bilan a montré." },
            { from: 50, to: 60, title: "Conseils", body: "Exercices et gestes pour la suite." },
        ],
        practicalHeading: "Bon à savoir",
        practical: [
            {
                title: "Ce qu'il faut prévoir",
                body: "Une tenue souple dans laquelle vous êtes à l'aise. Si vous avez des examens récents (radio, IRM), apportez-les.",
            },
            {
                title: "Remboursement",
                body: "L'ostéopathie n'est pas remboursée par la Sécurité sociale, mais beaucoup de mutuelles prennent en charge une partie des séances. Votre ostéopathe vous remet la facture pour votre mutuelle au cabinet.",
            },
            {
                title: "Annulation et retard",
                body: "Annulation gratuite jusqu'à 24 h avant, depuis le lien reçu par e-mail. En cas de retard, la séance se termine à l'heure prévue.",
            },
        ],
    },
    en: {
        heading: "A session, minute by minute",
        intro: "What happens during the hour, from the first conversation to advice for afterwards.",
        steps: [
            { from: 0, to: 10, title: "Talk", body: "Your pain, your activity, your history." },
            { from: 10, to: 20, title: "Assessment", body: "Mobility tests to find where the tension comes from." },
            { from: 20, to: 50, title: "Treatment", body: "Manual techniques fitted to what the assessment showed." },
            { from: 50, to: 60, title: "Advice", body: "Exercises and habits for afterwards." },
        ],
        practicalHeading: "Good to know",
        practical: [
            {
                title: "What to bring",
                body: "Loose clothes you're comfortable in. If you have recent scans (X-ray, MRI), bring them along.",
            },
            {
                title: "Reimbursement",
                body: "Osteopathy isn't covered by French national health insurance, but many top-up insurers cover part of each session. Your osteopath gives you the invoice for your insurer at the practice.",
            },
            {
                title: "Cancellation and lateness",
                body: "Free cancellation up to 24 h before, from the link in your email. If you're late, the session still ends at the planned time.",
            },
        ],
    },
};

export type InfoBlock = (typeof infoBlocks)["fr"];
