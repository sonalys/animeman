<script lang="ts">
	import { apiFetch } from '$lib/api';
	import type {
		ErrorResponse,
		FieldError,
		IndexerConfig as IndexerClientConfig,
		TransferClientConfig,
		WatchlistConfig
	} from '$lib/api/types';
	import Stepper from '$lib/components/MultiStep.svelte';
	import SegmentedControl from '$lib/components/SegmentedControl.svelte';
	import { slide } from 'svelte/transition';

	let isLoading = $state(false);

	function createErrorState() {
		let fieldMap = $state<Record<string, FieldError>>({});
		let globalDetails = $state<string | null>(null);

		const ERROR_MESSAGES: Record<string, string> = {
			alreadyExists: 'Already in use',
			minLength: 'Too short',
			maxLength: 'Too long',
			required: 'Required',
			invalidFormat: 'Invalid format',
			invalid: 'Invalid value',
			unknown: 'Unexpected error'
		};

		return {
			get: (field: string) => {
				const err = fieldMap[field];
				if (!err) return null;

				if (err.message.length == 0) {
					return ERROR_MESSAGES[err.code];
				}

				return `${ERROR_MESSAGES[err.code]}: ${err.message}`;
			},
			get details() {
				return globalDetails;
			},
			get hasErrors() {
				return Object.keys(fieldMap).length > 0;
			},

			set: (response: ErrorResponse) => {
				globalDetails = response.details || null;

				fieldMap = (response.fieldErrors || []).reduce(
					(acc, err) => {
						acc[err.field] = err;
						return acc;
					},
					{} as Record<string, FieldError>
				);

				console.log(fieldMap);
			},

			clear: () => {
				fieldMap = {};
				globalDetails = null;
			}
		};
	}

	const errors = createErrorState();

	let formState = $state({
		indexingClient: {
			type: 'prowlarr',
			hostname: 'http://localhost:9696',
			auth: { type: 'apiKey', key: '' }
		},
		transferClient: {
			type: 'qbittorrent',
			hostname: 'http://localhost:8080',
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

	interface FormInputProps {
		label: string;
		id: string;
		type?: string;
		placeholder?: string;
		validationName: string;
		target: any;
		key: string;
	}
</script>

{#snippet FormInput({
	id,
	label,
	type = 'text',
	target,
	key,
	placeholder,
	validationName
}: FormInputProps)}
	<div class="field-group">
		<label for={id}>{label}</label>
		<input
			{id}
			{type}
			{placeholder}
			class:error={errors.get(validationName)}
			value={target[key]}
			oninput={(e: any) => (target[key] = e.currentTarget.value)}
		/>
		{#if errors.get(validationName)}
			<span class="error-msg" transition:slide={{ duration: 200 }}>
				{errors.get(validationName)}
			</span>
		{/if}
	</div>
{/snippet}

{#snippet indexerStep({ back, next }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Configure Indexer</h2>
			<p>For finding your favorite episode</p>
		</div>

		{@render FormInput({
			id: 'url',
			label: 'Instance URL',
			validationName: 'hostname',
			target: formState.indexingClient,
			key: 'hostname',
			placeholder: 'http://localhost:9696'
		})}

		<SegmentedControl
			bind:active={formState.indexingClient.auth.type}
			options={authOptions}
			name="Auth Type"
		/>

		{#if formState.indexingClient.auth.type === 'apiKey'}
			{@render FormInput({
				id: 'apiKey',
				label: 'API Key',
				validationName: 'auth.key',
				target: formState.indexingClient.auth,
				key: 'key',
				placeholder: ''
			})}
		{:else}
			<div class="field-row">
				{@render FormInput({
					id: 'username',
					label: 'Username',
					validationName: 'auth.username',
					target: formState.indexingClient.auth,
					key: 'username',
					placeholder: ''
				})}
				{@render FormInput({
					id: 'password',
					label: 'Password',
					validationName: 'auth.password',
					target: formState.indexingClient.auth,
					key: 'password',
					placeholder: ''
				})}
			</div>
		{/if}

		<div class="btn-group">
			{#if back}
				<button class="btn-ghost" onclick={back}>Back</button>
			{/if}
			<button
				class="btn-primary"
				class:loading={isLoading}
				disabled={isLoading}
				onclick={async () => {
					isLoading = true;

					try {
						await apiFetch('/indexing-clients/test', {
							method: 'POST',
							body: formState.indexingClient
						});

						next();
					} catch (error) {
						const data = error as ErrorResponse;
						errors.set(data);
					} finally {
						isLoading = false;
					}
				}}
			>
				{#if isLoading}
					<div class="loader-container">
						<span class="spinner"></span>
						<span>Verifying...</span>
					</div>
				{:else}
					<span>Continue</span>
				{/if}
			</button>
		</div>

		{#if errors.details}
			<div class="error-banner" transition:slide>
				<strong>System Error</strong>
				<p>{errors.details}</p>
			</div>
		{/if}
	</div>
{/snippet}

{#snippet transferStep({ next, back }: any)}
	<div class="step-content">
		<div class="step-header">
			<h2>Configure Transfer Client</h2>
			<p>For downloading joy</p>
		</div>

		{@render FormInput({
			id: 'url',
			label: 'Instance URL',
			validationName: 'hostname',
			target: formState.transferClient,
			key: 'hostname',
			placeholder: 'http://localhost:8080'
		})}

		<SegmentedControl
			bind:active={formState.transferClient.auth.type}
			options={authOptions}
			name="Auth Type"
		/>

		{#if formState.transferClient.auth.type === 'apiKey'}
			{@render FormInput({
				id: 'apiKey',
				label: 'API Key',
				validationName: 'auth.key',
				target: formState.transferClient.auth,
				key: 'key',
				placeholder: ''
			})}
		{:else}
			<div class="field-row">
				{@render FormInput({
					id: 'username',
					label: 'Username',
					validationName: 'auth.username',
					target: formState.transferClient.auth,
					key: 'username',
					placeholder: ''
				})}
				{@render FormInput({
					id: 'password',
					label: 'Password',
					validationName: 'auth.password',
					target: formState.transferClient.auth,
					key: 'password',
					placeholder: ''
				})}
			</div>
		{/if}

		<div class="btn-group">
			{#if back}
				<button class="btn-ghost" onclick={back}>Back</button>
			{/if}
			<button
				class="btn-primary"
				class:loading={isLoading}
				disabled={isLoading}
				onclick={async () => {
					isLoading = true;

					try {
						await apiFetch('/transfer-clients/test', {
							method: 'POST',
							body: formState.transferClient
						});

						next();
					} catch (error) {
						const data = error as ErrorResponse;
						errors.set(data);
					} finally {
						isLoading = false;
					}
				}}
			>
				{#if isLoading}
					<div class="loader-container">
						<span class="spinner"></span>
						<span>Verifying...</span>
					</div>
				{:else}
					<span>Continue</span>
				{/if}
			</button>
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
			{#if back}
				<button class="btn-ghost" onclick={back}>Back</button>
			{/if}
			<button class="btn-primary" onclick={next}>Continue</button>
		</div>
	</div>
{/snippet}

{#snippet successStep({ next, back }: any)}
	<div class="success-screen">
		<div class="icon-check">✓</div>
		<h1>Ready!</h1>
		<div class="btn-group">
			{#if back}
				<button class="btn-ghost" onclick={back}>Back</button>
			{/if}
			<button class="btn-primary" onclick={next}> Continue </button>
		</div>
	</div>
{/snippet}

<Stepper
	steps={[
		{ name: 'Transfer', component: transferStep },
		{ name: 'Indexing', component: indexerStep },
		{ name: 'Watchlist', component: watchlistStep },
		{ name: 'Finish', component: successStep }
	]}
/>

<style>
	input.error {
		border-color: #ef4444;
		background: rgba(239, 68, 68, 0.05);
	}

	input.error:focus {
		outline: none;
		box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
	}

	.error-msg {
		color: #ef4444;
		font-size: 0.75rem;
		margin-top: 4px;
		display: block;
		font-weight: 500;
		animation: shake 0.2s ease-in-out;
	}

	@keyframes shake {
		0%,
		100% {
			transform: translateX(0);
		}
		25% {
			transform: translateX(4px);
		}
		75% {
			transform: translateX(-4px);
		}
	}

	.btn-primary.loading {
		cursor: wait;
		transform: scale(0.98);
		overflow: hidden;
		box-shadow:
			0 0 0 2px rgba(var(--accent-rgb, 124, 58, 237), 0.1),
			0 0 8px var(--accent);
		transform: box-shadow 1s;
	}

	.btn-primary.loading::before {
		z-index: -1;
		content: ''; /* Must be an empty string */
		display: block; /* Ensure it has dimensions */
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, var(--accent-hover), transparent);
		/* Start the element completely to the left of the button */
		transform: translateX(-100%);
		animation: shimmer 1.5s infinite linear;
		pointer-events: none; /* Ensure it doesn't block clicks */
	}

	.loader-container {
		z-index: 1;
		display: flex;
		gap: 0 10px;
		justify-content: center;
	}

	.spinner {
		width: 14px;
		height: 14px;
		border: 2px solid currentColor;
		border-bottom-color: transparent;
		border-radius: 50%;
		display: inline-block;
		animation: rotation 0.6s linear infinite;
	}

	@keyframes rotation {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	@keyframes shimmer {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	/* Subtle hover for the non-loading state */
	.btn-primary:hover:not(:disabled) {
		filter: brightness(1.1);
		box-shadow: 0 4px 12px rgba(56, 189, 248, 0.3);
	}

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
