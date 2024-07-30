import { redirect } from "@sveltejs/kit"

type Month = "Jan" | "Fev" | "Mar" | "Avr" | "Mai" | "Juin" | "Juil" | "Aout" | "Sept" | "Oct" | "Nov" | "Dec"
type NextVote = {
    month: Month,
    eventCount: number,
}

async function getNextEventId(userId: string) {
    console.log("fetch the next event with the user id")
    return "23r23asf3r32"
}

async function getNextVotes() {
    // TODO: Find that data using a fetch request on the backend golang.
    const nextVotes: NextVote[] = [
        {
            month: "Jan",
            eventCount: 5,
        },
        {
            month: "Fev",
            eventCount: 2,
        },
        {
            month: "Mar",
            eventCount: 9,
        },
    ]
    return nextVotes
}

export async function load({ locals }) {
    if (!locals.user) {
        console.log("redirect user")
        throw redirect(301, "/")
    }
    const nextVotes = await getNextVotes()
    const userId = locals.user.id
    const eventId = await getNextEventId(userId)
    return { nextVotes, eventId }
}
