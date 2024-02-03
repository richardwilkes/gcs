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
	import type { Directory } from '$lib/files.ts';
	import { tweened } from 'svelte/motion';
	import { slide } from 'svelte/transition';
	import FileNode from '$lib/filetree/FileNode.svelte';
	import OpenFolder from '$lib/svg/OpenFolder.svelte';
	import ClosedFolder from '$lib/svg/ClosedFolder.svelte';
	import CircledChevronRight from '$lib/svg/CircledChevronRight.svelte';

	export let dir: Directory;
	export let selectedFile: string | undefined;
	export let callback: (file: string) => void;

	const rotation = tweened(0, { duration: 200 });
	let opened = false;
</script>

<div class='dir'>
	<button on:click={() => { opened = !opened; rotation.set(opened ? 90 : 0); } }>
		<CircledChevronRight rotation={$rotation}/>
		{#if opened}
			<OpenFolder />
		{:else}
			<ClosedFolder />
		{/if}
		{dir.name}
	</button>
	{#if opened}
		<div class='children' transition:slide={{delay: 0, duration: 200, axis: 'y'}}>
			{#each dir.dirs || [] as subDir}
				<svelte:self dir={subDir} {selectedFile} {callback} />
			{/each}
			{#each dir.files || [] as file}
				<FileNode name={file} path={dir.path + '/' + file} {selectedFile} {callback} />
			{/each}
		</div>
	{/if}
</div>

<style>
	.dir {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
	}

	.children {
		margin-left: 1.5em;
	}

	button {
		background-color: var(--color-background);
		color: var(--color-on-background);
		border: none;
		font-size: 1.2em;
		text-align: left;
		display: flex;
		column-gap: 0.5em;
		align-items: center;
	}

	/*button:hover {*/
	/*	background-color: var(--color-link-rollover);*/
	/*	color: var(--color-on-link-rollover);*/
	/*}*/
</style>
