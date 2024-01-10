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
	import Shell from '$lib/shell/Shell.svelte';
</script>

<svelte:head>
	<title>GURPS Character Sheet</title>
</svelte:head>

<Shell>
	<div slot='content' class='content'>
		<form action='/login' method='post' class='panel'>
			<div class='title'>GURPS Character Sheet</div>
			<div class='subtitle'>by Richard A. Wilkes</div>
			<label for='name'>Name</label>
			<!-- svelte-ignore a11y-autofocus -->
			<input type='text' autofocus />
			<label for='password'>Password</label>
			<input type='text' />
			<button type='submit'>Login</button>
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
		box-shadow: 0 0 64px 8px var(--color-divider);
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
	}

	button {
		font: var(--font-login);
		align-self: stretch;
		margin-top: 20px;
	}
</style>
