import type { PageServerLoad, Actions } from "./$types";
import { env } from "$env/dynamic/private";
import { fail } from "@sveltejs/kit";

export const load: PageServerLoad = async ({ fetch }) => {
    const apiUrl = env.API_URL;

    // Best-effort: these are public settings endpoints. If any fail, the page
    // still renders — the "other ways to reach us" panel just shows less.
    const [emailRes, addressRes, instagramRes] = await Promise.allSettled([
        fetch(`${apiUrl}/settings/email`),
        fetch(`${apiUrl}/settings/address`),
        fetch(`${apiUrl}/settings/instagram`),
    ]);

    const companyEmail =
        emailRes.status === "fulfilled" && emailRes.value.ok
            ? (await emailRes.value.json()).email
            : null;
    const companyAddress =
        addressRes.status === "fulfilled" && addressRes.value.ok
            ? (await addressRes.value.json()).address
            : null;
    const companyInstagram =
        instagramRes.status === "fulfilled" && instagramRes.value.ok
            ? (await instagramRes.value.json()).instagram
            : null;

    return { companyEmail, companyAddress, companyInstagram };
};

export const actions = {
    default: async ({ request, fetch }) => {
        const formData = await request.formData();

        const name = (formData.get("name") as string)?.trim() ?? "";
        const contactEmail = (formData.get("contact_email") as string)?.trim() ?? "";
        const contactPhone = (formData.get("contact_phone") as string)?.trim() ?? "";
        const message = (formData.get("message") as string)?.trim() ?? "";

        const errors: Record<string, string> = {};
        if (!name) errors.name = "Le nom est requis";
        if (!message) errors.message = "Le message est requis";
        if (!contactEmail && !contactPhone) {
            errors.contact_email = "L'email ou le téléphone est requis";
            errors.contact_phone = "L'email ou le téléphone est requis";
        }

        if (Object.keys(errors).length > 0) {
            return fail(400, {
                errors,
                values: { name, contactEmail, contactPhone, message },
            });
        }

        const res = await fetch(`${env.API_URL}/contact`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                name,
                contact_email: contactEmail || undefined,
                contact_phone: contactPhone || undefined,
                message,
            }),
        });

        if (!res.ok) {
            let serverMessage = "Échec de l'envoi. Veuillez réessayer.";
            try {
                const body = await res.json();
                if (body?.error) serverMessage = body.error;
            } catch {
                // ignore — use the default message
            }
            const serverErrors: Record<string, string> = { _form: serverMessage };
            return fail(res.status === 429 ? 429 : 400, {
                errors: serverErrors,
                values: { name, contactEmail, contactPhone, message },
            });
        }

        return { success: true };
    },
} satisfies Actions;
