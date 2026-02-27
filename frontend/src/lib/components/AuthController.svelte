<script lang="ts">
	import { apiFetch } from '$lib/api';
	import { userId } from '$lib/userStore';
	import type { ErrorResponse, FieldError } from '$lib/api/types';
	import { scale } from 'svelte/transition';

	let username = $state('');
	let password = $state('');
	let passwordConfirmation = $state('');
	let loading = $state(false);
	let errorMessage = $state('');
	let fieldErrors: Record<string, string> = $state({});

	let currentStep = $state(0);
	let stepHeights = $state([0, 0]);
	let activeHeight = $derived(stepHeights[currentStep]);

	const next = () => currentStep++;
	const back = () => currentStep--;

	async function handleSubmit() {
		loading = true;
		errorMessage = '';
		fieldErrors = {};

		if (currentStep === 1 && password !== passwordConfirmation) {
			fieldErrors['password'] = 'Password confirmation mismatch';
			return;
		}

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
							<input
								id="pass"
								type="password"
								bind:value={passwordConfirmation}
								placeholder="••••••••"
							/>
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

	.content-area {
		position: relative;
		overflow: hidden;
		transition: height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		width: 100%;
	}

	.step-tray {
		display: flex;
		width: 100%;
		transition: transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
		will-change: transform;
		align-items: flex-start;
	}

	.step-wrapper {
		min-width: 100%;
		width: 100%;
		box-sizing: border-box;
		padding: 2rem;
	}

	.setup-wrapper {
		display: grid;
		place-items: center;
	}

	.container {
		width: 100%;
		max-width: 460px;
		background: var(--bg-terciary);
		backdrop-filter: blur(16px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 28px;
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		overflow: hidden;
	}

	header {
		padding: 1.5rem 2rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: var(--bg-secondary);
	}

	.brand {
		font-weight: 800;
		font-size: 1.4rem;
	}

	.brand span {
		color: #38bdf8;
		font-size: 1rem;
	}

	.btn-primary {
		margin-top: 2rem;
	}

	p {
		margin: 0;
		color: #94a3b8;
		font-size: 0.95rem;
		line-height: 1.5;
	}

	.input-group {
		margin-bottom: 10px;
	}
</style>
