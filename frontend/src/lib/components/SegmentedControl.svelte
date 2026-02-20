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
	label {
		display: block;
		font-size: 0.75rem;
		font-weight: 700;
		color: #64748b;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin-bottom: 5px;
	}

	.segmented-control {
		position: relative;
		display: grid;
		/* Create a column for every option */
		grid-template-columns: repeat(var(--total), 1fr);
		background: rgba(15, 23, 42, 0.6);
		padding: 4px;
		border-radius: 14px;
		border: 1px solid rgba(255, 255, 255, 0.05);
		isolation: isolate;
	}

	.pill {
		position: absolute;
		top: 4px;
		bottom: 4px;
		left: 4px;
		width: calc((100% - 8px) / var(--total));
		background: #38bdf8;
		border-radius: 10px;
		transform: translateX(calc(100% * var(--index)));
		transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		z-index: -1;
	}

	button {
		background: transparent;
		border: none;
		color: #94a3b8;
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
		color: #0f172a;
	}
</style>
