import { env } from "$env/dynamic/private";
import { parseDeposit, type Deposit } from "$lib/deposit";
import { mockPartners, mockProducts, mockPrices, mockUsers } from "$lib/data/mockData";

// Server-side loading for the practice page (issue 009), shared by `/` (fr)
// and `/en`. The page is fed by real data: the practice page settings name
// the partner and product; their public endpoints provide the profile, the
// product and the price; settings provide the practice address and email.
// Everything degrades to null placeholders so the page still renders when a
// setting is unset or the backend is briefly unreachable.

export type PracticePartner = {
    firstName: string;
    lastName: string;
    occupation: string;
    bio: string;
};

export type PracticeProduct = {
    name: string;
    durationMin: number;
    priceCents: number | null;
    /** Optional deposit (acompte, issue 013); null: the full price is paid online. */
    deposit: Deposit | null;
    /** Hours before the session until which cancellation is free (seed placeholder). */
    freeCancellationHours: number;
};

export type PracticePageData = {
    partner: PracticePartner | null;
    product: PracticeProduct | null;
    address: string;
    email: string;
    videoUrl: string | null;
};

const FALLBACK_COMPANY_EMAIL = "contact@leviosa.care";

async function getJson<T>(fetchFn: typeof fetch, path: string): Promise<T | null> {
    try {
        const res = await fetchFn(`${env.API_URL}${path}`);
        if (!res.ok) return null;
        return (await res.json()) as T;
    } catch {
        return null;
    }
}

type PartnerSettings = { partner_id: string };
type ProductSettings = { product_id: string };
type VideoSettings = { video_url: string };
type AddressSettings = { address: string };
type EmailSettings = { email: string };

type PublicPartner = {
    id: string;
    first_name: string;
    last_name: string;
    occupation: string;
    bio: string;
};

type PublicProduct = {
    product: {
        id: string;
        name: string;
        duration: number;
        cancellationHours: number;
        deposit?: unknown;
    };
};

type PublicPrice = {
    amount: number;
    interval: string;
};

function mockPracticeData(): PracticePageData {
    const product = mockProducts.find((p) => p.status === "published") ?? null;
    const partner = mockPartners[0] ?? null;
    const user = partner ? mockUsers.find((u) => u.id === partner.user_id) : null;
    const price =
        product && mockPrices.find((p) => p.productId === product.id && p.isActive && p.interval === "one_time");

    return {
        partner: partner
            ? {
                  firstName: user?.first_name ?? "",
                  lastName: user?.last_name ?? "",
                  occupation: partner.occupation ?? "",
                  bio: partner.bio ?? "",
              }
            : null,
        product: product
            ? {
                  name: product.name,
                  durationMin: product.duration,
                  priceCents: price?.amount ?? null,
                  deposit: parseDeposit(product.deposit),
                  freeCancellationHours: 24,
              }
            : null,
        address: "12 rue des Lilas, 75011 Paris",
        email: FALLBACK_COMPANY_EMAIL,
        videoUrl: null, // mock mode has no practice video: the placeholder frame shows
    };
}

export async function loadPracticePage(fetchFn: typeof fetch): Promise<PracticePageData> {
    if (env.USE_MOCK_DATA === "true") {
        return mockPracticeData();
    }

    const [partnerSettings, productSettings, videoSettings, addressSettings, emailSettings] = await Promise.all([
        getJson<PartnerSettings>(fetchFn, "/settings/practice-page/partner"),
        getJson<ProductSettings>(fetchFn, "/settings/practice-page/product"),
        getJson<VideoSettings>(fetchFn, "/settings/practice-page/video"),
        getJson<AddressSettings>(fetchFn, "/settings/address"),
        getJson<EmailSettings>(fetchFn, "/settings/email"),
    ]);

    let partner: PracticePartner | null = null;
    if (partnerSettings?.partner_id) {
        const publicPartner = await getJson<PublicPartner>(fetchFn, `/partners/${partnerSettings.partner_id}`);
        if (publicPartner) {
            partner = {
                firstName: publicPartner.first_name ?? "",
                lastName: publicPartner.last_name ?? "",
                occupation: publicPartner.occupation ?? "",
                bio: publicPartner.bio ?? "",
            };
        }
    }

    let product: PracticeProduct | null = null;
    if (productSettings?.product_id) {
        const publicProduct = await getJson<PublicProduct>(fetchFn, `/products/${productSettings.product_id}`);
        if (publicProduct?.product) {
            const prices = await getJson<PublicPrice[]>(fetchFn, `/products/${productSettings.product_id}/prices`);
            const price = prices?.find((p) => p.interval === "one_time") ?? prices?.[0] ?? null;
            product = {
                name: publicProduct.product.name,
                durationMin: publicProduct.product.duration,
                priceCents: price?.amount ?? null,
                deposit: parseDeposit(publicProduct.product.deposit),
                freeCancellationHours: publicProduct.product.cancellationHours ?? 24,
            };
        }
    }

    return {
        partner,
        product,
        address: addressSettings?.address ?? "",
        email: emailSettings?.email || FALLBACK_COMPANY_EMAIL,
        videoUrl: videoSettings?.video_url || null,
    };
}
