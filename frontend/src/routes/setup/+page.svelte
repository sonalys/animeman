<script lang="ts">
	import { scale, slide } from 'svelte/transition';
	import { apiFetch } from '$lib/api';
	import IndexerStep from './IndexerStep.svelte';
	import TransferStep from './TransferStep.svelte';
	import type { IndexerConfig } from '$lib/api/types';
	import Stepper from '$lib/components/MultiStep.svelte';

	// Reactive state for the whole form
	let formState = $state({
		indexingClient: {
			url: 'http://localhost:9696',
			auth: { type: 'apiKey', key: '' }
		},
		transferClient: {}
	} as {
		indexingClient: IndexerConfig;
		transferClient: {};
	});

	function handleSubmit() {
		console.log('Final Submission:', $state.snapshot(formState));
		// Call your SvelteKit Action here
	}
</script>

{#snippet indexerStep({ next }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Configure Indexer</h2>
			<p>Connect your primary search provider.</p>
		</div>

		<div class="field-group">
			<label for="url">Instance URL</label>
			<input
				id="url"
				type="url"
				bind:value={formState.indexingClient.url}
				placeholder="http://localhost:9696"
			/>
		</div>

		<div class="field-group">
			<label for="authMethod">Auth Method</label>
			<div class="segmented-control" id="authMethod">
				<button
					type="button"
					class:active={formState.indexingClient.auth.type === 'apiKey'}
					onclick={() => (formState.indexingClient.auth.type = 'apiKey')}>API Key</button
				>
				<button
					type="button"
					class:active={formState.indexingClient.auth.type === 'userPassword'}
					onclick={() => (formState.indexingClient.auth.type = 'userPassword')}>Credentials</button
				>
			</div>
		</div>

		{#if formState.indexingClient.auth.type === 'apiKey'}
			<div class="field-group">
				<label for="key">API Key</label>
				<input
					id="key"
					type="password"
					bind:value={formState.indexingClient.auth.key}
					placeholder="Paste key here..."
				/>
			</div>
		{:else}
			<div class="field-row">
				<div class="field-group">
					<label for="user">User</label>
					<input
						id="user"
						bind:value={formState.indexingClient.auth.username}
						placeholder="Username"
					/>
				</div>
				<div class="field-group">
					<label for="pass">Pass</label>
					<input
						id="pass"
						type="password"
						bind:value={formState.indexingClient.auth.password}
						placeholder="••••••"
					/>
				</div>
			</div>
		{/if}

		<button class="btn-primary" onclick={next}> Continue </button>
	</div>
{/snippet}

{#snippet transferStep({ next, back }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Configure Indexer</h2>
			<p>Connect your primary search provider.</p>
		</div>

		<div class="field-group">
			<label for="url">Instance URL</label>
			<input
				id="url"
				type="url"
				bind:value={formState.indexingClient.url}
				placeholder="http://localhost:9696"
			/>
		</div>

		<div class="field-group">
			<label for="authMethod">Auth Method</label>
			<div class="segmented-control" id="authMethod">
				<button
					type="button"
					class:active={formState.indexingClient.auth.type === 'apiKey'}
					onclick={() => (formState.indexingClient.auth.type = 'apiKey')}>API Key</button
				>
				<button
					type="button"
					class:active={formState.indexingClient.auth.type === 'userPassword'}
					onclick={() => (formState.indexingClient.auth.type = 'userPassword')}>Credentials</button
				>
			</div>
		</div>

		{#if formState.indexingClient.auth.type === 'apiKey'}
			<div class="field-group">
				<label for="key">API Key</label>
				<input
					id="key"
					type="password"
					bind:value={formState.indexingClient.auth.key}
					placeholder="Paste key here..."
				/>
			</div>
		{:else}
			<div class="field-row">
				<div class="field-group">
					<label for="user">User</label>
					<input
						id="user"
						bind:value={formState.indexingClient.auth.username}
						placeholder="Username"
					/>
				</div>
				<div class="field-group">
					<label for="pass">Pass</label>
					<input
						id="pass"
						type="password"
						bind:value={formState.indexingClient.auth.password}
						placeholder="••••••"
					/>
				</div>
			</div>
		{/if}

		<div class="btn-group">
			<button class="btn-ghost" onclick={back}>Back</button>
			<button class="btn-primary" onclick={next}> Continue </button>
		</div>
	</div>
{/snippet}

{#snippet successStep()}
	<div class="success-screen">
		<div class="icon-check">✓</div>
		<h1>Ready!</h1>
		<button class="btn-primary" onclick={() => alert('Done!')}>Enter Dashboard</button>
	</div>
{/snippet}

<Stepper
	onComplete={handleSubmit}
	steps={[
		{ name: 'Indexer', component: indexerStep },
		{ name: 'Library', component: transferStep },
		{ name: 'Finish', component: successStep }
	]}
/>

<style>
	.field-group {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.segmented-control {
		display: grid;
		grid-template-columns: 1fr 1fr;
		background: rgba(15, 23, 42, 0.6);
		padding: 4px;
		border-radius: 14px;
		border: 1px solid rgba(255, 255, 255, 0.05);
		margin-bottom: 20px;
	}

	.segmented-control button {
		background: transparent;
		border: none;
		color: #94a3b8;
		padding: 10px;
		border-radius: 10px;
		cursor: pointer;
		font-weight: 600;
	}

	.segmented-control button.active {
		background: #38bdf8;
		color: #0f172a;
	}

	label {
		display: block;
		font-size: 0.75rem;
		font-weight: 700;
		color: #64748b;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.step-content {
		justify-items: stretch;
	}

	/* Styles for the inside of your snippets */
	.step-content h2 {
		margin: 0 0 8px 0;
		font-size: 1.5rem;
	}
	.step-content p {
		color: #94a3b8;
		margin-bottom: 24px;
	}

	input {
		width: 100%;
		box-sizing: border-box;
		padding: 14px;
		background: rgba(15, 23, 42, 0.6);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		color: white;
		margin-bottom: 15px;
	}

	.btn-group {
		display: flex;
		gap: 12px;
	}

	.btn-primary {
		flex: 1;
		width: 100%;
		padding: 14px;
		background: #38bdf8;
		border: none;
		border-radius: 12px;
		font-weight: 700;
		color: #0f172a;
		cursor: pointer;
		transition: transform 0.2s;
	}
	.btn-primary:hover {
		transform: translateY(-2px);
		background: #7dd3fc;
	}

	.btn-ghost {
		padding: 14px 24px;
		background: transparent;
		color: #94a3b8;
		border: 1px solid #334155;
		border-radius: 12px;
		cursor: pointer;
	}

	.success-screen {
		text-align: center;
	}
	.icon-check {
		width: 64px;
		height: 64px;
		background: #059669;
		border-radius: 50%;
		display: grid;
		place-items: center;
		margin: 0 auto 16px;
		font-size: 2rem;
	}
</style>
