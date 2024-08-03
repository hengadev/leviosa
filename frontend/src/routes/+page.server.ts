import { redirect, fail } from "@sveltejs/kit"
import type { Action, Actions, PageServerLoad } from "./$types"

// type DomainName = "fr" | "com"
// type Email = `${string}@${string}.${DomainName}`

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
    // get info from the form
    const formData = await request.formData()
    const email = String(formData.get("email"))
    const password = String(formData.get("password"))
    // validate the email and password client side
    validate(email, password)
    const body = JSON.stringify({ email, password })
    // make request to the server.
    const res = await fetch("http://localhost:3500/api/v1/signin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body
    })
    if (res.ok) {
        // parse the reponse to reform the cookie.
        const cookieParsed = parseCookie(res.headers.getSetCookie()[0])
        cookies.set(cookieName, cookieParsed.Value, {
            path: "/",
            maxAge: 60 * 60 * 24 * 30,
            ...cookieParsed,
        })
        // redirect to sign in page.
        throw redirect(302, "/app")
    }
    // return data, in case we did not redirect the page and the fetching failed for some reason.
    return {
        email,
        success: false,
        message: "Error in setting the cookie or fetching the data."
    }
}

export const load: PageServerLoad = ({ locals }) => {
    if (locals.user) {
        throw redirect(301, "/app")
    }
}

export const actions: Actions = { register }
