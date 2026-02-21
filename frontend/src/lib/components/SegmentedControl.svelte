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

	let activeIndex = $derived(options.findIndex((opt: any) => opt.value === active));

	function handleKeyDown(e: KeyboardEvent) {
		if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight') return;

		e.preventDefault();
		let newIndex = activeIndex;

		if (e.key === 'ArrowLeft') {
			newIndex = (activeIndex - 1 + options.length) % options.length;
		} else if (e.key === 'ArrowRight') {
			newIndex = (activeIndex + 1) % options.length;
		}

		const nextOption = options[newIndex];
		if (active !== nextOption.value) error = undefined;
		active = nextOption.value;

		(e.currentTarget as HTMLElement).querySelectorAll('button')[newIndex]?.focus();
	}
</script>

<div onkeydown={handleKeyDown} role="radiogroup" tabindex="0" class="container">
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
				role="radio"
				aria-checked={active === option.value}
				class:active={active === option.value}
				tabindex={active === option.value ? 0 : -1}
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
		padding: 0.25rem;
		border-radius: 14px;
		outline: none;
		isolation: isolate;
	}

	.pill {
		position: absolute;
		top: 0.3rem;
		bottom: 0.3rem;
		left: 0.3rem;
		width: calc((100% - 0.5rem) / var(--total));
		background: var(--accent);
		border-radius: 10px;
		transform: translateX(calc(100% * var(--index)));
		transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
	}

	button {
		background: transparent;
		border: none;
		color: var(--text-muted);
		padding: 0.6rem;
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

	.container {
		outline: none;
	}

	.container:focus-within .pill {
		filter: brightness(1.1);
		box-shadow: 0 0px 0.8rem rgb(from var(--accent) r g b / 0.3);
	}
</style>
