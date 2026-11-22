import type { Actions, RequestEvent } from "./$types"
import { fail, redirect } from "@sveltejs/kit"
import * as env from "$app/env/private"

import { setError, superValidate } from 'sveltekit-superforms';
import { arktype } from 'sveltekit-superforms/adapters';
import { type } from "arktype"

import { forwardAuthCookies } from "#lib/utils/auth-helpers.ts";

function sanitizeRedirect(value: string | null): string | null {
    if (!value) return null;
    // Accept only paths starting with / but not // (protocol-relative)
    if (value.startsWith('/') && !value.startsWith('//')) return value;
    return null;
}

const loginSchema = type({
    email: "string.email",
    password: "8 < string < 64",
})
const loginDefaults = { email: "", password: "" }

// Login only: client accounts are out of scope in v1, so there is no
// registration and no Google/Apple sign-in here (unknown actions are refused
// by SvelteKit with a 404).

export const load = async ({ url }: RequestEvent) => {
    // Store the redirect target for use after successful login
    const redirectTo = sanitizeRedirect(url.searchParams.get("redirect"));

    const email = url.searchParams.get("email") // get email is from redirect (via the user.locals if hooks)
    const init = email ? { email, password: "" } : undefined

    const loginForm = await superValidate(init, arktype(loginSchema, { defaults: loginDefaults }))
    return { loginForm, redirectTo }
}

export const actions = {
    login: async ({ request, cookies, url, locals }: RequestEvent) => {
        const form = await superValidate(request, arktype(loginSchema, { defaults: loginDefaults }))

        if (!form.valid) {
            if (form.errors.email) {
                setError(form, "email", "L'adresse email saisie n'est pas valide. Veuillez vérifier et réessayer.")
            }
            if (form.errors.password) {
                setError(form, "password", "Le mot de passe saisi n'est pas valide. Veuillez vérifier et réessayer.")
            }
            return fail(400, { loginForm: form })
        }

        if (env.USE_MOCK_DATA === 'true') {
            const mockCredentials = [
                { email: env.MOCK_ADMIN_EMAIL, password: env.MOCK_ADMIN_PASSWORD, role: 'administrator', dest: '/admin' },
                { email: env.MOCK_PARTNER_EMAIL, password: env.MOCK_PARTNER_PASSWORD, role: 'partner', dest: '/staff' },
                { email: env.MOCK_CLIENT_EMAIL, password: env.MOCK_CLIENT_PASSWORD, role: 'standard', dest: '/client' },
            ];
            const match = mockCredentials.find(
                c => c.email && c.password && c.email === form.data.email && c.password === form.data.password
            );
            if (!match) {
                return setError(form, "Identifiants incorrects. Veuillez vérifier votre email et mot de passe.");
            }
            cookies.set('mock_session', match.role, {
                path: '/',
                httpOnly: true,
                sameSite: 'lax',
                maxAge: 60 * 60 * 24 * 7,
            });
            redirect(302, match.dest);
        }

        const res = await fetch(`${env.API_URL}/auth/login`, {
            method: "POST",
            headers: { 'Content-Type': "application/json" },
            body: JSON.stringify({
                email: form.data.email,
                password: form.data.password,
            })
        })

        if (res.status !== 201) {
            switch (res.status) {
                case 400:
                    return setError(form, "Format d'adresse e-mail ou de mot de passe invalide. Veuillez vérifier.");
                case 401:
                    return setError(form, "Identifiants incorrects. Veuillez vérifier votre email et mot de passe.");
                case 403:
                    return setError(form, "Votre compte n'est pas encore approuvé ou est inactif. Veuillez contacter l'administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté. Veuillez réessayer.");
                case 423:
                    return setError(form, "Votre compte est verrouillé suite à trop de tentatives échouées. Veuillez réessayer plus tard.");
                case 429:
                    return setError(form, "Trop de tentatives. Veuillez réessayer plus tard.");
                case 500:
                    return setError(form, "Une erreur est survenue sur le serveur. Nous travaillons à résoudre le problème au plus vite.");
                case 503:
                    return setError(form, "Le service est temporairement indisponible. Veuillez réessayer dans quelques instants.");
                default:
                    return setError(form, "Une erreur inattendue est survenue. Veuillez réessayer.");
            }
        }

        // Extract and forward authentication cookies from backend response to client.
        // The access token is stored under locals.sessionCookieName so staging and
        // production sessions don't collide on the shared .leviosa.care domain.
        forwardAuthCookies(res, cookies, locals.sessionCookieName, locals.cookieDomain, locals.refreshCookieName);

        // Get redirect target from URL params
        const redirectTo = sanitizeRedirect(url.searchParams.get("redirect"));

        // Fetch user to determine role-based redirect.
        // Read from sessionCookieName but always send as leviosa_access_token to backend.
        let finalRedirect = "/";
        try {
            const userRes = await fetch(`${env.API_URL}/users/me`, {
                headers: {
                    'Cookie': `leviosa_access_token=${cookies.get(locals.sessionCookieName) || ""}`
                }
            });

            if (userRes.ok) {
                const user = await userRes.json();

                // Role-based redirect: always send users to the appropriate page for their role
                // This prevents staff users from getting 403 errors if they tried to access /admin
                if (user.role === "administrator") {
                    finalRedirect = "/admin";
                } else if (user.role === "partner") {
                    finalRedirect = "/staff";
                } else if (redirectTo) {
                    // For other roles, respect the redirect param if provided
                    finalRedirect = redirectTo;
                }
            } else if (redirectTo) {
                // Fallback: use redirect param if user fetch fails
                finalRedirect = redirectTo;
            }
        } catch (e) {
            // If user fetch fails, use redirect param or default to home
            if (redirectTo) {
                finalRedirect = redirectTo;
            }
        }

        // Success - redirect to appropriate destination
        redirect(302, finalRedirect);
    },
} satisfies Actions
