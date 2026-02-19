<script lang="ts">
	import { configStore } from '$lib/stores/config';

	let { formState = $bindable() } = $props();

	// Local validation state
	let error = $state(''); // This fixed the warning!

	function validateAndNext() {
		const { url, username, password } = $configStore.transfer;

		if (!url.startsWith('http')) {
			error = 'Please enter a valid URL (including http/https)';
			return;
		}

		if (!username || !password) {
			error = 'Username and Password are required for qBittorrent';
			return;
		}

		error = '';
	}
</script>

<div class="step-container">
	<h2>Transfer Client (qBittorrent)</h2>
	<p class="description">Enter the details for your qBittorrent instance.</p>

	<div class="field">
		<label for="url">Instance URL</label>
		<input id="url" bind:value={formState.url} placeholder="http://localhost:9696" />
	</div>

	<div class="field">
		<label for="user">Username</label>
		<input id="user" type="text" bind:value={$configStore.transfer.username} placeholder="admin" />
	</div>
	<div class="field">
		<label for="pass">Password</label>
		<input id="pass" type="password" bind:value={$configStore.transfer.password} />
	</div>

	{#if error}
		<p class="error-message">{error}</p>
	{/if}

	<div class="actions">
		<button class="primary" onclick={validateAndNext}>Continue</button>
	</div>
</div>

<style>
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
</style>
