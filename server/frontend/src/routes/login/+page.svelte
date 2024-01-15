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
	import { goto } from '$app/navigation';
	import { apiPrefix } from '$lib/dev.ts';
	import Shell from '$lib/shell/Shell.svelte';
	import { onMount } from 'svelte';
	import { session, updateSessionFromResponse } from '$lib/session.ts';

	let form: HTMLFormElement;
	let nameInput: HTMLInputElement;
	let passwordInput: HTMLInputElement;
	let disabled = true;
	let nameEmpty = true;
	let passwordEmpty = true;
	let errorMsg = '';

	function updateNameEmpty(event: Event) {
		nameEmpty = (event.target as HTMLInputElement).value === '';
	}

	function updatePasswordEmpty(event: Event) {
		passwordEmpty = (event.target as HTMLInputElement).value === '';
	}

	function submit(event: Event) {
		event.preventDefault();
		fetch(apiPrefix('/login'), {
			method: 'POST',
			body: new FormData(form),
			cache: 'no-store'
		}).then(rsp => {
			if (!rsp.ok) {
				console.log(rsp.status + " " + rsp.statusText);
				handleFailure();
			} else {
					updateSessionFromResponse(rsp);
					if ($session) {
						// goto(new URLSearchParams(window.location.search).get('next') || '/');
						goto('/');
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

	session.set(null);

	$: disabled = nameEmpty || passwordEmpty;

	onMount(() => {
		nameInput.select();
		nameInput.focus();
	});
</script>

<svelte:head>
	<title>GURPS Character Sheet</title>
</svelte:head>

<Shell>
	<div slot='content' class='content'>
		<form class='panel' bind:this={form} on:submit={submit}>
			<img class='logo' src='/app.png' alt='GURPS Character Sheet' />
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
</Shell>

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
		font: var(--font-login);
		font-size: 200%;
		color: var(--color-accent);
	}

	.subtitle {
		font: var(--font-login);
		margin-bottom: 20px;
		color: var(--color-on-background);
		opacity: 60%;
		font-style: italic;
	}

	.logo {
		width: 128px;
		height: 128px;
	}

	label {
		font: var(--font-login);
		font-variant: small-caps;
		color: var(--color-on-background);
		opacity: 60%;
		margin-top: 8px;
		align-self: flex-start;
	}

	input {
		font: var(--font-login);
		align-self: stretch;
		padding: 0.5em;
	}

	button {
		font: var(--font-login);
		align-self: stretch;
		margin-top: 20px;
		padding: 0.5em;
	}

	.error {
		font: var(--font-login);
		color: var(--color-on-error);
		background-color: var(--color-error);
		margin-bottom: 10px;
		padding: 0.5em;
		text-align: center;
		align-self: stretch;
	}
</style>
