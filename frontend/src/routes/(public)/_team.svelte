<script lang="ts">
    import { reveal } from "$lib/actions/reveal";

    type TeamMember = {
        id: string;
        firstname: string;
        lastname: string;
        occupation: string;
        quote: string;
        tags: string[];
        picture?: string;
    };

    let { members }: { members: TeamMember[] } = $props();
</script>

<section class="py-20 lg:py-24 bg-background-section">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <!-- Section Header -->
        <div
            class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12"
            use:reveal={{ preset: "fade-up", delay: 100 }}
        >
            <div class="max-w-xl">
                <span
                    class="text-sm font-semibold text-muted-foreground uppercase tracking-wider"
                    >Notre Équipe</span
                >
                <h2
                    class="mt-3 text-3xl md:text-4xl font-semibold tracking-tight text-foreground"
                >
                    Des experts à votre service
                </h2>
                <p class="mt-4 text-base text-muted-foreground">
                    Chaque praticien est soigneusement sélectionné et vérifié
                    pour garantir la qualité de votre suivi.
                </p>
            </div>
            <a
                href="/team"
                class="group inline-flex items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors px-4 py-2 rounded-full hover:bg-surface-hover whitespace-nowrap"
            >
                Voir toute l'équipe
                <span
                    class="iconify group-hover:translate-x-0.5 transition-transform text-muted-foreground group-hover:text-foreground"
                    data-icon="lucide:arrow-right"
                    data-width="16"
                    data-stroke-width="2"
                ></span>
            </a>
        </div>

        <!-- Expert Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8">
            {#each members as member, i}
                <div
                    class="group flex flex-col bg-white rounded-3xl border border-border-input hover:border-border-input-hover hover:shadow-card transition-all duration-300 overflow-hidden p-6 md:p-8"
                    use:reveal={{ preset: "fade-up", delay: 150 + i * 75 }}
                >
                    <!-- Header: photo + name/occupation -->
                    <div class="flex gap-4 items-center mb-4">
                        {#if member.picture}
                            <img
                                src={member.picture}
                                alt="{member.firstname} {member.lastname}"
                                class="rounded-2xl aspect-square w-14 md:w-16 flex-shrink-0 object-cover"
                            />
                        {:else}
                            <div
                                class="bg-surface-hover rounded-2xl aspect-square w-14 md:w-16 flex-shrink-0 flex items-center justify-center text-sm font-semibold text-muted-foreground"
                            >
                                {member.firstname[0]}{member.lastname[0]}
                            </div>
                        {/if}
                        <div>
                            <h4
                                class="text-foreground font-semibold text-lg tracking-tight"
                            >
                                {member.firstname} {member.lastname}
                            </h4>
                            <p class="text-muted-foreground text-sm font-medium mt-0.5">
                                {member.occupation}
                            </p>
                        </div>
                    </div>

                    <!-- Tags -->
                    <div class="flex flex-wrap gap-2 mb-4">
                        {#each member.tags as tag}
                            <div
                                class="inline-flex items-center bg-surface border border-border-input rounded-full px-3 py-1.5"
                            >
                                <span class="text-xs font-medium text-foreground-alt"
                                    >{tag}</span
                                >
                            </div>
                        {/each}
                    </div>

                    <!-- Quote -->
                    {#if member.quote}
                    <p
                        class="text-muted-foreground text-sm leading-relaxed mb-6 flex-1"
                    >
                        "{member.quote}"
                    </p>
                    {/if}

                    <!-- CTAs -->
                    <div class="flex flex-col sm:flex-row gap-3 mt-auto">
                        <a
                            href="/book?partner={member.id}"
                            class="inline-flex justify-center items-center gap-2 bg-foreground hover:bg-foreground-alt text-white text-sm font-medium px-6 py-3 rounded-xl transition-all duration-200 shadow-mini hover:shadow-card"
                        >
                            Réserver
                            <span
                                class="iconify group-hover:translate-x-0.5 transition-transform"
                                data-icon="lucide:arrow-right"
                                data-width="16"
                                data-stroke-width="2"
                            ></span>
                        </a>
                        <a
                            href="/team/{member.id}"
                            class="inline-flex justify-center items-center gap-2 bg-transparent hover:bg-surface border border-border-input-hover hover:border-border-input-hover text-foreground-alt hover:text-foreground text-sm font-medium px-6 py-3 rounded-xl transition-all duration-200"
                        >
                            Voir le profil
                        </a>
                    </div>
                </div>
            {/each}
        </div>
    </div>
</section>
