import type { Actions, PageServerLoad } from "./$types"
import { fail, redirect } from "@sveltejs/kit"
import { env } from "$env/dynamic/private"

import { setError, superValidate } from 'sveltekit-superforms';
import { arktype } from 'sveltekit-superforms/adapters';

import {
    companyNameSchema,
    companyNameDefaults,
    companyEmailSchema,
    companyEmailDefaults,
    companyPhoneSchema,
    companyPhoneDefaults,
    companyAddressSchema,
    companyAddressDefaults,
    companyInstagramSchema,
    companyInstagramDefaults,
    otpDurationSchema,
    otpDurationDefaults,
    otpLengthSchema,
    otpLengthDefaults,
    otpMaxAttemptsSchema,
    otpMaxAttemptsDefaults,
    accessTokenDurationSchema,
    accessTokenDurationDefaults,
    refreshTokenDurationSchema,
    refreshTokenDurationDefaults,
    practicePageSchema,
    practicePageDefaults,
    bookingPolicySchema,
    bookingPolicyDefaults,
} from "./schemas"
// Enable mock mode for staging
const MOCK_MODE = env.USE_MOCK_DATA === 'true'

type Settings = {
    company: {
        name: string
        email: string
        telephone: string
        address: string
        instagram: string
        logo_url: string
        logo_content_type: string
    }
    otp: {
        duration: number
        length: number
        max_attempts: number
    }
    tokens: {
        access_duration: number
        refresh_duration: number
    }
    practicePage: {
        partnerId: string
        productId: string
        videoUrl: string
    }
}

// Booking policy applied by the slot engine to the public calendar and to
// booking creation. The backend answers the defaults while it is unset.
type BookingPolicySettings = { minimumNoticeHours: number; horizonWeeks: number; paymentHoldMinutes: number }

// Options for the practice page selectors: the partners and products the
// administrator can put on the public page.
type PracticeOption = { id: string; label: string }

