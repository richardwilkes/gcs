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
<script lang="ts">
	import ThemeSwitch from '$lib/ThemeSwitch.svelte';
	import { session } from '$lib/session.ts';
	import { sheet } from '$lib/sheet.ts';
	import { apiPrefix } from '$lib/dev.ts';
	import { navTo } from './nav';

	async function logout() {
		if ($session) {
			await fetch(apiPrefix('/logout'), {
				method: 'POST',
				headers: { 'X-Session': $session.ID },
				cache: 'no-store',
			});
		}
		$sheet = undefined;
		session.set(null);
		navTo('login');
	}
</script>

<div class="toolbar">
	<div class="fill">
		<slot />
	</div>
	<ThemeSwitch />
	{#if $session?.User}
		<button class="logout" title="Logout" on:click={logout}>
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" class="svg">
				<path
					d="m377.9 105.9 122.8 122.8c7.2 7.2 11.3 17.1 11.3 27.3s-4.1 20.1-11.3 27.3L377.9 406.1c-6.4 6.4-15 9.9-24 9.9-18.7 0-33.9-15.2-33.9-33.9V320H192c-17.7 0-32-14.3-32-32v-64c0-17.7 14.3-32 32-32h128v-62.1c0-18.7 15.2-33.9 33.9-33.9 9 0 17.6 3.6 24 9.9zM160 96H96c-17.7 0-32 14.3-32 32v256c0 17.7 14.3 32 32 32h64c17.7 0 32 14.3 32 32s-14.3 32-32 32H96c-53 0-96-43-96-96V128c0-53 43-96 96-96h64c17.7 0 32 14.3 32 32s-14.3 32-32 32z" />
			</svg>
		</button>
	{/if}
</div>

<style>
	.toolbar {
		background-color: var(--color-surface);
		border-bottom: 1px solid var(--color-outline-variant);
		padding: 6px 8px;
		display: flex;
		justify-content: flex-start;
		align-items: center;
		column-gap: 8px;
		row-gap: 4px;
	}

	.fill {
		display: flex;
		justify-content: flex-start;
		align-items: center;
		column-gap: 8px;
		row-gap: 4px;
		flex-grow: 1;
	}

	.logout {
		border: none;
		background: none;
		display: flex;
		cursor: pointer;
	}

	.svg {
		width: 1.2em;
		height: 1.2em;
		fill: var(--color-on-surface);
	}
</style>
