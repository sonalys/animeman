<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api/index';
	import { userId, authChecked } from '$lib/userStore';

	onMount(async () => {
		try {
			const data = await apiFetch<{ userID: string }>('/authentication/whoami');
			userId.set(data.userID);
		} catch (e) {
			userId.set(null);
		} finally {
			authChecked.set(true);
		}
	});
</script>

{#if $authChecked}
	<slot />
{:else}
	<div class="splash">
		<div class="loader"></div>
		<p>Synchronizing with Overlord...</p>
	</div>
{/if}

<style>
	.splash {
		height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background: var(--bg-primary);
		color: var(--accent);
	}
	.loader {
		width: 48px;
		height: 48px;
		border: 3px solid var(--bg-secondary);
		border-bottom-color: var(--accent);
		border-radius: 50%;
		animation: rotation 1s linear infinite;
		margin-bottom: 1rem;
	}
	@keyframes rotation {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>
