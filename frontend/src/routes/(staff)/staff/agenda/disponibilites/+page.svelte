<script lang="ts">
	import type { PageProps } from './$types';
	import type { Availability, RoomAllocation } from './+page.server';
	import {
		CalendarClock,
		Clock,
		MapPin,
		Plus,
		X,
		ChevronLeft,
		ChevronRight,
		Pencil,
		AlertTriangle,
		Check,
		Loader2,
		Repeat,
		Building2,
		CalendarOff
	} from '@lucide/svelte';
	import { invalidateAll } from '$app/navigation';
	import Modal from '$lib/ui/Modal.svelte';
	import { FormInput } from '$lib/ui/bits-components';
	import { FormSelect } from '$lib/ui/bits-components';
	import { PRACTICE_TIME_ZONE } from '$lib/i18n/practice';
	import { practiceDate, practiceInstant, type AvailabilityException } from './occurrences';

	let { data }: PageProps = $props();

	// --- Filter state ---
	let filterStatus = $state<'all' | 'available' | 'booked' | 'cancelled'>('all');

	// --- Modal states ---
	let createModalOpen = $state(false);
	let editModalOpen = $state(false);
	let cancelConfirmOpen = $state(false);
	let selectedSlot: Availability | null = $state(null);

	// --- Create form state ---
	let formDate = $state('');
	let formStartTime = $state('09:00');
	let formEndTime = $state('10:00');
	let formRoomId = $state('');
	let formMaxCapacity = $state('1');
	let formNotes = $state('');
	let formIsRecurring = $state(false);
	let formRecurrenceType = $state<'daily' | 'weekly' | 'monthly'>('weekly');
	let formRecurrenceInterval = $state('1');
	let formRecurrenceEndDate = $state('');
	let formDaysOfWeek = $state<number[]>([]);

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

	// --- Loading / error states ---
	let submitting = $state(false);
	let conflictWarning = $state<string | null>(null);
	let checkingConflict = $state(false);
	let formError = $state<string | null>(null);

	// --- Derived data ---
	const activeAllocations = $derived(
		(data.allocations as RoomAllocation[])?.filter((a) => a.isActive) ?? []
	);

	const filteredDays = $derived(
		data.availabilities
			.map((day) => ({
				...day,
				slots: day.slots.filter((s) => filterStatus === 'all' || s.status === filterStatus)
			}))
			.filter((day) => day.slots.length > 0)
	);

	const totalAvailable = $derived(
		data.availabilities.flatMap((d) => d.slots).filter((s) => s.status === 'available').length
	);
	const totalBooked = $derived(
		data.availabilities.flatMap((d) => d.slots).filter((s) => s.status === 'booked').length
	);

	// --- Helpers ---
	// Dates (YYYY-MM-DD) are calendar days; instants are shown in the practice time zone.
	function formatDate(dateStr: string): string {
		return new Date(`${dateStr.slice(0, 10)}T12:00:00Z`).toLocaleDateString('fr-FR', {
			day: 'numeric',
			month: 'short',
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

	function todayDate(): string {
		return practiceDate(new Date().toISOString()).toString();
	}

	function statusBadge(status: string): string {
		switch (status) {
			case 'available':
				return 'bg-green-100 text-green-700';
			case 'booked':
				return 'bg-blue-100 text-blue-700';
			case 'cancelled':
				return 'bg-red-100 text-red-700';
			case 'blocked':
				return 'bg-gray-200 text-gray-600';
			default:
				return 'bg-gray-100 text-gray-700';
		}
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'available':
				return 'Disponible';
			case 'booked':
				return 'Réservé';
			case 'cancelled':
				return 'Annulé';
			case 'blocked':
				return 'Bloqué';
			default:
				return status;
		}
	}

	const weekdayNames = ['dimanche', 'lundi', 'mardi', 'mercredi', 'jeudi', 'vendredi', 'samedi'];

	function recurrenceLabel(slot: Availability): string {
		if (!slot.recurrencePattern) return '';
		const p = slot.recurrencePattern;
		const n = p.interval > 1 ? p.interval : 1;
		switch (p.type) {
			case 'daily':
				return n > 1 ? `Tous les ${n} jours` : 'Tous les jours';
			case 'weekly': {
				const days = [...(p.days_of_week ?? [])]
					.sort((a, b) => ((a + 6) % 7) - ((b + 6) % 7))
					.map((d) => weekdayNames[d])
					.join(', ');
				return n > 1 ? `Toutes les ${n} semaines, le ${days}` : `Chaque semaine, le ${days}`;
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

	// --- Allocation dropdown options ---
	const roomOptions = $derived(
		activeAllocations.length > 0
			? activeAllocations.map((a) => ({
					value: a.roomId,
					label: `Salle ${a.roomId.slice(0, 8)}… (${a.allocationType === 'dedicated' ? 'Dédiée' : 'Partagée'})`
				}))
			: [{ value: '', label: 'Aucune salle allouée' }]
	);

	const recurrenceTypeOptions = [
		{ value: 'daily', label: 'Quotidien' },
		{ value: 'weekly', label: 'Hebdomadaire' },
		{ value: 'monthly', label: 'Mensuel' }
	];

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

	// --- Reset form ---
	function resetCreateForm() {
		formDate = todayDate();
		formStartTime = '09:00';
		formEndTime = '10:00';
		formRoomId = activeAllocations[0]?.roomId ?? '';
		formMaxCapacity = '1';
		formNotes = '';
		formIsRecurring = false;
		formRecurrenceType = 'weekly';
		formRecurrenceInterval = '1';
		formRecurrenceEndDate = '';
		formDaysOfWeek = [new Date(`${formDate}T12:00:00Z`).getUTCDay()];
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
		submitting = true;
		formError = null;
		conflictWarning = null;

		if (!formDate || !formStartTime || !formEndTime || !formRoomId) {
			formError = 'Date, horaires et salle sont obligatoires.';
			submitting = false;
			return;
		}

		const startIso = toDatetime(formDate, formStartTime);
		const endIso = toDatetime(formDate, formEndTime);

		if (startIso >= endIso) {
			formError = "L'heure de début doit être antérieure à l'heure de fin.";
			submitting = false;
			return;
		}

		if (startIso <= new Date().toISOString()) {
			formError = 'Le créneau doit commencer dans le futur.';
			submitting = false;
			return;
		}

		if (formIsRecurring && formRecurrenceEndDate && formRecurrenceEndDate < formDate) {
			formError = 'La date de fin doit suivre le premier jour.';
			submitting = false;
			return;
		}

		if (formIsRecurring && formRecurrenceType === 'weekly' && formDaysOfWeek.length === 0) {
			formError = 'Sélectionnez au moins un jour de répétition.';
			submitting = false;
			return;
		}

		// Check conflict
		checkingConflict = true;
		const hasConflict = await checkConflict(data.partnerId, startIso, endIso);
		checkingConflict = false;

		if (hasConflict) {
			conflictWarning = 'Ce créneau est en conflit avec un créneau existant.';
			submitting = false;
			return;
		}

		try {
			const body: any = {
				room_id: formRoomId,
				start_time: startIso,
				end_time: endIso,
				max_capacity: parseInt(formMaxCapacity, 10),
				notes: formNotes || undefined
			};

			if (formIsRecurring) {
				body.pattern = {
					type: formRecurrenceType,
					interval: parseInt(formRecurrenceInterval, 10) || 1,
					// The end date is inclusive: the series' last day.
					until: formRecurrenceEndDate ? toDatetime(formRecurrenceEndDate, '23:59') : undefined,
					days_of_week: formRecurrenceType === 'weekly' ? formDaysOfWeek : undefined
				};

				const res = await fetch('/api/availabilities/recurring', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(body)
				});

				if (!res.ok) {
					const err = await res.json().catch(() => ({}));
					formError = err.message || 'Erreur lors de la création de la disponibilité récurrente.';
					submitting = false;
					return;
				}
			} else {
				const res = await fetch('/api/availabilities', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(body)
				});

				if (!res.ok) {
					const err = await res.json().catch(() => ({}));
					formError = err.message || 'Erreur lors de la création de la disponibilité.';
					submitting = false;
					return;
				}
			}

			createModalOpen = false;
			resetCreateForm();
			await refreshPage();
		} catch (e: any) {
			formError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
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
			await refreshPage();
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
			await refreshPage();
		} catch (e: any) {
			formError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	// Open modals
	function openCreateModal() {
		resetCreateForm();
		createModalOpen = true;
	}

	// A recurring availability is edited from a first day on (the series'
	// first occurrence moves there); by default the later of its current
	// first day and tomorrow. Existing bookings never move.
	function openEditModal(slot: Availability) {
		const series = slot.isRecurring
			? ((data.series as Availability[]).find((a) => a.id === slot.id) ?? slot)
			: slot;
		selectedSlot = series;
		resetEditForm();
		if (series.isRecurring) {
			const tomorrow = practiceDate(new Date(Date.now() + 24 * 3600 * 1000).toISOString()).toString();
			const first = practiceDate(series.startTime).toString();
			editDate = first > tomorrow ? first : tomorrow;
		} else {
			editDate = practiceDate(series.startTime).toString();
		}
		editStartTime = practiceClock(series.startTime);
		editEndTime = practiceClock(series.endTime);
		editNotes = series.notes ?? '';
		editModalOpen = true;
	}

	function openCancelConfirm(slot: Availability) {
		selectedSlot = slot.isRecurring
			? ((data.series as Availability[]).find((a) => a.id === slot.id) ?? slot)
			: slot;
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
			const res = await fetch(`/api/availabilities/${exceptionSeries.id}/exceptions`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					start_date: exceptionStart,
					end_date: exceptionEnd || undefined
				})
			});
			if (!res.ok) {
				const err = await res.json().catch(() => ({}));
				exceptionError = err.message || 'Erreur lors de la fermeture des dates.';
				return;
			}
			const created = await res.json();
			exceptionNotice = affectedBookingsNotice(created.affected_bookings ?? 0);
			exceptionModalOpen = false;
			exceptionSeries = null;
			await invalidateAll();
		} catch (e: any) {
			exceptionError = e.message || 'Erreur réseau.';
		} finally {
			submitting = false;
		}
	}

	async function handleDeleteException(serie: Availability, e: AvailabilityException) {
		deletingExceptionId = e.id;
		exceptionNotice = null;
		try {
			const res = await fetch(`/api/availabilities/${serie.id}/exceptions/${e.id}`, {
				method: 'DELETE'
			});
			if (!res.ok && res.status !== 404) {
				const err = await res.json().catch(() => ({}));
				exceptionNotice = err.message || 'Erreur lors de la réouverture des dates.';
				return;
			}
			await invalidateAll();
		} catch (err: any) {
			exceptionNotice = err.message || 'Erreur réseau.';
		} finally {
			deletingExceptionId = null;
		}
	}

	async function refreshPage() {
		// Full page reload to refetch from server
		window.location.reload();
	}
</script>

<svelte:head>
	<title>Disponibilités | Staff</title>
</svelte:head>

<div class="p-6 lg:p-10">
	<div class="mb-10 flex flex-col sm:flex-row sm:items-end sm:justify-between gap-4">
		<div>
			<p class="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted-foreground mb-3">Agenda</p>
			<h1 class="font-display text-4xl lg:text-5xl font-semibold tracking-tight text-foreground leading-[1.1]">Disponibilités</h1>
			<p class="text-muted-foreground mt-3 text-sm">Vos disponibilités ponctuelles et récurrentes, semaine par semaine</p>
			<div class="mt-4 h-px w-16 bg-foreground/20"></div>
		</div>
		<button
			onclick={openCreateModal}
			class="inline-flex items-center gap-2 px-4 py-2.5 bg-foreground text-background rounded-lg text-sm font-medium hover:opacity-90 transition-opacity"
		>
			<Plus size={16} />
			Nouvelle disponibilité
		</button>
	</div>

	<!-- Recurring availabilities -->
	{#if data.series.length > 0}
		<div class="mb-8 bg-background rounded-xl border border-border overflow-hidden">
			<div class="bg-muted/50 px-5 py-3 border-b border-border flex items-center gap-2">
				<Repeat size={15} class="text-muted-foreground" />
				<h2 class="font-semibold text-foreground">Disponibilités récurrentes</h2>
			</div>
			{#if exceptionNotice}
				<div
					class="px-5 py-3 border-b border-border bg-amber-50 text-amber-800 text-sm flex items-start gap-2"
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
			<div class="divide-y divide-border">
				{#each data.series as serie (serie.id)}
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
								title="Fermer des dates"
								aria-label="Fermer des dates"
							>
								<CalendarOff size={14} />
							</button>
							<button
								onclick={() => openEditModal(serie)}
								class="p-1.5 rounded-md text-muted-foreground hover:text-blue-600 hover:bg-blue-50 transition-colors"
								title="Modifier la série"
								aria-label="Modifier la série"
							>
								<Pencil size={14} />
							</button>
							<button
								onclick={() => openCancelConfirm(serie)}
								class="p-1.5 rounded-md text-muted-foreground hover:text-red-600 hover:bg-red-50 transition-colors"
								title="Supprimer la série"
								aria-label="Supprimer la série"
							>
								<X size={14} />
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Room Allocations Info -->
	{#if activeAllocations.length > 0}
		<div class="mb-6 bg-background rounded-xl border border-border p-4">
			<div class="flex items-center gap-2 mb-2">
				<Building2 size={16} class="text-muted-foreground" />
				<h3 class="text-sm font-semibold text-foreground">Salles allouées</h3>
			</div>
			<div class="flex flex-wrap gap-2">
				{#each activeAllocations as alloc}
					<span
						class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-medium {alloc.allocationType === 'dedicated'
							? 'bg-purple-100 text-purple-700'
							: 'bg-amber-100 text-amber-700'}"
					>
						<MapPin size={12} />
						Salle {alloc.roomId.slice(0, 8)}…
						({alloc.allocationType === 'dedicated' ? 'Dédiée' : 'Partagée'})
						{#if alloc.startDate}
							— du {formatDate(alloc.startDate)}
							{#if alloc.endDate}au {formatDate(alloc.endDate)}{/if}
						{/if}
					</span>
				{/each}
			</div>
		</div>
	{/if}

	<!-- Summary Cards -->
	<div class="grid grid-cols-3 gap-4 mb-8">
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Libres</p>
			<p class="text-2xl font-bold text-green-600 dark:text-green-400 tabular-nums">{totalAvailable}</p>
		</div>
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Réservés</p>
			<p class="text-2xl font-bold text-blue-600 dark:text-blue-400 tabular-nums">{totalBooked}</p>
		</div>
		<div class="bg-background rounded-2xl border border-border p-5">
			<p class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">Occupation</p>
			<p class="text-2xl font-bold text-foreground tabular-nums">
				{totalBooked + totalAvailable > 0
					? Math.round((totalBooked / (totalBooked + totalAvailable)) * 100)
					: 0}%
			</p>
		</div>
	</div>

	<!-- Week Navigation -->
	<div
		class="flex items-center justify-between mb-6 bg-background rounded-xl border border-border p-4"
	>
		{#if data.weekOffset > 0}
			<a
				href="?semaine={data.weekOffset - 1}"
				class="p-2 hover:bg-muted rounded-md transition-colors"
				aria-label="Semaine précédente"
			>
				<ChevronLeft size={20} class="text-muted-foreground" />
			</a>
		{:else}
			<span class="p-2 opacity-40" aria-hidden="true"><ChevronLeft size={20} class="text-muted-foreground" /></span>
		{/if}
		<div class="flex items-center gap-2">
			<CalendarClock size={18} class="text-muted-foreground" />
			<span class="font-semibold text-foreground">
				{#if data.weekOffset === 0}
					7 prochains jours
				{:else if data.availabilities.length > 0}
					À partir du {formatDate(data.availabilities[0].date)}
				{/if}
			</span>
		</div>
		<a
			href="?semaine={data.weekOffset + 1}"
			class="p-2 hover:bg-muted rounded-md transition-colors"
			aria-label="Semaine suivante"
		>
			<ChevronRight size={20} class="text-muted-foreground" />
		</a>
	</div>

	<!-- Status Filter -->
	<div class="flex gap-2 mb-6 flex-wrap">
		{#each [['all', 'Tous'], ['available', 'Libres'], ['booked', 'Réservés'], ['cancelled', 'Annulés']] as [val, label]}
			<button
				class="px-4 py-2 rounded-md text-sm font-medium transition-colors {filterStatus === val
					? 'bg-foreground text-background'
					: 'bg-card text-foreground border border-border hover:bg-muted'}"
				onclick={() => (filterStatus = val as typeof filterStatus)}
			>
				{label}
			</button>
		{/each}
	</div>

	<!-- Days and Slots -->
	<div class="space-y-6">
		{#each filteredDays as day (day.date)}
			<div class="bg-background rounded-xl border border-border overflow-hidden">
				<div class="bg-muted/50 px-5 py-3 border-b border-border">
					<h2 class="font-semibold text-foreground capitalize">
						{day.dayName} {formatDate(day.date)}
					</h2>
				</div>
				<div class="divide-y divide-border">
					{#each day.slots as slot (slot.key)}
						<div class="p-4 sm:p-5 hover:bg-muted/20 transition-colors">
							<div class="flex flex-col sm:flex-row sm:items-center gap-3">
								<div
									class="flex items-center gap-2 text-muted-foreground min-w-fit"
								>
									<Clock size={15} />
									<span class="font-medium text-sm">
										{formatTime(slot.startTime)} – {formatTime(slot.endTime)}
									</span>
									{#if slot.isRecurring}
										<span class="text-xs text-muted-foreground inline-flex items-center gap-1" title={recurrenceLabel(slot)}>
											<Repeat size={12} class="inline" />
											<span class="sr-only">Récurrente :</span>{recurrenceLabel(slot)}
										</span>
									{/if}
								</div>
								<div class="flex-1">
									<div
										class="flex items-center gap-1.5 text-sm text-muted-foreground mb-0.5"
									>
										<MapPin size={13} />
										<span>Salle {slot.roomId.slice(0, 8)}…</span>
									</div>
									{#if slot.serviceType}
										<p class="text-sm text-foreground">{slot.serviceType}</p>
									{/if}
									{#if slot.notes}
										<p class="text-xs text-muted-foreground mt-0.5">{slot.notes}</p>
									{/if}
								</div>
								<div class="flex items-center gap-2">
									<span
										class="px-2.5 py-1 rounded-full text-xs font-medium {statusBadge(slot.status)}"
									>
										{statusLabel(slot.status)}
									</span>
									{#if slot.status === 'available' || slot.status === 'booked'}
										<button
											onclick={() => openEditModal(slot)}
											class="p-1.5 rounded-md text-muted-foreground hover:text-blue-600 hover:bg-blue-50 transition-colors"
											title={slot.isRecurring ? 'Modifier la série' : 'Modifier ce créneau'}
										>
											<Pencil size={14} />
										</button>
										<button
											onclick={() => openCancelConfirm(slot)}
											class="p-1.5 rounded-md text-muted-foreground hover:text-red-600 hover:bg-red-50 transition-colors"
											title={slot.isRecurring ? 'Supprimer la série' : 'Supprimer ce créneau'}
										>
											<X size={14} />
										</button>
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{:else}
			<div class="text-center py-12 text-muted-foreground">
				Aucun créneau pour cette période
			</div>
		{/each}
	</div>
</div>

<!-- ========== CREATE MODAL ========== -->
<Modal bind:isOpen={createModalOpen} title="Nouvelle disponibilité" description="Créez un créneau de disponibilité.">
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

		<!-- Room -->
		<FormSelect
			label="Salle"
			bind:value={formRoomId}
			options={roomOptions}
			placeholder="Choisir une salle"
			required
		/>

		<!-- Date (the first day of a recurring availability) -->
		<div class="space-y-2">
			<label class="text-sm font-medium text-foreground" for="create-date">
				{formIsRecurring ? 'Premier jour' : 'Date'} <span class="text-destructive ml-1">*</span>
			</label>
			<input
				id="create-date"
				type="date"
				bind:value={formDate}
				required
				class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm h-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 hover:border-input-hover"
			/>
		</div>

		<!-- Start / End time -->
		<div class="grid grid-cols-2 gap-3">
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground" for="create-start-time"
					>Début</label
				>
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
				<label class="text-sm font-medium text-foreground" for="create-end-time"
					>Fin</label
				>
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

		<!-- Max capacity -->
		<FormInput label="Capacité max" type="text" bind:value={formMaxCapacity} required />

		<!-- Notes -->
		<FormInput label="Notes" type="text" bind:value={formNotes} placeholder="Optionnel" />

		<!-- Recurring toggle -->
		<div class="flex items-center gap-3">
			<button
				type="button"
				role="switch"
				aria-label="Récurrent"
				aria-checked={formIsRecurring}
				onclick={() => (formIsRecurring = !formIsRecurring)}
				class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out {formIsRecurring
					? 'bg-foreground'
					: 'bg-muted'}"
			>
				<span
					class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out {formIsRecurring
						? 'translate-x-5'
						: 'translate-x-0'}"
				></span>
			</button>
			<span class="text-sm text-foreground">Récurrent</span>
		</div>

		{#if formIsRecurring}
			<div class="space-y-4 border-t border-border pt-4">
				<FormSelect
					label="Fréquence"
					bind:value={formRecurrenceType}
					options={recurrenceTypeOptions}
					required
				/>
				<FormInput
					label="Intervalle (toutes les N)"
					type="number"
					bind:value={formRecurrenceInterval}
				/>

				{#if formRecurrenceType === 'weekly'}
					<div class="space-y-2">
						<p class="text-sm font-medium text-foreground">
							Jours de répétition <span class="text-destructive ml-1">*</span>
						</p>
						<div class="flex flex-wrap gap-2">
							{#each [['Dim', 0], ['Lun', 1], ['Mar', 2], ['Mer', 3], ['Jeu', 4], ['Ven', 5], ['Sam', 6]] as [label, idx]}
								{@const checked = formDaysOfWeek.includes(idx as number)}
								<button
									type="button"
									onclick={() => {
										if (checked) {
											formDaysOfWeek = formDaysOfWeek.filter((d) => d !== idx);
										} else {
											formDaysOfWeek = [...formDaysOfWeek, idx as number];
										}
									}}
									class="px-3 py-1.5 rounded-md text-sm font-medium border transition-colors {checked
										? 'bg-foreground text-background border-foreground'
										: 'bg-background text-foreground border-border hover:bg-muted'}"
								>
									{label}
								</button>
							{/each}
						</div>
					</div>
				{/if}

				<div class="space-y-2">
					<label class="text-sm font-medium text-foreground" for="create-recurrence-end">
						Dernier jour (facultatif)
					</label>
					<input
						id="create-recurrence-end"
						type="date"
						bind:value={formRecurrenceEndDate}
						class="flex w-full rounded-input border border-input bg-background px-3 py-2 text-sm h-10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 hover:border-input-hover"
					/>
				</div>
			</div>
		{/if}

		<!-- Actions -->
		<div class="flex justify-end gap-3 pt-2">
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
	title={selectedSlot?.isRecurring ? 'Modifier la série' : 'Modifier le créneau'}
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

		{#if selectedSlot}
			<div class="text-sm text-muted-foreground mb-2">
				Salle {selectedSlot.roomId.slice(0, 8)}…
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

		<!-- Notes -->
		<FormInput label="Notes" type="text" bind:value={editNotes} placeholder="Optionnel" />

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
	title={selectedSlot?.isRecurring ? 'Supprimer la série' : 'Supprimer le créneau'}
	description={selectedSlot?.isRecurring
		? 'La série ne proposera plus aucun créneau. Les réservations déjà prises sont conservées : annulez-les une à une si besoin.'
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
				<p class="text-muted-foreground mt-1">
					Salle {selectedSlot.roomId.slice(0, 8)}…
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
	title="Fermer des dates"
	description="La série ne proposera aucun créneau ces jours-là. Les réservations déjà prises sont conservées."
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
