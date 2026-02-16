<script lang="ts">
	import { onMount } from 'svelte';
	import { apiFetch } from '$lib/api';
	import type { Indexer, IndexerConfig, AuthType, ErrorResponse } from '$lib/types';

	// State Management
	let indexers: Indexer[] = [];
	let loading = true;
	let globalError = '';
	let fieldErrors: Record<string, string> = {};

	// Form State
	let showCreate = false;
	let config: IndexerConfig = {
		type: 'prowlarr',
		url: '',
		auth: {
			type: 'apiKey',
			key: ''
		}
	};

	onMount(async () => {
		await fetchIndexers();
	});

	async function fetchIndexers() {
		try {
			indexers = await apiFetch<Indexer[]>('/indexers');
		} catch (e) {
			globalError = (e as ErrorResponse).details || 'Failed to load indexers';
		} finally {
			loading = false;
		}
	}

	function toggleAuthMode(type: AuthType) {
		if (type === 'apiKey') {
			config.auth = { type: 'apiKey', key: '' };
		} else {
			config.auth = { type: 'userPassword', username: '', password: '' };
		}
	}

	async function handleAddIndexer() {
		globalError = '';
		fieldErrors = {};

		try {
			const result = await apiFetch<Indexer>('/indexers', 'POST', config);
			indexers = [...indexers, result];
			showCreate = false;
			// Reset Form
			config = { type: 'prowlarr', url: '', auth: { type: 'apiKey', key: '' } };
		} catch (e) {
			const err = e as ErrorResponse;
			globalError = err.details || 'Failed to add indexer';
			err.fieldErrors?.forEach((fe) => (fieldErrors[fe.field] = fe.message));
		}
	}
</script>

<div class="page-container">
	<header class="header">
		<div>
			<h1>Indexer Nodes</h1>
			<p class="subtitle">Command and monitor your media indexer fleet.</p>
		</div>
		<button class="btn-primary" on:click={() => (showCreate = true)}> + Add New Node </button>
	</header>

	{#if loading}
		<div class="state-msg">Scanning network...</div>
	{:else if indexers.length === 0}
		<div class="empty-card">
			<div class="icon">🛰️</div>
			<p>No active nodes detected.</p>
			<button class="btn-link" on:click={() => (showCreate = true)}
				>Initialize your first indexer</button
			>
		</div>
	{:else}
		<div class="grid">
			{#each indexers as node}
				<div class="node-card">
					<div class="node-header">
						<span class="badge">{node.type}</span>
						<div class="status-indicator {node.status}">
							<div class="dot"></div>
							{node.status}
						</div>
					</div>
					<div class="node-body">
						<h3>{new URL(node.url).hostname}</h3>
						<code>{node.url}</code>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if globalError}
		<div class="error-banner">{globalError}</div>
	{/if}
</div>

{#if showCreate}
	<div class="modal-overlay" on:click|self={() => (showCreate = false)}>
		<div class="modal-card">
			<h2>Initialize Node</h2>
			<form on:submit|preventDefault={handleAddIndexer}>
				<div class="form-group">
					<label for="type">Indexer Type</label>
					<select id="type" bind:value={config.type}>
						<option value="prowlarr">Prowlarr</option>
						<option value="jackett">Jackett</option>
						<option value="torznab">Torznab</option>
					</select>
				</div>

				<div class="form-group">
					<label for="url">Base URL</label>
					<input
						id="url"
						type="url"
						placeholder="http://192.168.1.x:9696"
						bind:value={config.url}
						required
					/>
					{#if fieldErrors.url}<small>{fieldErrors.url}</small>{/if}
				</div>

				<div class="form-group">
					<label>Authentication Protocol</label>
					<div class="tab-group">
						<button
							type="button"
							class:active={config.auth.type === 'apiKey'}
							on:click={() => toggleAuthMode('apiKey')}
						>
							API Key
						</button>
						<button
							type="button"
							class:active={config.auth.type === 'userPassword'}
							on:click={() => toggleAuthMode('userPassword')}
						>
							User / Pass
						</button>
					</div>
				</div>

				{#if config.auth.type === 'apiKey'}
					<div class="form-group">
						<label for="key">API Key</label>
						<input id="key" type="password" bind:value={config.auth.key} required />
					</div>
				{:else}
					<div class="form-row">
						<div class="form-group">
							<label for="user">Username</label>
							<input id="user" type="text" bind:value={config.auth.username} required />
						</div>
						<div class="form-group">
							<label for="pass">Password</label>
							<input id="pass" type="password" bind:value={config.auth.password} required />
						</div>
					</div>
				{/if}

				<div class="modal-footer">
					<button type="button" class="btn-ghost" on:click={() => (showCreate = false)}
						>Abort</button
					>
					<button type="submit" class="btn-primary">Establish Connection</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.page-container {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 3rem;
	}

	h1 {
		font-size: 2rem;
		margin: 0;
		color: var(--text-main);
	}
	.subtitle {
		color: var(--text-muted);
		margin: 0.5rem 0 0;
	}

	/* Grid Styling */
	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.5rem;
	}

	.node-card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 1.5rem;
		transition: all 0.2s;
	}

	.node-card:hover {
		border-color: var(--accent);
		transform: translateY(-2px);
	}

	.node-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
	}

	.badge {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		background: var(--bg-primary);
		padding: 4px 8px;
		border-radius: 4px;
		border: 1px solid var(--border);
		color: var(--accent);
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.75rem;
		text-transform: uppercase;
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #475569;
	}
	.online .dot {
		background: #10b981;
		box-shadow: 0 0 8px #10b981;
	}
	.offline .dot {
		background: var(--error);
	}

	.node-body h3 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
	}
	code {
		font-size: 0.8rem;
		color: var(--text-muted);
		word-break: break-all;
	}

	/* Modal & Form Styling */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.85);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
		backdrop-filter: blur(4px);
	}

	.modal-card {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		padding: 2.5rem;
		border-radius: 16px;
		width: 100%;
		max-width: 500px;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1.25rem;
	}
	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	label {
		font-size: 0.85rem;
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

	.tab-group {
		display: flex;
		background: var(--bg-primary);
		padding: 4px;
		border-radius: 8px;
		border: 1px solid var(--border);
	}

	.tab-group button {
		flex: 1;
		background: transparent;
		border: none;
		color: var(--text-muted);
		padding: 0.5rem;
		cursor: pointer;
		border-radius: 6px;
		font-size: 0.85rem;
	}

	.tab-group button.active {
		background: var(--bg-secondary);
		color: var(--accent);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		margin-top: 2rem;
	}

	.btn-primary {
		background: var(--accent);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
	}

	.btn-ghost {
		background: transparent;
		color: var(--text-muted);
		border: none;
		cursor: pointer;
	}
	.error-banner {
		margin-top: 2rem;
		color: var(--error);
		text-align: center;
	}
	.state-msg {
		text-align: center;
		margin-top: 4rem;
		color: var(--text-muted);
	}
</style>
