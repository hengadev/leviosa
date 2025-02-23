import type { Message } from '$lib/types';

import { messages } from '$lib/data';

type PageRes = { messages: Message[] };
export function load({ params }): PageRes {
	// TODO: use that to do the fetching for the conversation
	// const id = params.id

	return { messages };
}
