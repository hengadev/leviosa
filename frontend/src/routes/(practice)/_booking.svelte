<script lang="ts">
    import { ChevronLeft, ChevronRight } from "@lucide/svelte";
    import { cap, dictionary, fmt, type Locale } from "$lib/i18n/practice";
    import { dayKey as keyOf, mockDays, openSlots } from "./_mock-slots";

    // The booking card's calendar area. STATIC for now: the days and times
    // below are the mock availability from _mock-slots.ts — the real free
    // slots arrive with the public slot endpoint (issue 010). Picking a slot
    // only marks it; no details or payment step exists yet (issue 014).
    let { locale }: { locale: Locale } = $props();

    const t = $derived(dictionary[locale]);

    const today = new Date();
    const days = mockDays(56);
    const byKey = new Map(days.map((day) => [day.key, day]));

    let monthOffset = $state(0);
    let selectedDayKey = $state(days.find((day) => openSlots(day).length > 0)?.key ?? "");
    let slotId = $state("");

    const day = $derived(byKey.get(selectedDayKey));
    const month = $derived(new Date(today.getFullYear(), today.getMonth() + monthOffset, 1));
    const lastDay = days[days.length - 1].date;
    const canNext = $derived(
        new Date(month.getFullYear(), month.getMonth() + 1, 1).getTime() <= lastDay.getTime(),
    );

    /** The month as Monday-first weeks; null pads the first week. */
    const cells = $derived.by(() => {
        const pad = (month.getDay() + 6) % 7;
        const count = new Date(month.getFullYear(), month.getMonth() + 1, 0).getDate();
        const list: (Date | null)[] = Array(pad).fill(null);
        for (let d = 1; d <= count; d++) list.push(new Date(month.getFullYear(), month.getMonth(), d));
        return list;
    });
</script>

<div class="grid gap-8 md:grid-cols-[minmax(0,1fr)_11rem]">
    <div>
        <div class="mb-4 flex items-center justify-between">
            <p class="font-semibold">{cap(fmt.month(month, locale))}</p>
            <div class="flex gap-1">
                <button
                    type="button"
                    aria-label={t.prevMonth}
                    disabled={monthOffset === 0}
                    onclick={() => monthOffset--}
                    class="flex size-9 cursor-pointer items-center justify-center rounded-lg border border-neutral-200 transition-colors hover:bg-neutral-100 focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
                >
                    <ChevronLeft class="size-4" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    aria-label={t.nextMonth}
                    disabled={!canNext}
                    onclick={() => monthOffset++}
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
                    {@const data = byKey.get(keyOf(cell))}
                    {@const open = data ? openSlots(data).length : 0}
                    {@const selected = selectedDayKey === keyOf(cell)}
                    <button
                        type="button"
                        disabled={open === 0}
                        aria-pressed={selected}
                        aria-label="{cap(fmt.long(cell, locale))}{open ? `, ${t.slotsLeft(open)}` : ''}"
                        onclick={() => {
                            selectedDayKey = keyOf(cell);
                            slotId = "";
                        }}
                        class="relative flex aspect-square cursor-pointer flex-col items-center justify-center rounded-lg text-sm tabular-nums transition-colors focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none disabled:cursor-default disabled:text-neutral-400 {selected
                            ? 'bg-neutral-950 font-semibold text-white'
                            : open
                              ? 'bg-neutral-100 font-semibold hover:bg-neutral-200'
                              : ''}"
                    >
                        {cell.getDate()}
                        {#if keyOf(cell) === keyOf(today)}
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
        {#if day}
            <p class="mb-4 font-semibold">
                {cap(fmt.weekdayShort(day.date, locale))}
                <span class="font-normal text-neutral-600">{fmt.dayMonth(day.date, locale)}</span>
            </p>
            {#if openSlots(day).length === 0}
                <p class="text-sm text-neutral-600">{t.noSlots}</p>
            {:else}
                <div class="grid max-h-[22rem] grid-cols-3 gap-2 overflow-y-auto md:grid-cols-1" aria-label={t.timeLegend}>
                    {#each openSlots(day) as candidate (candidate.id)}
                        {@const selected = slotId === candidate.id}
                        <button
                            type="button"
                            aria-pressed={selected}
                            onclick={() => (slotId = candidate.id)}
                            class="h-11 flex-1 cursor-pointer rounded-lg border text-sm font-semibold tabular-nums transition-all focus-visible:ring-3 focus-visible:ring-neutral-950/20 focus-visible:outline-none {selected
                                ? 'border-neutral-950 bg-neutral-950 text-white'
                                : 'border-neutral-300 hover:border-neutral-950'}"
                        >
                            {fmt.time(candidate.start, locale)}
                        </button>
                    {/each}
                </div>
            {/if}
        {:else}
            <p class="text-sm text-neutral-600">{t.pickDay}</p>
        {/if}
    </div>
</div>
