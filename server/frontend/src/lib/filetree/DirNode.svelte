<!--
  - Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang="ts">
	import DirNode from './DirNode.svelte';
	import type { Directory } from '$lib/files.ts';
	import { tweened } from 'svelte/motion';
	import { slide } from 'svelte/transition';
	import FileNode from '$lib/filetree/FileNode.svelte';
	import OpenFolderSVG from '$svg/OpenFolder.svg?raw';
	import ClosedFolderSVG from '$svg/ClosedFolder.svg?raw';
	import CircledChevronRightSVG from '$svg/CircledChevronRight.svg?raw';

	interface Props {
		dir: Directory;
		selectedFile: string | undefined;
		callback: (file: string, finish?: boolean) => void;
	}

	let { dir, selectedFile, callback }: Props = $props();

	const rotation = tweened(0, { duration: 200 });
	let opened = $state(false);
</script>

<div class="dir">
	<div class="node">
		<button
			class="chevron"
			style="transform: rotate({$rotation}deg)"
			onclick={() => {
				opened = !opened;
				rotation.set(opened ? 90 : 0);
			}}
		>
			{@html CircledChevronRightSVG}
		</button>
		<button
			class="item"
			onclick={() => {}}
			ondblclick={() => {
				opened = !opened;
				rotation.set(opened ? 90 : 0);
			}}
		>
			{#if opened}
				{@html OpenFolderSVG}
			{:else}
				{@html ClosedFolderSVG}
			{/if}
			{dir.name}
		</button>
	</div>
	{#if opened}
		<div class="children" transition:slide={{ delay: 0, duration: 200, axis: 'y' }}>
			{#each dir.dirs || [] as subDir}
				<DirNode dir={subDir} {selectedFile} {callback} />
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

	.node {
		display: flex;
		column-gap: 0.4em;
		align-items: center;
		text-align: left;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		padding: 0.2em;
	}

	.chevron,
	.item {
		padding: 0;
		border: none;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		-user-select: none;
		-webkit-user-select: none; /* Chrome/Safari */
		-moz-user-select: none; /* Firefox */
	}

	.chevron {
		display: flex;
		justify-content: center;
		align-items: center;
	}

	.chevron > :global(svg) {
		height: 0.75em;
	}

	.item > :global(svg) {
		height: 0.75em;
	}

	.children {
		margin-left: 1.2em;
	}
</style>
