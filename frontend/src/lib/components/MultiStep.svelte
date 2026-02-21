<script lang="ts">
	import { type Snippet } from 'svelte';
	import { scale } from 'svelte/transition';

	interface Step {
		name: string;
		component: Snippet<[arg: { next: () => void; back: () => void }]>;
	}

	let { steps } = $props<{
		steps: Step[];
	}>();

	let currentStep = $state(0);
	let stepHeights = $state<number[]>([]);

	$effect(() => {
		if (stepHeights.length !== steps.length) {
			// We preserve existing heights if possible, or reset
			stepHeights = new Array(steps.length).fill(0);
		}
	});

	// This remains derived because it's a read-only calculation based on state
	let activeHeight = $derived(stepHeights[currentStep] ?? 0);

	const next = () => {
		if (currentStep < steps.length - 1) currentStep++;
	};
	const back = () => {
		if (currentStep > 0) currentStep--;
	};
</script>

<div class="setup-wrapper">
	<div class="container" in:scale={{ start: 0.9, duration: 400 }}>
		<header>
			<div class="brand">Animeman <span>{steps[currentStep].name}</span></div>
			<div class="dots">
				{#each steps as _, i}
					<div class="dot" class:active={currentStep >= i} class:current={currentStep === i}></div>
				{/each}
			</div>
		</header>

		<div class="content-area" style="height: {activeHeight}px;">
			<div class="step-tray" style="transform: translateX(-{currentStep * 100}%);">
				{#each steps as step, i}
					<div class="step-wrapper" bind:clientHeight={stepHeights[i]}>
						{@render step.component({ next, back })}
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style>
	.setup-wrapper {
		display: grid;
		place-items: center;
		min-height: 100dvh;
		font-family: 'Inter', system-ui, sans-serif;
		color: #f8fafc;
		background-color: transparent;
	}

	.container {
		width: 100%;
		max-width: 460px;
		background: rgba(30, 41, 59, 0.6);
		backdrop-filter: blur(16px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 28px;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		overflow: hidden;
	}

	header {
		padding: 24px 32px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: var(--bg-secondary);
	}

	.brand {
		font-weight: 800;
		font-size: 1.1rem;
		letter-spacing: -0.5px;
	}
	.brand span {
		color: var(--accent);
		text-transform: uppercase;
		font-size: 0.8rem;
		margin-left: 4px;
		opacity: 0.8;
	}

	.dots {
		display: flex;
		gap: 6px;
	}
	.dot {
		width: 6px;
		height: 6px;
		background: #334155;
		border-radius: 10px;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}
	.dot.active {
		background: var(--accent);
	}
	.dot.current {
		width: 20px;
		box-shadow: 0 0 12px rgba(56, 189, 248, 0.4);
	}

	.content-area {
		position: relative;
		overflow: hidden;
		transition: height 0.2s linear;
		width: 100%;
	}

	.step-tray {
		display: flex;
		transition: transform 0.6s cubic-bezier(0.4, 0, 0.2, 1);
		align-items: flex-start;
	}

	.step-wrapper {
		min-width: 100%;
		padding: 32px;
		box-sizing: border-box;
	}
</style>
