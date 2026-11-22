<script lang="ts">
    import type { PageData } from "./$types";
    import { superForm } from "sveltekit-superforms";
    import { Dialog, Button, Label, Separator, Tabs } from "bits-ui";
    import { X, AlertCircle, Building2, Globe, CalendarClock, ShieldCheck, KeyRound } from "@lucide/svelte";
    import { invalidateAll, replaceState } from "$app/navigation";
    import { enhance } from "$app/forms";
    import { page } from "$app/state";
    import PracticeVideo from "./PracticeVideo.svelte";

    let { data }: { data: PageData } = $props();

    // One tab per group of settings, so the page never stacks every section.
    // The open tab is kept in `?tab=` so a reload or a shared link lands on it.
    const tabs = [
        { value: "company", label: "Entreprise", icon: Building2 },
        { value: "practice", label: "Page publique", icon: Globe },
        { value: "booking", label: "Réservation", icon: CalendarClock },
        { value: "access", label: "Accès", icon: KeyRound },
        { value: "security", label: "Sécurité", icon: ShieldCheck },
    ] as const;
    type TabValue = (typeof tabs)[number]["value"];

    function isTab(value: string | null): value is TabValue {
        return tabs.some((tab) => tab.value === value);
    }

    const requestedTab = page.url.searchParams.get("tab");
    let activeTab: string = $state(isTab(requestedTab) ? requestedTab : "company");

    function selectTab(value: string) {
        const url = new URL(page.url.href);
        url.searchParams.set("tab", value);
        replaceState(url, page.state);
    }

    let retrying = $state(false);

    async function retryLoad() {
        retrying = true;
        try {
            await invalidateAll();
        } finally {
            retrying = false;
        }
    }

    // Company Name Form
    const {
        form: companyNameForm,
        enhance: companyNameEnhance,
        delayed: companyNameDelayed,
    } = superForm(data.companyNameForm, {
        resetForm: false,
    });

    // Company Email Form
    const {
        form: companyEmailForm,
        enhance: companyEmailEnhance,
        delayed: companyEmailDelayed,
    } = superForm(data.companyEmailForm, {
        resetForm: false,
    });

    // Company Phone Form
    const {
        form: companyPhoneForm,
        enhance: companyPhoneEnhance,
        delayed: companyPhoneDelayed,
    } = superForm(data.companyPhoneForm, {
        resetForm: false,
    });

    // Company Address Form
    const {
        form: companyAddressForm,
        enhance: companyAddressEnhance,
        delayed: companyAddressDelayed,
    } = superForm(data.companyAddressForm, {
        resetForm: false,
    });

    // Company Instagram Form
    const {
        form: companyInstagramForm,
        enhance: companyInstagramEnhance,
        delayed: companyInstagramDelayed,
    } = superForm(data.companyInstagramForm, {
        resetForm: false,
    });

    // OTP Duration Form
    const {
        form: otpDurationForm,
        enhance: otpDurationEnhance,
        delayed: otpDurationDelayed,
    } = superForm(data.otpDurationForm, {
        resetForm: false,
    });

    // OTP Length Form
    const {
        form: otpLengthForm,
        enhance: otpLengthEnhance,
        delayed: otpLengthDelayed,
    } = superForm(data.otpLengthForm, {
        resetForm: false,
    });

    // OTP Max Attempts Form
    const {
        form: otpMaxAttemptsForm,
        enhance: otpMaxAttemptsEnhance,
        delayed: otpMaxAttemptsDelayed,
    } = superForm(data.otpMaxAttemptsForm, {
        resetForm: false,
    });

    // Access Token Duration Form
    const {
        form: accessTokenDurationForm,
        enhance: accessTokenDurationEnhance,
        delayed: accessTokenDurationDelayed,
    } = superForm(data.accessTokenDurationForm, {
        resetForm: false,
    });

    // Refresh Token Duration Form
    const {
        form: refreshTokenDurationForm,
        enhance: refreshTokenDurationEnhance,
        delayed: refreshTokenDurationDelayed,
    } = superForm(data.refreshTokenDurationForm, {
        resetForm: false,
    });

    // Practice Page Form (page publique du cabinet)
    const {
        form: practicePageForm,
        enhance: practicePageEnhance,
        delayed: practicePageDelayed,
    } = superForm(data.practicePageForm, {
        resetForm: false,
    });

    // Booking Policy Form (délai minimum, horizon de réservation)
    let bookingPolicyDialogOpen = $state(false);
    const {
        form: bookingPolicyForm,
        errors: bookingPolicyErrors,
        enhance: bookingPolicyEnhance,
        delayed: bookingPolicyDelayed,
    } = superForm(data.bookingPolicyForm, {
        resetForm: false,
        onUpdated: ({ form }) => {
            if (form.valid) bookingPolicyDialogOpen = false;
        },
    });

    // Practice Access Form (code de la porte, instructions d'accès). Private:
    // the door code stays masked until "Afficher".
    let practiceAccessDialogOpen = $state(false);
    let showDoorCode = $state(false);
    const {
        form: practiceAccessForm,
        errors: practiceAccessErrors,
        enhance: practiceAccessEnhance,
        delayed: practiceAccessDelayed,
    } = superForm(data.practiceAccessForm, {
        resetForm: false,
        onUpdated: ({ form }) => {
            if (form.valid) {
                practiceAccessDialogOpen = false;
                accessNotice = typeof form.message === "string" ? form.message : null;
            }
        },
    });

    // Access update: the upcoming confirmed bookings with a client email the
    // door code can be sent to, and the result of the last save or send.
    let accessRecipients = $derived(data.accessUpdate?.recipients ?? 0);
    let accessNotice: string | null = $state(null);
    let accessSending = $state(false);
    let doorCodeChanged = $derived(
        $practiceAccessForm.door_code.trim() !== (data.practiceAccess?.doorCode ?? ""),
    );

    function upcomingClients(n: number): string {
        return n === 1 ? "au client" : `aux ${n} clients`;
    }

    function formatWeeks(weeks: number): string {
        return weeks === 1 ? "1 semaine" : `${weeks} semaines`;
    }

    // Dialog states
    let companyNameDialogOpen = $state(false);
    let companyEmailDialogOpen = $state(false);
    let companyPhoneDialogOpen = $state(false);
    let companyAddressDialogOpen = $state(false);
    let companyInstagramDialogOpen = $state(false);
    let otpDurationDialogOpen = $state(false);
    let otpLengthDialogOpen = $state(false);
    let otpMaxAttemptsDialogOpen = $state(false);
    let accessTokenDurationDialogOpen = $state(false);
    let refreshTokenDurationDialogOpen = $state(false);
    let practicePageDialogOpen = $state(false);

    // Labels of the partner and product currently on the public practice page.
    const practicePartnerLabel = $derived(
        data.partnerOptions.find((option) => option.id === data.practicePage?.partnerId)?.label ?? "Non défini",
    );
    const practiceProductLabel = $derived(
        data.productOptions.find((option) => option.id === data.practicePage?.productId)?.label ?? "Non défini",
    );

    // Helper function to format duration in seconds to human-readable
    function formatSeconds(seconds: number): string {
        if (seconds < 60) return `${seconds}s`;
        const minutes = Math.floor(seconds / 60);
        return `${minutes} min`;
    }

    // Helper function to format duration in minutes to human-readable
    function formatMinutes(minutes: number): string {
        if (minutes < 60) return `${minutes} min`;
        const hours = Math.floor(minutes / 60);
        return `${hours}h`;
    }

    // Helper function to format duration in hours to human-readable
    function formatHours(hours: number): string {
        if (hours < 24) return `${hours}h`;
        const days = Math.floor(hours / 24);
        return `${days} jours`;
    }
