import type { PageServerLoad, Actions } from "./$types";
import { env } from "$env/dynamic/private";
import { error, redirect } from "@sveltejs/kit";
import { fail } from "@sveltejs/kit";
import { getCookieDomain } from "$lib/server/hostname";

export const load: PageServerLoad = async ({ fetch, url, locals }) => {
    const apiUrl = env.API_URL;

    // Fetch published categories for Step 1
    const categoriesRes = await fetch(`${apiUrl}/categories`);
    if (!categoriesRes.ok) {
        throw error(500, "Impossible de charger les catégories");
    }
    const categories = await categoriesRes.json();

    // If ?product=<id> is provided, pre-fetch product details and partners
    let preselectedProduct: any = null;
    let preselectedCategory: any = null;
    let preselectedPartners: any[] = [];

    const productId = url.searchParams.get("product");
    if (productId) {
        const productRes = await fetch(`${apiUrl}/products/${productId}`);
        if (productRes.ok) {
            preselectedProduct = await productRes.json();
            preselectedCategory = preselectedProduct.category ?? null;

            const partnersRes = await fetch(`${apiUrl}/partners/products/${productId}`);
            if (partnersRes.ok) {
                preselectedPartners = await partnersRes.json();
            }
        }
    }

    return {
        categories,
        preselectedProduct,
        preselectedCategory,
        preselectedPartners,
        user: locals.user ?? null,
    };
};

