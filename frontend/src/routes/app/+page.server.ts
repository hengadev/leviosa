import { redirect } from "@sveltejs/kit"
import { cookieName } from "$lib/types"
// import { API_URL } from "$env/static/private"

type Month = "Jan" | "Fev" | "Mar" | "Avr" | "Mai" | "Juin" | "Juil" | "Aout" | "Sept" | "Oct" | "Nov" | "Dec"
type NextVote = {
    month: Month,
    eventCount: number,
}

// async function getNextEventId(sessionID: string) {
async function getNextEventId() {
    // const url = `${API_URL}/events?n=1/${sessionID}`
    // const res = await fetch(url)
    // if (!res.ok) {
    //     throw new Error("")
    // }
    return "23r23asf3r32"
}

// async function getNextVotes(sessionID: string) {
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

export async function load({ locals, cookies }) {
    if (!locals.user) {
        console.log("redirect user")
        throw redirect(301, "/")
    }
    // const userId = locals.user.id
    // const sessionID = cookies.get(cookieName)
    try {
        // const eventId = await getNextEventId(sessionID)
        // const nextVotes = await getNextVotes(sessionID)
        const eventId = await getNextEventId()
        const nextVotes = await getNextVotes()
        console.log("the next votes are :", nextVotes)
        return { nextVotes, eventId }
    } catch (error) {
        console.error("error loading home page", error.message)
    }
}
