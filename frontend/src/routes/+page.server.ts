import { redirect, fail } from '@sveltejs/kit'; import type { Action, Actions, PageServerLoad } from './$types';
import { API_URL } from '$env/static/private';
import { describe, it, expect } from "vitest"

import { validate } from "$lib/scripts/credentials"

type CookieParsed = {
    Name: string;
    Value: string;
    Expires: Date;
    HttpOnly: boolean;
    Secure: boolean;
};

import { cookieName } from '$lib/types';

// TODO: I need to use that function is some sort of library and test it.

/**
 * Parse cookie into a CookieParsed object
 *
 * @param   cookie  the cookie in string format. 
 * @returns A cookie parsed into an exploitable CookieParsed object.
 */
function parseCookie(cookie: string): CookieParsed {
    // the default cookie object
    const res: CookieParsed = {
        Name: cookieName,
        Value: '',
        Expires: new Date(),
        HttpOnly: true,
        Secure: true
    };
    cookie.split(';').map((field) => {
        const split = field.trim().split('=');
        if (split[0] === 'Expires') {
            res.Expires = new Date(split[1]);
        }
        if (split[0] === cookieName) {
            res.Value = split[1];
        }
    });
    return res;
}

describe("parseCookie", () => {
    // just an example for the test in here and for me to use all my imports
    const cookie: string = ""
    it.todo("should do nothing for now", () => {
        const cookieParsed = parseCookie(cookie)
        // TODO: find the right format for the cookoie that I am going to exploit
        expect(cookieParsed).toHaveTextContent(/sometextcontent/iu)
    })
    expect(2 + 2).toBe(4)
})

// TODO: should I test that function with some mocking thing ?
const register: Action = async ({ request, cookies }) => {
    console.log('in the register action');
    const url = `${API_URL}/api/v1/signin`;
    // get info from the form
    const formData = await request.formData();
    const email = String(formData.get('email'));
    const password = String(formData.get('password'));
    console.log('the url that I am fetching is :', url);

    try {
        // validate the email and password client side
        if (!validate(email, password)) {
            throw new Error('email or password invalid');
        }
        const body = JSON.stringify({ email, password });
        // make request to the server.
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body,
            credentials: 'include'
        });
        if (!res.ok) {
            throw new Error(`Failed fetch with response status: ${res.statusText}`);
        }
        // parse reponse to use cookie.
        const cookie = res.headers.getSetCookie()[0];
        if (!cookie) {
            throw new Error(`No cookie found with response status: ${res.statusText}`);
        }
        const cookieParsed = parseCookie(cookie);
        cookies.set(cookieName, cookieParsed.Value, {
            path: '/',
            maxAge: 60 * 60 * 24 * 30,
            ...cookieParsed
        });
        // redirect to home page.
        throw redirect(302, '/app');
    } catch (error) {
        console.error(error.message);
    }
};

describe("register User", () => {
    it.todo("should register user with the right credentials")
})

export const load: PageServerLoad = ({ locals }) => {
    if (locals.user) {
        throw redirect(301, '/app');
    }
};

export const actions: Actions = { register };
