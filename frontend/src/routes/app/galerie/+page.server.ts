// TODO: put these types somewhere else

import type { EventPhotos, EventVideos } from '$lib/types';

import { eventsVideos, eventsPhotos } from '$lib/data/media';

type PageRes = { eventsPhotos: EventPhotos[]; eventsVideos: EventVideos[] };
export function load(): PageRes {
	return { eventsPhotos, eventsVideos };
}
