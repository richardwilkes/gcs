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
	import Modal from '$lib/Modal.svelte';
	import { apiPrefix } from '$lib/dev.ts';
	import { session } from '$lib/session.ts';
	import { type Directory, fillPathsForDir } from '$lib/files.ts';
	import DirNode from '$lib/filetree/DirNode.svelte';

	export let showModal = false;
	export let title = 'Select a File';
	export let path: string;
	export let callback: (file: string) => void;

	let modal: Modal;
	let awaitingLoad = false;
	let dirs: Directory[] | undefined;
	let selectedFile: string | undefined;

	$: if (showModal && modal && !dirs && !awaitingLoad) {
		awaitingLoad = true;
		(async function loadFiles() {
			const rsp = await fetch(apiPrefix(path), {
				method: 'GET',
				headers: { 'X-Session': $session?.ID ?? '' },
				cache: 'no-store'
			});
			if (rsp.ok) {
				let data = await rsp.json();
				for (const dir of data) {
					fillPathsForDir(dir, "");
				}
				dirs = data;
			} else {
				console.log(rsp.status + ' ' + rsp.statusText);
			}
		})();
	}

	function done(ok: boolean) {
		dirs = undefined;
		awaitingLoad = false;
		if (ok && selectedFile) callback(selectedFile);
	}
</script>

<Modal bind:this={modal} bind:showModal callback={(ok) => done(ok)}>
	<div slot='title'>{title}</div>
	<div class='tree'>
		{#if dirs}
			{#each dirs as dir}
				<div>
					<DirNode {dir} {selectedFile} callback={(file) => selectedFile = file}/>
				</div>
			{/each}
		{:else}
			Loading...
		{/if}
	</div>
</Modal>

<style>
	.tree {
		min-width: 50vw;
		min-height: 50vh;
	}
</style>