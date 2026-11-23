<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Ban, CalendarX, Phone } from '@lucide/svelte';
	import PageActionsMenu from '#lib/components/admin/PageActionsMenu.svelte';
	import PractitionerProfileMissing from '#lib/components/admin/PractitionerProfileMissing.svelte';
	import SmsButton from '#lib/components/practice/SmsButton.svelte';
	import type { PageData } from './$types';
	import type { TodaySession } from './+page.server';

	let { data }: { data: PageData } = $props();

	// Sessions drop out of the list as they end, without a reload.
	let now = $state(Date.now());
	onMount(() => {
		now = Date.now();
		const timer = setInterval(() => (now = Date.now()), 30_000);
		return () => clearInterval(timer);
	});

	const remaining = $derived(
		data.state === 'ready'
			? data.sessions.filter((s) => new Date(s.endsAt).getTime() > now)
			: []
	);
	const next = $derived(remaining[0]);
	const later = $derived(remaining.slice(1));

	function countLabel(count: number): string {
		if (count === 0) return 'aucune séance';
		return `${count} ${count > 1 ? 'séances' : 'séance'}`;
	}

	/** First word of the stored client name, for SMS button labels. */
	function firstName(name: string): string {
		return name.trim().split(/\s+/)[0] || name;
	}

	/** "Prochaine séance · dans 25 min", or "En cours" once it has started. */
	function nextHeading(session: TodaySession): string {
		const minutes = Math.ceil((new Date(session.startsAt).getTime() - now) / 60_000);
		if (minutes <= 0) return 'En cours';
		if (minutes < 60) return `Prochaine séance · dans ${minutes} min`;
		const hours = Math.floor(minutes / 60);
		const rest = minutes % 60;
		return `Prochaine séance · dans ${hours} h${rest > 0 ? ` ${String(rest).padStart(2, '0')}` : ''}`;
	}

	// Bookings being closed or reopened: "id+action" keys, to disable the button.
	const pending = $state(new Set<string>());

	/** POSTs one of the booking actions (complete, no-show, reopen) and
	 * reloads the day. */
	async function runBookingAction(id: string, action: 'complete' | 'no-show' | 'reopen') {
		const key = `${id}:${action}`;
		pending.add(key);
		try {
			const res = await fetch(`/api/bookings/${id}/${action}`, { method: 'POST' });
			if (!res.ok) {
				console.error(`Failed to ${action} the booking ${id}:`, res.status);
			}
			await invalidateAll();
		} catch (err) {
			console.error(`Failed to ${action} the booking ${id}:`, err);
		} finally {
			pending.delete(key);
		}
	}
</script>

<svelte:head>
	<title>Aujourd'hui | Leviosa Admin</title>
</svelte:head>

