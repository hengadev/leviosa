<script lang="ts">
    import { reveal } from "$lib/actions/reveal";
    import { enhance } from "$app/forms";
    import type { PageProps } from "./$types";
    import {
        Mail,
        MapPin,
        Instagram,
        Send,
        AlertCircle,
        CheckCircle2,
        Loader2,
    } from "@lucide/svelte";

    let { data, form }: PageProps = $props();

    let submitting = $state(false);

    let values = $derived(
        form?.values ?? { name: "", contactEmail: "", contactPhone: "", message: "" },
    );
    let errors = $derived(form?.errors ?? ({} as Record<string, string>));
    let success = $derived(form?.success === true);
</script>

<svelte:head>
    <title>Contact | Leviosa</title>
    <meta
        name="description"
        content="Une question avant de réserver ? Écrivez à l'équipe Leviosa, on vous répond directement."
    />
</svelte:head>

<div class="min-h-screen bg-surface py-24 md:py-32 px-4 lg:px-8">
    <div class="max-w-5xl mx-auto" use:reveal={{ preset: "fade-up", delay: 100 }}>
        <div class="text-center mb-12 max-w-2xl mx-auto">
            <h1 class="text-3xl md:text-4xl font-bold text-foreground mb-3">
                Une question ? Écrivez-nous.
            </h1>
            <p class="text-foreground-alt">
                Pas encore prêt·e à réserver mais besoin d'un renseignement ? Ce
                message va droit dans la boîte de l'équipe.
            </p>
        </div>

        <div class="grid lg:grid-cols-[1fr_320px] gap-6 items-start">
            <!-- Form / success -->
            <div class="bg-white rounded-3xl p-6 md:p-8 shadow-mini">
                {#if success}
                    <div class="text-center py-10">
                        <div
                            class="w-14 h-14 rounded-full bg-green-50 border border-green-200 flex items-center justify-center mx-auto mb-5"
                        >
                            <CheckCircle2 class="text-green-600" size={28} />
                        </div>
                        <h2 class="text-xl font-semibold text-foreground mb-2">
                            Message envoyé
                        </h2>
                        <p class="text-foreground-alt max-w-sm mx-auto">
                            Merci, votre message est bien parti. L'équipe Leviosa
                            vous répond dès que possible.
                        </p>
                        <a
                            href="/"
                            class="inline-block mt-6 text-sm font-medium text-foreground underline underline-offset-4 hover:no-underline"
                        >
                            Retour à l'accueil
                        </a>
                    </div>
                {:else}
                    {#if errors._form}
                        <div
                            class="flex items-center gap-2 px-4 py-3 mb-6 rounded-lg bg-red-50 border border-red-200 text-red-700 text-sm"
                            role="alert"
                        >
                            <AlertCircle size={16} class="flex-shrink-0" />
                            {errors._form}
                        </div>
                    {/if}

                    <form
                        method="POST"
                        class="space-y-5"
                        use:enhance={() => {
                            submitting = true;
                            return async ({ update }) => {
                                await update();
                                submitting = false;
                            };
                        }}
                    >
                        <div>
                            <label
                                for="name"
                                class="block text-sm font-medium text-foreground-alt mb-1.5"
                                >Nom</label
                            >
                            <input
                                id="name"
                                name="name"
                                type="text"
                                value={values.name}
                                placeholder="Jean Dupont"
                                maxlength="200"
                                required
                                aria-invalid={errors.name ? "true" : undefined}
                                aria-describedby={errors.name ? "name-error" : undefined}
                                class="w-full px-4 py-3 rounded-xl border {errors.name
                                    ? 'border-red-300'
                                    : 'border-border-input-hover'} text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent text-sm"
                            />
                            {#if errors.name}
                                <p id="name-error" class="text-red-500 text-sm mt-1.5">
                                    {errors.name}
                                </p>
                            {/if}
                        </div>

                        <div class="grid sm:grid-cols-2 gap-5">
                            <div>
                                <label
                                    for="contact_email"
                                    class="block text-sm font-medium text-foreground-alt mb-1.5"
                                    >Email</label
                                >
                                <input
                                    id="contact_email"
                                    name="contact_email"
                                    type="email"
                                    value={values.contactEmail}
                                    placeholder="jean@exemple.com"
                                    aria-invalid={errors.contact_email ? "true" : undefined}
                                    aria-describedby={errors.contact_email
                                        ? "email-error"
                                        : undefined}
                                    class="w-full px-4 py-3 rounded-xl border {errors.contact_email
                                        ? 'border-red-300'
                                        : 'border-border-input-hover'} text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent text-sm"
                                />
                                {#if errors.contact_email}
                                    <p id="email-error" class="text-red-500 text-sm mt-1.5">
                                        {errors.contact_email}
                                    </p>
                                {/if}
                            </div>
                            <div>
                                <label
                                    for="contact_phone"
                                    class="block text-sm font-medium text-foreground-alt mb-1.5"
                                    >Téléphone</label
                                >
                                <input
                                    id="contact_phone"
                                    name="contact_phone"
                                    type="tel"
                                    value={values.contactPhone}
                                    placeholder="06 12 34 56 78"
                                    aria-invalid={errors.contact_phone ? "true" : undefined}
                                    aria-describedby={errors.contact_phone
                                        ? "phone-error"
                                        : undefined}
                                    class="w-full px-4 py-3 rounded-xl border {errors.contact_phone
                                        ? 'border-red-300'
                                        : 'border-border-input-hover'} text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent text-sm"
                                />
                                {#if errors.contact_phone}
                                    <p id="phone-error" class="text-red-500 text-sm mt-1.5">
                                        {errors.contact_phone}
                                    </p>
                                {/if}
                            </div>
                        </div>
                        <p class="text-xs text-muted-foreground -mt-2.5">
                            Email ou téléphone requis pour qu'on puisse vous répondre.
                        </p>

                        <div>
                            <label
                                for="message"
                                class="block text-sm font-medium text-foreground-alt mb-1.5"
                                >Message</label
                            >
                            <textarea
                                id="message"
                                name="message"
                                rows="5"
                                placeholder="Comment pouvons-nous vous aider ?"
                                maxlength="5000"
                                required
                                aria-invalid={errors.message ? "true" : undefined}
                                aria-describedby={errors.message ? "message-error" : undefined}
                                class="w-full px-4 py-3 rounded-xl border {errors.message
                                    ? 'border-red-300'
                                    : 'border-border-input-hover'} text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-foreground focus:border-transparent text-sm resize-y"
                                >{values.message}</textarea
                            >
                            {#if errors.message}
                                <p id="message-error" class="text-red-500 text-sm mt-1.5">
                                    {errors.message}
                                </p>
                            {/if}
                        </div>

                        <button
                            type="submit"
                            disabled={submitting}
                            class="w-full flex items-center justify-center gap-2 bg-foreground hover:bg-foreground-alt text-white py-3 rounded-xl font-medium transition-colors disabled:opacity-50 cursor-pointer"
                        >
                            {#if submitting}
                                <Loader2 size={18} class="animate-spin" />
                                Envoi...
                            {:else}
                                <Send size={18} />
                                Envoyer le message
                            {/if}
                        </button>
                    </form>
                {/if}
            </div>

            <!-- Other ways to reach us -->
            <div class="bg-foreground text-white rounded-3xl p-6 md:p-8">
                <h2 class="text-lg font-semibold mb-5">Nous joindre autrement</h2>
                <div class="flex flex-col gap-4">
                    {#if data.companyEmail}
                        <a
                            href="mailto:{data.companyEmail}"
                            class="flex items-start gap-3 text-white/70 hover:text-white transition-colors"
                        >
                            <Mail size={18} class="mt-0.5 flex-shrink-0" />
                            <span class="text-sm break-all">{data.companyEmail}</span>
                        </a>
                    {/if}
                    {#if data.companyAddress}
                        <div class="flex items-start gap-3 text-white/70">
                            <MapPin size={18} class="mt-0.5 flex-shrink-0" />
                            <span class="text-sm">{data.companyAddress}</span>
                        </div>
                    {/if}
                    {#if data.companyInstagram}
                        <a
                            href={data.companyInstagram}
                            target="_blank"
                            rel="noopener noreferrer"
                            class="flex items-start gap-3 text-white/70 hover:text-white transition-colors"
                        >
                            <Instagram size={18} class="mt-0.5 flex-shrink-0" />
                            <span class="text-sm">Instagram</span>
                        </a>
                    {/if}
                </div>
                <div class="mt-6 pt-6 border-t border-white/10">
                    <a
                        href="/book"
                        class="text-sm font-medium text-white underline underline-offset-4 hover:no-underline"
                    >
                        Prêt·e à réserver directement ?
                    </a>
                </div>
            </div>
        </div>
    </div>
</div>
