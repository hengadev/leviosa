import { redirect } from '@sveltejs/kit';
import type { PageServerLoad, Action, Actions } from './$types';

export const load: PageServerLoad = ({ locals }) => {
	if (!locals.user) {
		throw redirect(301, '/');
	}
	if (locals.user.role !== 'admin') {
		throw redirect(301, '/');
	}
	// TODO: return the last result for event recently open for votings.
};

const viewMonth: Action = ({ request }) => {
	const formData = request.formData();
	const date = formData.get('date');
};

const viewDayInMonth: Action = ({ request }) => {
	const formData = request.formData();
	const date = formData.get('date');
};

// I want to select month or a specific data in a month
export const actions: Actions = { viewMonth, viewDayInMonth };
