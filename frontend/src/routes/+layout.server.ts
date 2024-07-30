import type { LayoutServerLoad } from "./$types"

// get locals.user and pass it to the page store, to make it available for all the pages.
export const load: LayoutServerLoad = async ({ locals }) => {
    return {
        user: locals.user
    }
}
