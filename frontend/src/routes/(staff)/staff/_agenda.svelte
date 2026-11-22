<script lang="ts">
    import type { TodaySlot } from './+page.server';

    let { slots }: { slots: TodaySlot[] } = $props();

    const statusConfig: Record<string, { label: string; dot: string; text: string }> = {
        // Black and white: what is still to happen is solid, what is settled recedes.
        confirmed: { label: 'À venir', dot: 'bg-foreground', text: 'text-foreground' },
        in_progress: { label: 'En cours', dot: 'bg-foreground', text: 'text-foreground' },
        completed: { label: 'Terminé', dot: 'bg-muted-foreground', text: 'text-muted-foreground' },
        no_show: { label: 'Absent', dot: 'bg-muted-foreground', text: 'text-muted-foreground' },
        cancelled: { label: 'Annulé', dot: 'bg-muted-foreground/40', text: 'text-muted-foreground' },
    };

    function getStatus(status: string) {
        return statusConfig[status] ?? { label: status, dot: 'bg-muted-foreground', text: 'text-muted-foreground' };
    }
</script>

<div class="bg-background rounded-2xl border border-border overflow-hidden">
    <div class="px-5 py-4 flex items-center justify-between">
        <h2 class="text-base font-semibold text-foreground">Agenda du jour</h2>
        <span class="text-[11px] text-muted-foreground tabular-nums">{slots.length} séance{slots.length !== 1 ? 's' : ''}</span>
    </div>

    {#if slots.length === 0}
        <div class="px-5 py-10 text-center">
            <p class="text-sm text-muted-foreground">Aucune séance aujourd'hui</p>
        </div>
    {:else}
        <div class="relative">
            <!-- Timeline rail -->
            <div class="absolute left-[3.25rem] top-0 bottom-0 w-px bg-border"></div>

            {#each slots as slot}
                {@const s = getStatus(slot.status)}
                <div class="relative px-5 py-4 flex items-center gap-4 hover:bg-muted/30 transition-colors duration-300 group">
                    <!-- Timeline dot -->
                    <div class="relative z-10 w-2.5 h-2.5 rounded-full {s.dot} ring-4 ring-background group-hover:scale-125 transition-transform duration-300"></div>

                    <!-- Time -->
                    <div class="flex flex-col min-w-[2.5rem]">
                        <span class="text-sm font-semibold text-foreground tabular-nums leading-none">{slot.startTime}</span>
                        <span class="text-[11px] text-muted-foreground tabular-nums">{slot.endTime}</span>
                    </div>

                    <!-- Content -->
                    <div class="flex-1 min-w-0">
                        <p class="text-sm font-medium text-foreground truncate">{slot.productName}</p>
                        <p class="text-[11px] text-muted-foreground mt-0.5">{slot.clientName}</p>
                    </div>

                    <!-- Status pill -->
                    <span class="text-[11px] font-semibold {s.text} whitespace-nowrap tracking-wide">
                        {s.label}
                    </span>
                </div>
            {/each}
        </div>
    {/if}

    <a
        href="/staff/agenda/reservations"
        class="block text-center py-3.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors border-t border-border"
    >
        Agenda complet →
    </a>
</div>