<div class="container mx-auto max-w-3xl px-4 py-8 lg:py-12">
	<div class="mb-6">
		<div class="mb-1 flex items-center justify-between gap-4">
			<h1 class="text-3xl font-bold lg:text-4xl">Aujourd'hui</h1>
			<!-- The sidebar, which holds the theme toggle, shows only from lg up -->
			<div class="lg:hidden">
				<PageActionsMenu />
			</div>
		</div>
		{#if data.state === 'ready'}
			<p class="text-muted-foreground">{data.dateLabel} · {countLabel(data.sessionCount)}</p>
		{/if}
	</div>

	{#if data.state === 'not-set-up'}
		<PractitionerProfileMissing />
	{:else}
		<div class="space-y-6">
			{#if next}
				<section class="rounded-2xl bg-foreground p-4 text-background sm:p-5">
					<p class="text-xs uppercase tracking-wider opacity-70">{nextHeading(next)}</p>
					<p class="mt-2 text-2xl font-semibold tabular-nums">{next.startLabel} – {next.endLabel}</p>
					<a href="/staff/clients/{next.id}" class="mt-1 block break-words text-lg hover:underline">{next.clientName}</a>
					<p class="break-words text-sm opacity-70">
						{next.productName}{next.paymentLabel ? ` · ${next.paymentLabel}` : ''}
					</p>
					{#if next.phone}
						<div class="mt-4 grid grid-cols-[1fr_2.75rem] gap-2">
							<a
								href="tel:{next.phone}"
								class="flex h-11 items-center justify-center gap-2 rounded-xl bg-background font-medium text-foreground"
							>
								<Phone size={18} />
								Appeler {next.phoneLabel}
							</a>
							<SmsButton phone={next.phone} firstName={firstName(next.clientName)} variant="wide" />
						</div>
					{/if}
				</section>
			{:else if data.state === 'ready' && data.toClose.length === 0}
				<section class="rounded-2xl border border-border-card p-6 text-center">
					<CalendarX size={28} class="mx-auto mb-3 text-muted-foreground" />
					<p class="font-medium text-foreground">
						{data.state === 'ready' && data.sessionCount > 0
							? "Plus de séance aujourd'hui"
							: "Aucune séance aujourd'hui"}
					</p>
				</section>
			{/if}

			{#if data.state === 'ready' && data.toClose.length > 0}
				<section>
					<h2 class="px-1 pb-1 text-sm font-semibold">À clôturer</h2>
					<ul class="space-y-2">
						{#each data.toClose as session (session.id)}
							<li class="rounded-2xl border border-border-card p-4">
								<p class="text-sm font-semibold tabular-nums">{session.startLabel} – {session.endLabel}</p>
								<a href="/staff/clients/{session.id}" class="-my-2.5 block truncate py-2.5 font-medium hover:underline">{session.clientName}</a>
								<p class="truncate text-xs text-muted-foreground">{session.productName}</p>
								<p class="mt-2 text-sm text-muted-foreground">Séance terminée ?</p>
								<div class="mt-2 grid grid-cols-2 gap-2">
									<button
										type="button"
										class="flex h-11 items-center justify-center rounded-xl bg-foreground font-medium text-background disabled:opacity-50"
										disabled={pending.has(`${session.id}:complete`)}
										onclick={() => runBookingAction(session.id, 'complete')}
									>
										Terminée
									</button>
									<button
										type="button"
										class="flex h-11 items-center justify-center rounded-xl border border-border-card font-medium hover:bg-muted disabled:opacity-50"
										disabled={pending.has(`${session.id}:no-show`)}
										onclick={() => runBookingAction(session.id, 'no-show')}
									>
										Absent
									</button>
								</div>
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			{#if later.length > 0}
				<section>
					<h2 class="px-1 pb-1 text-sm font-semibold">Ensuite</h2>
					<ul class="divide-y divide-border-card rounded-2xl border border-border-card">
						{#each later as session (session.id)}
							<li class="flex items-center gap-3 px-4 py-3">
								<div class="w-12 shrink-0 text-sm font-semibold tabular-nums">
									{session.startLabel}
								</div>
								<div class="min-w-0 flex-1">
									<a href="/staff/clients/{session.id}" class="-my-2.5 block truncate py-2.5 font-medium hover:underline">{session.clientName}</a>
									{#if session.paymentLabel}
										<p class="truncate text-xs text-muted-foreground">{session.paymentLabel}</p>
									{/if}
								</div>
								{#if session.phone}
									<a
										href="tel:{session.phone}"
										class="flex size-11 shrink-0 items-center justify-center rounded-full border border-border-card hover:bg-muted"
										aria-label="Appeler {session.clientName}"
									>
										<Phone size={18} />
									</a>
									<SmsButton phone={session.phone} firstName={firstName(session.clientName)} />
								{/if}
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			{#if data.state === 'ready' && data.closed.length > 0}
				<section>
					<h2 class="px-1 pb-1 text-sm font-semibold">Terminées aujourd'hui</h2>
					<ul class="divide-y divide-border-card rounded-2xl border border-border-card">
						{#each data.closed as session (session.id)}
							<li class="flex items-center gap-3 px-4 py-3">
								<div class="w-12 shrink-0 text-sm font-semibold tabular-nums">
									{session.startLabel}
								</div>
								<div class="min-w-0 flex-1">
									<a href="/staff/clients/{session.id}" class="-my-2.5 block truncate py-2.5 font-medium hover:underline">{session.clientName}</a>
									<p class="truncate text-xs text-muted-foreground">{session.resultLabel}</p>
								</div>
								{#if session.canUndo}
									<button
										type="button"
										class="flex h-11 items-center justify-center rounded-xl border border-border-card px-4 text-sm font-medium text-muted-foreground hover:bg-muted disabled:opacity-50"
										disabled={pending.has(`${session.id}:reopen`)}
										onclick={() => runBookingAction(session.id, 'reopen')}
									>
										Annuler
									</button>
								{/if}
							</li>
						{/each}
					</ul>
				</section>
			{/if}

			{#if data.nextWorkingDay}
				<section>
					<h2 class="px-1 pb-1 text-sm font-semibold">
						Prochain jour · {data.nextWorkingDay.dayLabel}
					</h2>
					{#if data.nextWorkingDay.sessions.length > 0}
						<ul class="divide-y divide-border-card rounded-2xl border border-border-card">
							{#each data.nextWorkingDay.sessions as session (session.id)}
								<li class="flex items-center gap-3 px-4 py-3">
									<div class="w-12 shrink-0 text-sm font-semibold tabular-nums">
										{session.startLabel}
									</div>
									<div class="min-w-0 flex-1">
										<a href="/staff/clients/{session.id}" class="-my-2.5 block truncate py-2.5 font-medium hover:underline">{session.clientName}</a>
										{#if session.paymentLabel}
											<p class="truncate text-xs text-muted-foreground">{session.paymentLabel}</p>
										{/if}
									</div>
									{#if session.phone}
										<a
											href="tel:{session.phone}"
											class="flex size-11 shrink-0 items-center justify-center rounded-full border border-border-card hover:bg-muted"
											aria-label="Appeler {session.clientName}"
										>
											<Phone size={18} />
										</a>
										<SmsButton phone={session.phone} firstName={firstName(session.clientName)} />
									{/if}
								</li>
							{/each}
						</ul>
					{:else}
						<p class="rounded-2xl border border-border-card px-4 py-3 text-sm text-muted-foreground">
							Aucune réservation pour l'instant
						</p>
					{/if}
				</section>
			{/if}

			<a
				href="/staff/agenda/disponibilites"
				class="flex h-12 w-full items-center justify-center gap-2 rounded-xl border border-border-card font-medium hover:bg-muted"
			>
				<Ban size={18} />
				Bloquer un créneau
			</a>
		</div>
	{/if}
</div>
