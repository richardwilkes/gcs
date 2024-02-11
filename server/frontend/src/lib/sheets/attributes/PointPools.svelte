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
	import { type AttributeDef, AttributeType } from '$lib/entity.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import EditableNumberField from '$lib/sheets/widget/EditableNumberField.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import PointsNoteField from '$lib/sheets/attributes/PointsNoteField.svelte';
	import { pc } from '$lib/entity.ts';

	let attrList: { name: string; current: number; max: number; points: number }[] = [];

	$: {
		attrList = [];
		if ($pc?.attributes) {
			let attrDefMap = new Map();
			let orderedAttrDefs: AttributeDef[] = [];
			$pc.settings?.attributes?.forEach((attrDef) => {
				if (attrDef.type === AttributeType.Pool) {
					attrDefMap.set(attrDef.id, attrDef);
					orderedAttrDefs.push(attrDef);
				}
			});
			let attrMap = new Map();
			$pc.attributes?.forEach((attr) => {
				if (attrDefMap.has(attr.attr_id)) {
					attrMap.set(attr.attr_id, attr);
				}
			});
			orderedAttrDefs.forEach((attrDef) => {
				let attr = attrMap.get(attrDef.id);
				if (attr === undefined) {
					attr = {
						attr_id: attrDef.id,
						adj: 0
					};
				}
				attrList.push({
					name: attrDef.full_name ? `${attrDef.full_name} (${attrDef.name})` : attrDef.name,
					current: attr.calc?.current ?? 0,
					max: attr.calc?.value ?? 0,
					points: attr.calc?.points ?? 0
				});
			});
		}
	}
</script>

<div class="content">
	<Header>Point Pools</Header>
	<div class="fields">
		{#each attrList as attr}
			<PointsNoteField value={attr.points} />
			<EditableNumberField name={attr.name + '.current'} value={attr.current} right={true} />
			<Label title="of" left={true} />
			<EditableNumberField name={attr.name + '.max'} value={attr.max} right={true} />
			<Label title={attr.name} left={true} />
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
