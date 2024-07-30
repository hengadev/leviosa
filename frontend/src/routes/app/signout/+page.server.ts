import { redirect, invalid } from "@sveltejs/kit"
import type { PageServerLoad, Action, Actions } from "./$types"

export const load: PageServerLoad = async () => {
    // only used for the api in the layout file
    throw redirect(302, "/")
}

export const signout: Action = async ({ cookies }) => {
}

// NOTE: How to hit that thing ?
export const actions: Actions = {
    default: async ({ cookies }) => {
        // eat the cookie
        cookies.set("sessionId", "", {
            path: "/",
            expires: new Date(0),
        })
        // redirect to the sign in page
        throw redirect(302, "/")
    }
}


