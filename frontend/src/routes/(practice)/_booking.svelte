<script lang="ts">
    import { ChevronLeft, ChevronRight } from "@lucide/svelte";
    import { onMount, tick } from "svelte";
    import { cap, dictionary, fmt, type Locale } from "$lib/i18n/practice";
    import type { PracticeBooking, PracticeSlot } from "./practice";
    import Checkout from "./_checkout.svelte";
    import {
        addMonths,
        compareMonths,
        dayKey,
        dayOf,
        monthCells,
        monthDate,
        monthOf,
        type CalendarMonth,
    } from "./_calendar";

    // The booking card's calendar area, fed by the slot engine (GET /slots,
    // issue 010): it opens on the first month with a free slot, marks the
    // days that have one, lists the selected day's times and cannot page
    // past the booking horizon. Picking a time leads to the details step and
    // the payment (issue 014, _checkout.svelte).
    let {
        locale,
        booking,
        priceCents,
    }: { locale: Locale; booking: PracticeBooking | null; priceCents: number | null } = $props();

    const t = $derived(dictionary[locale]);

    type Candidate = { slot: PracticeSlot; start: Date };

    /** Slots another client took while this one was entering their details. */
    let takenStarts = $state<string[]>([]);

    /** The free slots grouped by practice-time-zone day, in order. */
    const byDay = $derived.by(() => {
        const days = new Map<string, Candidate[]>();
        for (const slot of booking?.slots ?? []) {
            if (takenStarts.includes(slot.start)) continue;
            const start = new Date(slot.start);
            const key = dayKey(start);
            const list = days.get(key);
            if (list) list.push({ slot, start });
            else days.set(key, [{ slot, start }]);
        }
        return days;
    });

    const todayKey = dayKey(new Date());
    const firstMonth = monthOf(todayKey);
    const lastMonth = $derived(booking ? monthOf(dayKey(new Date(booking.horizonEnd))) : firstMonth);

    /** The first day with a free slot, if any: where the calendar opens. */
    const firstOpenKey = (): string => byDay.keys().next().value ?? "";

    const openingKey = firstOpenKey();
    let selectedDayKey = $state(openingKey);
    let month = $state<CalendarMonth>(openingKey ? monthOf(openingKey) : firstMonth);
    let chosen = $state<Candidate | null>(null);
    let slotTaken = $state(false);
    /** Back from a redirect-based payment method (?booking=<token>): resume polling. */
    let resumeToken = $state<string | null>(null);
    let calendarHeading = $state<HTMLElement | null>(null);

    const canPrev = $derived(compareMonths(month, firstMonth) > 0);
    const canNext = $derived(compareMonths(month, lastMonth) < 0);
    const cells = $derived(monthCells(month));
    const daySlots = $derived(byDay.get(selectedDayKey) ?? []);

    function choose(candidate: Candidate) {
        slotTaken = false;
        chosen = candidate;
    }

    async function backToCalendar(taken = false) {
        // Another client got the slot first: stop offering it here.
        if (taken && chosen) takenStarts = [...takenStarts, chosen.slot.start];
        slotTaken = taken;
        chosen = null;
        resumeToken = null;
        await tick();
        calendarHeading?.focus();
    }

    onMount(() => {
        const url = new URL(window.location.href);
        const token = url.searchParams.get("booking");
        if (!token) return;
        // Stripe appends redirect_status; a failed redirect payment leaves the
        // client on the calendar rather than waiting for a confirmation.
        if (url.searchParams.get("redirect_status") !== "failed") resumeToken = token;
        url.searchParams.delete("booking");
        url.searchParams.delete("payment_intent");
        url.searchParams.delete("payment_intent_client_secret");
        url.searchParams.delete("redirect_status");
        history.replaceState(history.state, "", url);
    });
</script>

