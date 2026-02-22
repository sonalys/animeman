<script lang="ts">
	import { apiFetch } from '$lib/api';
	import {
		type OnboardingStatus,
		type ErrorResponse,
		type IndexerConfig as IndexerClientConfig,
		type TransferClientConfig,
		type WatchlistConfig,
		type FieldError,
		ERROR_MESSAGES
	} from '$lib/api/types';
	import Stepper, { type Step } from '$lib/components/MultiStep.svelte';
	import SegmentedControl from '$lib/components/SegmentedControl.svelte';
	import { slide } from 'svelte/transition';

	function createErrorState() {
		let fieldMap = $state<Record<string, FieldError>>({});
		let globalDetails = $state<string | null>(null);

		return {
			get: (field: string): string | undefined => {
				const err = fieldMap[field];
				if (!err) return;

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
			},

			clear: () => {
				fieldMap = {};
				globalDetails = null;
			}
		};
	}

	const errors = createErrorState();
	const onboardingStatus = apiFetch<OnboardingStatus>('/setup');

	let isLoading = $state(false);

	let formState = $state({
		indexingClient: {
			type: 'prowlarr',
			hostname: 'http://localhost:9696',
			auth: { type: 'none' }
		},
		transferClient: {
			type: 'qbittorrent',
			hostname: 'http://localhost:8080',
			auth: { type: 'none' }
		},
		watchlist: {
			externalID: '',
			syncFrequencySeconds: 900
		}
	} as {
		indexingClient: IndexerClientConfig;
		transferClient: TransferClientConfig;
		watchlist: WatchlistConfig;
	});

	const authOptions = [
		{ label: 'None', value: 'none' },
		{ label: 'API Key', value: 'apiKey' },
		{ label: 'Credentials', value: 'userPassword' }
	];

	const sourceOptions = [
		{ label: 'None', value: undefined },
		{ label: 'AniList', value: 'anilist' },
		{ label: 'MAL', value: 'mal' },
		{ label: 'Local', value: 'local' }
	];

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

	const getSteps = (status: OnboardingStatus): Step[] => {
		const steps = status.missingSteps.reduce<Step[]>((acc, step) => {
			switch (step) {
				case 'watchlist':
					acc.push({ name: 'Watchlist', component: watchlistStep });
					return acc;
				case 'transfer':
					acc.push({ name: 'Transfer', component: transferStep });
					return acc;
				case 'indexing':
					acc.push({ name: 'Indexing', component: indexerStep });
					return acc;
				default:
					return acc;
			}
		}, []);

		if (steps.length > 0) steps.push({ name: 'Finish', component: successStep });

		return steps;
	};
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
			id: 'indexingClientHostname',
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
				id: 'indexingClientAPIKey',
				label: 'API Key',
				validationName: 'auth.key',
				target: formState.indexingClient.auth,
				key: 'key',
				placeholder: ''
			})}
		{:else if formState.indexingClient.auth.type === 'userPassword'}
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
						await apiFetch('/indexing-clients', {
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
			id: 'transferClientHostname',
			label: 'Instance URL',
			validationName: 'hostname',
			target: formState.transferClient,
			key: 'hostname',
			placeholder: 'http://localhost:8080'
		})}

		<SegmentedControl
			bind:active={formState.transferClient.auth.type}
			options={authOptions}
			error={errors.get('auth.type')}
			name="Auth Type"
		/>

		{#if formState.transferClient.auth.type === 'apiKey'}
			{@render FormInput({
				id: 'transferClientAPIKey',
				label: 'API Key',
				validationName: 'auth.key',
				target: formState.transferClient.auth,
				key: 'key',
				placeholder: ''
			})}
		{:else if formState.transferClient.auth.type === 'userPassword'}
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
						await apiFetch('/transfer-clients', {
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

				<SegmentedControl
					bind:active={formState.watchlist.syncFrequencySeconds}
					options={frequencies}
					name="Sync Frequency"
				/>
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

{#await onboardingStatus}
	<p>Loading...</p>
{:then status}
	<Stepper steps={getSteps(status)} />
{/await}

<style>
	.error-banner {
		color: var(--error);
	}

	input.error {
		border: 1px solid var(--error);
		background: rgba(239, 68, 68, 0.05);
	}

	input.error:focus {
		outline: none;
		box-shadow: 0 0 0 0.2rem rgba(239, 68, 68, 0.2);
	}

	.error-msg {
		color: #ef4444;
		font-size: 0.75rem;
		margin-top: 0.2rem;
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
			transform: translateX(0.3rem);
		}
		75% {
			transform: translateX(-0.3rem);
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
		gap: 0 0.6rem;
		justify-content: center;
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid currentColor;
		border-bottom-color: transparent;
		border-radius: 50%;
		display: inline-block;
		animation: spin 0.6s linear infinite;
	}

	@keyframes shimmer {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}

	.step-content {
		justify-items: stretch;
		display: flex;
		flex-direction: column;
	}

	/* Styles for the inside of your snippets */
	.step-content h2 {
		font-size: 1.5rem;
		margin: 0;
	}

	.step-content p {
		margin: 0;
		margin-bottom: 1.2rem;
		color: #94a3b8;
	}

	.btn-group {
		display: flex;
		gap: 0.8rem;
		margin-top: 2rem;
	}

	.success-screen {
		text-align: center;
	}

	.icon-check {
		width: 4rem;
		height: 4rem;
		background: #059669;
		border-radius: 50%;
		display: grid;
		place-items: center;
		margin: 0 auto 16px;
		font-size: 2rem;
	}
</style>
