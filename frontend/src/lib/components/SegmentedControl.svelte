<script lang="ts">
	let {
		options,
		active = $bindable(),
		name
	} = $props<{
		options: { label: string; value: any }[];
		active: any;
		name: string;
	}>();

	// Calculate index for the sliding logic
	let activeIndex = $derived(options.findIndex((opt: any) => opt.value === active));
</script>

<div>
	<label for="group">{name}</label>
	<div
		id="group"
		class="segmented-control"
		style="--total: {options.length}; --index: {activeIndex};"
	>
		<div class="pill"></div>

		{#each options as option}
			<button
				type="button"
				class:active={active === option.value}
				onclick={() => (active = option.value)}
			>
				{option.label}
			</button>
		{/each}
	</div>
</div>

<style>
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
