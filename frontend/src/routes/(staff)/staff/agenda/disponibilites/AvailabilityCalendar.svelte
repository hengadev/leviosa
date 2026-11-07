<script lang="ts">
	// A month of the practitioner's agenda, Monday first. The page uses it to
	// show which days are open (and pick one), the creation form to pick the
	// days a new availability covers. Each day's look comes from `info`.
	import type { CalendarDate } from '@internationalized/date';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
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

	const toneClass: Record<DayTone, string> = {
		none: 'text-foreground hover:bg-muted',
		open: 'bg-emerald-50 text-emerald-900 hover:bg-emerald-100 dark:bg-emerald-950/40 dark:text-emerald-100 dark:hover:bg-emerald-950/60',
		closed: 'bg-muted/60 text-muted-foreground line-through hover:bg-muted',
		picked: 'bg-foreground text-background hover:bg-foreground/90',
		covered: 'bg-foreground/10 text-foreground hover:bg-foreground/15'
	};
</script>

<div>
	<div class="flex items-center justify-between mb-3">
		<button
			type="button"
			onclick={() => (month = month.subtract({ months: 1 }))}
			disabled={!canGoBack}
			class="p-2 rounded-md hover:bg-muted transition-colors disabled:opacity-30 disabled:hover:bg-transparent"
			aria-label="Mois précédent"
		>
			<ChevronLeft size={18} class="text-muted-foreground" />
		</button>
		<p class="font-semibold text-foreground capitalize {compact ? 'text-sm' : ''}">{title}</p>
		<button
			type="button"
			onclick={() => (month = month.add({ months: 1 }))}
			class="p-2 rounded-md hover:bg-muted transition-colors"
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

	<div class="grid grid-cols-7 gap-1">
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
						<span class="mt-0.5 h-1 w-1 rounded-full bg-emerald-600" aria-hidden="true"></span>
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
</div>
