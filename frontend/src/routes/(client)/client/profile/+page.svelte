<script lang="ts">
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';
	import { Dialog } from 'bits-ui';
	import {
		CheckCircle,
		ArrowLeft,
		Save,
		MessageSquareText,
		AlertTriangle,
		X,
		CreditCard,
		ExternalLink
	} from '@lucide/svelte';

	let { data, form }: PageProps = $props();

	// Use updated user from form response if available, otherwise use initial data
	let user = $derived(
		form?.success && form.user ? form.user : data.user
	);

	let gender = $state(user?.gender ?? '');
	let birthdate = $state(user?.birthdate ? user.birthdate.split('T')[0] : '');
	let address1 = $state(user?.address1 ?? '');
	let address2 = $state(user?.address2 ?? '');
	let postalCode = $state(user?.postal_code ?? '');
	let city = $state(user?.city ?? '');

	// Issue 004: "request my data" affordance
	let requestingData = $state(false);

	// Issue 003: delete-account, guarded by a type-to-confirm dialog
	const DELETE_CONFIRM_WORD = 'SUPPRIMER';
	let deleteDialogOpen = $state(false);
	let deleteConfirmInput = $state('');
	let deletingAccount = $state(false);
	let deleteConfirmed = $derived(deleteConfirmInput.trim() === DELETE_CONFIRM_WORD);

	function closeDeleteDialog() {
		deleteDialogOpen = false;
		deleteConfirmInput = '';
	}

	// Issue 013: manage-payment-method link (Stripe Billing Portal)
	let billingPortalLoading = $state(false);
	let billingPortalError = $state<string | null>(null);

	async function openBillingPortal() {
		billingPortalLoading = true;
		billingPortalError = null;
		try {
			const res = await fetch('/api/users/me/stripe/billing-portal-link', { method: 'POST' });
			if (res.ok) {
				const { url } = await res.json();
				window.location.href = url;
			} else {
				billingPortalError = 'Impossible de générer le lien de gestion du paiement. Veuillez réessayer.';
			}
		} catch {
			billingPortalError = 'Une erreur est survenue. Veuillez réessayer.';
		} finally {
			billingPortalLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Mon profil | Leviosa</title>
</svelte:head>

<div class="space-y-8">
	<!-- Header -->
	<div>
		<h1 class="text-3xl lg:text-4xl font-bold text-foreground mb-1">Mon profil</h1>
		<p class="text-muted-foreground">
			{#if user?.profile_incomplete}
				<span class="text-amber-600 font-medium">Profil incomplet</span> — Complétez vos informations personnelles
			{:else}
				Gérez vos informations personnelles
			{/if}
		</p>
	</div>

	{#if form?.success}
		<div class="bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg p-4 flex items-center gap-3">
			<CheckCircle size={18} class="text-green-600 dark:text-green-400 flex-shrink-0" />
			<p class="text-sm text-green-800 dark:text-green-300">Profil mis à jour avec succès.</p>
		</div>
	{/if}

	{#if form?.error}
		<div class="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-center gap-3">
			<p class="text-sm text-red-700 dark:text-red-400">{form.error}</p>
		</div>
	{/if}

	{#if user}
		<form method="POST" action="?/updateProfile" use:enhance class="space-y-8">
			<!-- Personal info -->
			<div class="bg-card rounded-lg border border-border p-6 space-y-6">
				<h2 class="text-lg font-semibold text-foreground">Informations personnelles</h2>

				<!-- Read-only fields -->
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-muted-foreground mb-1">Prénom</label>
						<p class="text-foreground">{user.first_name || '—'}</p>
					</div>
					<div>
						<label class="block text-sm font-medium text-muted-foreground mb-1">Nom</label>
						<p class="text-foreground">{user.last_name || '—'}</p>
					</div>
					<div>
						<label class="block text-sm font-medium text-muted-foreground mb-1">Email</label>
						<p class="text-foreground">{user.email || '—'}</p>
					</div>
					<div>
						<label class="block text-sm font-medium text-muted-foreground mb-1">Téléphone</label>
						<p class="text-foreground">{user.telephone || '—'}</p>
					</div>
				</div>

				<div class="border-t border-border pt-6 space-y-4">
					<!-- Gender -->
					<div>
						<label for="gender" class="block text-sm font-medium text-foreground mb-1.5">
							Genre
							{#if !gender}
								<span class="text-amber-600 text-xs">(requis)</span>
							{/if}
						</label>
						<select
							name="gender"
							id="gender"
							bind:value={gender}
							class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
						>
							<option value="">Sélectionnez votre genre</option>
							<option value="man">Homme</option>
							<option value="woman">Femme</option>
							<option value="non_binary">Non binaire</option>
							<option value="prefer_not_to_say">Je préfère ne pas le dire</option>
							<option value="custom">Je préfère décrire mon genre</option>
						</select>
					</div>

					<!-- Birthdate -->
					<div>
						<label for="birthdate" class="block text-sm font-medium text-foreground mb-1.5">
							Date de naissance
							{#if !birthdate}
								<span class="text-amber-600 text-xs">(requis)</span>
							{/if}
						</label>
						<input
							type="date"
							name="birthdate"
							id="birthdate"
							bind:value={birthdate}
							class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
						/>
					</div>
				</div>
			</div>

			<!-- Address -->
			<div class="bg-card rounded-lg border border-border p-6 space-y-6">
				<h2 class="text-lg font-semibold text-foreground">
					Adresse
					{#if !address1}
						<span class="text-amber-600 text-sm font-normal">(requis)</span>
					{/if}
				</h2>

				<div class="space-y-4">
					<div>
						<label for="address1" class="block text-sm font-medium text-foreground mb-1.5">Adresse</label>
						<input
							type="text"
							name="address1"
							id="address1"
							bind:value={address1}
							placeholder="Numéro et nom de rue"
							class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
						/>
					</div>
					<div>
						<label for="address2" class="block text-sm font-medium text-foreground mb-1.5">Complément d'adresse</label>
						<input
							type="text"
							name="address2"
							id="address2"
							bind:value={address2}
							placeholder="Bâtiment, appartement, etc."
							class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
						/>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label for="postalCode" class="block text-sm font-medium text-foreground mb-1.5">Code postal</label>
							<input
								type="text"
								name="postalCode"
								id="postalCode"
								bind:value={postalCode}
								placeholder="75000"
								class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
							/>
						</div>
						<div>
							<label for="city" class="block text-sm font-medium text-foreground mb-1.5">Ville</label>
							<input
								type="text"
								name="city"
								id="city"
								bind:value={city}
								placeholder="Paris"
								class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-foreground/20"
							/>
						</div>
					</div>
				</div>
			</div>

			<!-- Submit -->
			<div class="flex justify-end">
				<button
					type="submit"
					class="inline-flex items-center gap-2 px-6 py-2.5 rounded-lg bg-foreground text-background text-sm font-medium hover:bg-foreground/90 transition-colors disabled:opacity-50"
				>
					<Save size={16} />
					Enregistrer
				</button>
			</div>
		</form>

		<!-- Payment method -->
		<div class="bg-card rounded-lg border border-border p-6 space-y-4">
			<h2 class="text-lg font-semibold text-foreground">Moyen de paiement</h2>
			<p class="text-sm text-muted-foreground">
				Mettez à jour ou remplacez la carte enregistrée sur votre compte via l'espace sécurisé de
				notre prestataire de paiement.
			</p>
			<button
				type="button"
				onclick={openBillingPortal}
				disabled={billingPortalLoading}
				class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm font-medium hover:bg-muted transition-colors disabled:opacity-50"
			>
				<CreditCard size={16} />
				{billingPortalLoading ? 'Chargement...' : 'Gérer mon moyen de paiement'}
				{#if !billingPortalLoading}
					<ExternalLink size={14} />
				{/if}
			</button>
			{#if billingPortalError}
				<p class="text-sm text-red-700 dark:text-red-400">{billingPortalError}</p>
			{/if}
		</div>
		<!-- RGPD: request my data -->
		<div class="bg-card rounded-lg border border-border p-6 space-y-4">
			<div>
				<h2 class="text-lg font-semibold text-foreground">Mes données personnelles</h2>
				<p class="text-sm text-muted-foreground mt-1">
					Vous pouvez demander une copie de vos données personnelles conformément au RGPD.
					Votre demande sera transmise via votre messagerie et traitée sous un délai d'un mois.
				</p>
			</div>

			{#if form?.dataRequestSent}
				<div class="bg-green-50 dark:bg-green-950 border border-green-200 dark:border-green-800 rounded-lg p-4 flex items-start gap-3">
					<CheckCircle size={18} class="text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" />
					<div class="text-sm text-green-800 dark:text-green-300">
						<p>Votre demande a été envoyée à votre praticien via la messagerie.</p>
						<a href="/client/messages?thread={form.threadId}" class="underline font-medium">
							Voir la conversation
						</a>
					</div>
				</div>
			{:else if form?.dataRequestNoThread}
				<div class="bg-amber-50 dark:bg-amber-950 border border-amber-200 dark:border-amber-800 rounded-lg p-4 text-sm text-amber-800 dark:text-amber-300">
					Vous n'avez pas encore de conversation active avec un praticien. Écrivez-nous
					directement à
					<a href="mailto:{data.companyEmail}" class="underline font-medium">{data.companyEmail}</a>
					pour faire votre demande.
				</div>
			{:else if form?.dataRequestError}
				<div class="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4 text-sm text-red-700 dark:text-red-400">
					{form.dataRequestError}
				</div>
			{/if}

			<form
				method="POST"
				action="?/requestData"
				use:enhance={() => {
					requestingData = true;
					return async ({ update }) => {
						await update();
						requestingData = false;
					};
				}}
			>
				<button
					type="submit"
					disabled={requestingData}
					class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm font-medium hover:bg-muted transition-colors disabled:opacity-50"
				>
					<MessageSquareText size={16} />
					{requestingData ? 'Envoi en cours…' : 'Demander mes données'}
				</button>
			</form>
		</div>

		<!-- Danger zone: delete account -->
		<div class="bg-card rounded-lg border border-destructive/30 p-6 space-y-4">
			<div>
				<h2 class="text-lg font-semibold text-foreground">Zone de danger</h2>
				<p class="text-sm text-muted-foreground mt-1">
					La suppression de votre compte est définitive et irréversible : vos informations
					personnelles, votre historique de réservations et vos messages seront effacés.
				</p>
			</div>

			{#if form?.deleteError}
				<div class="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4 text-sm text-red-700 dark:text-red-400">
					{form.deleteError}
				</div>
			{/if}

			<button
				type="button"
				onclick={() => (deleteDialogOpen = true)}
				class="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-destructive text-white text-sm font-medium hover:bg-destructive/90 focus-visible:ring-2 focus-visible:ring-destructive focus-visible:ring-offset-2 transition-colors"
			>
				Supprimer mon compte
			</button>
		</div>

		<Dialog.Root bind:open={deleteDialogOpen}>
			<Dialog.Portal>
				<Dialog.Overlay
					class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
				/>
				<Dialog.Content
					class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[480px]"
				>
					<div class="flex items-start gap-4">
						<div class="w-10 h-10 rounded-full bg-destructive/10 flex items-center justify-center flex-shrink-0">
							<AlertTriangle size={20} class="text-destructive" />
						</div>
						<div>
							<Dialog.Title class="text-xl font-semibold tracking-tight text-foreground">
								Supprimer votre compte ?
							</Dialog.Title>
							<Dialog.Description class="text-sm text-muted-foreground mt-2">
								Cette action est irréversible. Toutes vos données personnelles, réservations et
								messages seront définitivement supprimés et vous serez déconnecté.
							</Dialog.Description>
						</div>
					</div>

					<form
						method="POST"
						action="?/deleteAccount"
						class="mt-6"
						use:enhance={() => {
							deletingAccount = true;
							return async ({ update }) => {
								await update();
								deletingAccount = false;
							};
						}}
					>
						<label for="delete-confirm" class="block text-sm font-medium text-foreground mb-1.5">
							Tapez <span class="font-mono font-semibold">{DELETE_CONFIRM_WORD}</span> pour confirmer
						</label>
						<input
							type="text"
							id="delete-confirm"
							name="confirm"
							bind:value={deleteConfirmInput}
							autocomplete="off"
							class="w-full px-3 py-2.5 rounded-lg border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-destructive/40"
						/>

						<div class="flex w-full justify-end gap-3 mt-6">
							<button
								type="button"
								onclick={closeDeleteDialog}
								class="inline-flex items-center justify-center px-6 py-2.5 rounded-lg border border-border text-sm font-medium hover:bg-muted transition-colors"
							>
								Annuler
							</button>
							<button
								type="submit"
								disabled={!deleteConfirmed || deletingAccount}
								class="inline-flex items-center justify-center px-6 py-2.5 rounded-lg bg-destructive text-white text-sm font-semibold hover:bg-destructive/90 focus-visible:ring-2 focus-visible:ring-destructive focus-visible:ring-offset-2 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
							>
								{deletingAccount ? 'Suppression…' : 'Supprimer définitivement'}
							</button>
						</div>
					</form>

					<Dialog.Close
						onclick={closeDeleteDialog}
						class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 cursor-pointer"
					>
						<X class="text-foreground size-5" />
						<span class="sr-only">Fermer</span>
					</Dialog.Close>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	{:else}
		<div class="bg-card rounded-lg border border-border p-8 text-center text-muted-foreground">
			Impossible de charger vos informations. Veuillez réessayer.
		</div>
	{/if}
</div>
