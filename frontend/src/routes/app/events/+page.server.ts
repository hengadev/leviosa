import type { PageServerLoad } from "./$types"
import { redirect } from "@sveltejs/kit"

// TODO: make the right type based on the backend
// type Event = {
//     Id: string; Location: string;
//     PlaceCount: number;
//     Date: Date;
// }

export const load: PageServerLoad = async ({ locals, cookies }) => {
    if (!locals.user) {
        throw redirect(302, "/")
    }
    const sessionId = cookies.get("sessionId");
    const res = await fetch("http://localhost:5000/api/v1/events", {
        method: "GET",
        headers: {
            "Authorization": `Bearer ${sessionId}`,
        },
    })
    const { pastEvents, nextEvents, incomingEvents } = await res.json()
    return { pastEvents, nextEvents, incomingEvents }
}

// TODO: Should I use a form action or just a <a> balise ?
export const actions = {
    default: async ({ request }) => {
        const formData = await request.formData()
        const eventId = formData.get("id")
        throw redirect(302, `/app/events/${eventId}`)
    }
}
