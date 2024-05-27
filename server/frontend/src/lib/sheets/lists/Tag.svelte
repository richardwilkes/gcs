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
	import WarningSVG from '$svg/Warning.svg?raw';

	export let warning = false;
	export let tip = '';
</script>

<div class="tag" class:warning title={tip}>
	{#if warning}
		<div class="icon">{@html WarningSVG}</div>
	{/if}
	<slot />
</div>

<style>
	.tag {
		margin: 0 0.25em;
		display: inline-flex;
		align-items: center;
		padding: 0 0.5em;
		gap: 0.25em;
		width: fit-content;
		border-radius: 0.5em;
		color: var(--color-on-focus);
		background-color: var(--color-focus);
		font-size: 85%;
	}

	.warning {
		color: var(--color-on-error);
		background-color: var(--color-error);
	}

	.icon {
		width: 0.85em;
	}
</style>
