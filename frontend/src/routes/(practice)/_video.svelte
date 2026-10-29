<script lang="ts">
    import { Play } from "@lucide/svelte";

    // The practice's own video once one is configured in settings; until then
    // an honest placeholder frame. The prototype's third-party hero.mp4 never
    // ships (issue 009).
    let {
        label,
        src = null,
        class: className = "",
    }: { label: string; src?: string | null; class?: string } = $props();
</script>

{#if src}
    <video
        src={src}
        class="h-full w-full object-cover {className}"
        playsinline
        muted
        autoplay
        loop
    ></video>
{:else}
    <div
        class="relative flex h-full w-full items-center justify-center overflow-hidden bg-neutral-900 {className}"
        role="img"
        aria-label={label}
    >
        <!-- Diagonal stripes: reads as "media goes here" without faking a frame. -->
        <div
            class="absolute inset-0 opacity-40"
            style="background-image: repeating-linear-gradient(135deg, transparent 0 22px, rgb(255 255 255 / 0.05) 22px 23px);"
        ></div>
        <span class="relative flex items-center gap-2 rounded-full border border-white/20 px-3.5 py-1.5 text-sm text-neutral-300">
            <Play class="size-4" aria-hidden="true" />
            {label}
        </span>
    </div>
{/if}
