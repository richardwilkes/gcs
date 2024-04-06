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
	import { fetchSheet, sheet } from '$lib/sheet.ts';
	import { navTo } from '$lib/nav';
	import Lists from '$lib/sheets/lists/Lists.svelte';
	import Personal from '$lib/sheets/personal/Personal.svelte';
	import Attributes from '$lib/sheets/attributes/Attributes.svelte';

	export let path = '';

	let failed = false;

	async function load() {
		if (path) {
			const s = await fetchSheet(path);
			$sheet = s;
			if (!$sheet) {
				failed = true;
			}
		} else {
			navTo('#');
		}
	}

	$: {
		path; // For reactivity...
		load();
	}
</script>

<div class="content">
	<div class="sheet">
		{#if $sheet}
			<Personal />
			<Attributes />
			<Lists />
		{:else if failed}
			<div class="failed">Failed to load sheet</div>
		{:else}
			<div class="loading">Loading...</div>
		{/if}
	</div>
</div>

<style>
	.content {
		background-color: var(--color-surface);
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		align-items: stretch;
		align-content: stretch;
		flex-grow: 1;
	}

	.sheet {
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		align-items: stretch;
		align-content: stretch;
		padding: 5px;
		gap: var(--section-gap);
	}

	.loading {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		font-size: 300%;
	}

	.failed {
		display: flex;
		justify-content: center;
		align-items: center;
		height: 100%;
		font-size: 300%;
		color: var(--color-error);
	}
</style>
