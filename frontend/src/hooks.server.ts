import { redirect, type Handle } from '@sveltejs/kit';
import { cookieName } from "$lib/data/const"
import { role } from "$lib/data/role"
// TODO: that thing should be an env variable !
import { getAPIURL } from '$lib/data/const/api_url';
// TODO: add something so that I can add to the request made to the server the client remote address
// this is for logging on the golang backend.Since in my flow I make request to the sveltekit client
// which makes request to the golang backend, this is the only way for me to get the remote address


const environment = import.meta.env.VITE_ENVIRONMENT

export const handle: Handle = async ({ event, resolve }) => {
    if (event.url.pathname === '/soon') {
        return resolve(event);
    }
    // enrich the fetch with the custom header with the client IP
    event.fetch = async (input, init = {}) => {
        // Ensure headers object exists
        init.headers = {
            ...init.headers,
            'X-Client-IP': getClientIP(event.request)
        };

        // Proceed with the original fetch
        return fetch(input, init);
    };
    if (environment === "development") {
        event.locals.user = {
            firstname: "John",
            lastname: "DOE",
            city: "Paris",
            role: role,
        }
    }
    // else {
    //     const sessionID = event.cookies.get(cookieName);
    //     if (!sessionID) {
    //         throw redirect(302, "/")
    //     }
    //
    //     // NOTE: just trying to do some try catch in here
    //     try {
    //         const backendURL = getAPIURL()
    //         if (backendURL === "") {
    //             throw Error("")
    //         }
    //
    //         // const res = await fetch(`${API_URL}/api/v1/me`, {
    //         const res = await fetch(`${backendURL}/api/v1/me`, {
    //             headers: {
    //                 Authorization: `Bearer ${sessionID}`
    //             }
    //         });
    //
    //         if (res.status === 401) {
    //             console.log('the status is so that I should redirect the user and cancel the previous cookie.');
    //             event.cookies.set(cookieName, '', {
    //                 path: '/',
    //                 expires: new Date(0)
    //             });
    //             throw redirect(302, '/');
    //         }
    //
    //         const user = await res.json();
    //         if (user) {
    //             event.locals.user = { ...user };
    //         }
    //     } catch (error) {
    //         // TODO: make some better error with redirect to page to say that the server is not set properly
    //         console.log(error)
    //     }
    // }
    return await resolve(event);
}

function getClientIP(request: Request): string {
    const headers = request.headers;
    return headers.get('cf-connecting-ip') || // Cloudflare
        headers.get('x-real-ip') || // Nginx
        headers.get('x-forwarded-for')?.split(',')[0] || // Standard
        'unknown';
}