export const actions = {
    default: async ({ request, fetch, locals, cookies, url }) => {
        const formData = await request.formData();

        const availabilityId = formData.get("availability_id") as string;
        const productId = formData.get("product_id") as string;
        const slotStartTime = formData.get("slot_start_time") as string;
        const guestFirstName = formData.get("guest_first_name") as string;
        const guestLastName = formData.get("guest_last_name") as string;
        const guestEmail = formData.get("guest_email") as string;
        const guestPhone = formData.get("guest_phone") as string;

        // Validate required fields
        const errors: Record<string, string> = {};
        if (!availabilityId) errors.availability_id = "Requis";
        if (!productId) errors.product_id = "Requis";
        if (!slotStartTime) errors.slot_start_time = "Requis";

        // Read authenticated user from session — never trust client_id from form data
        const authenticatedUserId = locals.user?.id ?? null;

        if (!authenticatedUserId) {
            if (!guestFirstName?.trim()) errors.guest_first_name = "Le prénom est requis";
            if (!guestLastName?.trim()) errors.guest_last_name = "Le nom est requis";
            if (!guestEmail?.trim() && !guestPhone?.trim()) {
                errors.guest_email = "L'email ou le téléphone est requis";
                errors.guest_phone = "L'email ou le téléphone est requis";
            }
        }

        if (Object.keys(errors).length > 0) {
            return fail(400, { errors });
        }

        // Build request body
        const body: Record<string, any> = {
            availability_id: availabilityId,
            product_id: productId,
            slot_start_time: slotStartTime,
        };

        if (authenticatedUserId) {
            body.client_id = authenticatedUserId;
        } else {
            body.guest_first_name = guestFirstName;
            body.guest_last_name = guestLastName;
            if (guestEmail) body.guest_email = guestEmail;
            if (guestPhone) body.guest_phone = guestPhone;
        }

        // POST /bookings is public — no auth required
        const res = await fetch(`${env.API_URL}/bookings`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        });

        if (!res.ok) {
            const errorData = await res.text();
            console.error("Booking creation failed:", res.status, errorData);
            return fail(res.status === 409 ? 409 : 400, {
                errors: { _form: "Échec de la création de la réservation. Veuillez réessayer." },
            });
        }

        const booking = await res.json();

        // Build confirmation URL — no PII in query params
        const params = new URLSearchParams({ booking_id: booking.id });
        if (booking.product_id) params.set("product_id", booking.product_id);
        if (booking.slot_start_time) params.set("slot_start_time", booking.slot_start_time);
        if (booking.slot_end_time) params.set("slot_end_time", booking.slot_end_time);
        if (booking.total_price_cents) params.set("price_cents", String(booking.total_price_cents));

        // For guest bookings, store guest contact info in a short-lived
        // httponly cookie so the confirmation page can pre-fill the inline
        // account creation card without leaking PII into the URL.
        if (!authenticatedUserId) {
            const cookieDomain = getCookieDomain(url.hostname);
            cookies.set("guest_booking_info", JSON.stringify({
                guest_first_name: guestFirstName || "",
                guest_last_name: guestLastName || "",
                guest_email: guestEmail || "",
                guest_phone: guestPhone || "",
            }), {
                path: "/book/confirmation",
                maxAge: 300, // 5 minutes
                httpOnly: true,
                secure: !url.hostname.startsWith('localhost'),
                sameSite: "lax",
                ...(cookieDomain && { domain: cookieDomain }),
            });
        }

        throw redirect(302, `/book/confirmation?${params.toString()}`);
    },

    // Join the waitlist for a product/date range when no matching slot is
    // available (issue 014). Kept separate from `default` (booking creation)
    // so the same page can host both forms.
    joinWaitlist: async ({ request, fetch, locals }) => {
        const formData = await request.formData();

        const productId = formData.get("product_id") as string;
        const desiredRangeStart = formData.get("desired_range_start") as string;
        const desiredRangeEnd = formData.get("desired_range_end") as string;
        const notes = (formData.get("notes") as string) || "";
        const guestFirstName = formData.get("guest_first_name") as string;
        const guestLastName = formData.get("guest_last_name") as string;
        const guestEmail = formData.get("guest_email") as string;
        const guestPhone = formData.get("guest_phone") as string;

        const errors: Record<string, string> = {};
        if (!productId) errors.product_id = "Requis";
        if (!desiredRangeStart) errors.desired_range_start = "Requis";
        if (!desiredRangeEnd) errors.desired_range_end = "Requis";

        const authenticatedUserId = locals.user?.id ?? null;

        if (!authenticatedUserId) {
            if (!guestFirstName?.trim()) errors.guest_first_name = "Le prénom est requis";
            if (!guestLastName?.trim()) errors.guest_last_name = "Le nom est requis";
            if (!guestEmail?.trim() && !guestPhone?.trim()) {
                errors.guest_email = "L'email ou le téléphone est requis";
                errors.guest_phone = "L'email ou le téléphone est requis";
            }
        }

        if (Object.keys(errors).length > 0) {
            return fail(400, { waitlistErrors: errors });
        }

        const body: Record<string, any> = {
            product_id: productId,
            // Dates from <input type="date"> are bare YYYY-MM-DD; the backend
            // expects RFC3339. Treat the start as start-of-day and the end as
            // end-of-day (inclusive) in UTC.
            desired_range_start: new Date(`${desiredRangeStart}T00:00:00Z`).toISOString(),
            desired_range_end: new Date(`${desiredRangeEnd}T23:59:59Z`).toISOString(),
        };
        if (notes) body.notes = notes;

        if (authenticatedUserId) {
            body.client_id = authenticatedUserId;
        } else {
            body.guest_first_name = guestFirstName;
            body.guest_last_name = guestLastName;
            if (guestEmail) body.guest_email = guestEmail;
            if (guestPhone) body.guest_phone = guestPhone;
        }

        // POST /waitlist is public — no auth required (mirrors POST /bookings)
        const res = await fetch(`${env.API_URL}/waitlist`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        });

        if (!res.ok) {
            const errorData = await res.text();
            console.error("Join waitlist failed:", res.status, errorData);
            const submitErrors: Record<string, string> = {
                _form: "Échec de l'inscription à la liste d'attente. Veuillez réessayer.",
            };
            return fail(400, { waitlistErrors: submitErrors });
        }

        return { waitlistSuccess: true };
    },
} satisfies Actions;
