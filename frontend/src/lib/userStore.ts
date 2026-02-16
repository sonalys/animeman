import { writable } from 'svelte/store';

// Holds the UUID if logged in, null if guest
export const userId = writable<string | null>(null);
// Track if we've finished the initial check to avoid "flicker"
export const authChecked = writable(false);