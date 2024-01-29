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
	export let name:string;
	export let path:string;
	export let selected = false;
	export let callback : (path: string) => void;
</script>

<div>
	<button class:selected on:click={() => callback(path)}>{name}</button>
</div>

<style>
	div {
		padding-left: 1em;
	}

	button {
		background-color: var(--color-background);
		color: var(--color-on-background);
		border: none;
	}

	button:hover {
		background-color: var(--color-selection);
		color: var(--color-on-selection);
	}
</style>