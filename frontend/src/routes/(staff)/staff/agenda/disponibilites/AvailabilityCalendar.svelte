<script lang="ts">
	// A month of the practitioner's agenda, Monday first. The page uses it to
	// show which days are open (and pick one), the creation form to pick the
	// days a new availability covers. Each day's look comes from `info`.
	import type { CalendarDate } from '@internationalized/date';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
	import { fly } from 'svelte/transition';
	import { prefersReducedMotion } from 'svelte/motion';
	import { monthGrid } from './calendar';

	export type DayTone = 'none' | 'open' | 'closed' | 'picked' | 'covered';

	interface Props {
		/** Any day of the month shown. */
		month: CalendarDate;
		today: CalendarDate;
		/** `marked`: a dot under the day (already open, in the form). */
		info: (day: CalendarDate) => { tone: DayTone; hint?: string; marked?: boolean };
		onPick: (day: CalendarDate) => void;
		/** The day outlined (the one whose details are shown). */
		selected?: string | null;
		/** Days before it cannot be picked. */
		min?: CalendarDate;
		/** Smaller cells, without hints (inside a form). */
		compact?: boolean;
	}

	let { month = $bindable(), today, info, onPick, selected = null, min, compact = false }: Props = $props();

	const weeks = $derived(monthGrid(month));
	const title = $derived(
		new Date(Date.UTC(month.year, month.month - 1, 1)).toLocaleDateString('fr-FR', {
			month: 'long',
			year: 'numeric',
			timeZone: 'UTC'
		})
	);
	// « 30 octobre 2026 »: a day of the next or previous month says its own month.
	function dayName(day: CalendarDate): string {
		return new Date(Date.UTC(day.year, day.month - 1, day.day)).toLocaleDateString('fr-FR', {
			day: 'numeric',
			month: 'long',
			year: 'numeric',
			timeZone: 'UTC'
		});
	}

	const canGoBack = $derived(!min || month.set({ day: 1 }).compare(min.set({ day: 1 })) > 0);

	// The way the last change went: the new month slides in from that side.
	let direction = $state<1 | -1>(1);
	function previous() {
		if (!canGoBack) return;
		direction = -1;
		month = month.subtract({ months: 1 });
	}
	function next() {
		direction = 1;
		month = month.add({ months: 1 });
	}

	// Swipe on the days (touch screens): left for the next month, right for the
	// previous one. Only a clearly horizontal stroke counts, so scrolling the
	// page up or down over the calendar never changes the month; a swipe moves
	// too far for the browser to treat it as a tap on a day.
	let start: { x: number; y: number } | null = null;
	function onTouchStart(event: TouchEvent) {
		start = event.touches.length === 1 ? { x: event.touches[0].clientX, y: event.touches[0].clientY } : null;
	}
	function onTouchEnd(event: TouchEvent) {
		if (!start) return;
		const touch = event.changedTouches[0];
		const dx = touch.clientX - start.x;
		const dy = touch.clientY - start.y;
		start = null;
		if (Math.abs(dx) < 50 || Math.abs(dx) < Math.abs(dy) * 1.5) return;
		if (dx < 0) next();
		else previous();
	}

	const toneClass: Record<DayTone, string> = {
		none: 'text-foreground hover:bg-muted',
		open: 'bg-muted text-foreground font-medium hover:bg-surface-hover',
		closed: 'border border-dashed border-border-input text-muted-foreground line-through hover:bg-muted',
		picked: 'bg-foreground text-background hover:bg-foreground/90',
		covered: 'bg-foreground/10 text-foreground hover:bg-foreground/15'
	};
</script>

<div>
	<div class="flex items-center justify-between mb-3">
		<button
			type="button"
			onclick={previous}
			disabled={!canGoBack}
			class="flex size-11 items-center justify-center rounded-md hover:bg-muted transition-colors disabled:opacity-30 disabled:hover:bg-transparent"
			aria-label="Mois précédent"
		>
			<ChevronLeft size={18} class="text-muted-foreground" />
		</button>
		<p class="font-semibold text-foreground capitalize {compact ? 'text-sm' : ''}">{title}</p>
		<button
			type="button"
			onclick={next}
			class="flex size-11 items-center justify-center rounded-md hover:bg-muted transition-colors"
			aria-label="Mois suivant"
		>
			<ChevronRight size={18} class="text-muted-foreground" />
		</button>
	</div>

	<div class="grid grid-cols-7 gap-1 text-center text-[11px] font-medium uppercase tracking-wide text-muted-foreground mb-1">
		{#each ['Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam', 'Dim'] as name}
			<span>{name}</span>
		{/each}
	</div>

	<div
		class="overflow-hidden touch-pan-y"
		role="group"
		aria-label="Jours de {title}"
		ontouchstart={onTouchStart}
		ontouchend={onTouchEnd}
		ontouchcancel={() => (start = null)}
	>
	{#key month.toString().slice(0, 7)}
	<div
		class="grid grid-cols-7 gap-1"
		in:fly={{ x: prefersReducedMotion.current ? 0 : direction * 24, duration: prefersReducedMotion.current ? 0 : 180 }}
	>
		{#each weeks as week}
			{#each week as day (day.toString())}
				{@const inMonth = day.month === month.month}
				{@const past = min ? day.compare(min) < 0 : false}
				{@const look = info(day)}
				<button
					type="button"
					onclick={() => onPick(day)}
					disabled={past}
					aria-pressed={selected === day.toString()}
					aria-label="{dayName(day)}{look.hint ? ` : ${look.hint}` : ''}{look.marked ? ' : déjà ouvert' : ''}"
					class="relative flex flex-col items-center justify-start rounded-lg transition-colors disabled:cursor-not-allowed disabled:opacity-35
						{compact ? 'h-10 pt-2' : 'h-14 sm:h-20 pt-1.5 sm:pt-2'}
						{toneClass[look.tone]}
						{inMonth ? '' : 'opacity-40'}
						{selected === day.toString() ? 'ring-2 ring-foreground ring-offset-1 ring-offset-background' : ''}"
				>
					<span class="text-sm tabular-nums {day.compare(today) === 0 ? 'font-bold underline underline-offset-4' : ''}">
						{day.day}
					</span>
					{#if compact && look.marked}
						<span class="mt-0.5 h-1 w-1 rounded-full bg-foreground" aria-hidden="true"></span>
					{/if}
					{#if !compact && look.hint}
						<span class="hidden sm:block mt-1 px-1 text-[11px] leading-tight no-underline truncate max-w-full">
							{look.hint}
						</span>
						<span
							class="sm:hidden mt-1 h-1.5 w-1.5 rounded-full {look.tone === 'closed' ? 'bg-muted-foreground/50' : 'bg-current'}"
							aria-hidden="true"
						></span>
					{/if}
				</button>
			{/each}
		{/each}
	</div>
	{/key}
	</div>
</div>
