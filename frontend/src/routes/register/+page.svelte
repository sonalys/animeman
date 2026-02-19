<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { userId } from '$lib/userStore';
	import { goto } from '$app/navigation';
	import type { AuthResponse, ErrorResponse, UserRegistration } from '$lib/api/types';

	// Strictly typed state based on your OpenAPI schema
	let username = '';
	let password = '';

	// Error handling state
	let globalError = '';
	let fieldErrors: Record<string, string> = {};
	let loading = false;

	async function handleRegister() {
		loading = true;
		globalError = '';
		fieldErrors = {};

		try {
			await apiFetch<{ id: string }>('/register', { method: 'POST', body: { username, password } });

			const auth = await apiFetch<AuthResponse>('/authentication/whoami');

			userId.set(auth.userID);

			goto('/');
		} catch (e) {
			const err = e as ErrorResponse;
			globalError = err.details || 'An unexpected error occurred.';

			if (err.fieldErrors) {
				err.fieldErrors.forEach((fe) => {
					fieldErrors[fe.field] = fe.message;
				});
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="register-container">
	<h2>Create your Animeman Account</h2>

	<form on:submit|preventDefault={handleRegister}>
		<div class="field">
			<label for="username">Username</label>
			<input
				id="username"
				type="text"
				bind:value={username}
				placeholder="jdoe_secure"
				pattern="^[a-zA-Z0-9_]+$"
				required
				disabled={loading}
			/>
			{#if fieldErrors.username}
				<span class="error-text">{fieldErrors.username}</span>
			{/if}
		</div>

		<div class="field">
			<label for="password">Password</label>
			<input
				id="password"
				type="password"
				bind:value={password}
				minlength={8}
				maxlength={72}
				required
				disabled={loading}
			/>
			{#if fieldErrors.password}
				<span class="error-text">{fieldErrors.password}</span>
			{/if}
		</div>

		{#if globalError}
			<p class="error-banner">{globalError}</p>
		{/if}

		<button type="submit" disabled={loading}>
			{loading ? 'Registering...' : 'Sign Up'}
		</button>
	</form>
</div>

<style>
	.register-container {
		max-width: 400px;
		margin: 2rem auto;
		padding: 1rem;
		border: 1px solid #ccc;
		border-radius: 8px;
	}
	.field {
		margin-bottom: 1.2rem;
		display: flex;
		flex-direction: column;
	}
	.error-text {
		color: #d32f2f;
		font-size: 0.85rem;
		margin-top: 0.25rem;
	}
	.error-banner {
		background: #ffebee;
		color: #c62828;
		padding: 0.5rem;
		border-radius: 4px;
	}
	button {
		width: 100%;
		padding: 0.75rem;
		background: #4a90e2;
		color: white;
		border: none;
		cursor: pointer;
	}
	button:disabled {
		background: #ccc;
	}
</style>
