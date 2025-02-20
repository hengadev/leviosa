import type { Message, SessionNote } from "$lib/types"

import { messages, notes } from "$lib/data"

type LayoutRes = { messages: Message[], notes: SessionNote[] }
export function load({ params }): LayoutRes {
    // TODO: make a request to get the last conversation that I had
    return { messages, notes }
}
