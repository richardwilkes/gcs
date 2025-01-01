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
	import Header from '$lib/sheets/widget/Header.svelte';
	import EditableTextField from '$lib/sheets/widget/EditableTextField.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import PointsNoteField from '$lib/sheets/attributes/PointsNoteField.svelte';
	import { sheet } from '$lib/sheet.ts';

	// TODO: Use remaining fields
</script>

<div class="content">
	<Header>Secondary Attributes</Header>
	<div class="fields">
		{#each $sheet?.SecondaryAttributes || [] as attr, i}
			{@const banding = i % 2 === 1}
			<div class:banding><PointsNoteField value={attr.Points} /></div>
			<div class:banding><EditableTextField right key="SecondaryAttributes.{attr.Key}" /></div>
			<div class:banding><Label left>{attr.Name}</Label></div>
		{/each}
	</div>
</div>

<style>
	.content {
		grid-area: secondary-attributes;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 0fr 1fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-below-surface);
		color: var(--color-on-below-surface);
		padding-bottom: 2px;
	}

	.fields > div {
		display: flex;
		align-items: center;
	}
</style>
