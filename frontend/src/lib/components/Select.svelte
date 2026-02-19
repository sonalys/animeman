<script lang="ts">
	export let label: string;
	export let value: string;
	export let options: { value: string; label: string }[];
	export let id: string = Math.random().toString(36).substring(2, 9);
	export let autofocus: boolean = false;

	function handleKeyDown(e: KeyboardEvent) {
		const index = options.findIndex((opt) => opt.value === value);

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			const next = options[index + 1];
			if (next) value = next.value;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			const prev = options[index - 1];
			if (prev) value = prev.value;
		}
	}

	function focusOnMount(node: HTMLSelectElement) {
		if (autofocus) node.focus();
	}
</script>

<div class="select-group">
	<label for={id}>{label}</label>
	<select {id} bind:value use:focusOnMount on:keydown={handleKeyDown}>
		{#each options as option}
			<option value={option.value}>{option.label}</option>
		{/each}
	</select>
</div>

<style>
	.select-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		width: 100%;
	}

	label {
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	select {
		background: var(--bg-primary);
		border: 1px solid var(--border);
		color: white;
		padding: 0.75rem;
		border-radius: 8px;
		cursor: pointer;
		appearance: none; /* Removes default OS arrow to allow for custom styling */
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='white'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 1rem center;
		background-size: 1rem;
		padding-right: 2.5rem;
		transition: all 0.2s ease;
	}

	select:focus {
		outline: none;
		border-color: var(--accent);
		box-shadow:
			0 0 0 2px rgba(124, 58, 237, 0.2),
			0 0 8px var(--accent);
	}
</style>
