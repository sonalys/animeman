<script lang="ts">
	import { onMount } from 'svelte';
	import { userId } from '$lib/userStore';
	import { apiFetch } from '$lib/api';
	import AuthController from '$lib/components/AuthController.svelte';
	import type { Indexer, OnboardingStatus } from '$lib/api/types';
	import { goto } from '$app/navigation';

	let indexers: Indexer[] = [];
	let loading = true;

	onMount(async () => {
		const onboardingStatus = await apiFetch<OnboardingStatus>('/setup');

		if (!onboardingStatus.isCompleted) {
			goto('/setup');
		}

		if ($userId) await refreshIndexers();
		loading = false;
	});

	async function refreshIndexers() {
		try {
			indexers = await apiFetch<Indexer[]>('/indexing-clients');
		} catch (e) {
			console.error('Failed to load indexers', e);
		}
	}
</script>

<div class="page-layout" class:centered={!$userId}>
	{#if $userId}
		<header class="dashboard-header">
			<div class="header-text">
				<h1>Animeman Dashboard</h1>
				<p class="id-badge">User ID: {$userId}</p>
			</div>
		</header>

		<div class="indexer-grid">
			{#each indexers as node}
				<div class="node-card">
					<div class="node-top">
						<span class="badge">{node.type}</span>
					</div>
					<code class="url-sub">{node.url}</code>
				</div>
			{/each}

			{#if indexers.length === 0 && !loading}
				<div class="empty-state">
					<p>No indexers active. Initialize a new node to begin.</p>
				</div>
			{/if}
		</div>
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
		padding: 2rem;
		box-sizing: border-box;
	}

	.page-layout.centered {
		justify-content: center;
		align-items: center;
	}

	.dashboard-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		width: 100%;
		margin: 0 auto 3rem;
	}

	.hero {
		display: flex;
		flex-direction: column;
		align-items: center;
		max-width: 400px;
	}

	.logo-glow {
		font-size: 4rem;
		text-shadow: 0 0 20px var(--accent);
		margin-bottom: 1rem;
	}

	.subtitle {
		color: var(--text-muted);
		margin-bottom: 2rem;
		text-align: center;
	}

	.id-badge {
		background: var(--bg-secondary);
		padding: 0.3rem 0.8rem;
		border-radius: 12px;
		border: 1px solid var(--border);
		font-family: monospace;
		font-size: 0.8rem;
		margin-top: 0.5rem;
		display: inline-block;
	}

	.indexer-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
		width: 100%;
		max-width: 1000px;
		margin: 0 auto;
	}

	.node-card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		padding: 1.5rem;
		border-radius: 12px;
	}

	.node-top {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1rem;
	}

	.url-sub {
		font-size: 0.8rem;
		color: var(--text-muted);
	}
</style>
