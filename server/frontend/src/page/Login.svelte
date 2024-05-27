<!--
  - Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang="ts">
	import { apiPrefix } from '$lib/dev.ts';
	import { session, updateSessionFromResponse } from '$lib/session.ts';
	import { onMount } from 'svelte';
	import { url } from '$lib/url.ts';
	import { navTo } from '$lib/nav';

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
			cache: 'no-store',
		})
			.then((rsp) => {
				if (!rsp.ok) {
					console.log(rsp.status + ' ' + rsp.statusText);
					handleFailure();
				} else {
					updateSessionFromResponse(rsp);
					if ($session) {
						navTo($url.searchParams.get('next') || '#', undefined, undefined, true);
					} else {
						handleFailure();
					}
				}
			})
			.catch((error) => {
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

<div class="content">
	<form class="panel" bind:this={form} on:submit={submit}>
		<img class="logo" src="/app.webp" alt="GURPS Character Sheet" />
		<div class="title">GURPS Character Sheet</div>
		<div class="subtitle">by Richard A. Wilkes</div>
		{#if errorMsg}
			<div class="error">{errorMsg}</div>
		{/if}
		<label for="name">Name</label>
		<input type="text" id="name" name="name" bind:this={nameInput} on:input={updateNameEmpty} required />
		<label for="password">Password</label>
		<input
			type="password"
			id="password"
			name="password"
			bind:this={passwordInput}
			on:input={updatePasswordEmpty}
			required />
		<button type="submit" {disabled}>Login</button>
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
		background-color: var(--color-surface);
		box-shadow: 0 0 64px 0 var(--color-shadow);
	}

	.title {
		font-weight: bold;
		font-size: 200%;
		color: green;
	}

	.subtitle {
		font-style: italic;
		margin-bottom: 20px;
		color: darkgreen;
	}

	.logo {
		width: 128px;
		height: 128px;
	}

	label {
		font-weight: bold;
		font-variant: small-caps;
		color: var(--color-on-surface);
		margin-top: 8px;
		align-self: flex-start;
	}

	input {
		align-self: stretch;
		background-color: var(--color-deep-below-surface);
		color: var(--color-on-deep-below-surface);
	}

	button {
		align-self: stretch;
		margin-top: 20px;
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
