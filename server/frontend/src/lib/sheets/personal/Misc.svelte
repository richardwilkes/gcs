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
	import { formatDateStamp } from '$lib/datetime.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import EditableTextField from '$lib/sheets/widget/EditableTextField.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<Header>Miscellaneous</Header>
	<div class="fields">
		<div><Label>Created</Label></div>
		<div><Field style="width:100%;">{formatDateStamp($sheet?.Misc.Created)}</Field></div>
		<div class="banding"><Label>Modified</Label></div>
		<div class="banding"><Field style="width:100%;">{formatDateStamp($sheet?.Misc.Modified)}</Field></div>
		<div><Label>Player</Label></div>
		<div><EditableTextField key="Misc.Player" /></div>
	</div>
</div>

<style>
	.content {
		grid-area: miscellaneous;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 1fr;
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
		justify-content: end;
	}
</style>
