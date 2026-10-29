// Phone numbers on the practice page's booking form (issue 014). The backend
// accepts French numbers only (10 digits starting 01–07, no spaces), so the
// form normalises what people actually type before checking it.

/** The backend's format: 0 then 1–7, then 8 digits. */
export const FRENCH_PHONE = /^0[1-7]\d{8}$/;

/** Accepts spaces, dots, dashes, brackets and a +33 / 0033 prefix; answers 0XXXXXXXXX. */
export function normalizePhone(raw: string): string {
    let phone = raw.replace(/[\s.\-()]/g, '');
    if (phone.startsWith('+33')) phone = '0' + phone.slice(3);
    else if (phone.startsWith('0033')) phone = '0' + phone.slice(4);
    return phone;
}

export const isFrenchPhone = (raw: string) => FRENCH_PHONE.test(normalizePhone(raw));
