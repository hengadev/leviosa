import type { Event } from "$lib/types"
import { events } from "$lib/data"

type PageRes = { events: Event[], eventID: string }

// TODO: that thing needs to be fetched from the server brother

export function load({ params }): PageRes {
    // TODO: do the fetching for that thing brother and send back the user events.
    return { events, eventID: params.eventID }
}
