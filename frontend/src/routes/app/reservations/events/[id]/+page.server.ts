import type { EventInformation } from "$lib/types"

import { eventInformation } from "$lib/data"

type PageRes = { eventID: string, eventInformation: EventInformation }
export function load({ params }): PageRes {
    // TODO: do the fetching for that thing brother and send back the user events.
    return { eventID: params.id, eventInformation }
}
