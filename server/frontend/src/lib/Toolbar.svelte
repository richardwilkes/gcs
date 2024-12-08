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
	import ThemeSwitch from '$lib/ThemeSwitch.svelte';
	import { session } from '$lib/session.ts';
	import { sheet } from '$lib/sheet.ts';
	import { apiPrefix } from '$lib/dev.ts';
	import { navTo } from '$lib/nav';
	import LogoutSVG from '$svg/Logout.svg?raw';
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();

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
		{@render children?.()}
	</div>
	<ThemeSwitch />
	{#if $session?.User}
		<button class="logout svg" title="Logout" onclick={logout}>
			{@html LogoutSVG}
		</button>
	{/if}
</div>

<style>
	.toolbar {
		background-color: var(--color-surface);
		border-bottom: 1px solid var(--color-surface-edge);
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

	.logout > :global(svg) {
		width: 1.2em;
		height: 1.2em;
		color: var(--color-on-surface);
	}
</style>
