<script lang="ts">
    import { enhance } from "$app/forms";
    import { invalidateAll } from "$app/navigation";
    import { Upload, Trash2, AlertCircle } from "@lucide/svelte";

    // The practice page video: the administrator uploads a file (MP4 or WebM,
    // 50 MB at most) instead of pasting a link. The file goes to the upload
    // endpoint next to this page, which streams it to the backend.
    let { videoUrl }: { videoUrl: string } = $props();

    const MAX_BYTES = 50 * 1024 * 1024;
    const ACCEPTED_TYPES = ["video/mp4", "video/webm"];

    let fileInput: HTMLInputElement | undefined = $state();
    let progress: number | null = $state(null);
    let error: string | null = $state(null);
    let removing = $state(false);

    function formatMegabytes(bytes: number): string {
        return `${(bytes / (1024 * 1024)).toFixed(1).replace(".", ",")} Mo`;
    }

    function messageFor(status: number): string {
        if (status === 413) return "Le fichier dépasse 50 Mo.";
        if (status === 400) return "Ce fichier n'est pas une vidéo MP4 ou WebM lisible.";
        if (status === 401 || status === 403) return "Votre session a expiré. Reconnectez-vous puis réessayez.";
        return "L'envoi a échoué. Veuillez réessayer.";
    }

    function upload(file: File) {
        error = null;
        if (!ACCEPTED_TYPES.includes(file.type)) {
            error = "Choisissez une vidéo au format MP4 ou WebM (les vidéos .mov d'iPhone doivent d'abord être exportées en MP4).";
            return;
        }
        if (file.size > MAX_BYTES) {
            error = `Cette vidéo fait ${formatMegabytes(file.size)} : la limite est de 50 Mo. Exportez-la en 1080p ou moins, sans le son.`;
            return;
        }

        // XMLHttpRequest rather than fetch: it reports upload progress.
        const body = new FormData();
        body.append("video", file);
        const request = new XMLHttpRequest();
        request.open("POST", "/admin/settings/video");
        request.upload.onprogress = (event) => {
            if (event.lengthComputable) progress = Math.round((event.loaded / event.total) * 100);
        };
        request.onload = async () => {
            if (request.status >= 200 && request.status < 300) {
                await invalidateAll();
            } else {
                error = messageFor(request.status);
            }
            progress = null;
        };
        request.onerror = () => {
            error = "L'envoi a été interrompu. Vérifiez votre connexion puis réessayez.";
            progress = null;
        };
        progress = 0;
        request.send(body);
    }

    function onFileChosen(event: Event) {
        const input = event.currentTarget as HTMLInputElement;
        const file = input.files?.[0];
        input.value = "";
        if (file) upload(file);
    }
</script>

<div class="py-3 space-y-3">
    <div class="flex items-start justify-between gap-4">
        <div class="flex-1 min-w-0">
            <div class="font-medium text-sm">Vidéo de la page</div>
            <p class="text-sm text-foreground-alt mt-1">
                MP4 ou WebM, 50 Mo maximum. Elle est lue en boucle et sans le son : une vidéo courte
                (20 à 40 secondes), en 1080p ou moins, se charge vite sur téléphone.
            </p>
        </div>
    </div>

    {#if videoUrl}
        <!-- svelte-ignore a11y_media_has_caption -->
        <video
            src={videoUrl}
            class="aspect-video w-full max-w-md rounded-card-sm bg-neutral-900 object-cover"
            muted
            playsinline
            controls
            preload="metadata"
        ></video>
    {:else}
        <p class="text-sm text-foreground-alt">Aucune vidéo : la page affiche une image d'attente.</p>
    {/if}

    <input
        bind:this={fileInput}
        type="file"
        accept="video/mp4,video/webm"
        class="sr-only"
        tabindex="-1"
        aria-hidden="true"
        onchange={onFileChosen}
    />

    {#if progress !== null}
        <div class="max-w-md space-y-1" aria-live="polite">
            <div class="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div class="h-full bg-dark transition-[width]" style="width: {progress}%"></div>
            </div>
            <p class="text-xs text-foreground-alt">
                {progress < 100 ? `Envoi en cours… ${progress} %` : "Enregistrement de la vidéo…"}
            </p>
        </div>
    {/if}

    {#if error}
        <p class="flex items-start gap-2 text-sm text-red-700" role="alert">
            <AlertCircle class="size-4 shrink-0 mt-0.5" aria-hidden="true" />
            {error}
        </p>
    {/if}

    <div class="flex flex-wrap gap-3">
        <button
            type="button"
            class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center gap-2 px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all disabled:cursor-not-allowed disabled:opacity-50"
            disabled={progress !== null || removing}
            onclick={() => fileInput?.click()}
        >
            <Upload class="size-4" aria-hidden="true" />
            {videoUrl ? "Remplacer la vidéo" : "Choisir une vidéo"}
        </button>

        {#if videoUrl}
            <form
                method="POST"
                action="?/removePracticePageVideo"
                use:enhance={() => {
                    removing = true;
                    error = null;
                    return async ({ result }) => {
                        removing = false;
                        if (result.type === "success") {
                            await invalidateAll();
                        } else {
                            error = "La vidéo n'a pas pu être retirée. Veuillez réessayer.";
                        }
                    };
                }}
            >
                <button
                    type="submit"
                    class="h-input rounded-input border border-border-input hover:bg-dark-04 focus-visible:ring-dark focus-visible:ring-offset-background focus-visible:outline-hidden inline-flex items-center justify-center gap-2 px-4 text-sm font-medium focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] cursor-pointer transition-all disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={progress !== null || removing}
                >
                    <Trash2 class="size-4" aria-hidden="true" />
                    {removing ? "Retrait…" : "Retirer la vidéo"}
                </button>
            </form>
        {/if}
    </div>
</div>
