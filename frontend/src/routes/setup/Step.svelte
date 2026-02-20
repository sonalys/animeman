<script lang="ts">
	// Define the interface for your "events"
	interface Props {
		onLoaded: () => void; // No argument
		onChange: (value: string) => void; // Required string
	}

	// Use the $props() rune
	let { onLoaded, onChange }: Props = $props();

	function handleAction() {
		onChange('New Value'); // Just call it like a function
	}
</script>

<button onclick={handleAction}>Update</button>
<button onclick={() => onLoaded()}>Done</button>
