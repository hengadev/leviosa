// The client of a practitioner booking (issue 041): a guest entered by hand
// (first name, last name and phone; email optional). Issue 043 adds picking
// an existing client to the same step.

export interface GuestClient {
	firstName: string;
	lastName: string;
	phone: string;
	email: string;
}

export function emptyGuestClient(): GuestClient {
	return { firstName: '', lastName: '', phone: '', email: '' };
}

/** One result of the client search (GET /api/partners/clients): a guest
 * client or one with an account, whose client_id books for that account. */
export interface ClientSummary {
	client_id?: string;
	name: string;
	first_name?: string;
	last_name?: string;
	phone?: string;
	email?: string;
}

/** A picked search result as form fields: the guest parts when the backend
 * sent them, else the display name split (first word / the rest). */
export function guestFieldsOf(summary: ClientSummary): GuestClient {
	let firstName = (summary.first_name ?? '').trim();
	let lastName = (summary.last_name ?? '').trim();
	if (!firstName && !lastName) {
		const parts = summary.name.trim().split(/\s+/);
		firstName = parts[0] ?? '';
		lastName = parts.slice(1).join(' ');
	}
	return { firstName, lastName, phone: summary.phone ?? '', email: summary.email ?? '' };
}

/** "06 12 34 56 78", "06.12.34.56.78" or "+33 6 12 34 56 78" become
 * "0612345678", the form the backend accepts. */
export function normalizePhone(phone: string): string {
	let digits = phone.trim().replace(/[\s.\-()]/g, '');
	if (digits.startsWith('+33')) digits = '0' + digits.slice(3);
	else if (digits.startsWith('0033')) digits = '0' + digits.slice(4);
	return digits;
}

/** "0612345678" → "06 12 34 56 78". */
export function displayPhone(phone: string): string {
	const digits = normalizePhone(phone);
	return /^0\d{9}$/.test(digits) ? digits.replace(/(\d{2})(?=\d)/g, '$1 ') : phone.trim();
}

export type ClientErrors = Partial<Record<keyof GuestClient, string>>;

/** The problems of the client fields, in French; empty when they are valid. */
export function validateClient(client: GuestClient): ClientErrors {
	const errors: ClientErrors = {};
	if (!client.firstName.trim()) errors.firstName = 'Indiquez le prénom.';
	if (!client.lastName.trim()) errors.lastName = 'Indiquez le nom.';
	if (!client.phone.trim()) {
		errors.phone = 'Indiquez le téléphone.';
	} else if (!/^0[1-7]\d{8}$/.test(normalizePhone(client.phone))) {
		errors.phone = 'Un numéro français à 10 chiffres, par exemple 06 12 34 56 78.';
	}
	const email = client.email.trim();
	if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
		errors.email = 'Cette adresse email ne semble pas valide.';
	}
	return errors;
}
