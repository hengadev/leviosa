<script lang="ts">
	import type { PageData } from "./$types";
	import Alert from "#lib/ui/Alert.svelte";

	type EmailAudience = "client" | "partner" | "practice";
	type EmailLocale = "fr" | "en";

	let { data }: { data: PageData } = $props();

	const groups: { audience: EmailAudience; label: string }[] = [
		{ audience: "client", label: "Client" },
		{ audience: "partner", label: "Praticien" },
		{ audience: "practice", label: "Cabinet" },
	];

	const localeLabels: Record<EmailLocale, string> = { fr: "FR", en: "EN" };
	const localeNames: Record<EmailLocale, string> = { fr: "français", en: "anglais" };

	const selected = $derived(data.entries.find((e) => e.id === data.selectedId) ?? null);
	const singleLocale = $derived(selected !== null && selected.locales.length < 2);

	// Fits the frame to the email, so a short email leaves no blank area and
	// a long one has no scroll of its own. Keeps the CSS height if the email
	// can't be measured.
	function fitToContent(event: Event) {
		const frame = event.currentTarget as HTMLIFrameElement;
		const doc = frame.contentDocument;
		if (!doc) return;
		// Measured from a zero height: otherwise the document is at least as
		// tall as the frame and a short email never shrinks it.
		frame.style.height = "0px";
		const height = doc.documentElement.scrollHeight;
		frame.style.height = height ? `${height}px` : "";
	}

	const href = (id: string, lang: EmailLocale) =>
		`?email=${encodeURIComponent(id)}&lang=${lang}`;
</script>

<svelte:head>
	<title>Emails | Admin</title>
</svelte:head>

<div class="container mx-auto px-4 py-8 lg:py-12">
	<div class="mb-8">
		<h1 class="text-3xl lg:text-4xl font-bold mb-2 text-foreground">Emails</h1>
		<p class="text-muted-foreground">
			Les emails envoyés par la plateforme, tels que les reçoivent clients et praticiens.
			Les aperçus utilisent des données d'exemple : aucun email n'est envoyé.
		</p>
	</div>

	{#if data.entries.length === 0}
		<Alert type={data.error ? "error" : "info"} message={data.error ?? "Aucun email à afficher."} />
	{:else}
		<div class="grid gap-6 lg:grid-cols-[16rem_minmax(0,1fr)] lg:items-start">
			<nav aria-label="Liste des emails" class="bg-card rounded-lg border border-border p-3">
				{#each groups as group (group.audience)}
					{@const items = data.entries.filter((e) => e.audience === group.audience)}
					{#if items.length > 0}
						<h2 class="px-2 pt-3 pb-1.5 mb-1 border-b border-border text-sm font-bold text-foreground">
							{group.label}
						</h2>
						<ul class="mb-2 space-y-0.5">
							{#each items as entry (entry.id)}
								{@const current = entry.id === data.selectedId}
								<li>
									<a
										href={href(entry.id, data.locale)}
										aria-current={current ? "page" : undefined}
										class="block rounded-md px-2 py-1.5 text-sm transition-colors {current
											? 'bg-foreground text-background font-medium'
											: 'text-foreground hover:bg-muted'}"
									>
										{entry.label}
									</a>
								</li>
							{/each}
						</ul>
					{/if}
				{/each}
			</nav>

			<section aria-label="Aperçu" class="min-w-0 space-y-4">
				{#if selected}
					<div class="flex flex-wrap items-center gap-3">
						<div role="group" aria-label="Langue" class="inline-flex rounded-md border border-border bg-card p-0.5">
							{#each ["fr", "en"] as const as lang (lang)}
								{@const available = selected.locales.includes(lang)}
								{@const current = lang === data.locale}
								{#if available}
									<a
										href={href(selected.id, lang)}
										aria-current={current ? "true" : undefined}
										class="rounded px-3 py-1 text-sm font-medium transition-colors {current
											? 'bg-foreground text-background'
											: 'text-foreground hover:bg-muted'}"
									>
										{localeLabels[lang]}
									</a>
								{:else}
									<span
										aria-disabled="true"
										class="cursor-not-allowed rounded px-3 py-1 text-sm font-medium text-muted-foreground/60"
									>
										{localeLabels[lang]}
									</span>
								{/if}
							{/each}
						</div>
						{#if singleLocale}
							<p class="text-sm text-muted-foreground">
								Cet email n'existe qu'en {localeNames[selected.locales[0]]}.
							</p>
						{/if}
					</div>
				{/if}

				{#if data.error}
					<Alert type="error" message={data.error} />
				{:else if data.preview}
					<div class="bg-card rounded-lg border border-border p-4 space-y-1">
						<p class="text-sm text-muted-foreground">Objet</p>
						<p class="font-medium text-foreground break-words">{data.preview.subject}</p>
						{#if data.preview.attachments.length > 0}
							<p class="pt-2 text-sm text-muted-foreground">
								Pièces jointes : <span class="text-foreground">{data.preview.attachments.join(", ")}</span>
							</p>
						{/if}
					</div>

					<!-- No allow-scripts: nothing in the email can run. allow-same-origin
					     only lets this page read the email's height to fit the frame. -->
					<iframe
						title="Aperçu de l'email : {data.preview.subject}"
						srcdoc={data.preview.html}
						sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
						onload={fitToContent}
						class="block w-full h-[1400px] rounded-lg border border-border bg-white"
					></iframe>
				{/if}
			</section>
		</div>
	{/if}
</div>
