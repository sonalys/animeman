<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { userId } from '$lib/userStore';
	import type { ErrorResponse, FieldError } from '$lib/types';

	let isLogin = true; // State toggle
	let username = '';
	let password = '';
	let loading = false;
	let errorMessage = '';
	let fieldErrors: Record<string, string> = {};

	async function handleSubmit() {
		loading = true;
		errorMessage = '';
		fieldErrors = {};

		const endpoint = isLogin ? '/authentication/login' : '/register';

		try {
			// 1. Perform Auth Action
			await apiFetch(endpoint, 'POST', { username, password });

			// 2. Refresh Auth State
			const auth = await apiFetch<{ userID: string }>('/authentication/whoami');
			userId.set(auth.userID);
		} catch (e) {
			const err = e as ErrorResponse;
			errorMessage = err.details || 'Authentication failed';
			err.fieldErrors?.forEach((f: FieldError) => (fieldErrors[f.field] = f.message));
		} finally {
			loading = false;
		}
	}
</script>

<div class="auth-card">
	<h2>{isLogin ? 'Welcome Back' : 'Create Identity'}</h2>

	<form on:submit|preventDefault={handleSubmit}>
		<div class="input-group">
			<label for="user">Username</label>
			<input id="user" bind:value={username} placeholder="Enter username..." />
			{#if fieldErrors.username}<span class="err">{fieldErrors.username}</span>{/if}
		</div>

		<div class="input-group">
			<label for="pass">Password</label>
			<input id="pass" type="password" bind:value={password} placeholder="••••••••" />
			{#if fieldErrors.password}<span class="err">{fieldErrors.password}</span>{/if}
		</div>

		<button type="submit" disabled={loading} class="btn-primary">
			{loading ? 'Processing...' : isLogin ? 'Authenticate' : 'Register'}
		</button>
	</form>

	<p class="footer">
		{isLogin ? 'New to the system?' : 'Already registered?'}
		<button class="btn-link" on:click={() => (isLogin = !isLogin)}>
			{isLogin ? 'Create Account' : 'Login'}
		</button>
	</p>
</div>

<style>
	.auth-card {
		background: var(--bg-secondary);
		padding: 2.5rem;
		border-radius: 16px;
		box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.3);
		border: 1px solid var(--border);
		width: 100%;
		max-width: 400px;
	}
	h2 {
		margin-bottom: 1.5rem;
		color: var(--text-main);
	}
	.input-group {
		text-align: left;
		margin-bottom: 1rem;
	}
	label {
		display: block;
		font-size: 0.8rem;
		color: var(--text-muted);
		margin-bottom: 0.4rem;
	}

	input {
		width: 100%;
		padding: 0.8rem;
		background: var(--bg-primary);
		border: 1px solid var(--border);
		border-radius: 8px;
		color: white;
		box-sizing: border-box;
	}
	input:focus {
		outline: 2px solid var(--accent);
		border-color: transparent;
	}

	.btn-primary {
		width: 100%;
		padding: 0.8rem;
		background: var(--accent);
		color: white;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
		margin-top: 1rem;
		transition: background 0.2s;
	}
	.btn-primary:hover {
		background: var(--accent-hover);
	}

	.btn-link {
		background: none;
		border: none;
		color: var(--accent);
		cursor: pointer;
		text-decoration: underline;
	}
	.err {
		color: var(--error);
		font-size: 0.75rem;
		margin-top: 0.25rem;
		display: block;
	}
	.footer {
		margin-top: 1.5rem;
		color: var(--text-muted);
		font-size: 0.9rem;
	}
</style>
