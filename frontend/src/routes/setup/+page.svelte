<script lang="ts">
	import { onMount } from 'svelte';
	import { scale } from 'svelte/transition';
	import { apiFetch } from '$lib/api';
	import IndexerStep from './IndexerStep.svelte';
	import TransferStep from './TransferStep.svelte';
	import type { IndexerConfig } from '$lib/api/types';

	let steps = ['Indexer', 'Transfer', 'Library'];
	let currentStep = $state(0);

	let stepHeights = $state([0, 0, 0]);
	let activeHeight = $derived(stepHeights[currentStep]);

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

	const next = () => currentStep++;
	const back = () => currentStep--;
</script>

<div class="setup-wrapper">
	<div class="container" in:scale={{ start: 0.9, duration: 400 }}>
		<header>
			<div class="brand">Animeman <span>Setup</span></div>
			<div class="dots">
				{#each steps as _, i}
					<div class="dot" class:active={currentStep >= i}></div>
				{/each}
			</div>
		</header>

		<div class="content-area" style="height: {activeHeight}px;">
			<div class="step-tray" style="transform: translateX(-{currentStep * 100}%);">
				<div class="step-wrapper" bind:clientHeight={stepHeights[0]}>
					<IndexerStep bind:formState={formState.indexingClient} onNext={next} />
				</div>

				<div class="step-wrapper" bind:clientHeight={stepHeights[1]}>
					<TransferStep bind:formState={formState.indexingClient} onNext={next} onBack={back} />
				</div>

				<div class="step-wrapper success-screen" bind:clientHeight={stepHeights[2]}>
					<div class="icon-check">✓</div>
					<h1>You're all set!</h1>
					<p>Everything is configured and ready.</p>
					<button class="btn-primary" onclick={() => (window.location.href = '/')}>
						Enter Dashboard
					</button>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	/* The outer box that hides the overflow */
	.content-area {
		position: relative;
		overflow: hidden;
		transition: height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		width: 100%;
	}

	/* The long horizontal strip containing all steps */
	.step-tray {
		display: flex;
		width: 100%;
		transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
		will-change: transform;
		align-items: flex-start; /* Ensures height is measured correctly per step */
	}

	/* Each individual step is exactly 100% of the container's width */
	.step-wrapper {
		min-width: 100%;
		width: 100%;
		box-sizing: border-box;
		padding: 32px; /* Move padding here so height is measured accurately */
	}

	/* Rest of your existing pretty styles */
	:global(body) {
		background: radial-gradient(circle at top right, #1e293b, #0f172a);
		color: #f8fafc;
		font-family: 'Inter', system-ui, sans-serif;
		margin: 0;
	}

	.setup-wrapper {
		display: grid;
		place-items: center;
		min-height: 100dvh;
	}

	.container {
		width: 100%;
		max-width: 440px;
		background: rgba(30, 41, 59, 0.7);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 24px;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
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
</style>
