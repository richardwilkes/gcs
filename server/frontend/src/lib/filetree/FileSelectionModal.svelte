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
	import Waiting from '$lib/Waiting.svelte';

	export let showModal = false;
	export let title = 'Select a File';
	export let path: string;
	export let callback: (file: string, finish?: boolean) => void;

	let modal: Modal;
	let pending = false;
	let error = false;
	let dirs: Directory[] | undefined;
	let selectedFile: string | undefined;

	$: if (showModal && modal && !dirs && !pending) {
		error = false;
		pending = true;
		(async function loadFiles() {
			const rsp = await fetch(apiPrefix(path), {
				method: 'GET',
				headers: { 'X-Session': $session?.ID ?? '' },
				cache: 'no-store'
			});
			if (rsp.ok) {
				let data = await rsp.json();
				for (const dir of data) {
					fillPathsForDir(dir, '');
				}
				dirs = data;
				pending = false;
			} else {
				error = true;
			}
		})();
	}

	function done(ok: boolean) {
		dirs = undefined;
		pending = false;
		if (ok && selectedFile) callback(selectedFile);
	}
</script>

<Modal bind:this={modal} bind:showModal callback={(ok) => done(ok)}>
	<div slot='title'>{title}</div>
	<div class='content'>
		{#if pending}
			<div class='pending' class:error>
				{#if error}
					Error loading file list
				{:else}
					<Waiting />
				{/if}
			</div>
		{:else}
			<div class='tree'>
				<div class='inner'>
					{#each dirs || [] as dir}
						<DirNode {dir} {selectedFile} callback={(file, finish) => {
							selectedFile = file;
							if (finish) {
								modal.close(true);
							}
						}} />
					{/each}
				</div>
			</div>
		{/if}
	</div>
</Modal>

<style>
	.content {
		min-width: 50vw;
		min-height: 50vh;
		display: flex;
	}

	.tree {
		background-color: var(--color-content);
		flex-grow: 1;
		border: 1px solid var(--color-divider);
	}

	.inner {
		padding: 0.5em;
		display: flex;
		flex-direction: column;
		overflow: auto;
		max-height: 50vh;
		max-width: 50vw;
	}

	.pending {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		flex-grow: 1;
		font-size: 2em;
	}

	.error {
		color: red;
	}
</style>