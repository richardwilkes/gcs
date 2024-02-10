<!--
  - Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang='ts'>
	import { apiPrefix } from '$lib/dev.ts';
	import { session, updateSessionFromResponse } from '$lib/session.ts';
	import { page } from '$lib/page.ts';
	import { onMount } from 'svelte';

	let form: HTMLFormElement;
	let nameInput: HTMLInputElement;
	let passwordInput: HTMLInputElement;
	let disabled = true;
	let nameEmpty = true;
	let passwordEmpty = true;
	let errorMsg = '';

	function submit(event: Event) {
		event.preventDefault();
		fetch(apiPrefix('/login'), {
			method: 'POST',
			body: new FormData(form),
			cache: 'no-store'
		}).then(rsp => {
			if (!rsp.ok) {
				console.log(rsp.status + ' ' + rsp.statusText);
				handleFailure();
			} else {
				updateSessionFromResponse(rsp);
				if ($session) {
					const prev = $page;
					prev.ID = prev.NextID;
					$page = prev;
				} else {
					handleFailure();
				}
			}
		}).catch(error => {
			console.log(error);
			handleFailure();
		});
	}

	function handleFailure() {
		errorMsg = 'Login failed';
		passwordInput.select();
		passwordInput.focus();
	}

	function updateNameEmpty(event: Event) {
		nameEmpty = (event.target as HTMLInputElement).value === '';
	}

	function updatePasswordEmpty(event: Event) {
		passwordEmpty = (event.target as HTMLInputElement).value === '';
	}

	onMount(() => {
		nameInput.select();
		nameInput.focus();
	});

	$: disabled = nameEmpty || passwordEmpty;
</script>

<div class='content'>
	<form class='panel' bind:this={form} on:submit={submit}>
		<img class='logo' src='/app.webp' alt='GURPS Character Sheet' />
		<div class='title'>GURPS Character Sheet</div>
		<div class='subtitle'>by Richard A. Wilkes</div>
		{#if errorMsg}
			<div class='error'>{errorMsg}</div>
		{/if}
		<label for='name'>Name</label>
		<input type='text' id='name' name='name' bind:this={nameInput} on:input={updateNameEmpty} required />
		<label for='password'>Password</label>
		<input type='password' id='password' name='password' bind:this={passwordInput} on:input={updatePasswordEmpty} required />
		<button type='submit' {disabled}>Login</button>
	</form>
</div>

<style>
	.content {
		display: flex;
		flex-flow: column nowrap;
		justify-content: center;
		align-items: center;
		flex-grow: 1;
	}

	.panel {
		display: flex;
		flex-flow: column nowrap;
		align-items: center;
		flex-direction: column;
		border-radius: 20px;
		margin: 20px;
		padding: 20px;
		box-shadow: 0 0 128px 4px var(--color-divider);
	}

	.title {
		font-weight: bold;
		font-size: 200%;
		color: var(--color-accent);
	}

	.subtitle {
		font-style: italic;
		margin-bottom: 20px;
		color: var(--color-on-background);
		opacity: 60%;
	}

	.logo {
		width: 128px;
		height: 128px;
	}

	label {
		font-weight: bold;
		font-variant: small-caps;
		color: var(--color-on-background);
		opacity: 60%;
		margin-top: 8px;
		align-self: flex-start;
	}

	input {
		font-weight: bold;
		align-self: stretch;
		padding: 0.5em;
		background-color: var(--color-editable);
		color: var(--color-on-editable);
		border: 1px solid var(--color-divider);
		border-radius: 8px;
	}

	button {
		font-weight: bold;
		align-self: stretch;
		margin-top: 20px;
		padding: 0.5em;
		color: var(--color-on-control);
		background-color: var(--color-control);
		border: 1px solid var(--color-control-edge);
		border-radius: 8px;
	}

	button:active {
		background-color: var(--color-control-pressed);
		color: var(--color-on-control-pressed);
	}

	button:disabled {
		background-color: var(--color-control-disabled); /* TODO: Fix. Doesn't currently exist */
		color: var(--color-on-control-disabled); /* TODO: Fix. Doesn't currently exist */
	}

	.error {
		font-weight: bold;
		color: var(--color-on-error);
		background-color: var(--color-error);
		margin-bottom: 10px;
		padding: 0.5em;
		text-align: center;
		align-self: stretch;
	}
</style>