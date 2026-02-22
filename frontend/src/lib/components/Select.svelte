<script lang="ts">
	let {
		label,
		options,
		value = $bindable(),
		id = crypto.randomUUID(),
		autofocus = false
	} = $props<{
		label: string;
		value: string;
		options: { value: string; label: string }[];
		id?: string;
		autofocus?: boolean;
	}>();

	function handleKeyDown(e: KeyboardEvent) {
		const index = options.findIndex((opt: any) => opt.value === value);

		if (e.key === 'ArrowDown' && index < options.length - 1) {
			e.preventDefault();
			value = options[index + 1].value;
		} else if (e.key === 'ArrowUp' && index > 0) {
			e.preventDefault();
			value = options[index - 1].value;
		}
	}

	function focusOnMount(node: HTMLSelectElement) {
		if (autofocus) node.focus();
	}
</script>

<div class="select-group">
	<label for={id}>{label}</label>

	<div class="select-wrapper">
		<select {id} bind:value use:focusOnMount onkeydown={handleKeyDown}>
			{#each options as option}
				<option value={option.value}>{option.label}</option>
			{/each}
		</select>
		<span class="chevron" aria-hidden="true">▼</span>
	</div>
</div>

<style>
	.select-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		width: 100%;
		font-family: sans-serif;
	}

	label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-muted, #94a3b8);
	}

	.select-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}

	select {
		width: 100%;
		appearance: none;
		background: var(--bg-primary, #1e293b);
		border: 1px solid var(--border, #334155);
		color: var(--text-main);
		padding: 0.75rem 1rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
	}

	select:focus {
		outline: 2px solid transparent;
		border-color: var(--accent, #7c3aed);
		box-shadow: 0 0 0 2px rgba(124, 58, 237, 0.2);
	}

	/* Separating the icon from the background-image for easier styling */
	.chevron {
		position: absolute;
		right: 1rem;
		pointer-events: none;
		font-size: 0.7rem;
		color: var(--text-muted, #94a3b8);
	}
</style>
