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
	interface Props {
		span?: number;
		tip?: string;
		extra_style?: string;
		children?: import('svelte').Snippet;
	}

	let {
		span = 1,
		tip = '',
		extra_style = '',
		children
	}: Props = $props();
</script>

<div class="header" title={tip} style="grid-column-end: span {span};{extra_style}">
	{#if children}{@render children()}{:else}&nbsp;{/if}
</div>

<style>
	.header {
		background-color: var(--color-header);
		color: var(--color-on-header);
		fill: var(--color-on-header);
		font-variant: small-caps;
		padding: var(--padding-standard);
		display: flex;
		align-items: center;
		justify-content: center;
	}
</style>
