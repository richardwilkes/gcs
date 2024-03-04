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
	import { sheet, updateSheetField } from '$lib/sheet.ts';
	import { page } from '$lib/page.ts';

	async function updateName(event: FocusEvent) {
		if ($page.Sheet) {
			let target = event.target as HTMLElement;
			const text = target.innerText;
			const updatedSheet = await updateSheetField($page.Sheet, "Identity.Name", text);
			sheet.update((_) => updatedSheet);
			if (updatedSheet && updatedSheet.Identity.Name != text) {
				target.innerText = updatedSheet.Identity.Name;
			}
		}
	}
</script>

<div class='content'>
	<Header>Identity</Header>
	<div class='fields'>
		<div class='banding'><Label>Name</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;' on:blur={updateName}>{$sheet?.Identity.Name ?? ''}</Field>
		</div>
		<div><Label>Title</Label></div>
		<div>
			<Field editable style='width:100%;'>{$sheet?.Identity.Title ?? ''}</Field>
		</div>
		<div class='banding'><Label>Organization</Label></div>
		<div class='banding'>
			<Field editable style='width:100%;'>{$sheet?.Identity.Organization ?? ''}</Field>
		</div>
	</div>
</div>

<style>
	.content {
		grid-area: identity;
		flex-grow: 1;
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
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		padding-bottom: 2px;
	}

	.fields > div {
		display: flex;
		align-items: center;
		justify-content: end;
	}
</style>
