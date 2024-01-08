<script lang="ts">
	import { type AttributeDef, AttributeType } from '$lib/entity.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import EditableNumberField from '$lib/sheets/widget/EditableNumberField.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import PointsNoteField from '$lib/sheets/attributes/PointsNoteField.svelte';
	import { pc } from '$lib/entity.ts';

	let attrList: { name: string; value: number; points: number }[] = [];

	$: {
		attrList = [];
		if ($pc.attributes) {
			let attrDefMap = new Map();
			let orderedAttrDefs: AttributeDef[] = [];
			$pc.settings?.attributes?.forEach((attrDef) => {
				if (
					!attrDef.attribute_base?.includes('$') &&
					(attrDef.type === AttributeType.Integer ||
						attrDef.type === AttributeType.IntegerRef ||
						attrDef.type === AttributeType.Decimal ||
						attrDef.type === AttributeType.DecimalRef)
				) {
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
					value: attr.calc?.value ?? 0,
					points: attr.calc?.points ?? 0
				});
			});
		}
	}
</script>

<div class="content">
	<Header>Primary Attributes</Header>
	<div class="fields">
		{#each attrList as attr}
			<PointsNoteField value={attr.points} />
			<EditableNumberField name={attr.name} value={attr.value} right={true} />
			<Label title={attr.name} left={true} />
		{/each}
	</div>
</div>

<style>
	.content {
		grid-area: primary-attributes;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 0fr 1fr;
		align-content: space-between;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
		column-gap: 2px;
		padding-bottom: 2px;
	}
</style>
