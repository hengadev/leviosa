import { redirect } from '@sveltejs/kit';

type PageRes = { role: import("$lib/types").Role }

export function load({ locals }): PageRes {
    if (locals.user.role !== "admin") throw redirect(302, "/app")
    return { role: locals.user.role }
}
