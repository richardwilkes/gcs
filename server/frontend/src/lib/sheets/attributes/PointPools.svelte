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
	import EditableNumberField from '$lib/sheets/widget/EditableNumberField.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import PointsNoteField from '$lib/sheets/attributes/PointsNoteField.svelte';
	import { sheet } from '$lib/sheet.ts';

	// TODO: Use remaining fields in PointPool
</script>

<div class="content">
	<Header>Point Pools</Header>
	<div class="fields">
		{#each $sheet?.PointPools || [] as pool}
			<PointsNoteField value={pool.Points} />
			<EditableNumberField name={pool.Key + '.current'} value={pool.Value} right={true} />
			<Label title="of" left={true} />
			<EditableNumberField name={pool.Key + '.max'} value={pool.Max} right={true} />
			<Label title={pool.Name} left={true} />
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
		align-content: space-between;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		column-gap: 2px;
		padding-bottom: 2px;
	}
</style>
