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
	import File from '$lib/svg/File.svelte';
	import SheetFile from '$lib/svg/SheetFile.svelte';

	export let name: string;
	export let path: string;
	export let selectedFile: string | undefined;
	export let callback: (path: string, finish?: boolean) => void;
</script>

<div class='file' class:selected={path === selectedFile}>
	<button class:selected={path === selectedFile} on:click={() => callback(path)} on:dblclick={() => callback(path, true)}>
		{#if path.toLowerCase().endsWith('.gcs')}
			<SheetFile />
		{:else}
			<File />
		{/if}
		{name}
	</button>
</div>

<style>
	.file {
		margin-left: 0.8em;
		padding: 0.2em;
	}

	button {
		padding: 0;
		border: none;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		user-select: none;
		align-items: center;
	}

	.selected {
		background-color: var(--color-tertiary);
		color: var(--color-on-tertiary);
	}
</style>