</script>

{#if data.settingsError}
    <div class="flex-1 overflow-y-auto p-6">
        <div class="max-w-4xl mx-auto">
            <div class="bg-red-50 border border-red-200 rounded-lg p-6">
                <div class="flex items-center gap-3">
                    <AlertCircle class="text-red-600 shrink-0" size={20} />
                    <div>
                        <h2 class="font-semibold text-red-900">Erreur de chargement</h2>
                        <p class="text-sm text-red-700 mt-1">
                            Les paramètres n'ont pas pu être chargés. Vérifiez votre connexion et réessayez.
                        </p>
                        <button
                            onclick={retryLoad}
                            disabled={retrying}
                            class="mt-3 h-input rounded-input border border-red-300 bg-white text-red-700 hover:bg-red-50 inline-flex items-center justify-center px-4 text-sm font-medium active:scale-[0.98] cursor-pointer transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {retrying ? "Chargement..." : "Réessayer"}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
{:else}
<div class="flex-1 overflow-y-auto p-6">
    <div class="max-w-4xl mx-auto">
    <Tabs.Root bind:value={activeTab} onValueChange={selectTab} class="space-y-6">
        <Tabs.List
            class="flex items-center w-fit max-w-full overflow-x-auto bg-transparent gap-2 text-sm font-semibold border-b border-border-card p-1"
        >
            {#each tabs as tab (tab.value)}
                <Tabs.Trigger
                    value={tab.value}
                    class="px-4 py-2 rounded-none bg-transparent border-b-2 data-[state=active]:shadow-none mb-[-2px] data-[state=active]:border-b-foreground data-[state=active]:text-foreground data-[state=inactive]:border-transparent data-[state=inactive]:text-muted-foreground data-[state=inactive]:hover:bg-transparent data-[state=inactive]:hover:text-foreground-alt transition-colors cursor-pointer flex items-center gap-2 whitespace-nowrap shrink-0"
                >
                    <tab.icon size={16} />
                    <span>{tab.label}</span>
                </Tabs.Trigger>
            {/each}
        </Tabs.List>

        <Tabs.Content value="company">
        <!-- Company Information Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card">
                <h2 class="text-lg font-semibold">Informations de l'entreprise</h2>
                <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                    Gérez les informations de contact et les réseaux sociaux
                </p>
            </div>
            <div class="p-6 space-y-4">
                <!-- Company Name -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Nom de l'entreprise</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.company?.name || "Non défini"}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $companyNameForm.name = data.settings?.company?.name || "";
                            companyNameDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Company Email -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Email de contact</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.company?.email || "Non défini"}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $companyEmailForm.email = data.settings?.company?.email || "";
                            companyEmailDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Company Phone -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Téléphone</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.company?.telephone || "Non défini"}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $companyPhoneForm.telephone = data.settings?.company?.telephone || "";
                            companyPhoneDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Company Address -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Adresse</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.company?.address || "Non définie"}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $companyAddressForm.address = data.settings?.company?.address || "";
                            companyAddressDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Company Instagram -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Instagram</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.company?.instagram || "Non défini"}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $companyInstagramForm.instagram = data.settings?.company?.instagram || "";
                            companyInstagramDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>
            </div>
        </div>
        </Tabs.Content>

        <Tabs.Content value="practice">
        <!-- Practice Page Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card">
                <h2 class="text-lg font-semibold">Page publique</h2>
                <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                    Le partenaire et le produit présentés sur la page du cabinet — un changement prend effet immédiatement, sans déploiement
                </p>
            </div>
            <div class="p-6 space-y-4">
                <!-- Partner shown -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Partenaire mis en avant</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {practicePartnerLabel}
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $practicePageForm.partner_id = data.practicePage?.partnerId || "";
                            $practicePageForm.product_id = data.practicePage?.productId || "";
                            practicePageDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Product shown -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Produit proposé</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {practiceProductLabel}
                        </div>
                    </div>
                </div>

                <!-- Video -->
                <PracticeVideo videoUrl={data.practicePage?.videoUrl ?? ""} />
            </div>
        </div>
        </Tabs.Content>

        <Tabs.Content value="booking">
        <!-- Booking Policy Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card">
                <h2 class="text-lg font-semibold">Réservation</h2>
                <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                    Les créneaux proposés sur la page publique et acceptés à la réservation — un changement prend effet immédiatement, sans déploiement
                </p>
            </div>
            <div class="p-6 space-y-4">
                {#if data.bookingPolicy}
                    <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                        <div class="flex-1 min-w-0">
                            <div class="font-medium text-sm">Délai minimum avant une séance</div>
                            <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                                {data.bookingPolicy.minimumNoticeHours} h
                            </div>
                        </div>
                        <Button.Root
                            type="button"
                            class="cursor-pointer shrink-0"
                            onclick={() => {
                                $bookingPolicyForm.minimum_notice_hours = data.bookingPolicy?.minimumNoticeHours ?? 12;
                                $bookingPolicyForm.horizon_weeks = data.bookingPolicy?.horizonWeeks ?? 8;
                                $bookingPolicyForm.payment_hold_minutes = data.bookingPolicy?.paymentHoldMinutes ?? 15;
                                bookingPolicyDialogOpen = true;
                            }}
                        >
                            <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                                Modifier
                            </div>
                        </Button.Root>
                    </div>
                    <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                        <div class="flex-1 min-w-0">
                            <div class="font-medium text-sm">Horizon de réservation</div>
                            <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                                {formatWeeks(data.bookingPolicy.horizonWeeks)}
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                        <div class="flex-1 min-w-0">
                            <div class="font-medium text-sm">Maintien du créneau pendant le paiement</div>
                            <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                                {data.bookingPolicy.paymentHoldMinutes} min
                            </div>
                        </div>
                    </div>
                {:else}
                    <p class="text-sm text-foreground-alt">Impossible de charger la politique de réservation.</p>
                {/if}
            </div>
        </div>
        </Tabs.Content>

        <Tabs.Content value="access">
        <!-- Practice Access Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card flex items-start justify-between gap-4">
                <div class="min-w-0">
                    <h2 class="text-lg font-semibold">Accès au cabinet</h2>
                    <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                        Ces informations sont privées : elles ne sont envoyées qu'aux clients ayant une réservation confirmée
                    </p>
                </div>
                {#if data.practiceAccess}
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $practiceAccessForm.door_code = data.practiceAccess?.doorCode ?? "";
                            $practiceAccessForm.access_instructions = data.practiceAccess?.accessInstructions ?? "";
                            $practiceAccessForm.notify_clients = true;
                            practiceAccessDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                {/if}
            </div>
            <div class="p-6 space-y-4">
                <div class="py-3 border-b border-border-card last:border-0">
                    <div class="font-medium text-sm">Adresse du cabinet</div>
                    <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                        {data.settings?.company?.address || "Non définie"}
                    </div>
                    <p class="text-xs text-foreground-alt mt-1">Se modifie dans l'onglet Entreprise.</p>
                </div>
                {#if data.practiceAccess}
                    <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                        <div class="flex-1 min-w-0">
                            <div class="font-medium text-sm">Code d'accès</div>
                            <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                                {#if !data.practiceAccess.doorCode}
                                    Non défini
                                {:else if showDoorCode}
                                    <span class="font-mono">{data.practiceAccess.doorCode}</span>
                                {:else}
                                    <span aria-label="Code masqué">••••••</span>
                                {/if}
                            </div>
                        </div>
                        {#if data.practiceAccess.doorCode}
                            <Button.Root
                                type="button"
                                class="cursor-pointer shrink-0"
                                aria-pressed={showDoorCode}
                                onclick={() => (showDoorCode = !showDoorCode)}
                            >
                                <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                                    {showDoorCode ? "Masquer" : "Afficher"}
                                </div>
                            </Button.Root>
                        {/if}
                    </div>
                    <div class="py-3 border-b border-border-card last:border-0">
                        <div class="font-medium text-sm">Instructions d'accès</div>
                        <div class="text-sm text-foreground-alt mt-1 whitespace-pre-line [overflow-wrap:anywhere]">
                            {data.practiceAccess.accessInstructions || "Non définies"}
                        </div>
                    </div>
                    {#if data.practiceAccess.doorCode && accessRecipients > 0}
                        <form
                            method="POST"
                            action="?/sendPracticeAccess"
                            class="pt-1"
                            use:enhance={() => {
                                accessSending = true;
                                accessNotice = null;
                                return async ({ result }) => {
                                    accessSending = false;
                                    if (result.type === "success" || result.type === "failure") {
                                        const text = result.data?.accessSendMessage;
                                        accessNotice = typeof text === "string" ? text : null;
                                    } else {
                                        accessNotice = "L'accès n'a pas pu être envoyé. Réessayez.";
                                    }
                                    await invalidateAll();
                                };
                            }}
                        >
                            <button
                                type="submit"
                                class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all disabled:cursor-not-allowed disabled:opacity-50"
                                disabled={accessSending}
                            >
                                {accessSending ? "Envoi..." : `Envoyer l'accès ${upcomingClients(accessRecipients)} à venir`}
                            </button>
                        </form>
                    {/if}
                {:else}
                    <p class="text-sm text-foreground-alt">Impossible de charger l'accès au cabinet.</p>
                {/if}
                {#if accessNotice}
                    <p class="text-sm text-foreground-alt [overflow-wrap:anywhere]" role="status">{accessNotice}</p>
                {/if}
            </div>
        </div>
        </Tabs.Content>

        <Tabs.Content value="security" class="space-y-6">
        <!-- OTP Configuration Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card">
                <h2 class="text-lg font-semibold">Configuration OTP</h2>
                <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                    Paramètres de sécurité pour les codes de vérification
                </p>
            </div>
            <div class="p-6 space-y-4">
                <!-- OTP Duration -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Durée de validité OTP</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {formatSeconds(data.settings?.otp?.duration || 600)} (60-3600 secondes)
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $otpDurationForm.duration = data.settings?.otp?.duration || 600;
                            otpDurationDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- OTP Length -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Longueur du code OTP</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.otp?.length || 6} chiffres (4-10 chiffres)
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $otpLengthForm.length = data.settings?.otp?.length || 6;
                            otpLengthDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- OTP Max Attempts -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Tentatives maximales</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {data.settings?.otp?.max_attempts || 5} tentatives (1-10)
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $otpMaxAttemptsForm.max_attempts = data.settings?.otp?.max_attempts || 5;
                            otpMaxAttemptsDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>
            </div>
        </div>

        <!-- Token Configuration Section -->
        <div class="bg-white rounded-lg border border-border-card overflow-hidden">
            <div class="px-6 py-4 border-b border-border-card">
                <h2 class="text-lg font-semibold">Configuration des tokens</h2>
                <p class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                    Paramètres de durée de vie des tokens d'authentification
                </p>
            </div>
            <div class="p-6 space-y-4">
                <!-- Access Token Duration -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Durée du token d'accès</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {formatMinutes(data.settings?.tokens?.access_duration || 15)} (1-240 minutes)
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $accessTokenDurationForm.duration = data.settings?.tokens?.access_duration || 15;
                            accessTokenDurationDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>

                <!-- Refresh Token Duration -->
                <div class="flex items-center justify-between gap-4 py-3 border-b border-border-card last:border-0">
                    <div class="flex-1 min-w-0">
                        <div class="font-medium text-sm">Durée du token de rafraîchissement</div>
                        <div class="text-sm text-foreground-alt mt-1 [overflow-wrap:anywhere]">
                            {formatHours(data.settings?.tokens?.refresh_duration || 168)} (1-720 heures)
                        </div>
                    </div>
                    <Button.Root
                        type="button"
                        class="cursor-pointer shrink-0"
                        onclick={() => {
                            $refreshTokenDurationForm.duration = data.settings?.tokens?.refresh_duration || 168;
                            refreshTokenDurationDialogOpen = true;
                        }}
                    >
                        <div class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Modifier
                        </div>
                    </Button.Root>
                </div>
            </div>
        </div>
        </Tabs.Content>
    </Tabs.Root>
    </div>
</div>

<!-- Company Name Dialog -->
<Dialog.Root bind:open={companyNameDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier le nom de l'entreprise
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez le nom officiel de votre entreprise
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateCompanyName" use:companyNameEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="company-name" class="text-sm font-medium">Nom</Label.Root>
                    <input
                        id="company-name"
                        name="name"
                        type="text"
                        bind:value={$companyNameForm.name}
                        placeholder="Leviosa Spa & Wellness"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$companyNameDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$companyNameDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Company Email Dialog -->
<Dialog.Root bind:open={companyEmailDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier l'email de contact
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez l'adresse email de contact de votre entreprise
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateCompanyEmail" use:companyEmailEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="company-email" class="text-sm font-medium">Email</Label.Root>
                    <input
                        id="company-email"
                        name="email"
                        type="email"
                        bind:value={$companyEmailForm.email}
                        placeholder="contact@leviosa-spa.com"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$companyEmailDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$companyEmailDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Company Phone Dialog -->
<Dialog.Root bind:open={companyPhoneDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier le téléphone
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez le numéro de téléphone de votre entreprise
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateCompanyPhone" use:companyPhoneEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="company-phone" class="text-sm font-medium">Téléphone</Label.Root>
                    <input
                        id="company-phone"
                        name="telephone"
                        type="tel"
                        bind:value={$companyPhoneForm.telephone}
                        placeholder="+33123456789"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$companyPhoneDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$companyPhoneDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Company Address Dialog -->
<Dialog.Root bind:open={companyAddressDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier l'adresse
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez l'adresse légale de votre entreprise
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateCompanyAddress" use:companyAddressEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="company-address" class="text-sm font-medium">Adresse</Label.Root>
                    <input
                        id="company-address"
                        name="address"
                        type="text"
                        bind:value={$companyAddressForm.address}
                        placeholder="123 Wellness Boulevard, 75001 Paris"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$companyAddressDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$companyAddressDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Company Instagram Dialog -->
<Dialog.Root bind:open={companyInstagramDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier le lien Instagram
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez l'URL de votre profil Instagram
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateCompanyInstagram" use:companyInstagramEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="company-instagram" class="text-sm font-medium">URL Instagram</Label.Root>
                    <input
                        id="company-instagram"
                        name="instagram"
                        type="url"
                        bind:value={$companyInstagramForm.instagram}
                        placeholder="https://instagram.com/leviosa_spa"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$companyInstagramDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$companyInstagramDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- OTP Duration Dialog -->
<Dialog.Root bind:open={otpDurationDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la durée de validité OTP
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez combien de temps un code OTP reste valide (en secondes)
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateOtpDuration" use:otpDurationEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="otp-duration" class="text-sm font-medium">Durée (secondes)</Label.Root>
                    <input
                        id="otp-duration"
                        name="duration"
                        type="number"
                        min="60"
                        max="3600"
                        bind:value={$otpDurationForm.duration}
                        placeholder="300"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">
                        Min: 60 secondes, Max: 3600 secondes
                    </p>
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$otpDurationDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$otpDurationDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- OTP Length Dialog -->
<Dialog.Root bind:open={otpLengthDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la longueur du code OTP
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez le nombre de chiffres dans le code OTP
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateOtpLength" use:otpLengthEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="otp-length" class="text-sm font-medium">Longueur (chiffres)</Label.Root>
                    <input
                        id="otp-length"
                        name="length"
                        type="number"
                        min="4"
                        max="10"
                        bind:value={$otpLengthForm.length}
                        placeholder="6"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">
                        Min: 4 chiffres, Max: 10 chiffres
                    </p>
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$otpLengthDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$otpLengthDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- OTP Max Attempts Dialog -->
<Dialog.Root bind:open={otpMaxAttemptsDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier les tentatives maximales
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez le nombre de tentatives avant le verrouillage du compte
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateOtpMaxAttempts" use:otpMaxAttemptsEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="otp-max-attempts" class="text-sm font-medium">Tentatives maximales</Label.Root>
                    <input
                        id="otp-max-attempts"
                        name="max_attempts"
                        type="number"
                        min="1"
                        max="10"
                        bind:value={$otpMaxAttemptsForm.max_attempts}
                        placeholder="5"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">
                        Min: 1 tentative, Max: 10 tentatives
                    </p>
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$otpMaxAttemptsDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$otpMaxAttemptsDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Access Token Duration Dialog -->
<Dialog.Root bind:open={accessTokenDurationDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la durée du token d'accès
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez combien de temps un token d'accès reste valide (en minutes)
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateAccessTokenDuration" use:accessTokenDurationEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="access-token-duration" class="text-sm font-medium">Durée (minutes)</Label.Root>
                    <input
                        id="access-token-duration"
                        name="duration"
                        type="number"
                        min="1"
                        max="240"
                        bind:value={$accessTokenDurationForm.duration}
                        placeholder="15"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">
                        Min: 1 minute, Max: 240 minutes (4 heures)
                    </p>
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$accessTokenDurationDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$accessTokenDurationDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Refresh Token Duration Dialog -->
<Dialog.Root bind:open={refreshTokenDurationDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la durée du token de rafraîchissement
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Définissez combien de temps un token de rafraîchissement reste valide (en heures)
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateRefreshTokenDuration" use:refreshTokenDurationEnhance class="grid gap-4">
                <div class="grid gap-2 py-4">
                    <Label.Root for="refresh-token-duration" class="text-sm font-medium">Durée (heures)</Label.Root>
                    <input
                        id="refresh-token-duration"
                        name="duration"
                        type="number"
                        min="1"
                        max="720"
                        bind:value={$refreshTokenDurationForm.duration}
                        placeholder="168"
                        class="h-input rounded-card-sm border-border-input bg-background placeholder:text-foreground-alt/50 hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">
                        Min: 1 heure, Max: 720 heures (30 jours)
                    </p>
                </div>
                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$refreshTokenDurationDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$refreshTokenDurationDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>
{/if}

<!-- Booking Policy Dialog -->
<Dialog.Root bind:open={bookingPolicyDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[490px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la politique de réservation
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Aucun créneau n'est proposé avant le délai minimum, ni au-delà de l'horizon. Un créneau reste réservé au client le temps du paiement
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updateBookingPolicy" use:bookingPolicyEnhance class="grid gap-4">
                <div class="grid gap-2 py-2">
                    <Label.Root for="booking-minimum-notice" class="text-sm font-medium">Délai minimum (heures)</Label.Root>
                    <input
                        id="booking-minimum-notice"
                        name="minimum_notice_hours"
                        type="number"
                        min="0"
                        max="168"
                        step="1"
                        bind:value={$bookingPolicyForm.minimum_notice_hours}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">De 0 à 168 h (une semaine). Par défaut : 12 h.</p>
                </div>

                <div class="grid gap-2 py-2">
                    <Label.Root for="booking-horizon" class="text-sm font-medium">Horizon (semaines)</Label.Root>
                    <input
                        id="booking-horizon"
                        name="horizon_weeks"
                        type="number"
                        min="1"
                        max="52"
                        step="1"
                        bind:value={$bookingPolicyForm.horizon_weeks}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">De 1 à 52 semaines. Par défaut : 8 semaines.</p>
                </div>

                <div class="grid gap-2 py-2">
                    <Label.Root for="booking-payment-hold" class="text-sm font-medium">Maintien pendant le paiement (minutes)</Label.Root>
                    <input
                        id="booking-payment-hold"
                        name="payment_hold_minutes"
                        type="number"
                        min="5"
                        max="60"
                        step="1"
                        bind:value={$bookingPolicyForm.payment_hold_minutes}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">De 5 à 60 min. Par défaut : 15 min. Passé ce délai sans paiement, le créneau est proposé à nouveau.</p>
                </div>

                {#if $bookingPolicyErrors._errors?.length}
                    <p class="text-sm text-destructive" role="alert">{$bookingPolicyErrors._errors[0]}</p>
                {/if}

                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$bookingPolicyDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$bookingPolicyDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Practice Access Dialog -->
<Dialog.Root bind:open={practiceAccessDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier l'accès au cabinet
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Informations privées, envoyées uniquement aux clients ayant une réservation confirmée. Laissez un champ vide pour l'effacer
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updatePracticeAccess" use:practiceAccessEnhance class="grid gap-4">
                <div class="grid gap-2 py-2">
                    <Label.Root for="practice-access-door-code" class="text-sm font-medium">Code d'accès</Label.Root>
                    <input
                        id="practice-access-door-code"
                        name="door_code"
                        type="text"
                        maxlength="50"
                        autocomplete="off"
                        spellcheck="false"
                        bind:value={$practiceAccessForm.door_code}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    />
                    <p class="text-xs text-foreground-alt">50 caractères au plus.</p>
                    {#if $practiceAccessErrors.door_code}
                        <p class="text-sm text-destructive" role="alert">{$practiceAccessErrors.door_code[0]}</p>
                    {/if}
                </div>

                <div class="grid gap-2 py-2">
                    <Label.Root for="practice-access-instructions" class="text-sm font-medium">Instructions d'accès</Label.Root>
                    <textarea
                        id="practice-access-instructions"
                        name="access_instructions"
                        rows="5"
                        maxlength="1000"
                        bind:value={$practiceAccessForm.access_instructions}
                        class="rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden w-full border px-4 py-3 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    ></textarea>
                    <p class="text-xs text-foreground-alt">Bâtiment, étage, interphone… 1000 caractères au plus.</p>
                    {#if $practiceAccessErrors.access_instructions}
                        <p class="text-sm text-destructive" role="alert">{$practiceAccessErrors.access_instructions[0]}</p>
                    {/if}
                </div>

                {#if doorCodeChanged && accessRecipients > 0}
                    <label class="flex items-start gap-3 text-sm">
                        <input
                            type="checkbox"
                            name="notify_clients"
                            bind:checked={$practiceAccessForm.notify_clients}
                            class="mt-0.5 size-4 cursor-pointer accent-dark"
                        />
                        <span>Envoyer le nouveau code {upcomingClients(accessRecipients)} ayant un rendez-vous à venir</span>
                    </label>
                {/if}

                {#if $practiceAccessErrors._errors?.length}
                    <p class="text-sm text-destructive" role="alert">{$practiceAccessErrors._errors[0]}</p>
                {/if}

                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$practiceAccessDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$practiceAccessDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>

<!-- Practice Page Dialog -->
{#if !data.settingsError}
<Dialog.Root bind:open={practicePageDialogOpen}>
    <Dialog.Portal>
        <Dialog.Overlay class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80" />
        <Dialog.Content class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full">
            <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                Modifier la page publique
            </Dialog.Title>
            <Dialog.Description class="text-foreground-alt !mt-1 text-sm">
                Choisissez le partenaire et le produit présentés sur la page du cabinet
            </Dialog.Description>

            <Separator.Root class="bg-muted mx-5 !mb-2 !mt-5 block h-px" />

            <form method="POST" action="?/updatePracticePage" use:practicePageEnhance class="grid gap-4">
                <div class="grid gap-2 py-2">
                    <Label.Root for="practice-partner" class="text-sm font-medium">Partenaire</Label.Root>
                    <select
                        id="practice-partner"
                        name="partner_id"
                        bind:value={$practicePageForm.partner_id}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    >
                        {#if !$practicePageForm.partner_id}
                            <option value="" disabled>Sélectionner un partenaire</option>
                        {/if}
                        {#each data.partnerOptions as option (option.id)}
                            <option value={option.id}>{option.label}</option>
                        {/each}
                    </select>
                </div>

                <div class="grid gap-2 py-2">
                    <Label.Root for="practice-product" class="text-sm font-medium">Produit</Label.Root>
                    <select
                        id="practice-product"
                        name="product_id"
                        bind:value={$practicePageForm.product_id}
                        class="h-input rounded-card-sm border-border-input bg-background hover:border-dark-40 focus:ring-foreground focus:ring-offset-background focus:outline-hidden inline-flex w-full items-center border px-4 text-base focus:ring-2 focus:ring-offset-2 sm:text-sm transition-all"
                    >
                        {#if !$practicePageForm.product_id}
                            <option value="" disabled>Sélectionner un produit</option>
                        {/if}
                        {#each data.productOptions as option (option.id)}
                            <option value={option.id}>{option.label}</option>
                        {/each}
                    </select>
                </div>

                <div class="flex w-full justify-end gap-3">
                    <Button.Root type="button" class="cursor-pointer">
                        <Dialog.Close class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-6 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            Annuler
                        </Dialog.Close>
                    </Button.Root>
                    <Button.Root type="submit" class="cursor-pointer" disabled={$practicePageDelayed}>
                        <div class="h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center px-8 text-sm font-semibold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all">
                            {$practicePageDelayed ? "Enregistrement..." : "Enregistrer"}
                        </div>
                    </Button.Root>
                </div>
            </form>

            <Button.Root type="button" class="cursor-pointer">
                <Dialog.Close class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer">
                    <X class="text-foreground size-5" />
                    <span class="sr-only">Close</span>
                </Dialog.Close>
            </Button.Root>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>
{/if}