{#if resumeToken && booking}
    <Checkout
        {locale}
        productId={booking.productId}
        availabilityId=""
        start={new Date()}
        {priceCents}
        {resumeToken}
        onBack={backToCalendar}
    />
{:else if chosen && booking}
    <Checkout
        {locale}
        productId={booking.productId}
        availabilityId={chosen.slot.availabilityId}
        start={chosen.start}
        {priceCents}
        onBack={backToCalendar}
    />
{:else if !booking}
    <p class="text-sm text-pretty text-neutral-600">{t.slotsUnavailable}</p>
{:else}
    {#if slotTaken}
        <p role="alert" class="mb-6 text-sm text-pretty text-rose-700">{t.slotTaken}</p>
    {/if}
    <div class="grid gap-8 md:grid-cols-[minmax(0,1fr)_11rem]">
        <div>
            <div class="mb-4 flex items-center justify-between">
                <p bind:this={calendarHeading} tabindex="-1" class="font-semibold focus:outline-none">{cap(fmt.month(monthDate(month), locale))}</p>
                <div class="flex gap-1">
                    <button
                        type="button"
                        aria-label={t.prevMonth}
                        disabled={!canPrev}
                        onclick={() => (month = addMonths(month, -1))}
                        class="flex size-9 cursor-pointer items-center justify-center rounded-lg border border-neutral-200 transition-colors hover:bg-neutral-100 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
                    >
                        <ChevronLeft class="size-4" aria-hidden="true" />
                    </button>
                    <button
                        type="button"
                        aria-label={t.nextMonth}
                        disabled={!canNext}
                        onclick={() => (month = addMonths(month, 1))}
                        class="flex size-9 cursor-pointer items-center justify-center rounded-lg border border-neutral-200 transition-colors hover:bg-neutral-100 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
                    >
                        <ChevronRight class="size-4" aria-hidden="true" />
                    </button>
                </div>
            </div>
            <div class="grid grid-cols-7 gap-1 text-center">
                {#each t.weekdays as weekday (weekday)}
                    <span class="pb-2 text-xs font-medium text-neutral-600 uppercase">{weekday}</span>
                {/each}
                {#each cells as cell, index (index)}
                    {#if cell}
                        {@const open = byDay.get(cell.key)?.length ?? 0}
                        {@const selected = selectedDayKey === cell.key}
                        <button
                            type="button"
                            disabled={open === 0}
                            aria-pressed={selected}
                            aria-label="{cap(fmt.long(cell.date, locale))}{open ? `, ${t.slotsLeft(open)}` : ''}"
                            onclick={() => (selectedDayKey = cell.key)}
                            class="relative flex aspect-square cursor-pointer flex-col items-center justify-center rounded-lg text-sm tabular-nums transition-colors focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none disabled:cursor-default disabled:text-neutral-400 {selected
                                ? 'bg-neutral-950 font-semibold text-white'
                                : open
                                  ? 'bg-neutral-100 font-semibold hover:bg-neutral-200'
                                  : ''}"
                        >
                            {cell.date.getUTCDate()}
                            {#if cell.key === todayKey}
                                <span class="absolute bottom-1.5 size-1 rounded-full {selected ? 'bg-white' : 'bg-neutral-950'}"></span>
                            {/if}
                        </button>
                    {:else}
                        <span></span>
                    {/if}
                {/each}
            </div>
        </div>

        <div>
            {#if byDay.size === 0}
                <p class="text-sm text-pretty text-neutral-600">{t.noSlotsYet}</p>
            {:else if selectedDayKey}
                {@const day = dayOf(selectedDayKey)}
                <p class="mb-4 font-semibold">
                    {cap(fmt.weekdayShort(day.date, locale))}
                    <span class="font-normal text-neutral-600">{fmt.dayMonth(day.date, locale)}</span>
                </p>
                {#if daySlots.length === 0}
                    <p class="text-sm text-neutral-600">{t.noSlots}</p>
                {:else}
                    <div class="grid max-h-[22rem] grid-cols-3 gap-2 overflow-y-auto md:grid-cols-1" role="group" aria-label={t.timeLegend}>
                        {#each daySlots as candidate (candidate.slot.start)}
                            <button
                                type="button"
                                onclick={() => choose(candidate)}
                                class="h-11 flex-1 cursor-pointer rounded-lg border border-neutral-300 text-sm font-semibold tabular-nums transition-all hover:border-neutral-950 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none"
                            >
                                {fmt.time(candidate.start, locale)}
                            </button>
                        {/each}
                    </div>
                    <p class="mt-3 text-xs text-neutral-600">{t.timeZoneNote}</p>
                {/if}
            {:else}
                <p class="text-sm text-neutral-600">{t.pickDay}</p>
            {/if}
        </div>
    </div>
{/if}
