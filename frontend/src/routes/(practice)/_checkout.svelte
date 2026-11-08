<script lang="ts">
    import { ChevronLeft, LoaderCircle, CircleCheck } from "@lucide/svelte";
    import { onDestroy, onMount, tick } from "svelte";
    import { env } from "$env/dynamic/public";
    import type { Stripe, StripeElements } from "@stripe/stripe-js";
    import { cap, dictionary, fmt, type Locale } from "$lib/i18n/practice";
    import { isFrenchPhone, normalizePhone } from "$lib/phone";

    // The booking card's details step and payment (issue 014, ADR-0018).
    // The client gives their details and accepts the terms of sale; the
    // booking is then created awaiting payment, holding the slot for the
    // payment hold, and the Stripe Payment Element is mounted with the
    // PaymentIntent's client secret. Once Stripe accepts the payment, the
    // card polls the booking by its token until the webhook has confirmed
    // it. No free-text field: Leviosa stores no health data (ADR-0019).
    //
    // When the payment hold runs out (issue 015) — the client is still on
    // the Payment Element, clicks pay too late, or Stripe refuses because the
    // expiry job canceled the intent — the card says so in the page's
    // language, never with Stripe's raw error, and sends the client back to
    // the calendar, whose slots are refreshed.
    type Props = {
        locale: Locale;
        productId: string;
        availabilityId: string;
        start: Date;
        priceCents: number | null;
        /** Back to the calendar; slotTaken: the chosen slot was taken meanwhile. */
        onBack: (slotTaken?: boolean) => void;
        /** Resume after a redirect-based payment method: poll this booking token. */
        resumeToken?: string | null;
    };
    let { locale, productId, availabilityId, start, priceCents, onBack, resumeToken = null }: Props = $props();

    const t = $derived(dictionary[locale]);
    const publishableKey = env.PUBLIC_STRIPE_PUBLISHABLE_KEY ?? "";

    type Phase = "details" | "payment" | "confirming" | "confirmed" | "delayed" | "lost" | "expired";
    let phase = $state<Phase>(resumeToken ? "confirming" : "details");

    let firstName = $state("");
    let lastName = $state("");
    let email = $state("");
    let phone = $state("");
    let acceptTerms = $state(false);
    let submitted = $state(false);
    let busy = $state(false);
    let formError = $state("");

    let token = $state(resumeToken ?? "");
    let amountDueCents = $state(0);
    let holdExpiresAt = $state<Date | null>(null);
    /** Expired phase: the late payment was refunded (the slot was taken meanwhile). */
    let refunded = $state(false);

    let heading = $state<HTMLElement | null>(null);
    let paymentMount = $state<HTMLElement | null>(null);
    let stripe: Stripe | null = null;
    let elements = $state.raw<StripeElements | null>(null);
    let pollTimer: ReturnType<typeof setTimeout> | null = null;
    let holdTimer: ReturnType<typeof setTimeout> | null = null;

    const holdElapsed = () => holdExpiresAt != null && Date.now() >= holdExpiresAt.getTime();

    async function showExpired(wasRefunded = false) {
        if (holdTimer) clearTimeout(holdTimer);
        refunded = wasRefunded;
        formError = "";
        phase = "expired";
        await focusHeading();
    }

    /** Leaves the payment step once the hold runs out, unless a payment is under way. */
    function watchHold() {
        if (holdTimer) clearTimeout(holdTimer);
        if (!holdExpiresAt) return;
        // setTimeout caps at about 24 days; a hold is minutes long.
        const delay = Math.max(0, Math.min(holdExpiresAt.getTime() - Date.now(), 2 ** 31 - 1));
        holdTimer = setTimeout(() => {
            if (phase !== "payment") return;
            if (busy) {
                watchHold(); // decide once Stripe answered
                return;
            }
            void showExpired();
        }, delay + 500);
    }

    const errors = $derived({
        firstName: firstName.trim() ? "" : t.requiredField,
        lastName: lastName.trim() ? "" : t.requiredField,
        email: !email.trim() ? t.requiredField : /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()) ? "" : t.invalidEmail,
        phone: !phone.trim() ? t.requiredField : isFrenchPhone(phone) ? "" : t.invalidPhone,
        terms: acceptTerms ? "" : t.termsRequired,
    });
    const valid = $derived(Object.values(errors).every((message) => !message));

    const payLabel = $derived.by(() => {
        const amount = fmt.price(amountDueCents, locale);
        return priceCents != null && amountDueCents < priceCents ? t.payDeposit(amount) : t.payFull(amount);
    });

    async function focusHeading() {
        await tick();
        heading?.focus();
    }

    async function submitDetails(event: SubmitEvent) {
        event.preventDefault();
        submitted = true;
        formError = "";
        if (!valid) {
            await tick();
            document.querySelector<HTMLElement>("#checkout [aria-invalid='true']")?.focus();
            return;
        }
        if (!publishableKey) {
            formError = t.paymentUnavailable;
            return;
        }

        busy = true;
        try {
            const res = await fetch("/api/bookings", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    availabilityId,
                    productId,
                    slotStart: start.toISOString(),
                    firstName: firstName.trim(),
                    lastName: lastName.trim(),
                    email: email.trim(),
                    phone,
                    acceptTerms,
                    locale,
                }),
            });
            if (res.status === 409) {
                onBack(true);
                return;
            }
            if (!res.ok) {
                formError = t.bookingFailed;
                return;
            }
            const created = (await res.json()) as {
                token: string;
                status: string;
                clientSecret: string;
                amountDueCents: number;
                holdExpiresAt: string | null;
            };
            token = created.token;
            amountDueCents = created.amountDueCents;
            holdExpiresAt = created.holdExpiresAt ? new Date(created.holdExpiresAt) : null;

            if (created.status === "confirmed" || !created.clientSecret) {
                // Nothing was due online: the booking is confirmed already.
                phase = "confirmed";
                await focusHeading();
                return;
            }

            phase = "payment";
            watchHold();
            await focusHeading();
            await mountPayment(created.clientSecret);
        } catch {
            formError = t.bookingFailed;
        } finally {
            busy = false;
        }
    }

    async function mountPayment(clientSecret: string) {
        try {
            const { loadStripe } = await import("@stripe/stripe-js");
            stripe = await loadStripe(publishableKey, { locale });
            if (!stripe || !paymentMount) throw new Error("stripe unavailable");
            elements = stripe.elements({
                clientSecret,
                locale,
                appearance: {
                    theme: "stripe",
                    variables: {
                        colorPrimary: "#0a0a0a",
                        colorText: "#0a0a0a",
                        colorDanger: "#e21d48",
                        borderRadius: "8px",
                        fontSizeBase: "16px",
                    },
                },
            });
            // The details step's values, so Link's "save my information"
            // form does not ask for them again. Link wants the phone in
            // international form: 0612345678 → +33612345678.
            elements
                .create("payment", {
                    layout: "tabs",
                    defaultValues: {
                        billingDetails: {
                            name: `${firstName.trim()} ${lastName.trim()}`,
                            email: email.trim(),
                            phone: "+33" + normalizePhone(phone).slice(1),
                        },
                    },
                })
                .mount(paymentMount);
        } catch {
            formError = t.paymentLoadFailed;
        }
    }

    async function pay(event: SubmitEvent) {
        event.preventDefault();
        if (!stripe || !elements || busy) return;
        if (holdElapsed()) {
            // The slot is no longer held: don't take a payment for it.
            await showExpired();
            return;
        }
        busy = true;
        formError = "";
        try {
            // Card payments complete here; a redirect-based method comes back
            // to the practice page with the booking token (see _booking.svelte).
            const returnUrl = new URL(window.location.href);
            returnUrl.search = "";
            returnUrl.hash = "reserver";
            returnUrl.searchParams.set("booking", token);
            const result = await stripe.confirmPayment({
                elements,
                redirect: "if_required",
                confirmParams: { return_url: returnUrl.toString() },
            });
            if (result.error) {
                // The intent was canceled by the expiry job, or the hold ran
                // out while the client was paying: the slot is gone, so say
                // that rather than Stripe's message.
                if (
                    result.error.code === "payment_intent_unexpected_state" ||
                    result.error.payment_intent?.status === "canceled" ||
                    holdElapsed()
                ) {
                    await showExpired();
                    return;
                }
                // Stripe's message is already in the page's language.
                formError = result.error.message ?? t.bookingFailed;
                return;
            }
            if (holdTimer) clearTimeout(holdTimer);
            phase = "confirming";
            await focusHeading();
            poll(0);
        } finally {
            busy = false;
        }
    }

    /** Polls the booking by token until the webhook confirmed it (about a minute at most). */
    function poll(attempt: number) {
        if (!token) {
            phase = "delayed";
            return;
        }
        pollTimer = setTimeout(async () => {
            try {
                const res = await fetch(`/api/bookings/lookup?token=${encodeURIComponent(token)}`);
                if (res.ok) {
                    const booking = (await res.json()) as { status: string; paymentStatus?: string };
                    if (booking.status === "confirmed") {
                        phase = "confirmed";
                        await focusHeading();
                        return;
                    }
                    // A payment that reached an expired booking is either
                    // confirmed (the slot still free) or refunded: until the
                    // webhook settled it, keep waiting.
                    if (booking.status === "expired" && booking.paymentStatus === "refunded") {
                        await showExpired(true);
                        return;
                    }
                    if (booking.status === "cancelled") {
                        phase = "lost";
                        await focusHeading();
                        return;
                    }
                }
            } catch {
                // keep polling: a transient network error is not an answer
            }
            if (attempt >= 30) {
                phase = "delayed";
                await focusHeading();
                return;
            }
            poll(attempt + 1);
        }, attempt === 0 ? 1000 : 2000);
    }

    onMount(() => {
        if (resumeToken) poll(0);
    });

    onDestroy(() => {
        if (pollTimer) clearTimeout(pollTimer);
        if (holdTimer) clearTimeout(holdTimer);
    });

    const inputClass =
        "mt-1.5 block h-12 w-full min-w-0 rounded-lg border bg-white px-3 text-base text-neutral-950 transition-colors placeholder:text-neutral-500 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none";
    const borderFor = (message: string) =>
        submitted && message ? "border-rose-600 focus-visible:ring-rose-600/25" : "border-neutral-300 focus-visible:border-neutral-950";
