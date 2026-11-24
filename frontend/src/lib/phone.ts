// Phone numbers on the practice page's booking form (issue 014). Clients pick
// their country's calling code and type the number as they usually write it;
// libphonenumber-js checks it against that country.
//
// Stored form, which the backend accepts (validation.ValidateGuestPhone): a
// French number stays in the compact domestic form the practice has always
// stored (0612345678), any other country is E.164 (+447400123456).

import { getCountries, getCountryCallingCode, getExampleNumber, parsePhoneNumberFromString, type CountryCode } from 'libphonenumber-js';
import examples from 'libphonenumber-js/mobile/examples';

export type { CountryCode };

export const DEFAULT_PHONE_COUNTRY: CountryCode = 'FR';

export type PhoneCountry = { code: CountryCode; name: string; callingCode: string };

/** Every country with its calling code, sorted by its name in the page's language. */
export function phoneCountries(locale: string): PhoneCountry[] {
    let names: Intl.DisplayNames | null = null;
    try {
        names = new Intl.DisplayNames([locale], { type: 'region' });
    } catch {
        // no region names in this runtime: the codes stand in
    }
    return getCountries()
        .map((code) => ({ code, name: names?.of(code) ?? code, callingCode: getCountryCallingCode(code) }))
        .sort((a, b) => a.name.localeCompare(b.name, locale));
}

/**
 * A mobile number of that country for the placeholder, written as it follows
 * the calling code: no national prefix (+33 then 6 12 34 56 78, not 06…).
 * A number typed with its national prefix is still accepted.
 */
export function examplePhone(country: CountryCode): string {
    const example = getExampleNumber(country, examples);
    if (!example) return '';
    return example.formatInternational().replace(`+${example.countryCallingCode}`, '').trim();
}

/** The country of a number typed in international form (+44…, 0044…), if recognised. */
export function countryOfInternational(raw: string): CountryCode | undefined {
    const compact = raw.replace(/[\s.\-()]/g, '');
    if (!compact.startsWith('+') && !compact.startsWith('00')) return undefined;
    return parsePhoneNumberFromString(compact.replace(/^00/, '+'))?.country;
}

/**
 * The stored form of a valid number, or null. A number typed in
 * international form wins over the chosen country.
 */
export function toStoredPhone(raw: string, country: CountryCode = DEFAULT_PHONE_COUNTRY): string | null {
    const compact = raw.trim().replace(/^00/, '+');
    const parsed = parsePhoneNumberFromString(compact, country);
    if (!parsed?.isValid()) return null;
    return parsed.countryCallingCode === '33' ? '0' + parsed.nationalNumber : parsed.number;
}

/** E.164 of a stored number (0612345678 → +33612345678), as Stripe Link wants it. */
export const storedToE164 = (stored: string) => (stored.startsWith('0') ? '+33' + stored.slice(1) : stored);
