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
	import EditableTextField from '$lib/sheets/widget/EditableTextField.svelte';
</script>

<div class="content">
	<div style="grid-area: header;">
		<Header>Description</Header>
	</div>
	<div class="fields" style="grid-area:block1;">
		<div class="banding"><Label>Gender</Label></div>
		<div class="banding">
			<EditableTextField key="Description.Gender" />
		</div>
		<div><Label>Age</Label></div>
		<div>
			<EditableTextField key="Description.Age" />
		</div>
		<div class="banding"><Label>Birthday</Label></div>
		<div class="banding">
			<EditableTextField key="Description.Birthday" />
		</div>
		<div><Label>Religion</Label></div>
		<div>
			<EditableTextField key="Description.Religion" />
		</div>
	</div>
	<div class="fields divider" style="grid-area:block2;">
		<div class="banding"><Label>Height</Label></div>
		<div class="banding">
			<EditableTextField key="Description.Height" />
		</div>
		<div><Label>Weight</Label></div>
		<div>
			<EditableTextField key="Description.Weight" />
		</div>
		<div class="banding"><Label>Size</Label></div>
		<div class="banding">
			<EditableTextField key="Description.SizeModifier" />
		</div>
		<div><Label>TL</Label></div>
		<div>
			<EditableTextField key="Description.TechLevel" />
		</div>
	</div>
	<div class="fields divider" style="grid-area:block3;">
		<div class="banding"><Label>Hair</Label></div>
		<div class="banding">
			<EditableTextField key="Description.Hair" />
		</div>
		<div><Label>Eyes</Label></div>
		<div>
			<EditableTextField key="Description.Eyes" />
		</div>
		<div class="banding"><Label>Skin</Label></div>
		<div class="banding">
			<EditableTextField key="Description.Skin" />
		</div>
		<div><Label>Hand</Label></div>
		<div>
			<EditableTextField key="Description.Hand" />
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
		background-color: var(--color-below-surface);
		color: var(--color-on-below-surface);
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
