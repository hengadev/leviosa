<script>
	export let data;
	$: ({ mails, notifications } = data);
	import * as Tabs from '$lib/components/ui/tabs';
	import Bell from 'lucide-svelte/icons/bell';
</script>

<div class="content">
	<h2 class="title">Emails et notifications</h2>
	{#if mails.length === 0}
		<p>no mails</p>
	{:else}
		<div class="block">
			<Tabs.Root value="messages" class="dark w-[100%]">
				<Tabs.List class="tabs w-[100%]">
					<Tabs.Trigger value="messages" class="w-[100%]">Messages</Tabs.Trigger>
					<Tabs.Trigger value="notifications" class="w-[100%]">Notifications</Tabs.Trigger>
				</Tabs.List>
				<Tabs.Content value="messages">
					<div class="mails">
						{#each mails as mail}
							<div class="mail">
								<div class="mail__image"></div>
								<div class="mail__text">
									<p><span class="mail__sender">Equipe MyApp </span> vous envoie un rappel.</p>
									<p>2 hours ago</p>
								</div>
							</div>
						{/each}
					</div>
				</Tabs.Content>
				<Tabs.Content value="notifications">
					<div class="notifications">
						{#each notifications as notification}
							<div class="notification">
								<Bell />
								<div class="notification__text">
									<p>
										<span class="notification__sender">Equipe MyApp </span> : nouvelle notification !
									</p>
									<p>2 hours ago</p>
									<!-- <p>{notification.content}</p> -->
								</div>
							</div>
						{/each}
					</div>
				</Tabs.Content>
			</Tabs.Root>
		</div>
	{/if}
</div>

<style>
	.content {
		padding: 2rem 1.5rem;
		margin-bottom: 4rem;
	}
	/* .tabs { */
	/* 	border: 1px solid rgba(60, 60, 67, 0.78); */
	/* } */
	.block {
		margin-top: 3rem;
		padding: 1rem;
		border-radius: 0.5rem;
		box-shadow:
			rgba(60, 60, 67, 0.78) 0px 1px 3px 0px,
			rgba(60, 60, 67, 0.78) 0px 0px 0px 1px;
		/* background-color: #202127; */
		/* border twitter 1px solid rgb(47,51,54) */
	}

	.title {
		font-size: 1.5rem;
		color: #f67373;
		font-weight: 700;
	}

	.mails,
	.notifications {
		margin-top: 1rem;
	}

	.mail,
	.notification {
		margin-bottom: 1rem;
		background-color: rgba(60, 60, 67, 0.78);
		background-color: #202127;
		padding: 0.75rem;
		border-radius: 0.5rem;
		/* border: 0.5px solid #bdbdbd; */
		display: flex;
		/* align-items: flex-start; */
		align-items: center;
		gap: 1.2rem;
		box-shadow:
			rgba(60, 60, 67, 0.78) 0px 1px 3px 0px,
			rgba(60, 60, 67, 0.78) 0px 0px 0px 1px;
	}

	.mail__image {
		background-color: white;
		width: 25px;
		aspect-ratio: 1;
		height: min-content;
		border-radius: 0.2rem;
	}
	.mail__sender,
	.notification__sender {
		color: white;
		font-weight: 600;
		font-size: 0.9rem;
	}
	.mail__text,
	.notification__text {
		font-size: 0.8rem;
	}
</style>
