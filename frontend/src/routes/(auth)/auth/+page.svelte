<script lang="ts">
    import type { PageProps } from "./$types";

    import { ArrowRight } from "@lucide/svelte";
    import { Button } from "bits-ui";
    import { superForm } from "sveltekit-superforms";

    import Input from "#lib/ui/Input.svelte";
    import Alert from "#lib/ui/Alert.svelte";

    import { page } from "$app/state";
    import { MESSAGES } from "#lib/utils/redirect.ts";

    // Login only, for the practice's administrator and practitioner: client
    // accounts are out of scope in v1 (clients book as guests and manage
    // their booking from the link in their email), so the page offers no
    // account creation and no Google/Apple sign-in (which would create one).

    let { data }: PageProps = $props();

    const redirectFrom = page.url.searchParams.get("redirectFrom");

    const { form, errors, enhance, constraints } = superForm(data.loginForm, {
        resetForm: true,
    });
</script>

<div class="grid place-content-center h-[100vh]">
    <div class="pt-6">
        <div
            class="rounded-card border-muted bg-background-alt shadow-card md:w-[540px] w-[360px] border p-8"
        >
            <div class="select-none pt-8">
                <div class="text-center mb-8 grid gap-4">
                    <h2 class="font-extrabold text-2xl">Retrouvez votre espace</h2>
                    <p class="text-muted-foreground">
                        Connectez-vous pour accéder à votre compte.
                    </p>
                </div>
                {#if redirectFrom && !$errors._errors}
                    <Alert message={MESSAGES.expiredSession} />
                {/if}
                {#if $errors._errors}
                    <Alert message={$errors._errors[0]} />
                {/if}
            </div>
            <div class="select-none pt-8">
                <form
                    method="POST"
                    action="?/login"
                    class="grid gap-4"
                    use:enhance
                >
                    <div class="grid gap-6">
                        <div>
                            <Input
                                label="email"
                                name="email"
                                type="email"
                                autocomplete="email"
                                bind:value={$form.email}
                                error={$errors.email ? $errors.email[0] : ""}
                                {...$constraints.email}
                            />
                        </div>
                        <div>
                            <Input
                                name="password"
                                label="Mot de passe"
                                type="password"
                                autocomplete="current-password"
                                bind:value={$form.password}
                                error={$errors.password
                                    ? $errors.password[0]
                                    : ""}
                                showForgotPasswordbutton={true}
                                {...$constraints.password}
                            />
                        </div>
                    </div>
                    <Button.Root
                        class="mt-6 justify-center gap-4 items-center h-input rounded-input bg-dark text-background shadow-mini hover:bg-dark/95 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex px-4 text-[15px] font-bold focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer"
                        type="submit"
                    >
                        Se connecter
                        <ArrowRight size={20} />
                    </Button.Root>
                </form>
            </div>
        </div>
    </div>
</div>
