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
	import Header from '$lib/sheets/widget/Header.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import PointsNoteField from '$lib/sheets/attributes/PointsNoteField.svelte';
	import { sheet } from '$lib/sheet.ts';

	// TODO: Use remaining fields in PointPool
</script>

<div class="content">
	<Header>Point Pools</Header>
	<div class="fields">
		{#each $sheet?.PointPools || [] as pool, i}
			{@const banding = i % 2 === 1}
			<div class:banding><PointsNoteField value={pool.Points} /></div>
			<div class:banding><Field editable right style='width: 100%;'>{pool.Value}</Field></div>
			<div class:banding><Label left>of</Label></div>
			<div class:banding><Field editable right style='width: 100%;'>{pool.Max}</Field></div>
			<div class:banding><Label left>{pool.Name}</Label></div>
		{/each}
	</div>
</div>

<style>
	.content {
		grid-area: pool-attributes;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 0fr 0fr 0fr 1fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		padding-bottom: 2px;
	}

	.fields > div {
		display: flex;
		align-items: center;
	}
</style>
