<!--
	@component
	Shown on the practitioner pages (Aujourd'hui, Réservations, Disponibilités)
	when the administrator has no active practitioner profile of their own yet.
-->
<script lang="ts">
	import { UserRoundX } from "@lucide/svelte";
</script>

<div class="rounded-2xl border border-border-card p-6 text-center">
	<UserRoundX size={28} class="mx-auto mb-3 text-muted-foreground" />
	<p class="font-medium text-foreground">Votre profil praticien n'est pas encore configuré</p>
	<p class="mt-1 text-sm text-muted-foreground">
		Activez-le dans Utilisateurs, onglet Partenaires, avec « Je suis aussi praticien ». Vos séances apparaîtront ici.
	</p>
	<a
		href="/admin/users"
		class="mt-4 inline-flex h-11 items-center justify-center rounded-xl border border-border-card px-4 text-sm font-medium hover:bg-muted"
	>
		Configurer mon profil
	</a>
</div>
