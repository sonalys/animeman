<script lang="ts">
	import { slide } from 'svelte/transition';

	let {
		options,
		active = $bindable(),
		name,
		error = $bindable()
	} = $props<{
		options: { label: string; value: any }[];
		active: any;
		name: string;
		error?: string;
	}>();

	// Calculate index for the sliding logic
	let activeIndex = $derived(options.findIndex((opt: any) => opt.value === active));
</script>

<div>
	<label for="group">{name}</label>
	<div
		id="group"
		class="segmented-control"
		class:error={!!error}
		style="--total: {options.length}; --index: {activeIndex};"
	>
		<div class="pill"></div>

		{#each options as option}
			<button
				type="button"
				class:active={active === option.value}
				onclick={() => {
					if (active !== option.value) error = undefined;
					active = option.value;
				}}
			>
				{option.label}
			</button>
		{/each}
	</div>
	{#if error}
		<span class="error-msg" transition:slide={{ duration: 200 }}>
			{error}
		</span>
	{/if}
</div>

<style>
	.error {
		> .pill {
			background: var(--error);
		}

		border: solid 1px var(--error);
		background: rgba(239, 68, 68, 0.05);
	}

	.segmented-control.error:hover .pill {
		background: rgb(255, 179, 179) !important;
	}

	.error-msg {
		color: var(--error);
		font-size: 0.75rem;
		margin-top: 4px;
		display: block;
		font-weight: 500;
		animation: shake 0.2s ease-in-out;
	}

	.segmented-control {
		position: relative;
		display: grid;
		grid-template-columns: repeat(var(--total), 1fr);
		background: var(--bg-primary);
		padding: 4px;
		border-radius: 14px;
		isolation: isolate;
	}

	.pill {
		position: absolute;
		top: 4px;
		bottom: 4px;
		left: 4px;
		width: calc((100% - 8px) / var(--total));
		background: var(--accent);
		border-radius: 10px;
		transform: translateX(calc(100% * var(--index)));
		transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	button {
		background: transparent;
		border: none;
		color: var(--text-muted);
		padding: 10px 4px;
		border-radius: 10px;
		cursor: pointer;
		font-weight: 600;
		font-size: 0.9rem;
		transition: color 0.3s ease;
		z-index: 1;
		white-space: nowrap;
	}

	button.active {
		color: var(--bg-primary);
	}

	.segmented-control:hover .pill {
		background: var(--accent-hover);
	}
</style>
