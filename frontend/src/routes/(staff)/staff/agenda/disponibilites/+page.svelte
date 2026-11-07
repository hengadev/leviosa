<script lang="ts">
	import type { PageProps } from './$types';
	import type { Availability } from './+page.server';
	import {
		Plus,
		X,
		Pencil,
		AlertTriangle,
		Loader2,
		Repeat,
		CalendarOff,
		CalendarPlus,
		Clock
	} from '@lucide/svelte';
	import { invalidateAll } from '$app/navigation';
	import { parseDate, type CalendarDate } from '@internationalized/date';
	import Modal from '$lib/ui/Modal.svelte';
	import { PRACTICE_TIME_ZONE } from '$lib/i18n/practice';
	import { practiceDate, practiceInstant, type AvailabilityException } from './occurrences';
	import { dayOpenings, firstWeeklyClash, firstWeekday, overlapping, weekday, type Opening } from './calendar';
	import AvailabilityCalendar, { type DayTone } from './AvailabilityCalendar.svelte';

	let { data }: PageProps = $props();

	// --- Calendar state ---
	const today: CalendarDate = practiceDate(new Date().toISOString());
	let month = $state<CalendarDate>(today);
	let selectedDay = $state<CalendarDate>(today);

	// --- Modal states ---
	let createModalOpen = $state(false);
	let editModalOpen = $state(false);
	let cancelConfirmOpen = $state(false);
	let selectedSlot: Availability | null = $state(null);

	// --- Create form state ---
	// « Chaque semaine »: the weekdays, from a first day to an optional last
	// one (a recurring availability). « Certains jours »: the dates picked,
	// one availability each.
	let formMode = $state<'weekly' | 'dates'>('weekly');
	let formStartTime = $state('09:00');
	let formEndTime = $state('13:00');
	let formDaysOfWeek = $state<number[]>([]);
	let formFrom = $state('');
	let formUntil = $state('');
	// Which end of the period the next tap on the calendar sets.
	let formPicking = $state<'from' | 'until'>('until');
	let formDates = $state<string[]>([]);
	let pickerMonth = $state<CalendarDate>(today);

	// --- Edit form state ---
	let editDate = $state('');
	let editStartTime = $state('');
	let editEndTime = $state('');
	let editNotes = $state('');

	// --- Availability exception (closed dates) state ---
	let exceptionModalOpen = $state(false);
	let exceptionSeries: Availability | null = $state(null);
	let exceptionStart = $state('');
	let exceptionEnd = $state('');
	let exceptionError = $state<string | null>(null);
	let exceptionNotice = $state<string | null>(null);
	let deletingExceptionId = $state<string | null>(null);

	// --- Day panel state ---
	// The opening whose hours are being changed for the selected day only.
	let dayEdit = $state<{ key: string; start: string; end: string } | null>(null);
	let dayBusy = $state(false);
	let dayError = $state<string | null>(null);

	// --- Loading / error states ---
	let submitting = $state(false);
	let conflictWarning = $state<string | null>(null);
	let checkingConflict = $state(false);
	let formError = $state<string | null>(null);

	// --- Derived data ---
	const series = $derived(data.series as Availability[]);
	const oneOffs = $derived(data.oneOffs as Availability[]);
	const selectedOpenings = $derived(dayOpenings(selectedDay, series, oneOffs));
	const selectedIsPast = $derived(selectedDay.compare(today) < 0);

	// --- Helpers ---
	// Dates (YYYY-MM-DD) are calendar days; instants are shown in the practice time zone.
	function formatDate(dateStr: string): string {
		return new Date(`${dateStr.slice(0, 10)}T12:00:00Z`).toLocaleDateString('fr-FR', {
			day: 'numeric',
			month: 'short',
			timeZone: 'UTC'
		});
	}

	function formatLongDate(day: CalendarDate): string {
		return new Date(Date.UTC(day.year, day.month - 1, day.day, 12)).toLocaleDateString('fr-FR', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			timeZone: 'UTC'
		});
	}

	function formatTime(iso: string): string {
		return new Date(iso).toLocaleTimeString('fr-FR', {
			hour: '2-digit',
			minute: '2-digit',
			timeZone: PRACTICE_TIME_ZONE
		});
	}

	// HH:MM of an instant in the practice time zone.
	function practiceClock(iso: string): string {
		return new Date(iso).toLocaleTimeString('fr-FR', {
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23',
			timeZone: PRACTICE_TIME_ZONE
		});
	}

	// 9:00 → « 9h », 9:30 → « 9h30 »: short enough for a calendar cell.
	function shortClock(iso: string): string {
		const [h, m] = practiceClock(iso).split(':');
		return `${Number(h)}h${m === '00' ? '' : m}`;
	}

	function todayDate(): string {
		return today.toString();
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'booked':
				return 'Réservé';
			case 'blocked':
				return 'Bloqué';
			default:
				return status;
		}
	}

	const weekdayNames = ['dimanche', 'lundi', 'mardi', 'mercredi', 'jeudi', 'vendredi', 'samedi'];
	// Monday first, as on the calendar.
	const weekdayChips: [string, number][] = [
		['Lun', 1],
		['Mar', 2],
		['Mer', 3],
		['Jeu', 4],
		['Ven', 5],
		['Sam', 6],
		['Dim', 0]
	];

	function joinFrench(items: string[]): string {
		return items.length <= 1 ? (items[0] ?? '') : `${items.slice(0, -1).join(', ')} et ${items[items.length - 1]}`;
	}

	function sortedWeekdays(days: number[]): number[] {
		return [...days].sort((a, b) => ((a + 6) % 7) - ((b + 6) % 7));
	}

	function recurrenceLabel(slot: Availability): string {
		if (!slot.recurrencePattern) return '';
		const p = slot.recurrencePattern;
		const n = p.interval > 1 ? p.interval : 1;
		switch (p.type) {
			case 'daily':
				return n > 1 ? `Tous les ${n} jours` : 'Tous les jours';
			case 'weekly': {
				const days = joinFrench(sortedWeekdays(p.days_of_week ?? []).map((d) => weekdayNames[d]));
				return n > 1 ? `Toutes les ${n} semaines, le ${days}` : `Chaque ${days}`;
			}
			case 'monthly': {
				const day = practiceDate(slot.startTime).day;
				return n > 1 ? `Tous les ${n} mois, le ${day}` : `Chaque mois, le ${day}`;
			}
			default:
				return p.type;
		}
	}

	// An RFC3339 instant from a date + HH:MM wall-clock time in the practice time zone.
	function toDatetime(date: string, time: string): string {
		return practiceInstant(date, time);
	}

	// --- Time options (10-min aligned) ---
	const timeOptions: string[] = (() => {
		const opts: string[] = [];
		for (let h = 6; h <= 22; h++) {
			for (let m = 0; m < 60; m += 10) {
				opts.push(`${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`);
			}
		}
		return opts;
	})();

	// --- The page calendar ---
	// A day is open when one of its openings is; closed when every opening
	// is closed by an exception.
	function dayInfo(day: CalendarDate): { tone: DayTone; hint?: string } {
		const openings = dayOpenings(day, series, oneOffs);
		if (openings.length === 0) return { tone: 'none' };
		const open = openings.filter((o) => !o.closedBy);
		if (open.length === 0) return { tone: 'closed', hint: 'Fermé' };
		const hint =
			open.length === 1
				? `${shortClock(open[0].start)}–${shortClock(open[0].end)}`
				: `${open.length} plages`;
		return { tone: 'open', hint };
	}

	let dayPanel: HTMLElement | null = $state(null);

	function selectDay(day: CalendarDate) {
		selectedDay = day;
		dayEdit = null;
		dayError = null;
		// Below the calendar on a phone: bring the day's details into view.
		if (dayPanel && !window.matchMedia('(min-width: 1024px)').matches) {
			dayPanel.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}

	// --- The creation form's calendar ---
	const formWeekdaysLabel = $derived(
		joinFrench(sortedWeekdays(formDaysOfWeek).map((d) => weekdayNames[d]))
	);

	// The days the new availability covers, and a dot on those already open.
	function pickerInfo(day: CalendarDate): { tone: DayTone; marked: boolean } {
		const date = day.toString();
		const marked = dayInfo(day).tone === 'open';
		if (formMode === 'dates') {
			return { tone: formDates.includes(date) ? 'picked' : 'none', marked };
		}
		if (date === formFrom || date === formUntil) return { tone: 'picked', marked };
		const inPeriod = formFrom && date >= formFrom && (!formUntil || date <= formUntil);
		if (inPeriod && formDaysOfWeek.includes(weekday(day))) return { tone: 'covered', marked };
		return { tone: 'none', marked };
	}

	function pickInForm(day: CalendarDate) {
		const date = day.toString();
		if (formMode === 'dates') {
			formDates = formDates.includes(date)
				? formDates.filter((d) => d !== date)
				: [...formDates, date].sort();
			return;
		}
		// Choosing a weekday here would be confusing: the calendar sets the
		// period, the chips the weekdays. A first day after the last one
		// drops the last one.
		if (formPicking === 'from' || !formFrom || date < formFrom) {
			formFrom = date;
			if (formUntil && formUntil < date) formUntil = '';
			formPicking = 'until';
		} else {
			formUntil = date;
		}
	}

	function toggleWeekday(idx: number) {
		formDaysOfWeek = formDaysOfWeek.includes(idx)
			? formDaysOfWeek.filter((d) => d !== idx)
			: [...formDaysOfWeek, idx];
	}

	// --- Conflict check ---
	async function checkConflict(
		partnerId: string,
		startTime: string,
		endTime: string,
		excludeId?: string
	): Promise<boolean> {
		const params = new URLSearchParams({ start_time: startTime, end_time: endTime });
		if (excludeId) params.set('exclude_id', excludeId);
		try {
			const res = await fetch(
				`/api/partners/${partnerId}/availabilities/conflict?${params.toString()}`
			);
			if (!res.ok) return false;
			const data = await res.json();
			return data.has_conflict === true;
		} catch {
			return false;
		}
	}

	// A one-off availability; null when created, otherwise why not. `replacing`:
	// the series just closed on that day, still open in the page's data.
	async function createOneOff(
		date: string,
		startTime: string,
		endTime: string,
		replacing?: string
	): Promise<string | null> {
		const startIso = toDatetime(date, startTime);
		const endIso = toDatetime(date, endTime);
		if (startIso <= new Date().toISOString()) return 'ce créneau est déjà commencé';
		if (
			overlapping(parseDate(date), startIso, endIso, series, oneOffs, replacing) ||
			(await checkConflict(data.partnerId, startIso, endIso))
		) {
			return 'en conflit avec une disponibilité existante';
		}
		const res = await fetch('/api/availabilities', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ room_id: data.roomId, start_time: startIso, end_time: endIso, max_capacity: 1 })
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			return err.message || 'erreur lors de la création';
		}
		return null;
	}

	// --- Reset form ---
	function resetCreateForm() {
		formMode = 'weekly';
		formStartTime = '09:00';
		formEndTime = '13:00';
		formDaysOfWeek = [];
		formFrom = todayDate();
		formUntil = '';
		formPicking = 'until';
		formDates = [];
		pickerMonth = today;
		conflictWarning = null;
		formError = null;
	}

	function resetEditForm() {
		editDate = '';
		editStartTime = '';
		editEndTime = '';
		editNotes = '';
		conflictWarning = null;
		formError = null;
	}

	// --- Actions ---
	async function handleCreate() {
		formError = null;
		conflictWarning = null;

		if (!data.roomId) {
			formError = 'Aucune salle ne vous est attribuée : demandez à l’administrateur de vous en attribuer une.';
			return;
		}
		if (formStartTime >= formEndTime) {
			formError = "L'heure de début doit être antérieure à l'heure de fin.";
			return;
		}

		submitting = true;
		try {
			if (formMode === 'weekly') {
				await createWeekly();
			} else {
				await createDates();
			}
		} catch (e: any) {
			formError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	async function createWeekly() {
		if (formDaysOfWeek.length === 0) {
			formError = 'Choisissez au moins un jour de la semaine.';
			return;
		}
		if (!formFrom) {
			formError = 'Choisissez le premier jour sur le calendrier.';
			return;
		}
		const until = formUntil ? parseDate(formUntil) : undefined;
		// The series starts on its first opening: the first chosen weekday
		// from the first day, skipping today when its hours have begun.
		let first = firstWeekday(parseDate(formFrom), formDaysOfWeek, until);
		if (first && toDatetime(first.toString(), formStartTime) <= new Date().toISOString()) {
			first = firstWeekday(first.add({ days: 1 }), formDaysOfWeek, until);
		}
		if (!first) {
			formError = 'Aucun des jours choisis ne tombe dans cette période.';
			return;
		}

		const startIso = toDatetime(first.toString(), formStartTime);
		const endIso = toDatetime(first.toString(), formEndTime);

		const clash = firstWeeklyClash(
			first,
			until,
			formDaysOfWeek,
			(d) => ({ start: toDatetime(d.toString(), formStartTime), end: toDatetime(d.toString(), formEndTime) }),
			series,
			oneOffs
		);
		if (clash) {
			conflictWarning = `Le ${formatDate(clash.toString())}, ces horaires chevauchent une disponibilité existante.`;
			return;
		}
		checkingConflict = true;
		const hasConflict = await checkConflict(data.partnerId, startIso, endIso);
		checkingConflict = false;
		if (hasConflict) {
			conflictWarning = `Le ${formatDate(first.toString())} est en conflit avec une disponibilité existante.`;
			return;
		}

		const res = await fetch('/api/availabilities/recurring', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				room_id: data.roomId,
				start_time: startIso,
				end_time: endIso,
				max_capacity: 1,
				pattern: {
					type: 'weekly',
					interval: 1,
					// The last day is inclusive.
					until: formUntil ? toDatetime(formUntil, '23:59') : undefined,
					days_of_week: formDaysOfWeek
				}
			})
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			formError = err.message || 'Erreur lors de la création de la disponibilité récurrente.';
			return;
		}

		createModalOpen = false;
		selectDay(first);
		month = first;
		await invalidateAll();
	}

	// One availability per date. The dates that fail stay picked, with why.
	async function createDates() {
		if (formDates.length === 0) {
			formError = 'Choisissez au moins un jour sur le calendrier.';
			return;
		}
		const failures: string[] = [];
		const failed: string[] = [];
		for (const date of formDates) {
			const why = await createOneOff(date, formStartTime, formEndTime);
			if (why) {
				failures.push(`${formatDate(date)} : ${why}`);
				failed.push(date);
			}
		}
		await invalidateAll();
		if (failures.length > 0) {
			formDates = failed;
			formError = `Non créé${failed.length > 1 ? 's' : ''} — ${failures.join(' · ')}`;
			return;
		}
		createModalOpen = false;
	}

	async function handleEdit() {
		if (!selectedSlot) return;
		submitting = true;
		formError = null;
		conflictWarning = null;

		if (!editDate || !editStartTime || !editEndTime) {
			formError = 'Date et horaires obligatoires.';
			submitting = false;
			return;
		}

		const startIso = toDatetime(editDate, editStartTime);
		const endIso = toDatetime(editDate, editEndTime);

		if (startIso >= endIso) {
			formError = "L'heure de début doit être antérieure à l'heure de fin.";
			submitting = false;
			return;
		}
		if (startIso <= new Date().toISOString()) {
			formError = selectedSlot.isRecurring
				? 'Choisissez un premier jour à venir : la modification s’applique à partir de cette date.'
				: 'Le créneau doit commencer dans le futur.';
			submitting = false;
			return;
		}

		// Check conflict (exclude current slot)
		checkingConflict = true;
		const hasConflict = await checkConflict(data.partnerId, startIso, endIso, selectedSlot.id);
		checkingConflict = false;

		if (hasConflict) {
			conflictWarning = 'Ce créneau est en conflit avec un créneau existant.';
			submitting = false;
			return;
		}

		try {
			const body: any = {
				start_time: startIso,
				end_time: endIso,
				notes: editNotes
			};

			const res = await fetch(`/api/availabilities/${selectedSlot.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(body)
			});

			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				formError = err.message || 'Erreur lors de la mise à jour.';
				submitting = false;
				return;
			}

			editModalOpen = false;
			selectedSlot = null;
			resetEditForm();
			await invalidateAll();
		} catch (e: any) {
			formError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	async function handleCancel() {
		if (!selectedSlot) return;
		submitting = true;
		formError = null;

		try {
			const res = await fetch(`/api/availabilities/${selectedSlot.id}/cancel`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' }
			});

			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				formError = err.message || 'Erreur lors de la suppression.';
				submitting = false;
				return;
			}

			cancelConfirmOpen = false;
			selectedSlot = null;
			await invalidateAll();
		} catch (e: any) {
			formError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	// Open modals
	function openCreateModal(date?: CalendarDate) {
		resetCreateForm();
		if (date) {
			// From a day of the calendar: that day, with the hours to set.
			formMode = 'dates';
			formDates = [date.toString()];
			pickerMonth = date;
		}
		createModalOpen = true;
	}

	// A recurring availability is edited from a first day on (the series'
	// first occurrence moves there); by default the later of its current
	// first day and tomorrow. Existing bookings never move.
	function openEditModal(slot: Availability) {
		selectedSlot = slot;
		resetEditForm();
		if (slot.isRecurring) {
			const tomorrow = today.add({ days: 1 }).toString();
			const first = practiceDate(slot.startTime).toString();
			editDate = first > tomorrow ? first : tomorrow;
		} else {
			editDate = practiceDate(slot.startTime).toString();
		}
		editStartTime = practiceClock(slot.startTime);
		editEndTime = practiceClock(slot.endTime);
		editNotes = slot.notes ?? '';
		editModalOpen = true;
	}

	function openCancelConfirm(slot: Availability) {
		selectedSlot = slot;
		formError = null;
		cancelConfirmOpen = true;
	}

	// --- Availability exceptions (closed dates of a recurring series) ---
	function exceptionLabel(e: AvailabilityException): string {
		return e.startDate === e.endDate
			? `Fermé le ${formatDate(e.startDate)}`
			: `Fermé du ${formatDate(e.startDate)} au ${formatDate(e.endDate)}`;
	}

	// Closing dates never cancels the bookings already on them.
	function affectedBookingsNotice(n: number): string | null {
		if (n <= 0) return null;
		return n === 1
			? '1 réservation sur ces dates — elle reste confirmée ; annulez-la si besoin.'
			: `${n} réservations sur ces dates — elles restent confirmées ; annulez-les une par une si besoin.`;
	}

	function openExceptionModal(serie: Availability) {
		exceptionSeries = serie;
		exceptionStart = todayDate();
		exceptionEnd = '';
		exceptionError = null;
		exceptionModalOpen = true;
	}

	// Closes dates of a series: the exception created, or why not.
	async function closeDates(
		serie: Availability,
		start: string,
		end?: string
	): Promise<{ exception?: AvailabilityException; error?: string }> {
		const res = await fetch(`/api/availabilities/${serie.id}/exceptions`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ start_date: start, end_date: end || undefined })
		});
		if (!res.ok) {
			const err = await res.json().catch(() => ({}));
			return { error: err.message || 'Erreur lors de la fermeture des dates.' };
		}
		const created = await res.json();
		exceptionNotice = affectedBookingsNotice(created.affected_bookings ?? 0);
		const e = created.exception;
		return { exception: { id: e.id, startDate: e.start_date, endDate: e.end_date } };
	}

	async function handleCreateException() {
		if (!exceptionSeries) return;
		exceptionError = null;
		if (!exceptionStart) {
			exceptionError = 'Choisissez le premier jour fermé.';
			return;
		}
		if (exceptionEnd && exceptionEnd < exceptionStart) {
			exceptionError = 'Le dernier jour doit suivre le premier.';
			return;
		}
		submitting = true;
		try {
			exceptionError = (await closeDates(exceptionSeries, exceptionStart, exceptionEnd)).error ?? null;
			if (exceptionError) return;
			exceptionModalOpen = false;
			exceptionSeries = null;
			await invalidateAll();
		} catch (e: any) {
			exceptionError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	// Reopens the dates; the error message, or null.
	async function reopenDates(serie: Availability, e: AvailabilityException): Promise<string | null> {
		const res = await fetch(`/api/availabilities/${serie.id}/exceptions/${e.id}`, {
			method: 'DELETE'
		});
		if (!res.ok && res.status !== 404) {
			const err = await res.json().catch(() => ({}));
			return err.message || 'Erreur lors de la réouverture des dates.';
		}
		return null;
	}

	async function handleDeleteException(serie: Availability, e: AvailabilityException) {
		deletingExceptionId = e.id;
		exceptionNotice = null;
		try {
			exceptionNotice = await reopenDates(serie, e);
			await invalidateAll();
		} catch (err: any) {
			exceptionNotice = err.message || 'Erreur réseau.';
		} finally {
			deletingExceptionId = null;
		}
	}

	// --- Day panel actions ---
	function openingKey(o: Opening): string {
		return `${o.availability.id}-${o.start}`;
	}

	async function dayAction(run: () => Promise<string | null>) {
		dayBusy = true;
		dayError = null;
		try {
			dayError = await run();
			if (!dayError) dayEdit = null;
			await invalidateAll();
		} catch (e: any) {
			dayError = e.message || 'Erreur réseau.';
		} finally {
			dayBusy = false;
		}
	}

	function closeDay(o: Opening) {
		dayAction(async () => (await closeDates(o.availability, selectedDay.toString())).error ?? null);
	}

	function reopenDay(o: Opening) {
		if (!o.closedBy) return;
		const closedBy = o.closedBy;
		dayAction(() => reopenDates(o.availability, closedBy));
	}

	// Other hours on this day only: the series closes that day and a one-off
	// availability takes its place. If the one-off cannot be created, the day
	// is reopened as it was.
	function changeDayHours(o: Opening) {
		if (!dayEdit) return;
		const { start, end } = dayEdit;
		if (start >= end) {
			dayError = "L'heure de début doit être antérieure à l'heure de fin.";
			return;
		}
		const date = selectedDay.toString();
		dayAction(async () => {
			const closed = await closeDates(o.availability, date);
			if (!closed.exception) return closed.error ?? null;
			const why = await createOneOff(date, start, end, o.availability.id);
			if (!why) return null;
			await reopenDates(o.availability, closed.exception);
			exceptionNotice = null;
			return `Horaires non modifiés : ${why}.`;
		});
	}
</script>

<svelte:head>
	<title>Disponibilités | Staff</title>
</svelte:head>

<div class="p-6 lg:p-10">
	<div class="mb-8 flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
		<div>
			<p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted-foreground mb-3">Agenda</p>
			<h1 class="font-display text-4xl lg:text-5xl font-semibold tracking-tight text-foreground leading-[1.1]">Disponibilités</h1>
			<p class="text-muted-foreground mt-3 text-sm">Les jours et horaires où vos clients peuvent réserver</p>
			<div class="mt-4 h-px w-16 bg-foreground/20"></div>
		</div>
		<button
			onclick={() => openCreateModal()}
			class="inline-flex items-center gap-2 px-4 py-2.5 bg-foreground text-background rounded-lg text-sm font-medium hover:opacity-90 transition-opacity"
		>
			<Plus size={16} />
			Nouvelle disponibilité
		</button>
	</div>

	{#if !data.roomId}
		<div class="mb-6 p-4 rounded-xl border border-amber-200 bg-amber-50 text-amber-800 text-sm flex items-start gap-2">
			<AlertTriangle size={16} class="mt-0.5 shrink-0" />
			<span>Aucune salle ne vous est attribuée : vous ne pouvez pas encore ajouter de disponibilités.</span>
		</div>
	{/if}

	{#if exceptionNotice}
		<div
			class="mb-6 px-4 py-3 rounded-xl border border-amber-200 bg-amber-50 text-amber-800 text-sm flex items-start gap-2"
			role="status"
		>
			<AlertTriangle size={16} class="mt-0.5 shrink-0" />
			<span class="flex-1">{exceptionNotice}</span>
			<button
				onclick={() => (exceptionNotice = null)}
				class="p-0.5 rounded hover:bg-amber-100"
				aria-label="Fermer le message"
			>
				<X size={14} />
			</button>
		</div>
	{/if}

	<!-- Calendar and the selected day -->
	<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem] mb-8">
		<div class="bg-background rounded-xl border border-border p-4 sm:p-5">
			<AvailabilityCalendar
				bind:month
				{today}
				info={dayInfo}
				onPick={selectDay}
				selected={selectedDay.toString()}
			/>
			<div class="mt-4 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
				<span class="inline-flex items-center gap-1.5">
					<span class="h-3 w-3 rounded bg-emerald-100 dark:bg-emerald-950/60"></span>Ouvert
				</span>
				<span class="inline-flex items-center gap-1.5">
					<span class="h-3 w-3 rounded bg-muted"></span>Fermé exceptionnellement
				</span>
			</div>
		</div>

		<div bind:this={dayPanel} class="bg-background rounded-xl border border-border overflow-hidden self-start scroll-mt-4">
			<div class="bg-muted/50 px-5 py-3 border-b border-border">
				<h2 class="font-semibold text-foreground first-letter:uppercase">{formatLongDate(selectedDay)}</h2>
			</div>
			<div class="p-5 space-y-3">
				{#if dayError}
					<div class="p-3 rounded-md bg-red-50 text-red-700 text-sm">{dayError}</div>
				{/if}

				{#each selectedOpenings as o (openingKey(o))}
					{@const key = openingKey(o)}
					<div class="rounded-lg border border-border p-3 {o.closedBy ? 'bg-muted/40' : ''}">
						<div class="flex items-start justify-between gap-2">
							<div>
								<p class="text-sm font-medium tabular-nums {o.closedBy ? 'line-through text-muted-foreground' : 'text-foreground'}">
									{formatTime(o.start)} – {formatTime(o.end)}
								</p>
								<p class="text-xs text-muted-foreground mt-0.5 inline-flex items-center gap-1">
									{#if o.availability.isRecurring}
										<Repeat size={11} />{recurrenceLabel(o.availability)}
									{:else}
										<Clock size={11} />Ce jour uniquement
									{/if}
								</p>
								{#if o.closedBy}
									<p class="text-xs text-muted-foreground mt-1 inline-flex items-center gap-1">
										<CalendarOff size={11} />{exceptionLabel(o.closedBy)}
									</p>
								{/if}
							</div>
							{#if o.availability.status !== 'available'}
								<span class="px-2 py-0.5 rounded-full text-xs font-medium bg-muted text-foreground">
									{statusLabel(o.availability.status)}
								</span>
							{/if}
						</div>

						{#if !selectedIsPast}
							{#if dayEdit?.key === key}
								<div class="mt-3 space-y-3">
									<div class="grid grid-cols-2 gap-2">
										<select
											bind:value={dayEdit.start}
											aria-label="Début"
											class="w-full rounded-input border border-input bg-background px-2 py-1.5 text-sm"
										>
											{#each timeOptions as opt}
												<option value={opt}>{opt}</option>
											{/each}
										</select>
										<select
											bind:value={dayEdit.end}
											aria-label="Fin"
											class="w-full rounded-input border border-input bg-background px-2 py-1.5 text-sm"
										>
											{#each timeOptions as opt}
												<option value={opt}>{opt}</option>
											{/each}
										</select>
									</div>
									<p class="text-xs text-muted-foreground">Ces horaires ne changent que pour ce jour.</p>
									<div class="flex justify-end gap-2">
										<button
											onclick={() => (dayEdit = null)}
											class="px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground"
										>
											Annuler
										</button>
										<button
											onclick={() => changeDayHours(o)}
											disabled={dayBusy}
											class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-foreground text-background rounded-md text-sm font-medium hover:opacity-90 disabled:opacity-50"
										>
											{#if dayBusy}<Loader2 size={13} class="animate-spin" />{/if}
											Enregistrer
										</button>
									</div>
								</div>
							{:else}
								<div class="mt-3 flex flex-wrap gap-2">
									{#if o.availability.isRecurring}
										{#if o.closedBy}
											<button
												onclick={() => reopenDay(o)}
												disabled={dayBusy}
												class="px-3 py-1.5 rounded-md border border-border text-xs font-medium hover:bg-muted disabled:opacity-50"
											>
												{o.closedBy.startDate === o.closedBy.endDate
													? 'Rouvrir ce jour'
													: `Rouvrir du ${formatDate(o.closedBy.startDate)} au ${formatDate(o.closedBy.endDate)}`}
											</button>
										{:else}
											<button
												onclick={() =>
													(dayEdit = { key, start: practiceClock(o.start), end: practiceClock(o.end) })}
												disabled={dayBusy}
												class="px-3 py-1.5 rounded-md border border-border text-xs font-medium hover:bg-muted disabled:opacity-50"
											>
												Autres horaires ce jour
											</button>
											<button
												onclick={() => closeDay(o)}
												disabled={dayBusy}
												class="px-3 py-1.5 rounded-md border border-border text-xs font-medium hover:bg-muted disabled:opacity-50"
											>
												Fermer ce jour
											</button>
										{/if}
									{:else}
										<button
											onclick={() => openEditModal(o.availability)}
											class="px-3 py-1.5 rounded-md border border-border text-xs font-medium hover:bg-muted"
										>
											Modifier
										</button>
										<button
											onclick={() => openCancelConfirm(o.availability)}
											class="px-3 py-1.5 rounded-md border border-border text-xs font-medium text-red-700 hover:bg-red-50"
										>
											Supprimer
										</button>
									{/if}
								</div>
							{/if}
						{/if}
					</div>
				{:else}
					<p class="text-sm text-muted-foreground">Aucune disponibilité ce jour.</p>
				{/each}

				{#if !selectedIsPast && data.roomId}
					<button
						onclick={() => openCreateModal(selectedDay)}
						class="w-full inline-flex items-center justify-center gap-2 px-3 py-2 rounded-lg border border-dashed border-border text-sm font-medium text-foreground hover:bg-muted transition-colors"
					>
						<CalendarPlus size={15} />
						Ajouter des horaires ce jour
					</button>
				{/if}
			</div>
		</div>
	</div>

	<!-- Recurring availabilities -->
	<div class="bg-background rounded-xl border border-border overflow-hidden">
		<div class="bg-muted/50 px-5 py-3 border-b border-border flex items-center gap-2">
			<Repeat size={15} class="text-muted-foreground" />
			<h2 class="font-semibold text-foreground">Horaires réguliers</h2>
		</div>
		<div class="divide-y divide-border">
			{#each series as serie (serie.id)}
				<div class="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center gap-3">
					<div class="flex-1">
						<p class="text-sm font-medium text-foreground">
							{recurrenceLabel(serie)} · {formatTime(serie.startTime)} – {formatTime(serie.endTime)}
						</p>
						<p class="text-xs text-muted-foreground mt-0.5">
							Depuis le {formatDate(serie.date)}
							{#if serie.recurrencePattern?.until}
								· jusqu’au {formatDate(practiceDate(serie.recurrencePattern.until).toString())}
							{:else}
								· sans date de fin
							{/if}
						</p>
						{#if serie.exceptions.length > 0}
							<ul class="mt-2 flex flex-wrap gap-1.5" aria-label="Dates fermées">
								{#each serie.exceptions as exception (exception.id)}
									<li
										class="inline-flex items-center gap-1 pl-2.5 pr-1 py-0.5 rounded-full bg-muted text-xs text-foreground"
									>
										<CalendarOff size={12} class="text-muted-foreground" />
										{exceptionLabel(exception)}
										<button
											onclick={() => handleDeleteException(serie, exception)}
											disabled={deletingExceptionId === exception.id}
											class="p-0.5 rounded-full text-muted-foreground hover:text-red-600 hover:bg-red-50 transition-colors disabled:opacity-50"
											title="Rouvrir ces dates"
											aria-label="Rouvrir : {exceptionLabel(exception)}"
										>
											{#if deletingExceptionId === exception.id}
												<Loader2 size={12} class="animate-spin" />
											{:else}
												<X size={12} />
											{/if}
										</button>
									</li>
								{/each}
							</ul>
						{/if}
					</div>
					<div class="flex items-center gap-2">
						<button
							onclick={() => openExceptionModal(serie)}
							class="p-1.5 rounded-md text-muted-foreground hover:text-amber-700 hover:bg-amber-50 transition-colors"
							title="Fermer une période (congés…)"
							aria-label="Fermer une période"
						>
							<CalendarOff size={14} />
						</button>
						<button
							onclick={() => openEditModal(serie)}
							class="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
							title="Modifier les horaires"
							aria-label="Modifier les horaires"
						>
							<Pencil size={14} />
						</button>
						<button
							onclick={() => openCancelConfirm(serie)}
							class="p-1.5 rounded-md text-muted-foreground hover:text-red-600 hover:bg-red-50 transition-colors"
							title="Supprimer ces horaires"
							aria-label="Supprimer ces horaires"
						>
							<X size={14} />
						</button>
					</div>
				</div>
			{:else}
				<p class="p-5 text-sm text-muted-foreground">
					Pas encore d’horaires réguliers. « Nouvelle disponibilité » les ajoute, par exemple chaque mardi et jeudi de 9h à 13h.
				</p>
			{/each}
		</div>
	</div>
</div>

<!-- ========== CREATE MODAL ========== -->
<Modal
	bind:isOpen={createModalOpen}
	title="Nouvelle disponibilité"
	description="Des horaires qui reviennent chaque semaine, ou seulement certains jours."
>
	<div class="space-y-5">
		{#if formError}
			<div class="p-3 rounded-md bg-red-50 text-red-700 text-sm">{formError}</div>
		{/if}
		{#if conflictWarning}
			<div class="p-3 rounded-md bg-amber-50 text-amber-700 text-sm flex items-start gap-2">
				<AlertTriangle size={16} class="mt-0.5 shrink-0" />
				<span>{conflictWarning}</span>
			</div>
		{/if}

		<!-- Mode -->
		<div class="grid grid-cols-2 p-1 rounded-lg bg-muted text-sm font-medium" role="tablist">
			{#each [['weekly', 'Chaque semaine'], ['dates', 'Certains jours']] as [value, label]}
				<button
					type="button"
					role="tab"
					aria-selected={formMode === value}
					onclick={() => {
						formMode = value as typeof formMode;
						formError = null;
					}}
					class="py-1.5 rounded-md transition-colors {formMode === value
						? 'bg-background text-foreground shadow-mini'
						: 'text-muted-foreground hover:text-foreground'}"
				>
					{label}
				</button>
			{/each}
		</div>

		<!-- Hours -->
		<div class="grid grid-cols-2 gap-3">
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="create-start-time">De</label>
				<select
					id="create-start-time"
					bind:value={formStartTime}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each timeOptions as opt}
						<option value={opt}>{opt}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="create-end-time">À</label>
				<select
					id="create-end-time"
					bind:value={formEndTime}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each timeOptions as opt}
						<option value={opt}>{opt}</option>
					{/each}
				</select>
			</div>
		</div>

		{#if formMode === 'weekly'}
			<!-- Weekdays -->
			<div class="space-y-2">
				<p class="text-sm font-medium text-foreground">Les jours</p>
				<div class="grid grid-cols-7 gap-1.5">
					{#each weekdayChips as [label, idx]}
						{@const checked = formDaysOfWeek.includes(idx)}
						<button
							type="button"
							aria-pressed={checked}
							onclick={() => toggleWeekday(idx)}
							class="py-1.5 rounded-md text-sm font-medium border transition-colors {checked
								? 'bg-foreground text-background border-foreground'
								: 'bg-background text-foreground border-border hover:bg-muted'}"
						>
							{label}
						</button>
					{/each}
				</div>
			</div>

			<!-- Period -->
			<div class="space-y-2">
				<p class="text-sm font-medium text-foreground">La période</p>
				<div class="grid grid-cols-2 gap-2">
					<button
						type="button"
						onclick={() => (formPicking = 'from')}
						aria-pressed={formPicking === 'from'}
						class="text-left px-3 py-2 rounded-lg border text-sm transition-colors {formPicking === 'from'
							? 'border-foreground ring-1 ring-foreground'
							: 'border-border hover:bg-muted'}"
					>
						<span class="block text-xs text-muted-foreground">À partir du</span>
						{formFrom ? formatDate(formFrom) : '—'}
					</button>
					<div
						class="relative px-3 py-2 rounded-lg border text-sm transition-colors {formPicking === 'until'
							? 'border-foreground ring-1 ring-foreground'
							: 'border-border hover:bg-muted'}"
					>
						<button type="button" onclick={() => (formPicking = 'until')} class="text-left w-full">
							<span class="block text-xs text-muted-foreground">Jusqu’au</span>
							{formUntil ? formatDate(formUntil) : 'Sans date de fin'}
						</button>
						{#if formUntil}
							<button
								type="button"
								onclick={() => (formUntil = '')}
								class="absolute top-1.5 right-1.5 p-1 rounded text-muted-foreground hover:text-foreground hover:bg-muted"
								aria-label="Retirer la date de fin"
							>
								<X size={13} />
							</button>
						{/if}
					</div>
				</div>
				<p class="text-xs text-muted-foreground">
					Touchez sur le calendrier {formPicking === 'from' ? 'le premier jour' : 'le dernier jour, ou laissez sans date de fin'}. Un point marque les jours déjà ouverts.
				</p>
			</div>
		{:else}
			<p class="text-sm text-muted-foreground">
				Touchez les jours voulus sur le calendrier ; touchez à nouveau pour en retirer un. Un point marque les jours déjà ouverts.
			</p>
		{/if}

		<div class="rounded-lg border border-border p-3">
			<AvailabilityCalendar
				bind:month={pickerMonth}
				{today}
				min={today}
				info={pickerInfo}
				onPick={pickInForm}
				compact
			/>
		</div>

		<!-- Summary -->
		<p class="text-sm text-foreground">
			{#if formMode === 'weekly'}
				{#if formDaysOfWeek.length === 0}
					<span class="text-muted-foreground">Choisissez les jours de la semaine.</span>
				{:else}
					Chaque {formWeekdaysLabel}, de {formStartTime} à {formEndTime}, à partir du {formatDate(formFrom)}{formUntil
						? ` jusqu’au ${formatDate(formUntil)}`
						: ', sans date de fin'}
				{/if}
			{:else if formDates.length === 0}
				<span class="text-muted-foreground">Aucun jour choisi.</span>
			{:else}
				{formDates.length} jour{formDates.length > 1 ? 's' : ''}, de {formStartTime} à {formEndTime} :
				{formDates.map(formatDate).join(', ')}
			{/if}
		</p>

		<!-- Actions -->
		<div class="flex justify-end gap-3 pt-1">
			<button
				onclick={() => (createModalOpen = false)}
				class="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
			>
				Annuler
			</button>
			<button
				onclick={handleCreate}
				disabled={submitting}
				class="inline-flex items-center gap-2 px-4 py-2 bg-foreground text-background rounded-lg text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
			>
				{#if submitting}
					<Loader2 size={14} class="animate-spin" />
				{/if}
				Créer
			</button>
		</div>
	</div>
</Modal>

<!-- ========== EDIT MODAL ========== -->
<Modal
	bind:isOpen={editModalOpen}
	title={selectedSlot?.isRecurring ? 'Modifier les horaires réguliers' : 'Modifier ce jour'}
	description={selectedSlot?.isRecurring
		? 'Les nouveaux horaires s’appliquent à partir du jour choisi. Les réservations déjà prises ne bougent pas.'
		: 'Modifiez les horaires de ce créneau. Les réservations déjà prises ne bougent pas.'}
>
	<div class="space-y-4">
		{#if formError}
			<div class="p-3 rounded-md bg-red-50 text-red-700 text-sm">{formError}</div>
		{/if}
		{#if conflictWarning}
			<div
				class="p-3 rounded-md bg-amber-50 text-amber-700 text-sm flex items-start gap-2"
			>
				<AlertTriangle size={16} class="mt-0.5 shrink-0" />
				<span>{conflictWarning}</span>
			</div>
		{/if}

		<!-- Date / Start / End time for edit (10-minute steps, ADR-0006) -->
		<div class="space-y-2">
			<label class="text-sm font-medium text-foreground" for="edit-date">
				{selectedSlot?.isRecurring ? 'À partir du' : 'Date'}
			</label>
			<input
				id="edit-date"
				type="date"
				bind:value={editDate}
				min={todayDate()}
				class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm h-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
			/>
		</div>
		<div class="grid grid-cols-2 gap-3">
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="edit-start">Début</label>
				<select
					id="edit-start"
					bind:value={editStartTime}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each timeOptions as opt}
						<option value={opt}>{opt}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="edit-end">Fin</label>
				<select
					id="edit-end"
					bind:value={editEndTime}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					{#each timeOptions as opt}
						<option value={opt}>{opt}</option>
					{/each}
				</select>
			</div>
		</div>

		<!-- Actions -->
		<div class="flex justify-end gap-3 pt-2">
			<button
				onclick={() => (editModalOpen = false)}
				class="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
			>
				Annuler
			</button>
			<button
				onclick={handleEdit}
				disabled={submitting}
				class="inline-flex items-center gap-2 px-4 py-2 bg-foreground text-background rounded-lg text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
			>
				{#if submitting}
					<Loader2 size={14} class="animate-spin" />
				{/if}
				Enregistrer
			</button>
		</div>
	</div>
</Modal>

<!-- ========== CANCEL CONFIRM MODAL ========== -->
<Modal
	bind:isOpen={cancelConfirmOpen}
	title={selectedSlot?.isRecurring ? 'Supprimer ces horaires réguliers' : 'Supprimer ce jour'}
	description={selectedSlot?.isRecurring
		? 'Ces horaires ne seront plus proposés, sur aucun jour. Les réservations déjà prises sont conservées : annulez-les une à une si besoin.'
		: 'Ce créneau ne sera plus proposé. Les réservations déjà prises sont conservées : annulez-les une à une si besoin.'}
>
	<div class="space-y-4">
		{#if formError}
			<div class="p-3 rounded-md bg-red-50 text-red-700 text-sm">{formError}</div>
		{/if}

		{#if selectedSlot}
			<div class="p-3 bg-muted rounded-md text-sm">
				<p class="font-medium text-foreground">
					{#if selectedSlot.isRecurring}{recurrenceLabel(selectedSlot)} ·{:else}{formatDate(selectedSlot.date)} ·{/if}
					{formatTime(selectedSlot.startTime)} – {formatTime(selectedSlot.endTime)}
				</p>
			</div>
		{/if}

		<div class="flex justify-end gap-3 pt-2">
			<button
				onclick={() => (cancelConfirmOpen = false)}
				class="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
			>
				Retour
			</button>
			<button
				onclick={handleCancel}
				disabled={submitting}
				class="inline-flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition-colors disabled:opacity-50"
			>
				{#if submitting}
					<Loader2 size={14} class="animate-spin" />
				{/if}
				Supprimer
			</button>
		</div>
	</div>
</Modal>

<!-- ========== CLOSE DATES (AVAILABILITY EXCEPTION) MODAL ========== -->
<Modal
	bind:isOpen={exceptionModalOpen}
	title="Fermer une période"
	description="Ces horaires ne seront pas proposés ces jours-là (congés…). Les réservations déjà prises sont conservées."
>
	<div class="space-y-4">
		{#if exceptionError}
			<div class="p-3 rounded-md bg-red-50 text-red-700 text-sm">{exceptionError}</div>
		{/if}

		{#if exceptionSeries}
			<div class="p-3 bg-muted rounded-md text-sm">
				<p class="font-medium text-foreground">
					{recurrenceLabel(exceptionSeries)} · {formatTime(exceptionSeries.startTime)} – {formatTime(exceptionSeries.endTime)}
				</p>
			</div>
		{/if}

		<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="exception-start">Du</label>
				<input
					id="exception-start"
					type="date"
					bind:value={exceptionStart}
					min={todayDate()}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm h-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				/>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="exception-end">
					Au (facultatif)
				</label>
				<input
					id="exception-end"
					type="date"
					bind:value={exceptionEnd}
					min={exceptionStart || todayDate()}
					class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm h-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
				/>
			</div>
		</div>
		<p class="text-xs text-muted-foreground">Sans dernier jour, seul le premier jour est fermé.</p>

		<div class="flex justify-end gap-3 pt-2">
			<button
				onclick={() => (exceptionModalOpen = false)}
				class="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
			>
				Annuler
			</button>
			<button
				onclick={handleCreateException}
				disabled={submitting}
				class="inline-flex items-center gap-2 px-4 py-2 bg-foreground text-background rounded-lg text-sm font-medium hover:opacity-90 transition-opacity disabled:opacity-50"
			>
				{#if submitting}
					<Loader2 size={14} class="animate-spin" />
				{/if}
				Fermer ces dates
			</button>
		</div>
	</div>
</Modal>
