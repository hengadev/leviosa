import type { Conversation, SessionNote } from '$lib/types';

import { conversations, notes } from '$lib/data';

type PageRes = { conversations: Conversation[]; notes: SessionNote[] };
export function load(): PageRes {
	return { conversations, notes };
}
