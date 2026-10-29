// Product deposit (French UI: *acompte*, never *arrhes*) — issue 013.
// Mirrors backend/internal/catalog/domain/deposit.go: the backend is the
// source of truth and validates every deposit; these helpers only shape the
// admin form and the amounts shown on the practice page.

export type DepositType = "fixed" | "percent";

/** fixed: `value` in cents; percent: `value` in whole percentage points of the price. */
export type Deposit = { type: DepositType; value: number };

/**
 * The amount paid online when booking: the deposit when one is set,
 * otherwise the full price, rounded to the cent (same rule as the backend's
 * AmountDueOnline).
 */
export function amountDueOnlineCents(deposit: Deposit | null | undefined, priceCents: number): number {
    if (!deposit) return priceCents;
    if (deposit.type === "fixed") return Math.min(deposit.value, priceCents);
    return Math.round((priceCents * deposit.value) / 100);
}

/** Reads a deposit coming from the API, dropping anything malformed. */
export function parseDeposit(raw: unknown): Deposit | null {
    if (!raw || typeof raw !== "object") return null;
    const { type, value } = raw as { type?: unknown; value?: unknown };
    if ((type === "fixed" || type === "percent") && typeof value === "number") return { type, value };
    return null;
}

/**
 * Turns the admin form's `depositType` / `depositValue` fields into the API
 * payload. A fixed amount is typed in euros and sent in cents.
 */
export function depositFromForm(
    type: FormDataEntryValue | null,
    value: FormDataEntryValue | null,
): { ok: true; deposit: Deposit | null } | { ok: false; error: string } {
    if (type == null || type === "" || type === "none") return { ok: true, deposit: null };
    const n = Number(String(value ?? "").replace(",", "."));
    if (type === "fixed") {
        if (!Number.isFinite(n) || n < 0) return { ok: false, error: "Le montant de l'acompte doit être positif." };
        return { ok: true, deposit: { type: "fixed", value: Math.round(n * 100) } };
    }
    if (type === "percent") {
        if (!Number.isInteger(n) || n <= 0 || n > 100) {
            return { ok: false, error: "Le pourcentage d'acompte doit être un entier entre 1 et 100." };
        }
        return { ok: true, deposit: { type: "percent", value: n } };
    }
    return { ok: false, error: "Type d'acompte invalide." };
}
