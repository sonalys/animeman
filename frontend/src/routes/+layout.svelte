<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api';
	import { userId, authChecked } from '$lib/userStore';

	onMount(async () => {
		try {
			// Validate the JWT cookie with the Go backend
			const data = await apiFetch<{ userID: string }>('/authentication/whoami');
			userId.set(data.userID);
		} catch (e) {
			// If 401 or invalid signature, we treat them as a guest
			userId.set(null);
		} finally {
			// This is the key: we only reveal the page once the server responds
			authChecked.set(true);
		}
	});
</script>

{#if !$authChecked}
	<div class="splash">
		<div class="loader"></div>
		<p>Synchronizing with Overlord...</p>
	</div>
{:else}
	<slot />
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
