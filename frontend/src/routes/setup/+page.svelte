<script lang="ts">
	import { onMount } from 'svelte';
	import { fly, fade, scale } from 'svelte/transition';
	import { apiFetch } from '$lib/api';
	import IndexerStep from './IndexerStep.svelte';
	import TransferStep from './TransferStep.svelte';

	let currentStep = $state(0); // 0 = Loading, 1 = Indexer, 2 = Transfer, 3 = Success
	let steps = ['Indexer', 'Transfer', 'Library'];

	onMount(async () => {
		try {
			const [indexers, transfers] = await Promise.all([
				apiFetch<any[]>({ path: '/indexers' }),
				apiFetch<any[]>({ path: '/transfers' })
			]);

			if (indexers.length === 0) currentStep = 1;
			else if (transfers.length === 0) currentStep = 2;
			else currentStep = 3;
		} catch (e) {
			currentStep = 1; // Fallback
		}
	});

	const next = () => currentStep++;
</script>

<div class="setup-wrapper">
	{#if currentStep === 0}
		<div class="loader" out:fade>
			<div class="orbit"></div>
			<p>Initializing setup...</p>
		</div>
	{:else}
		<div class="container" in:scale={{ start: 0.95, duration: 400 }}>
			<header>
				<div class="brand">Animeman <span>Setup</span></div>
				<div class="dots">
					{#each steps as _, i}
						<div class="dot" class:active={currentStep >= i + 1}></div>
					{/each}
				</div>
			</header>

			<div class="content-area">
				{#if currentStep === 1}
					<div in:fly={{ x: 30, duration: 500 }} out:fly={{ x: -30, duration: 300 }}>
						<IndexerStep onNext={next} />
					</div>
				{:else if currentStep === 2}
					<div in:fly={{ x: 30, duration: 500 }} out:fly={{ x: -30, duration: 300 }}>
						<TransferStep onNext={next} />
					</div>
				{:else}
					<div class="success-screen" in:fly={{ y: 20 }}>
						<div class="icon-check">✓</div>
						<h1>You're all set!</h1>
						<p>Everything is configured and ready.</p>
						<button class="btn-primary" onclick={() => (window.location.href = '/')}>
							Enter Dashboard
						</button>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	:global(body) {
		background: radial-gradient(circle at top right, #1e293b, #0f172a);
		color: #f8fafc;
		font-family: 'Inter', system-ui, sans-serif;
		margin: 0;
	}

	.setup-wrapper {
		display: grid;
		place-items: center;
		min-height: 100vh;
		padding: 20px;
	}

	.container {
		width: 100%;
		max-width: 440px;
		background: rgba(30, 41, 59, 0.7);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 24px;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		overflow: hidden;
	}

	header {
		padding: 24px 32px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.brand {
		font-weight: 800;
		letter-spacing: -0.5px;
		font-size: 1.1rem;
	}
	.brand span {
		color: #38bdf8;
	}

	.dots {
		display: flex;
		gap: 8px;
	}
	.dot {
		width: 8px;
		height: 8px;
		background: #334155;
		border-radius: 50%;
		transition: 0.3s;
	}
	.dot.active {
		background: #38bdf8;
		box-shadow: 0 0 10px #38bdf8;
	}

	.content-area {
		padding: 32px;
		position: relative;
		min-height: 380px;
	}

	/* Success Screen */
	.success-screen {
		text-align: center;
	}
	.icon-check {
		width: 64px;
		height: 64px;
		background: #059669;
		color: white;
		font-size: 32px;
		display: grid;
		place-items: center;
		border-radius: 50%;
		margin: 0 auto 24px;
	}

	.btn-primary {
		width: 100%;
		padding: 14px;
		background: #38bdf8;
		border: none;
		border-radius: 12px;
		font-weight: 700;
		color: #0f172a;
		cursor: pointer;
		margin-top: 24px;
	}

	/* Simple Loader */
	.loader {
		text-align: center;
		color: #94a3b8;
	}
	.orbit {
		width: 40px;
		height: 40px;
		border: 3px solid #334155;
		border-top-color: #38bdf8;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto 16px;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
