<script lang="ts">
    import { ArrowDown, Clock, CreditCard, MapPin } from "@lucide/svelte";
    import { dictionary, fmt, infoBlocks, type Locale } from "#lib/i18n/practice.ts";
    import Header from "./_header.svelte";
    import Video from "./_video.svelte";
    import Booking from "./_booking.svelte";
    import Timeline from "./_timeline.svelte";
    import Footer from "./_footer.svelte";
    import type { PracticePageData } from "./practice";
    import { amountDueOnlineCents } from "#lib/deposit.ts";

    // The practice page (issue 009), ported from the chosen prototype
    // (variant D): a black hero band, a booking card, the session explained
    // minute by minute, practical points, and the footer with the legal
    // links. Everything client-facing comes from `data` — the partner and
    // product chosen in settings — and every missing value renders as an
    // honest em dash rather than invented content.
    let { locale, data }: { locale: Locale; data: PracticePageData } = $props();

    const t = $derived(dictionary[locale]);
    const info = $derived(infoBlocks[locale]);

    const dash = (value: string) => (value.trim() ? value : "—");
    const partnerName = $derived(
        data.partner ? dash(`${data.partner.firstName} ${data.partner.lastName}`.trim()) : "—",
    );
    const price = $derived(data.product?.priceCents != null ? fmt.price(data.product.priceCents, locale) : "—");

    // Price, deposit and balance (issue 013). With a deposit below the price,
    // only the deposit is paid online and the balance at the practice; with
    // none (or one covering the whole price) the full price is paid online.
    const payment = $derived.by(() => {
        const priceCents = data.product?.priceCents;
        if (priceCents == null) return null;
        const onlineCents = amountDueOnlineCents(data.product?.deposit, priceCents);
        if (onlineCents >= priceCents) return { deposit: null, balance: null };
        return {
            deposit: fmt.price(onlineCents, locale),
            balance: fmt.price(priceCents - onlineCents, locale),
        };
    });

    const facts = $derived([
        { term: t.duration, value: data.product ? t.minutes(data.product.durationMin) : "—" },
        { term: t.price, value: payment?.deposit ? t.priceWithDeposit(price, payment.deposit) : price },
        { term: t.address, value: dash(data.address) },
        { term: t.payment, value: payment?.deposit ? t.paymentDeposit : t.paymentValue },
    ]);
</script>

<svelte:head>
    <title>{data.partner ? `${partnerName} — leviosa.care` : "leviosa.care"}</title>
    <meta name="description" content={data.partner?.bio || (data.partner ? `${partnerName} — ${t.bookSession}` : t.bookSession)} />
</svelte:head>

<Header {locale} />

