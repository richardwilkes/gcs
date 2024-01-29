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
	import { replaceState } from '$app/navigation';
	import Shell from '$lib/shell/Shell.svelte';
	import Personal from '$lib/sheets/personal/Personal.svelte';
	import Attributes from '$lib/sheets/attributes/Attributes.svelte';
	import Lists from '$lib/sheets/lists/Lists.svelte';
	import { pc } from '$lib/entity.ts';
	import PCLoaderButton from '$lib/sheets/widget/PCLoaderButton.svelte';
	import FileSelectionModal from '$lib/filetree/FileSelectionModal.svelte';
	import { apiPrefix } from '$lib/dev.ts';
	import { session } from '$lib/session.ts';

	let showModal = false;

	if ($page.url.searchParams.get('path')) {
		loadSheet();
	} else if (!$pc) {
		showModal = true;
	}

	function gotoSheet(path: string) {
		const url = $page.url;
		url.searchParams.set('path', path);
		replaceState(url, {});
		loadSheet();
	}

	async function loadSheet() {
		const rsp = await fetch(new URL(apiPrefix(`/sheet/${$page.url.searchParams.get('path')}`)), {
			method: 'GET',
			headers: { 'X-Session': $session?.ID ?? '' },
			cache: 'no-store'
		});
		if (rsp.ok) {
			$pc = await rsp.json();
		} else {
			console.log(rsp.status + ' ' + rsp.statusText);
		}
	}
</script>

<Shell>
	<PCLoaderButton slot='toolbar' />
	<div slot='content' class='content'>
		{#if $pc}
			<Personal />
			<Attributes />
			<Lists />
		{:else}
			<FileSelectionModal bind:showModal path='/sheets' callback={gotoSheet} />
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
</style>
