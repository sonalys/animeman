<script lang="ts">
	import { onMount } from 'svelte';
	import { userId } from '$lib/userStore';
	import { apiFetch } from '$lib/api';
	import AuthController from '$lib/components/AuthController.svelte';
	import type { Indexer, IndexerConfig, ErrorResponse } from '$lib/types';

	let indexers: Indexer[] = [];
	let loading = true;
	let showCreate = false;

	// Form State
	let newConfig: IndexerConfig = {
		type: 'prowlarr',
		url: '',
		auth: { type: 'apiKey', key: '' }
	};

	onMount(async () => {
		if ($userId) {
			await refreshIndexers();
		}
		loading = false;
	});

	async function refreshIndexers() {
		try {
			indexers = await apiFetch<Indexer[]>('/indexers');
		} catch (e) {
			console.error('Failed to load indexers', e);
		}
	}

	async function handleCreate() {
		try {
			const created = await apiFetch<Indexer>('/indexers', 'POST', newConfig);
			indexers = [...indexers, created];
			showCreate = false;
			newConfig = { type: 'prowlarr', url: '', auth: { type: 'apiKey', key: '' } };
		} catch (e) {
			alert('Failed to initialize node. Check console for details.');
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
			<button class="btn-primary" on:click={() => (showCreate = true)}> + Add Indexer </button>
		</header>

		<div class="indexer-grid">
			{#each indexers as node}
				<div class="node-card">
					<div class="node-top">
						<span class="badge">{node.type}</span>
						<div class="status-dot {node.status}"></div>
					</div>
					<h3>{new URL(node.url).hostname}</h3>
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

{#if showCreate}
	<div class="modal-backdrop" on:click|self={() => (showCreate = false)}>
		<div class="modal">
			<h2>New Indexer Connection</h2>
			<form on:submit|preventDefault={handleCreate}>
				<label for="type">Service Type</label>
				<select id="type" bind:value={newConfig.type}>
					<option value="prowlarr">Prowlarr</option>
					<option value="jackett">Jackett</option>
					<option value="torznab">Torznab</option>
				</select>

				<label for="url">Base URL</label>
				<input
					id="url"
					type="url"
					bind:value={newConfig.url}
					placeholder="http://192.168.1.x:9696"
					required
				/>

				<label>Authentication Method</label>
				<div class="auth-tabs">
					<button
						type="button"
						class:active={newConfig.auth.type === 'apiKey'}
						on:click={() => (newConfig.auth = { type: 'apiKey', key: '' })}>API Key</button
					>
					<button
						type="button"
						class:active={newConfig.auth.type === 'userPassword'}
						on:click={() => (newConfig.auth = { type: 'userPassword', username: '', password: '' })}
						>User/Pass</button
					>
				</div>

				{#if newConfig.auth.type === 'apiKey'}
					<input
						type="password"
						bind:value={newConfig.auth.key}
						placeholder="Enter API Key"
						required
					/>
				{:else}
					<input type="text" bind:value={newConfig.auth.username} placeholder="Username" required />
					<input
						type="password"
						bind:value={newConfig.auth.password}
						placeholder="Password"
						required
					/>
				{/if}

				<div class="actions">
					<button type="button" class="btn-ghost" on:click={() => (showCreate = false)}
						>Cancel</button
					>
					<button type="submit" class="btn-primary">Initialize Node</button>
				</div>
			</form>
		</div>
	</div>
{/if}

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
		max-width: 1000px;
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

	/* Indexer Grid */
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

	.status-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: #475569;
	}
	.status-dot.online {
		background: #10b981;
		box-shadow: 0 0 8px #10b981;
	}
	.status-dot.offline {
		background: var(--error);
	}

	.url-sub {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	/* Modal Styling */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.85);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal {
		background: var(--bg-secondary);
		padding: 2rem;
		border-radius: 16px;
		width: 100%;
		max-width: 450px;
		border: 1px solid var(--border);
	}

	form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		margin-top: 1.5rem;
	}
	label {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	input,
	select {
		background: var(--bg-primary);
		border: 1px solid var(--border);
		color: white;
		padding: 0.75rem;
		border-radius: 8px;
	}

	.auth-tabs {
		display: flex;
		gap: 0.5rem;
	}
	.auth-tabs button {
		flex: 1;
		padding: 0.5rem;
		background: var(--bg-primary);
		border: 1px solid var(--border);
		color: var(--text-muted);
		cursor: pointer;
		border-radius: 4px;
	}
	.auth-tabs button.active {
		border-color: var(--accent);
		color: white;
	}

	.btn-primary {
		background: var(--accent);
		color: white;
		border: none;
		padding: 0.75rem 1.2rem;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
	}

	.btn-ghost {
		background: transparent;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		margin-top: 1rem;
	}
</style>
