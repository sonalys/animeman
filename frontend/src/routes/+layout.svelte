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
	p {
		color: #38bdf8;
	}

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
		width: 2.5rem;
		height: 2.5rem;
		border: 0.15rem solid #334155;
		border-top-color: #38bdf8;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto 16px;
	}
</style>
