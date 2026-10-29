import type { PageServerLoad } from "./$types";
import { loadPracticePage } from "../practice";

export const load: PageServerLoad = async ({ fetch }) => loadPracticePage(fetch);
