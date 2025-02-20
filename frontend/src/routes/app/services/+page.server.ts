import type { PageLoad } from './$types';

import type { Offer } from "$lib/types"
import { offers } from "$lib/data"

export const load: PageLoad = (): { offers: Offer[] } => {
    return { offers }
}
