<script lang="ts">
    // A confirmation before a destructive action. The trigger is a button of
    // its own and the action (a form) sits in the dialog: a form inside the
    // trigger would put a button inside a button and break hydration.
    import type { Snippet } from "svelte";
    import { Dialog } from "bits-ui";
    import { X, Info } from "@lucide/svelte";

    type Props = {
        /** The trigger's accessible name, e.g. "Supprimer". */
        label: string;
        message: string;
        triggerClass?: string;
        /** The trigger's content (an icon). */
        trigger: Snippet;
        /** The confirming action, a form with its submit button. */
        children: Snippet;
    };

    let { label, message, triggerClass, trigger, children }: Props = $props();
</script>

<Dialog.Root>
    <Dialog.Trigger class={triggerClass} aria-label={label} title={label}>
        {@render trigger()}
    </Dialog.Trigger>
    <Dialog.Portal>
        <Dialog.Overlay
            class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-black/80"
        />
        <Dialog.Content
            class="rounded-card-lg bg-background shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden fixed left-[50%] top-[50%] z-50 w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] border p-8 sm:max-w-[540px] md:w-full"
        >
            <div class="flex items-start gap-6">
                <Info size={32} class="shrink-0" />
                <div>
                    <Dialog.Title class="w-full text-xl font-semibold tracking-tight">
                        Êtes-vous sûr ?
                    </Dialog.Title>
                    <Dialog.Description class="pt-2 pb-4 text-sm text-muted-foreground">
                        {message}
                    </Dialog.Description>
                </div>
            </div>
            <div class="flex w-full justify-end gap-3 mt-4">
                <Dialog.Close
                    class="px-4 py-2 rounded-lg border border-border-card text-sm font-medium text-foreground hover:bg-muted transition-colors cursor-pointer"
                >
                    Annuler
                </Dialog.Close>
                {@render children()}
            </div>
            <Dialog.Close
                class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer"
            >
                <X class="text-foreground size-5" />
                <span class="sr-only">Fermer</span>
            </Dialog.Close>
        </Dialog.Content>
    </Dialog.Portal>
</Dialog.Root>
