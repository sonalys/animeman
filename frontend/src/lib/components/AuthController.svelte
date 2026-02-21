<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { userId } from '$lib/userStore';
	import type { ErrorResponse, FieldError } from '$lib/api/types';
	import { scale } from 'svelte/transition';

	let username = $state('');
	let password = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let fieldErrors: Record<string, string> = $state({});

	let steps = ['Login', 'Register'];
	let currentStep = $state(0);
	let stepHeights = $state([0, 0]);
	let activeHeight = $derived(stepHeights[currentStep]);

	const next = () => currentStep++;
	const back = () => currentStep--;

	async function handleSubmit() {
		loading = true;
		errorMessage = '';
		fieldErrors = {};

		const endpoint = currentStep == 0 ? '/authentication/login' : '/register';

		try {
			await apiFetch(endpoint, { method: 'POST', body: { username, password } });

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

<div class="setup-wrapper">
	<div class="container" in:scale={{ start: 0.9, duration: 400 }}>
		<header>
			<div class="brand">Animeman <span>Authentication</span></div>
		</header>

		<div class="content-area" style="height: {activeHeight}px;">
			<div class="step-tray" style="transform: translateX(-{currentStep * 100}%);">
				<div class="step-wrapper" bind:clientHeight={stepHeights[0]} inert={currentStep !== 0}>
					<form onsubmit={handleSubmit}>
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

						<button type="submit" disabled={loading} class="btn-primary">Authenticate</button>
					</form>

					<p class="footer">
						New to the system?
						<button class="btn-link" onclick={next}>Create Account</button>
					</p>
				</div>

				<div class="step-wrapper" bind:clientHeight={stepHeights[1]} inert={currentStep !== 1}>
					<form onsubmit={handleSubmit}>
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

						<div class="input-group">
							<label for="pass">Confirm Password</label>
							<input id="pass" type="password" bind:value={password} placeholder="••••••••" />
							{#if fieldErrors.password}<span class="err">{fieldErrors.password}</span>{/if}
						</div>

						<button type="submit" disabled={loading} class="btn-primary">Register</button>
					</form>

					<p class="footer">
						Already registered?
						<button class="btn-link" onclick={back}>Login</button>
					</p>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.btn-link {
		background: none;
		border: none;
		color: var(--accent);
		cursor: pointer;
		text-decoration: underline;
	}

	/* The outer box that hides the overflow */
	.content-area {
		position: relative;
		overflow: hidden;
		transition: height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		width: 100%;
	}

	/* The long horizontal strip containing all steps */
	.step-tray {
		display: flex;
		width: 100%;
		transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
		will-change: transform;
		align-items: flex-start; /* Ensures height is measured correctly per step */
	}

	/* Each individual step is exactly 100% of the container's width */
	.step-wrapper {
		min-width: 100%;
		width: 100%;
		box-sizing: border-box;
		padding: 32px; /* Move padding here so height is measured accurately */
	}

	/* Rest of your existing pretty styles */
	:global(body) {
		color: #f8fafc;
		font-family: 'Inter', system-ui, sans-serif;
		margin: 0;
	}

	.setup-wrapper {
		display: grid;
		place-items: center;
		min-height: 100dvh;
	}

	.container {
		width: 100%;
		max-width: 440px;
		background: rgba(30, 41, 59, 0.7);
		backdrop-filter: blur(12px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 24px;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
	}

	header {
		padding: 24px 32px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.brand {
		font-weight: 800;
		font-size: 1.1rem;
	}
	.brand span {
		color: #38bdf8;
	}

	.btn-primary {
		width: 100%;
		padding: 14px;
		background: #38bdf8;
		border: none;
		border-radius: 12px;
		font-weight: 700;
		color: #0f172a;
		cursor: pointer;
		margin-top: 24px;
	}

	p {
		margin: 0 0 32px;
		color: #94a3b8;
		font-size: 0.95rem;
		line-height: 1.5;
	}

	label {
		display: block;
		font-size: 0.75rem;
		font-weight: 700;
		color: #64748b;
		text-transform: uppercase;
		margin-bottom: 8px;
		letter-spacing: 0.5px;
	}

	input {
		width: 100%;
		background: #0f172a;
		border: 1px solid #334155;
		padding: 12px 16px;
		border-radius: 12px;
		color: white;
		font-size: 1rem;
		box-sizing: border-box;
		transition: 0.2s;
	}
	input:focus {
		border-color: #38bdf8;
		outline: none;
		box-shadow: 0 0 0 4px rgba(56, 189, 248, 0.1);
	}

	button {
		width: 100%;
		padding: 14px;
		background: #38bdf8;
		border: none;
		border-radius: 12px;
		font-weight: 700;
		color: #0f172a;
		cursor: pointer;
		margin-top: 12px;
		transition: 0.2s;
	}
	button:disabled {
		background: #1e293b;
		color: #475569;
		cursor: not-allowed;
	}
	button:hover:not(:disabled) {
		transform: translateY(-1px);
		background: #7dd3fc;
	}
</style>