export const load: PageServerLoad = async ({ parent, fetch, cookies }) => {
    // The admin layout already lets administrators only; this is a second check.
    const { permissions } = await parent()

    if (!permissions.canAccessOps) {
        throw redirect(302, '/app')
    }

    // Initialize forms
    const companyNameForm = await superValidate(arktype(companyNameSchema, { defaults: companyNameDefaults }))
    const companyEmailForm = await superValidate(arktype(companyEmailSchema, { defaults: companyEmailDefaults }))
    const companyPhoneForm = await superValidate(arktype(companyPhoneSchema, { defaults: companyPhoneDefaults }))
    const companyAddressForm = await superValidate(arktype(companyAddressSchema, { defaults: companyAddressDefaults }))
    const companyInstagramForm = await superValidate(arktype(companyInstagramSchema, { defaults: companyInstagramDefaults }))
    const otpDurationForm = await superValidate(arktype(otpDurationSchema, { defaults: otpDurationDefaults }))
    const otpLengthForm = await superValidate(arktype(otpLengthSchema, { defaults: otpLengthDefaults }))
    const otpMaxAttemptsForm = await superValidate(arktype(otpMaxAttemptsSchema, { defaults: otpMaxAttemptsDefaults }))
    const accessTokenDurationForm = await superValidate(arktype(accessTokenDurationSchema, { defaults: accessTokenDurationDefaults }))
    const refreshTokenDurationForm = await superValidate(arktype(refreshTokenDurationSchema, { defaults: refreshTokenDurationDefaults }))
    const practicePageForm = await superValidate(arktype(practicePageSchema, { defaults: practicePageDefaults }))
    const bookingPolicyForm = await superValidate(arktype(bookingPolicySchema, { defaults: bookingPolicyDefaults }))

    let settings: Settings | null = null
    let settingsError = false

    const sessionCookie = cookies.get('session')

    // The backend's key names (internal/common/contracts/settings/key.go). The
    // token durations are not in the bulk endpoint: they have their own GETs.
    const bulkKeys = [
        'company_name',
        'support_email',
        'telephone',
        'headquarters',
        'instagram_path',
        'otp_duration',
        'otp_length',
        'otp_max_attempts',
    ].join(',')

    try {
        const headers = { 'Authorization': `Bearer ${sessionCookie}` }
        const [settingsRes, accessRes, refreshRes] = await Promise.all([
            fetch(`${env.API_URL}/admin/settings/bulk?keys=${bulkKeys}`, { headers }),
            fetch(`${env.API_URL}/admin/settings/tokens/access-duration`, { headers }),
            fetch(`${env.API_URL}/admin/settings/tokens/refresh-duration`, { headers }),
        ])
        const duration = async (res: Response) => (res.ok ? String((await res.json()).duration ?? '') : '')
        const [accessDuration, refreshDuration] = await Promise.all([duration(accessRes), duration(refreshRes)])

        if (settingsRes.ok) {
            // 200 is the list; 207 (some keys never set yet) wraps it as
            // { data, errors }, and an unset key just shows empty.
            type BulkSetting = { key: string; value: string }
            const body: BulkSetting[] | { data: BulkSetting[] } = await settingsRes.json()
            const bulkData = Array.isArray(body) ? body : body.data ?? []
            const values: Record<string, string> = {
                ...Object.fromEntries(bulkData.map((s) => [s.key, s.value])),
                access_token_duration: accessDuration,
                refresh_token_duration: refreshDuration,
            }
            const find = (key: string) => values[key] ?? ''

            settings = {
                company: {
                    name: find('company_name'),
                    email: find('support_email'),
                    telephone: find('telephone'),
                    address: find('headquarters'),
                    instagram: find('instagram_path'),
                    logo_url: '',
                    logo_content_type: '',
                },
                otp: {
                    duration: parseInt(find('otp_duration') || '300'),
                    length: parseInt(find('otp_length') || '6'),
                    max_attempts: parseInt(find('otp_max_attempts') || '5'),
                },
                tokens: {
                    access_duration: parseInt(find('access_token_duration') || '15'),
                    refresh_duration: parseInt(find('refresh_token_duration') || '168'),
                },
                practicePage: {
                    partnerId: '',
                    productId: '',
                    videoUrl: '',
                },
            }
        } else {
            settingsError = true
        }
    } catch {
        settingsError = true
    }

    // Practice page: current settings (public GETs; 404 = unset) plus the
    // partner and product lists for the selectors. Best-effort, like the
    // settings fetch above.
    let practicePage: Settings['practicePage'] | null = null
    let partnerOptions: PracticeOption[] = []
    let productOptions: PracticeOption[] = []
    try {
        const authHeaders = { 'Authorization': `Bearer ${sessionCookie}` }
        const [partnerSetting, productSetting, videoSetting, partnersRes, productsRes] = await Promise.all([
            fetch(`${env.API_URL}/settings/practice-page/partner`, { headers: authHeaders }),
            fetch(`${env.API_URL}/settings/practice-page/product`, { headers: authHeaders }),
            fetch(`${env.API_URL}/settings/practice-page/video`, { headers: authHeaders }),
            fetch(`${env.API_URL}/partners`, { headers: authHeaders }),
            fetch(`${env.API_URL}/products`, { headers: authHeaders }),
        ])

        const partnerId = partnerSetting.ok ? (await partnerSetting.json()).partner_id || '' : ''
        const productId = productSetting.ok ? (await productSetting.json()).product_id || '' : ''
        const videoUrl = videoSetting.ok ? (await videoSetting.json()).video_url || '' : ''
        practicePage = { partnerId, productId, videoUrl }

        if (partnersRes.ok) {
            const partners: { id: string; first_name?: string; last_name?: string }[] = await partnersRes.json()
            partnerOptions = partners.map((partner) => ({
                id: partner.id,
                label: `${partner.first_name ?? ''} ${partner.last_name ?? ''}`.trim() || partner.id,
            }))
        }
        if (productsRes.ok) {
            const products: { product?: { id: string; name: string }; id?: string; name?: string }[] = await productsRes.json()
            productOptions = products.map((entry) => ({
                id: entry.product?.id ?? entry.id ?? '',
                label: entry.product?.name ?? entry.name ?? '',
            }))
        }
    } catch {
        practicePage = null
    }

    // Booking policy: public GET, defaults included. Best-effort.
    let bookingPolicy: BookingPolicySettings | null = null
    try {
        const policyRes = await fetch(`${env.API_URL}/settings/booking-policy`)
        if (policyRes.ok) {
            const policy: { minimum_notice_hours: number; horizon_weeks: number; payment_hold_minutes?: number } = await policyRes.json()
            bookingPolicy = {
                minimumNoticeHours: policy.minimum_notice_hours,
                horizonWeeks: policy.horizon_weeks,
                paymentHoldMinutes: policy.payment_hold_minutes ?? bookingPolicyDefaults.payment_hold_minutes,
            }
        }
    } catch {
        bookingPolicy = null
    }
    if (MOCK_MODE && !bookingPolicy) {
        bookingPolicy = {
            minimumNoticeHours: bookingPolicyDefaults.minimum_notice_hours,
            horizonWeeks: bookingPolicyDefaults.horizon_weeks,
            paymentHoldMinutes: bookingPolicyDefaults.payment_hold_minutes,
        }
    }

    return {
        companyNameForm,
        companyEmailForm,
        companyPhoneForm,
        companyAddressForm,
        companyInstagramForm,
        otpDurationForm,
        otpLengthForm,
        otpMaxAttemptsForm,
        accessTokenDurationForm,
        refreshTokenDurationForm,
        practicePageForm,
        bookingPolicyForm,
        bookingPolicy,
        settings,
        settingsError,
        practicePage,
        partnerOptions,
        productOptions,
    }
}

