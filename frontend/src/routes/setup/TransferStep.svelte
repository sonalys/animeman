<script lang="ts">
	import { configStore } from '$lib/stores/config';
	let { onNext }: { onNext: () => void } = $props();

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
		onNext();
	}
</script>

<div class="step-container">
	<h2>Transfer Client (qBittorrent)</h2>
	<p class="description">Enter the details for your qBittorrent instance.</p>

	<div class="form-group">
		<label for="url">Connection URL</label>
		<input
			id="url"
			type="url"
			bind:value={$configStore.transfer.url}
			placeholder="http://192.168.1.219:8088"
		/>
	</div>

	<div class="form-row">
		<div class="form-group">
			<label for="user">Username</label>
			<input
				id="user"
				type="text"
				bind:value={$configStore.transfer.username}
				placeholder="admin"
			/>
		</div>

		<div class="form-group">
			<label for="pass">Password</label>
			<input id="pass" type="password" bind:value={$configStore.transfer.password} />
		</div>
	</div>

	{#if error}
		<p class="error-message">{error}</p>
	{/if}

	<div class="actions">
		<button class="primary" onclick={validateAndNext}>Continue</button>
	</div>
</div>

<style>
	.form-group {
		margin-bottom: 1rem;
		display: flex;
		flex-direction: column;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	label {
		font-weight: bold;
		font-size: 0.9rem;
		margin-bottom: 0.25rem;
	}

	input {
		padding: 0.5rem;
		border: 1px solid #ccc;
		border-radius: 4px;
	}

	.error-message {
		color: #ef4444;
		font-size: 0.85rem;
	}

	.actions {
		display: flex;
		justify-content: space-between;
		margin-top: 2rem;
	}

	button.primary {
		background-color: #2563eb;
		color: white;
		padding: 0.5rem 1.5rem;
		border-radius: 4px;
		border: none;
		cursor: pointer;
	}

	button.secondary {
		background: none;
		border: 1px solid #ccc;
		padding: 0.5rem 1.5rem;
		border-radius: 4px;
		cursor: pointer;
	}
</style>
