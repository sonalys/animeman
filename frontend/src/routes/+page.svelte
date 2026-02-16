<script lang="ts">
	import { userId } from '$lib/userStore';
	import AuthController from '$lib/components/AuthController.svelte';
</script>

<div class="page-layout">
	{#if $userId}
		<header class="dashboard-header">
			<h1>Animeman Dashboard</h1>
			<p class="id-badge">User ID: {$userId}</p>
		</header>
	{:else}
		<div class="hero">
			<div class="logo-glow">🏯</div>
			<h1>Animeman</h1>
			<p class="subtitle">Your anime media collection overlord.</p>
			<AuthController />
		</div>
	{/if}
</div>

<style>
	.page-layout {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 2rem;
	}
	.hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		margin-top: 10vh;
	}
	.logo-glow {
		font-size: 4rem;
		text-shadow: 0 0 20px var(--accent);
		margin-bottom: 1rem;
	}
	.subtitle {
		color: var(--text-muted);
		margin-bottom: 2rem;
	}
	.id-badge {
		text-align: center;
		background: var(--bg-secondary);
		padding: 0.5rem 1rem;
		border-radius: 20px;
		border: 1px solid var(--border);
		font-family: monospace;
	}
</style>
