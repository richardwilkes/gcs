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
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<Header>{$sheet?.Points.Total ?? 0} Points</Header>
	<div class="fields">
		<Field right>{$sheet?.Points.Unspent ?? 0}</Field>
		<Label left>Unspent</Label>
		<div class="banding">
			<Field right>{$sheet?.Points.Ancestry ?? 0}</Field>
		</div>
		<div class="banding"><Label left>Race</Label></div>
		<Field right>{$sheet?.Points.Attributes ?? 0}</Field>
		<Label left>Attributes</Label>
		<div class="banding">
			<Field right>{$sheet?.Points.Advantages ?? 0}</Field>
		</div>
		<div class="banding"><Label left>Advantages</Label></div>
		<Field right>{$sheet?.Points.Disadvantages ?? 0}</Field>
		<Label left>Disadvantages</Label>
		<div class="banding">
			<Field right>{$sheet?.Points.Quirks ?? 0}</Field>
		</div>
		<div class="banding"><Label left>Quirks</Label></div>
		<Field right>{$sheet?.Points.Skills ?? 0}</Field>
		<Label left>Skills</Label>
		<div class="banding">
			<Field right>{$sheet?.Points.Spells ?? 0}</Field>
		</div>
		<div class="banding"><Label left>Spells</Label></div>
	</div>
</div>

<style>
	.content {
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		flex-grow: 1;
		display: grid;
		grid-template-columns: auto 1fr;
		align-content: space-between;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-below-surface);
		color: var(--color-on-below-surface);
	}
</style>
