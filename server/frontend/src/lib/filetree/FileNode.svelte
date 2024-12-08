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
	import FileSVG from '$svg/File.svg?raw';
	import SheetFileSVG from '$svg/SheetFile.svg?raw';

	interface Props {
		name: string;
		path: string;
		selectedFile: string | undefined;
		callback: (path: string, finish?: boolean) => void;
	}

	let { name, path, selectedFile, callback }: Props = $props();
</script>

<div class="file" class:selected={path === selectedFile}>
	<button
		class="item"
		class:selected={path === selectedFile}
		onclick={() => callback(path)}
		ondblclick={() => callback(path, true)}
	>
		{#if path.toLowerCase().endsWith('.gcs')}
			{@html SheetFileSVG}
		{:else}
			{@html FileSVG}
		{/if}
		{name}
	</button>
</div>

<style>
	.file {
		margin-left: 0.8em;
		padding: 0.2em;
	}

	.item {
		padding: 0;
		border: none;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		-user-select: none;
		-webkit-user-select: none; /* Chrome/Safari */
		-moz-user-select: none; /* Firefox */
		align-items: center;
	}

	.item > :global(svg) {
		height: 0.75em;
	}

	.selected {
		background-color: var(--color-focus);
		color: var(--color-on-focus);
	}
</style>