export const actions = {
    updatePracticePage: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(practicePageSchema, { defaults: practicePageDefaults }))

        if (!form.valid) {
            return fail(400, { practicePageForm: form })
        }

        if (MOCK_MODE) {
            console.log('[MOCK] Updating practice page settings:', form.data.partner_id, form.data.product_id)
            return { practicePageForm: form }
        }

        const sessionCookie = cookies.get('session');

        const post = (path: string, body: unknown) =>
            fetch(`${env.API_URL}/admin/settings/practice-page/${path}`, {
                method: "POST",
                headers: {
                    'Content-Type': "application/json",
                    'Authorization': `Bearer ${sessionCookie}`,
                },
                body: JSON.stringify(body),
            })

        // An empty video URL is meaningful (it clears the video back to the
        // placeholder frame); partner and product are required UUIDs.
        const [partnerRes, productRes, videoRes] = await Promise.all([
            post("partner", { partner_id: form.data.partner_id }),
            post("product", { product_id: form.data.product_id }),
            post("video", { video_url: form.data.video_url }),
        ])

        if (!partnerRes.ok || !productRes.ok || !videoRes.ok) {
            return setError(form, "Partenaire, produit ou vidéo invalide.")
        }

        return { practicePageForm: form }
    },

    updateBookingPolicy: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(bookingPolicySchema, { defaults: bookingPolicyDefaults }))

        if (!form.valid) {
            return fail(400, { bookingPolicyForm: form })
        }

        if (MOCK_MODE) {
            console.log('[MOCK] Updating booking policy:', form.data.minimum_notice_hours, form.data.horizon_weeks, form.data.payment_hold_minutes)
            return { bookingPolicyForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/booking-policy`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({
                minimum_notice_hours: form.data.minimum_notice_hours,
                horizon_weeks: form.data.horizon_weeks,
                payment_hold_minutes: form.data.payment_hold_minutes,
            }),
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Valeurs invalides (délai : 0 à 168 h ; horizon : 1 à 52 semaines ; maintien : 5 à 60 min).")
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.")
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.")
                default:
                    return setError(form, "Erreur lors de l'enregistrement. Réessayez.")
            }
        }

        return { bookingPolicyForm: form }
    },

    updateCompanyName: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(companyNameSchema, { defaults: companyNameDefaults }))

        if (!form.valid) {
            return fail(400, { companyNameForm: form })
        }

        if (MOCK_MODE) {
            console.log('🏢 [MOCK] Updating company name:', form.data.name)
            return { companyNameForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/name`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ name: form.data.name })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Nom invalide.");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { companyNameForm: form }
    },

    updateCompanyEmail: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(companyEmailSchema, { defaults: companyEmailDefaults }))

        if (!form.valid) {
            return fail(400, { companyEmailForm: form })
        }

        if (MOCK_MODE) {
            console.log('📧 [MOCK] Updating company email:', form.data.email)
            return { companyEmailForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/email`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ email: form.data.email })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Email invalide.");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { companyEmailForm: form }
    },

    updateCompanyPhone: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(companyPhoneSchema, { defaults: companyPhoneDefaults }))

        if (!form.valid) {
            return fail(400, { companyPhoneForm: form })
        }

        if (MOCK_MODE) {
            console.log('📞 [MOCK] Updating company phone:', form.data.telephone)
            return { companyPhoneForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/phone`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ telephone: form.data.telephone })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Téléphone invalide.");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { companyPhoneForm: form }
    },

    updateCompanyAddress: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(companyAddressSchema, { defaults: companyAddressDefaults }))

        if (!form.valid) {
            return fail(400, { companyAddressForm: form })
        }

        if (MOCK_MODE) {
            console.log('🏠 [MOCK] Updating company address:', form.data.address)
            return { companyAddressForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/address`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ address: form.data.address })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Adresse invalide.");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { companyAddressForm: form }
    },

    updateCompanyInstagram: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(companyInstagramSchema, { defaults: companyInstagramDefaults }))

        if (!form.valid) {
            return fail(400, { companyInstagramForm: form })
        }

        if (MOCK_MODE) {
            console.log('📸 [MOCK] Updating company Instagram:', form.data.instagram)
            return { companyInstagramForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/instagram`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ instagram: form.data.instagram })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "URL Instagram invalide.");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { companyInstagramForm: form }
    },

    updateOtpDuration: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(otpDurationSchema, { defaults: otpDurationDefaults }))

        if (!form.valid) {
            return fail(400, { otpDurationForm: form })
        }

        if (MOCK_MODE) {
            console.log('⏱️ [MOCK] Updating OTP duration:', form.data.duration)
            return { otpDurationForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/otp/duration`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ duration: form.data.duration })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Durée invalide (60-3600 secondes).");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { otpDurationForm: form }
    },

    updateOtpLength: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(otpLengthSchema, { defaults: otpLengthDefaults }))

        if (!form.valid) {
            return fail(400, { otpLengthForm: form })
        }

        if (MOCK_MODE) {
            console.log('🔢 [MOCK] Updating OTP length:', form.data.length)
            return { otpLengthForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/otp/length`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ length: form.data.length })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Longueur invalide (4-10 chiffres).");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { otpLengthForm: form }
    },

    updateOtpMaxAttempts: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(otpMaxAttemptsSchema, { defaults: otpMaxAttemptsDefaults }))

        if (!form.valid) {
            return fail(400, { otpMaxAttemptsForm: form })
        }

        if (MOCK_MODE) {
            console.log('🚫 [MOCK] Updating OTP max attempts:', form.data.max_attempts)
            return { otpMaxAttemptsForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/otp/max-attempts`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ max_attempts: form.data.max_attempts })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Nombre de tentatives invalide (1-10).");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { otpMaxAttemptsForm: form }
    },

    updateAccessTokenDuration: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(accessTokenDurationSchema, { defaults: accessTokenDurationDefaults }))

        if (!form.valid) {
            return fail(400, { accessTokenDurationForm: form })
        }

        if (MOCK_MODE) {
            console.log('🔑 [MOCK] Updating access token duration:', form.data.duration)
            return { accessTokenDurationForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/tokens/access-duration`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ duration: form.data.duration })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Durée invalide (1-240 minutes).");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { accessTokenDurationForm: form }
    },

    updateRefreshTokenDuration: async ({ request, fetch, cookies }) => {
        const form = await superValidate(request, arktype(refreshTokenDurationSchema, { defaults: refreshTokenDurationDefaults }))

        if (!form.valid) {
            return fail(400, { refreshTokenDurationForm: form })
        }

        if (MOCK_MODE) {
            console.log('🔄 [MOCK] Updating refresh token duration:', form.data.duration)
            return { refreshTokenDurationForm: form }
        }

        const sessionCookie = cookies.get('session');

        const res = await fetch(`${env.API_URL}/admin/settings/tokens/refresh-duration`, {
            method: "POST",
            headers: {
                'Content-Type': "application/json",
                'Authorization': `Bearer ${sessionCookie}`,
            },
            body: JSON.stringify({ duration: form.data.duration })
        })

        if (!res.ok) {
            switch (res.status) {
                case 400:
                    return setError(form, "Durée invalide (1-720 heures).");
                case 401:
                    return setError(form, "Non autorisé. Veuillez vous reconnecter.");
                case 403:
                    return setError(form, "Accès refusé. Vous devez être administrateur.");
                case 415:
                    return setError(form, "Type de contenu non supporté.");
                case 500:
                    return setError(form, "Erreur serveur. Veuillez réessayer.");
                case 503:
                    return setError(form, "Service temporairement indisponible.");
                default:
                    return setError(form, "Une erreur inattendue est survenue.");
            }
        }

        return { refreshTokenDurationForm: form }
    },
} satisfies Actions
