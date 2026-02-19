<script lang="ts">
	import { apiFetch } from '$lib/api';
	let { onNext } = $props();

	let url = $state('');
	let apiKey = $state('');
	let loading = $state(false);
	let error = $state('');

	let canSubmit = $derived(url.length > 8 && apiKey.length > 5);

	async function handleSave() {
		loading = true;
		error = '';
		try {
			await apiFetch('/indexers', {
				method: 'POST',
				body: { type: 'prowlarr', url, apiKey }
			});
			onNext();
		} catch (e: any) {
			error = e.details || 'Connection failed. Check settings.';
		} finally {
			loading = false;
		}
	}
</script>

<div class="step-form">
	<h3>Indexer Config</h3>
	<p>Connect your Prowlarr instance to begin indexing.</p>

	<div class="field">
		<label for="url">Instance URL</label>
		<input id="url" bind:value={url} placeholder="http://localhost:9696" />
	</div>

	<div class="field">
		<label for="key">API Key</label>
		<input id="key" type="password" bind:value={apiKey} placeholder="Paste key here" />
	</div>

	{#if error}
		<div class="error-toast">{error}</div>
	{/if}

	<button onclick={handleSave} disabled={!canSubmit || loading}>
		{loading ? 'Verifying...' : 'Continue'}
	</button>
</div>

<style>
	h3 {
		margin: 0 0 8px;
		font-size: 1.5rem;
		font-weight: 700;
	}
	p {
		margin: 0 0 32px;
		color: #94a3b8;
		font-size: 0.95rem;
		line-height: 1.5;
	}

	.field {
		margin-bottom: 20px;
	}
	label {
		display: block;
		font-size: 0.75rem;
		font-weight: 700;
		color: #64748b;
		text-transform: uppercase;
		margin-bottom: 8px;
		letter-spacing: 0.5px;
	}

	input {
		width: 100%;
		background: #0f172a;
		border: 1px solid #334155;
		padding: 12px 16px;
		border-radius: 12px;
		color: white;
		font-size: 1rem;
		box-sizing: border-box;
		transition: 0.2s;
	}
	input:focus {
		border-color: #38bdf8;
		outline: none;
		box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.1);
	}

	button {
		width: 100%;
		padding: 14px;
		background: #38bdf8;
		border: none;
		border-radius: 12px;
		font-weight: 700;
		color: #0f172a;
		cursor: pointer;
		margin-top: 12px;
		transition: 0.2s;
	}
	button:disabled {
		background: #1e293b;
		color: #475569;
		cursor: not-allowed;
	}
	button:hover:not(:disabled) {
		transform: translateY(-1px);
		background: #7dd3fc;
	}

	.error-toast {
		background: rgba(239, 68, 68, 0.1);
		color: #f87171;
		padding: 12px;
		border-radius: 8px;
		font-size: 0.85rem;
		margin-bottom: 16px;
		border: 1px solid rgba(239, 68, 68, 0.2);
	}
</style>
