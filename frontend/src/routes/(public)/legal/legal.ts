// Shared facts of the legal pages (issue 001). Interim content on germinal's
// model: only the company's name, its contact email and the host are given
// until the business details (SIRET, address, publication director, ADELI or
// RPPS number) arrive and the text has been reviewed.

export const COMPANY_NAME = "Leviosa Care";
export const CONTACT_EMAIL = "contact@leviosa.care";

export const HOST = {
    name: "Hetzner Online GmbH",
    address: "Industriestr. 25, 91710 Gunzenhausen, Allemagne",
};

/** One section of a legal page: a heading, paragraphs, then an optional list. */
export type LegalSection = {
    title: string;
    paragraphs?: string[];
    items?: string[];
    /** Paragraphs shown after the list. */
    after?: string[];
};
