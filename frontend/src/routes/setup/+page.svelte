<script lang="ts">
	import { apiFetch } from '$lib/api';
	import type {
		IndexerConfig as IndexerClientConfig,
		TransferClientConfig,
		WatchlistConfig
	} from '$lib/api/types';
	import Stepper from '$lib/components/MultiStep.svelte';
	import SegmentedControl from '$lib/components/SegmentedControl.svelte';
	import { slide } from 'svelte/transition';

	// Reactive state for the whole form
	let formState = $state({
		indexingClient: {
			url: 'http://localhost:9696',
			auth: { type: 'apiKey', key: '' }
		},
		transferClient: {
			url: 'http://localhost:8080',
			auth: { type: 'userPassword', username: '', password: '' }
		},
		watchlist: {
			externalID: '',
			syncFrequencySeconds: 300
		}
	} as {
		indexingClient: IndexerClientConfig;
		transferClient: TransferClientConfig;
		watchlist: WatchlistConfig;
	});

	function handleSubmit() {
		console.log('Final Submission:', $state.snapshot(formState));
	}

	const authOptions = [
		{ label: 'API Key', value: 'apiKey' },
		{ label: 'Credentials', value: 'userPassword' }
	];

	const sourceOptions = [
		{ label: 'None', value: undefined },
		{ label: 'AniList', value: 'anilist' },
		{ label: 'MAL', value: 'mal' },
		{ label: 'Local', value: 'local' }
	];

	// Helper to display friendly frequency names
	const frequencies = [
		{ label: '15m', value: 900 },
		{ label: '1h', value: 3600 },
		{ label: '6h', value: 21600 },
		{ label: 'Daily', value: 86400 }
	];
</script>

{#snippet indexerStep({ next }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Configure Indexer</h2>
			<p>For finding your favorite episode</p>
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

		<SegmentedControl
			bind:active={formState.indexingClient.auth.type}
			options={authOptions}
			name="Auth Type"
		/>

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
			<h2>Configure Transfer Client</h2>
			<p>For downloading joy</p>
		</div>

		<div class="field-group">
			<label for="url">Instance URL</label>
			<input
				id="url"
				type="url"
				bind:value={formState.transferClient.url}
				placeholder="http://localhost:9696"
			/>
		</div>

		<SegmentedControl
			bind:active={formState.transferClient.auth.type}
			options={authOptions}
			name="Auth Type"
		/>

		{#if formState.transferClient.auth.type === 'apiKey'}
			<div class="field-group">
				<label for="key">API Key</label>
				<input
					id="key"
					type="password"
					bind:value={formState.transferClient.auth.key}
					placeholder="Paste key here..."
				/>
			</div>
		{:else}
			<div class="field-row">
				<div class="field-group">
					<label for="user">User</label>
					<input
						id="user"
						bind:value={formState.transferClient.auth.username}
						placeholder="Username"
					/>
				</div>
				<div class="field-group">
					<label for="pass">Pass</label>
					<input
						id="pass"
						type="password"
						bind:value={formState.transferClient.auth.password}
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

{#snippet watchlistStep({ next, back }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Watchlist</h2>
			<p>For keeping your stash up-to-date</p>
		</div>

		<div class="field-group">
			<SegmentedControl
				bind:active={formState.watchlist.source}
				options={sourceOptions}
				name="Provider"
			/>
		</div>

		{#if formState.watchlist.source == 'anilist' || formState.watchlist.source == 'mal'}
			<div class="external-fields" transition:slide>
				<div class="field-group">
					<label for="extId">Username</label>
					<input id="extId" bind:value={formState.watchlist.externalID} placeholder="Username" />
				</div>

				<div class="field-group">
					<label for="freq">Sync Frequency</label>
					<div class="frequency-grid" id="freq">
						{#each frequencies as freq}
							<button
								type="button"
								class="freq-btn"
								class:active={formState.watchlist.syncFrequencySeconds === freq.value}
								onclick={() => (formState.watchlist.syncFrequencySeconds = freq.value)}
							>
								{freq.label}
							</button>
						{/each}
					</div>
				</div>
			</div>
		{/if}

		<div class="btn-group">
			<button class="btn-ghost" onclick={back}>Back</button>
			<button class="btn-primary" onclick={next}>Continue</button>
		</div>
	</div>
{/snippet}

{#snippet successStep({ next, back }: any)}
	<div class="success-screen">
		<div class="icon-check">✓</div>
		<h1>Ready!</h1>
		<div class="btn-group">
			<button class="btn-ghost" onclick={back}>Back</button>
			<button class="btn-primary" onclick={next}> Continue </button>
		</div>
	</div>
{/snippet}

<Stepper
	steps={[
		{ name: 'Indexer', component: indexerStep },
		{ name: 'Library', component: transferStep },
		{ name: 'Watchlist', component: watchlistStep },
		{ name: 'Finish', component: successStep }
	]}
/>

<style>
	.field-group {
		margin-bottom: 5px;
	}

	/* Specific Frequency Button Grid */
	.frequency-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 8px;
		margin-bottom: 20px;
	}

	.freq-btn {
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid rgba(255, 255, 255, 0.05);
		color: #94a3b8;
		padding: 10px;
		border-radius: 10px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.freq-btn:hover {
		background: rgba(255, 255, 255, 0.08);
	}

	.freq-btn.active {
		background: rgba(56, 189, 248, 0.15);
		border-color: #38bdf8;
		color: #38bdf8;
	}

	.step-content {
		justify-items: stretch;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}

	/* Styles for the inside of your snippets */
	.step-content h2 {
		font-size: 1.5rem;
		margin: 0;
	}

	.step-content p {
		margin: 0;
		margin-bottom: 20px;
		color: #94a3b8;
	}

	.btn-group {
		display: flex;
		gap: 12px;
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