</script>

<div id="checkout" class="max-w-md min-w-0">
    {#if phase === "details" || phase === "payment"}
        <button
            type="button"
            onclick={() => onBack()}
            class="-ml-2 mb-6 inline-flex h-11 cursor-pointer items-center gap-1.5 rounded-lg px-2 text-sm font-medium text-neutral-700 transition-colors hover:bg-neutral-100 hover:text-neutral-950 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
        >
            <ChevronLeft class="size-4" aria-hidden="true" />
            {t.changeSlot}
        </button>
        <p class="font-semibold tabular-nums">
            {cap(fmt.long(start, locale))} · {fmt.time(start, locale)}
            <span class="font-normal text-neutral-600">({t.timeZoneNote})</span>
        </p>
    {/if}

    {#if phase === "details"}
        <h3 bind:this={heading} tabindex="-1" class="mt-6 text-lg font-semibold focus:outline-none">
            {t.detailsHeading}
        </h3>
        <p class="mt-1 text-sm text-pretty text-neutral-600">{t.detailsIntro}</p>

        <form class="mt-5 space-y-4" novalidate onsubmit={submitDetails}>
            <div class="grid gap-4 sm:grid-cols-2">
                <div class="min-w-0">
                    <label for="checkout-first-name" class="text-sm font-medium">{t.firstName}</label>
                    <input
                        id="checkout-first-name"
                        name="given-name"
                        autocomplete="given-name"
                        required
                        bind:value={firstName}
                        aria-invalid={submitted && !!errors.firstName}
                        aria-describedby={submitted && errors.firstName ? "checkout-first-name-error" : undefined}
                        class="{inputClass} {borderFor(errors.firstName)}"
                    />
                    {#if submitted && errors.firstName}
                        <p id="checkout-first-name-error" class="mt-1 text-sm text-rose-700">{errors.firstName}</p>
                    {/if}
                </div>
                <div class="min-w-0">
                    <label for="checkout-last-name" class="text-sm font-medium">{t.lastName}</label>
                    <input
                        id="checkout-last-name"
                        name="family-name"
                        autocomplete="family-name"
                        required
                        bind:value={lastName}
                        aria-invalid={submitted && !!errors.lastName}
                        aria-describedby={submitted && errors.lastName ? "checkout-last-name-error" : undefined}
                        class="{inputClass} {borderFor(errors.lastName)}"
                    />
                    {#if submitted && errors.lastName}
                        <p id="checkout-last-name-error" class="mt-1 text-sm text-rose-700">{errors.lastName}</p>
                    {/if}
                </div>
            </div>
            <div>
                <label for="checkout-email" class="text-sm font-medium">{t.email}</label>
                <input
                    id="checkout-email"
                    name="email"
                    type="email"
                    autocomplete="email"
                    inputmode="email"
                    required
                    bind:value={email}
                    aria-invalid={submitted && !!errors.email}
                    aria-describedby={submitted && errors.email ? "checkout-email-error" : undefined}
                    class="{inputClass} {borderFor(errors.email)}"
                />
                {#if submitted && errors.email}
                    <p id="checkout-email-error" class="mt-1 text-sm text-rose-700">{errors.email}</p>
                {/if}
            </div>
            <div>
                <label for="checkout-phone" class="text-sm font-medium">{t.phone}</label>
                <input
                    id="checkout-phone"
                    name="tel"
                    type="tel"
                    autocomplete="tel"
                    inputmode="tel"
                    required
                    bind:value={phone}
                    aria-invalid={submitted && !!errors.phone}
                    aria-describedby="checkout-phone-hint{submitted && errors.phone ? ' checkout-phone-error' : ''}"
                    class="{inputClass} {borderFor(errors.phone)}"
                />
                <p id="checkout-phone-hint" class="mt-1 text-sm text-neutral-600">{t.phoneHint}</p>
                {#if submitted && errors.phone}
                    <p id="checkout-phone-error" class="mt-1 text-sm text-rose-700">{errors.phone}</p>
                {/if}
            </div>
            <div>
                <div class="flex items-start gap-3">
                    <input
                        id="checkout-terms"
                        type="checkbox"
                        required
                        bind:checked={acceptTerms}
                        aria-invalid={submitted && !!errors.terms}
                        aria-describedby={submitted && errors.terms ? "checkout-terms-error" : undefined}
                        class="mt-0.5 size-5 shrink-0 cursor-pointer accent-neutral-950"
                    />
                    <label for="checkout-terms" class="text-sm text-pretty">
                        {t.acceptTerms}
                        <a
                            href="/legal/terms"
                            target="_blank"
                            rel="noopener"
                            class="font-medium underline underline-offset-2 hover:text-neutral-700 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
                            >{t.termsLink}</a
                        >.
                    </label>
                </div>
                {#if submitted && errors.terms}
                    <p id="checkout-terms-error" class="mt-1 text-sm text-rose-700">{errors.terms}</p>
                {/if}
            </div>

            {#if formError}
                <p role="alert" class="text-sm text-pretty text-rose-700">{formError}</p>
            {/if}

            <button
                type="submit"
                disabled={busy}
                class="inline-flex h-12 w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-neutral-950 px-5 text-base font-semibold text-white transition-colors hover:bg-neutral-800 focus-visible:ring-3 focus-visible:ring-neutral-950/30 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-70"
            >
                {#if busy}
                    <LoaderCircle class="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" />
                    {t.holdingSlot}
                {:else}
                    {t.continueToPayment}
                {/if}
            </button>
        </form>
    {:else if phase === "payment"}
        <h3 bind:this={heading} tabindex="-1" class="mt-6 text-lg font-semibold focus:outline-none">
            {t.paymentHeading}
        </h3>
        {#if holdExpiresAt}
            <p class="mt-1 text-sm text-pretty text-neutral-600">{t.holdNotice(fmt.time(holdExpiresAt, locale))}</p>
        {/if}

        <form class="mt-5 space-y-4" onsubmit={pay}>
            <div bind:this={paymentMount} class="min-h-40 min-w-0"></div>

            {#if formError}
                <p role="alert" class="text-sm text-pretty text-rose-700">{formError}</p>
            {/if}

            <button
                type="submit"
                disabled={busy || !elements}
                class="inline-flex h-12 w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-neutral-950 px-5 text-base font-semibold text-white transition-colors hover:bg-neutral-800 focus-visible:ring-3 focus-visible:ring-neutral-950/30 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-70"
            >
                {#if busy}
                    <LoaderCircle class="size-4 animate-spin motion-reduce:animate-none" aria-hidden="true" />
                    {t.paying}
                {:else}
                    {payLabel}
                {/if}
            </button>
            <p class="text-xs text-pretty text-neutral-600">{t.securePayment}</p>
        </form>
    {:else if phase === "confirming"}
        <div role="status" class="flex items-start gap-3">
            <LoaderCircle class="mt-1 size-5 shrink-0 animate-spin motion-reduce:animate-none" aria-hidden="true" />
            <div>
                <h3 bind:this={heading} tabindex="-1" class="text-lg font-semibold focus:outline-none">
                    {t.confirmingHeading}
                </h3>
                <p class="mt-1 text-sm text-pretty text-neutral-600">{t.confirmingBody}</p>
            </div>
        </div>
    {:else if phase === "confirmed"}
        <div role="status">
            <CircleCheck class="size-8 text-neutral-950" aria-hidden="true" />
            <h3 bind:this={heading} tabindex="-1" class="mt-4 text-2xl font-bold tracking-tight focus:outline-none">
                {t.confirmedHeading}
            </h3>
            {#if !resumeToken}
                <p class="mt-3 font-semibold tabular-nums">
                    {cap(fmt.long(start, locale))} · {fmt.time(start, locale)}
                    <span class="font-normal text-neutral-600">({t.timeZoneNote})</span>
                </p>
            {/if}
            <p class="mt-3 text-pretty text-neutral-600">
                {email.trim() ? t.confirmedEmail(email.trim()) : t.confirmedNoEmail}
            </p>
            {#if token}
                <a
                    href="/bookings?token={encodeURIComponent(token)}"
                    class="mt-6 inline-flex h-12 items-center rounded-lg border border-neutral-300 px-5 font-semibold transition-colors hover:border-neutral-950 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
                >
                    {t.viewBooking}
                </a>
            {/if}
        </div>
    {:else if phase === "expired"}
        <div role="alert">
            <h3 bind:this={heading} tabindex="-1" class="text-lg font-semibold focus:outline-none">
                {t.holdExpiredHeading}
            </h3>
            <p class="mt-2 text-pretty text-neutral-600">{refunded ? t.latePaymentRefunded : t.holdExpired}</p>
            <button
                type="button"
                onclick={() => onBack()}
                class="mt-6 inline-flex h-12 cursor-pointer items-center rounded-lg bg-neutral-950 px-5 font-semibold text-white transition-colors hover:bg-neutral-800 focus-visible:ring-3 focus-visible:ring-neutral-950/30 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
                {t.newBooking}
            </button>
        </div>
    {:else}
        <div role="status">
            <h3 bind:this={heading} tabindex="-1" class="text-lg font-semibold focus:outline-none">
                {phase === "delayed" ? t.confirmingHeading : t.paymentHeading}
            </h3>
            <p class="mt-2 text-pretty text-neutral-600">{phase === "delayed" ? t.confirmDelayed : t.bookingLost}</p>
            {#if phase === "lost"}
                <button
                    type="button"
                    onclick={() => onBack()}
                    class="mt-6 inline-flex h-12 cursor-pointer items-center rounded-lg border border-neutral-300 px-5 font-semibold transition-colors hover:border-neutral-950 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
                >
                    {t.newBooking}
                </button>
            {:else if token}
                <a
                    href="/bookings?token={encodeURIComponent(token)}"
                    class="mt-6 inline-flex h-12 items-center rounded-lg border border-neutral-300 px-5 font-semibold transition-colors hover:border-neutral-950 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
                >
                    {t.viewBooking}
                </a>
            {/if}
        </div>
    {/if}
</div>