<div class="bg-neutral-950 text-neutral-300 antialiased">
    <!-- Skip link: keyboard users reach the booking card without tabbing through the pitch. -->
    <a
        href="#reserver"
        class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:rounded-lg focus:bg-white focus:px-4 focus:py-2 focus:font-semibold focus:text-neutral-950"
    >
        {t.skipToBooking}
    </a>

    <div id="top"></div>

    <!-- Hero band: the practice video (or its placeholder frame), the partner's
         name, title and bio, the call to action, and the key facts. -->
    <section class="text-neutral-300">
        <div class="mx-auto w-full max-w-5xl px-4 pt-6 pb-10 sm:pt-10 sm:pb-14">
            <div class="grid gap-8 sm:gap-10 lg:grid-cols-[3fr_2fr] lg:items-end lg:gap-12">
                <div class="aspect-video overflow-hidden rounded-sm bg-neutral-900">
                    <Video label={t.videoMissing} src={data.videoUrl} />
                </div>
                <div class="flex flex-col items-start gap-6">
                    <h1 class="text-[2.5rem] leading-none font-black tracking-[-0.04em] text-white uppercase sm:text-5xl">
                        {partnerName}
                        {#if data.partner?.occupation}
                            <span class="block font-light text-neutral-400">/ {data.partner.occupation}</span>
                        {/if}
                    </h1>
                    {#if data.partner?.bio}
                        <p class="max-w-prose text-pretty">{data.partner.bio}</p>
                    {/if}
                    <a
                        href="#reserver"
                        class="inline-flex h-12 items-center gap-2 rounded-lg bg-white px-5 text-base font-semibold text-neutral-950 transition-colors hover:bg-neutral-200 focus-visible:ring-3 focus-visible:ring-white/50 focus-visible:outline-none"
                    >
                        {t.bookSession}
                        <ArrowDown class="size-4" aria-hidden="true" />
                    </a>
                </div>
            </div>

            <dl class="mt-10 grid grid-cols-2 gap-x-8 sm:mt-14 lg:grid-cols-4">
                {#each facts as fact (fact.term)}
                    <div class="border-t border-neutral-800 py-4">
                        <dt class="text-xs font-medium tracking-widest text-neutral-400 uppercase">{fact.term}</dt>
                        <dd class="mt-1.5 font-semibold text-white">{fact.value}</dd>
                    </div>
                {/each}
            </dl>
        </div>
    </section>

    <!-- Booking card: the heading, intro and session summary in the left
         panel; the calendar of free slots on the right. -->
    <section id="reserver" class="scroll-mt-16 bg-neutral-50 px-4 py-14 text-neutral-950 sm:py-20">
        <div
            class="mx-auto grid max-w-5xl overflow-hidden rounded-2xl border border-neutral-200 bg-white shadow-[0_1px_2px_rgb(0_0_0/0.04),0_8px_24px_-12px_rgb(0_0_0/0.12)] lg:grid-cols-[minmax(0,19rem)_minmax(0,1fr)]"
        >
            <aside class="space-y-5 border-b border-neutral-200 p-6 sm:p-8 lg:border-r lg:border-b-0">
                <h2 class="text-2xl font-bold tracking-tight">{t.bookSession}</h2>
                <p class="text-pretty text-neutral-600">
                    {t.intro(data.product?.fullRefundHours ?? null, payment?.deposit ?? null)}
                </p>
                <ul class="space-y-2.5 border-t border-neutral-200 pt-5 text-sm text-neutral-700">
                    <li class="flex items-center gap-2.5">
                        <Clock class="size-4 shrink-0 text-neutral-500" aria-hidden="true" />
                        {#if data.product}
                            {data.product.name} · {t.minutes(data.product.durationMin)} · {price}
                        {:else}
                            —
                        {/if}
                    </li>
                    <li class="flex items-center gap-2.5">
                        <MapPin class="size-4 shrink-0 text-neutral-500" aria-hidden="true" />
                        {dash(data.address)}
                    </li>
                    <li class="flex items-center gap-2.5">
                        <CreditCard class="size-4 shrink-0 text-neutral-500" aria-hidden="true" />
                        {#if payment?.deposit && payment.balance}
                            {t.paymentSummaryDeposit(payment.deposit, payment.balance)}
                        {:else if payment}
                            {t.paymentSummaryFull(price)}
                        {:else}
                            {t.paymentValue}
                        {/if}
                    </li>
                </ul>
            </aside>
            <div class="p-6 sm:p-8">
                <Booking {locale} booking={data.booking} priceCents={data.product?.priceCents ?? null} />
            </div>
        </div>
    </section>

    <!-- The session explained: the hour as a proportional timeline, then the
         practical points as a disclosure list (native details/summary, so it
         works without JavaScript). -->
    <section class="bg-white px-4 py-16 text-neutral-950 sm:py-24">
        <div class="mx-auto max-w-5xl">
            <div class="max-w-2xl">
                <h2 class="text-3xl font-black tracking-[-0.03em] text-balance uppercase sm:text-4xl">{info.heading}</h2>
                <p class="mt-3 text-pretty text-neutral-600">{info.intro}</p>
            </div>
            <div class="mt-10 sm:mt-12">
                <Timeline {locale} />
            </div>

            <div class="mt-16 grid gap-6 sm:mt-20 lg:grid-cols-[minmax(0,19rem)_minmax(0,1fr)] lg:gap-12">
                <h3 class="text-xl font-bold tracking-tight">{info.practicalHeading}</h3>
                <div class="divide-y divide-neutral-200 border-y border-neutral-200">
                    {#each info.practical as item (item.title)}
                        <details class="group">
                            <summary
                                class="flex cursor-pointer list-none items-center justify-between gap-4 py-4 font-semibold focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none [&::-webkit-details-marker]:hidden"
                            >
                                {item.title}
                                <span
                                    class="text-2xl leading-none font-light text-neutral-600 transition-transform duration-200 group-open:rotate-45"
                                    aria-hidden="true"
                                >+</span>
                            </summary>
                            <p class="max-w-prose pb-5 leading-relaxed text-pretty text-neutral-600">{item.body}</p>
                        </details>
                    {/each}
                </div>
            </div>
        </div>
    </section>

    <Footer {locale} {data} />
</div>
