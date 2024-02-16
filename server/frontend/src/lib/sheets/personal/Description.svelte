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
	import EditableNumberField from '$lib/sheets/widget/EditableNumberField.svelte';
	import EditableStringField from '$lib/sheets/widget/EditableStringField.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<div style="grid-area: header;">
		<Header>Description</Header>
	</div>
	<div class="fields" style="grid-area:block1;">
		<Label title="Gender" />
		<EditableStringField name="profile.gender" value={$sheet?.Description.Gender ?? ''} />
		<Label title="Age" />
		<EditableStringField name="profile.age" value={$sheet?.Description.Age ?? ''} />
		<Label title="Birthday" />
		<EditableStringField name="profile.birthday" value={$sheet?.Description.Birthday ?? ''} />
		<Label title="Religion" />
		<EditableStringField name="profile.religion" value={$sheet?.Description.Religion ?? ''} />
	</div>
	<div class="fields" style="grid-area:block2;">
		<Label title="Height" />
		<EditableStringField name="profile.height" value={$sheet?.Description.Height ?? ''} />
		<Label title="Weight" />
		<EditableStringField name="profile.weight" value={$sheet?.Description.Weight ?? ''} />
		<Label title="Size" />
		<EditableNumberField name="SM" value={$sheet?.Description.SizeModifier ?? 0} />
		<Label title="TL" />
		<EditableStringField name="profile.tl" value={$sheet?.Description.TechLevel ?? ''} />
	</div>
	<div class="fields" style="grid-area:block3;">
		<Label title="Hair" />
		<EditableStringField name="profile.hair" value={$sheet?.Description.Hair ?? ''} />
		<Label title="Eyes" />
		<EditableStringField name="profile.eyes" value={$sheet?.Description.Eyes ?? ''} />
		<Label title="Skin" />
		<EditableStringField name="profile.skin" value={$sheet?.Description.Skin ?? ''} />
		<Label title="Hand" />
		<EditableStringField name="profile.handedness" value={$sheet?.Description.Hand ?? ''} />
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
		grid-template-columns: 1fr 1fr 1fr;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		grid-template-columns: 0fr 1fr;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		padding-right: 4px;
		padding-bottom: 2px;
	}
</style>
