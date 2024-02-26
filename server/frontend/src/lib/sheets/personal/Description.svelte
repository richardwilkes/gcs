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
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class='content'>
	<div style='grid-area: header;'>
		<Header>Description</Header>
	</div>
	<div class='fields' style='grid-area:block1;'>
		<div class='banding'><Label>Gender</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.Gender ?? ''}</Field>
		</div>
		<div><Label>Age</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.Age ?? ''}</Field>
		</div>
		<div class='banding'><Label>Birthday</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.Birthday ?? ''}</Field>
		</div>
		<div><Label>Religion</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.Religion ?? ''}</Field>
		</div>
	</div>
	<div class='fields divider' style='grid-area:block2;'>
		<div class='banding'><Label>Height</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.Height ?? ''}</Field>
		</div>
		<div><Label>Weight</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.Weight ?? ''}</Field>
		</div>
		<div class='banding'><Label>Size</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.SizeModifier ?? 0}</Field>
		</div>
		<div><Label>TL</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.TechLevel ?? ''}</Field>
		</div>
	</div>
	<div class='fields divider' style='grid-area:block3;'>
		<div class='banding'><Label>Hair</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.Hair ?? ''}</Field>
		</div>
		<div><Label>Eyes</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.Eyes ?? ''}</Field>
		</div>
		<div class='banding'><Label>Skin</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Description.Skin ?? ''}</Field>
		</div>
		<div><Label>Hand</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Description.Hand ?? ''}</Field>
		</div>
	</div>
</div>

<style>
	.content {
		grid-area: description;
		flex-grow: 1;
		display: grid;
		grid-template:
			'header header header'
			'block1 block2 block3';
		grid-template-columns: auto auto auto;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr auto;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}

	.fields > div {
		display: flex;
		align-items: center;
		justify-content: end;
		padding: var(--padding-standard);
	}

	.divider {
		border-left: var(--standard-border);
	}
</style>
