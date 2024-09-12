import type { PageLoad } from './$types';

import { address, general } from './formConstructor';

export const load: PageLoad = () => {
	return { general, address };
};
