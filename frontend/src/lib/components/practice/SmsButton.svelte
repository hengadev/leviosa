<!--
	@component
	Round or wide "Envoyer un SMS" button opening a small menu of ready-made
	messages (plus an empty one). Each message opens the phone's Messages app
	via an `sms:` link — nothing goes through Leviosa.

	Props:
	- phone: the dialable number (E.164 as stored), already checked non-empty by the caller
	- firstName: the client's first name, used for the accessible label
	- variant: 'round' (44px circle, light surfaces) or 'wide' (square, dark next-session card)
-->
<script lang="ts">
	import { DropdownMenu } from 'bits-ui';
	import { MessageSquare } from '@lucide/svelte';

	let {
		phone,
		firstName,
		variant = 'round'
	}: { phone: string; firstName: string; variant?: 'round' | 'wide' } = $props();

	// One constant French list, reused as-is everywhere.
	const READY_MADE_MESSAGES = [
		"J'ai environ 10 minutes de retard.",
		'Je suis disponible pour vous rappeler.'
	] as const;

	// iOS form: `sms:<number>&body=<encoded text>`; no body for the empty message.
	function smsHref(text?: string): string {
		return text ? `sms:${phone}&body=${encodeURIComponent(text)}` : `sms:${phone}`;
	}

	const label = `Envoyer un SMS à ${firstName}`;

	const triggerClass =
		variant === 'wide'
			? 'flex size-11 shrink-0 items-center justify-center rounded-xl bg-background text-foreground'
			: 'flex size-11 shrink-0 items-center justify-center rounded-full border border-border-card hover:bg-muted';
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger class={triggerClass} aria-label={label}>
		<MessageSquare size={18} />
	</DropdownMenu.Trigger>
	<DropdownMenu.Portal>
		<DropdownMenu.Content
			align="end"
			sideOffset={4}
			class="z-50 min-w-56 max-w-[calc(100vw-2rem)] bg-background border border-border-card rounded-lg shadow-popover py-1"
		>
			<p class="px-3 pt-1.5 pb-1 text-xs font-medium text-muted-foreground">{label}</p>
			{#each READY_MADE_MESSAGES as message (message)}
				<a
					href={smsHref(message)}
					class="flex min-h-11 items-center px-3 text-sm text-foreground-alt hover:text-foreground hover:bg-muted transition-colors"
				>
					{message}
				</a>
			{/each}
			<div class="my-1 border-t border-border-card"></div>
			<a
				href={smsHref()}
				class="flex min-h-11 items-center px-3 text-sm text-foreground-alt hover:text-foreground hover:bg-muted transition-colors"
			>
				Message vide
			</a>
		</DropdownMenu.Content>
	</DropdownMenu.Portal>
</DropdownMenu.Root>
