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
	import { page } from '$app/stores';
	import { apiPrefix } from '$lib/dev.ts';
	import { session } from '$lib/session.ts';
	import { pc } from '$lib/entity.ts';
	import Lists from '$lib/sheets/lists/Lists.svelte';
	import PCLoaderButton from '$lib/sheets/widget/PCLoaderButton.svelte';
	import Personal from '$lib/sheets/personal/Personal.svelte';
	import Attributes from '$lib/sheets/attributes/Attributes.svelte';
	import Shell from '$lib/shell/Shell.svelte';

	let failed = false;

	fetch(new URL(apiPrefix(`/sheet/${$page.params.path}`)), {
		method: 'GET',
		headers: { 'X-Session': $session?.ID ?? '' },
		cache: 'no-store'
	}).then(async (rsp) => {
		if (rsp.ok) {
			$pc = await rsp.json();
		} else {
			failed = true;
			$pc = undefined;
			console.log(rsp.status + ' ' + rsp.statusText);
		}
	}).catch((err) => {
		failed = true;
		$pc = undefined;
		console.log(err);
	});
</script>

<svelte:head>
	<title>{$pc?.profile?.name ?? 'GURPS Character Sheet'}</title>
</svelte:head>

<Shell>
	<PCLoaderButton slot='toolbar' />
	<div slot='content' class='content'>
		{#if $pc}
			<Personal />
			<Attributes />
			<Lists />
		{:else}
			{#if failed}
				<div class='failed'>Failed to load sheet</div>
			{:else}
				<div class='loading'>Loading...</div>
			{/if}
		{/if}
	</div>
</Shell>

<style>
	.content {
		background-color: var(--color-page);
		color: var(--color-on-page);
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		align-items: stretch;
		align-content: stretch;
		gap: var(--section-gap);
		padding: 5px;
		overflow: auto;
		flex-grow: 1;
	}

	.loading {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		font-size: 3em;
	}

	.failed {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		font-size: 3em;
		color: var(--color-error);
	}
</style>
