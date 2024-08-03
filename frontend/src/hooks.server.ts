import { redirect, type Handle } from "@sveltejs/kit";

// TODO: What happens if my cookie is no longer valid ?
export const handle: Handle = async ({ event, resolve }) => {
    // get cookies from browser
    const sessionId = event.cookies.get("sessionId");
    // if no session then load page as normal
    if (!sessionId) {
        console.log("No session active")
        return await resolve(event)
    }
    const res = await fetch("http://localhost:5000/me", {
        headers: {
            "Authorization": `Bearer ${sessionId}`,
        },
    })
    // if there  is a sessionId but the session is no longer valid, remove the cookie and redirect to sign in.
    if (res.status === 401) {
        console.log("the status is so that I should redirect the user and cancel the previous cookie.")
        event.cookies.set("sessionId", "", {
            path: "/",
            expires: new Date(0),
        })
        throw redirect(302, "/")
    }

    // else we get back a user from successful request.
    const user = (await res.json()) as {
        email: string,
        lastname: string,
        firstname: string,
        gender: string,
        birthdate: string,
        telephone: string,
        address: string,
        city: string,
        postalcard: string,
    }
    // set the user with the information from got the server.
    if (user) { event.locals.user = { ...user } }
    // load page as normal if error
    return await resolve(event)
}
