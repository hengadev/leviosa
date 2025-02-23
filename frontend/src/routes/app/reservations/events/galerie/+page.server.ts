import type { EventPhotos, EventVideos } from '$lib/types';
import { eventsPhotos, eventsVideos } from '$lib/data';

type PageRes = { eventsPhotos: EventPhotos[]; eventsVideos: EventVideos[] };

export function load(): PageRes {
	return { eventsPhotos, eventsVideos };
}
