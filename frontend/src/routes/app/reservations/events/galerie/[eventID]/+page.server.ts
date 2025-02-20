import type { Event } from "$lib/types"
import { events } from "$lib/data"

type PageRes = { events: Event[], eventID: string }

export function load({ params }): PageRes {
    // TODO: do the fetching for that thing brother and send back the user events.
    return { events, eventID: params.eventID }
}
