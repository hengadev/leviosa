import { redirect } from '@sveltejs/kit';

type Event = {
	Id: string;
	Location: string;
	PlaceCount: number;
	FreePlace: number;
	BeginAt: string;
	SessionDuration: string;
	Day: number;
	Month: number;
	Year: number;
};

export async function load({ locals, cookies, params }) {
	if (!locals.user) {
		throw redirect(302, '/');
	}
	const eventId = params.eventId;
	const sessionId = cookies.get('sessionId');
	const res = await fetch(`http://localhost:5000/api/v1/events?id=${eventId}`, {
		method: 'GET',
		headers: { Authorization: `Bearer ${sessionId}`, 'Content-Type': 'application/json' }
	});
	const event = (await res.json()) as Event;
	return { event };
}
