import { redirect, fail } from "@sveltejs/kit"
import type { Action, Actions, PageServerLoad } from "./$types"
import { API_URL } from "$env/static/private"

// type DomainName = "fr" | "com"
// type Email = `${string}@${string}.${DomainName}`

// TODO: make a better email validation
function validate(email: string, password: string) {
    if (
        typeof email !== 'string' ||
        typeof password !== 'string' ||
        !email ||
        !password
    ) {
        return fail(400, { invalid: true })
    }
    return true
}

type CookieParsed = {
    Name: string,
    Value: string,
    Expires: Date,
    HttpOnly: boolean,
    Secure: boolean,
}

import { cookieName } from "$lib/types"

function parseCookie(cookie: string): CookieParsed {
    // the default cookie object
    const res: CookieParsed = {
        Name: cookieName,
        Value: "",
        Expires: new Date(),
        HttpOnly: true,
        Secure: true,
    };
    cookie.split(";").map((field) => {
        const split = field.trim().split("=");
        if (split[0] === "Expires") {
            res.Expires = new Date(split[1]);
        }
        if (split[0] === cookieName) {
            res.Value = split[1]
        }
    });
    return res;
}

const register: Action = async ({ request, cookies }) => {
    console.log("in the register action")
    const url = `${API_URL}/api/v1/signin`
    // get info from the form
    const formData = await request.formData()
    const email = String(formData.get("email"))
    const password = String(formData.get("password"))
    console.log("the url that I am fetching is :", url)

    try {
        // validate the email and password client side
        if (!validate(email, password)) {
            throw new Error("email or password invalid")
        }
        const body = JSON.stringify({ email, password })
        // make request to the server.
        const res = await fetch(url, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body,
            credentials: "include",
        })
        if (!res.ok) {
            throw new Error(`Failed fetch with response status: ${res.statusText}`)
        }
        // parse reponse to use cookie.
        const cookie = res.headers.getSetCookie()[0]
        if (!cookie) {
            throw new Error(`No cookie found with response status: ${res.statusText}`)
        }
        const cookieParsed = parseCookie(cookie)
        cookies.set(cookieName, cookieParsed.Value, {
            path: "/",
            maxAge: 60 * 60 * 24 * 30,
            ...cookieParsed,
        })
        // redirect to home page.
        throw redirect(302, "/app")
    } catch (error) {
        console.error(error.message)
    }
}

export const config = {
    csrf: false
}

export const load: PageServerLoad = ({ locals }) => {
    if (locals.user) {
        throw redirect(301, "/app")
    }
}

export const actions: Actions = { register }
