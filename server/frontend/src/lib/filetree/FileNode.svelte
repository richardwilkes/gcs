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
	export let callback: (path: string) => void;
</script>

<div>
	<button class:selected={path === selectedFile} on:click={() => callback(path)}>
		{#if path.toLowerCase().endsWith('.gcs')}
			<SheetFile />
		{:else}
			<File />
		{/if}
		{name}
	</button>
</div>

<style>
	div {
		padding-left: 1.5em;
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

	.selected {
		background-color: var(--color-selection);
		color: var(--color-on-selection);
	}
</style>