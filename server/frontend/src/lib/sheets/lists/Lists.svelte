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
	import List from '$lib/sheets/lists/List.svelte';
	import { sheet } from '$lib/sheet.ts';

	let layout: string;

	$: {
		layout = 'grid-template:';
		if ($sheet?.Reactions?.Rows.length && $sheet?.ConditionalModifiers?.Rows.length) {
			layout += '"reactions conditional_modifiers"';
		} else if ($sheet?.Reactions?.Rows.length) {
			layout += '"reactions reactions"';
		} else if ($sheet?.ConditionalModifiers?.Rows.length) {
			layout += '"conditional_modifiers conditional_modifiers"';
		}
		if ($sheet?.MeleeWeapons?.Rows.length) {
			layout += '"melee melee"';
		}
		if ($sheet?.RangedWeapons?.Rows.length) {
			layout += '"ranged ranged"';
		}
		if ($sheet?.Traits?.Rows.length && $sheet?.Skills?.Rows.length) {
			layout += '"traits skills"';
		} else if ($sheet?.Traits?.Rows.length) {
			layout += '"traits traits"';
		} else if ($sheet?.Skills?.Rows.length) {
			layout += '"skills skills"';
		}
		if ($sheet?.Spells?.Rows.length) {
			layout += '"spells spells"';
		}
		if ($sheet?.CarriedEquipment?.Rows.length) {
			layout += '"equipment equipment"';
		}
		if ($sheet?.OtherEquipment?.Rows.length) {
			layout += '"other_equipment other_equipment"';
		}
		if ($sheet?.Notes?.Rows.length) {
			layout += '"notes notes"';
		}
		layout += ';'
	}
</script>

<div class='lists' style={layout}>
	<List table={$sheet?.Reactions} area='reactions' />
	<List table={$sheet?.ConditionalModifiers} area='conditional_modifiers' />
	<List table={$sheet?.MeleeWeapons} area='melee' />
	<List table={$sheet?.RangedWeapons} area='ranged' />
	<List table={$sheet?.Traits} area='traits' />
	<List table={$sheet?.Skills} area='skills' />
	<List table={$sheet?.Spells} area='spells' />
	<List table={$sheet?.CarriedEquipment} area='equipment' />
	<List table={$sheet?.OtherEquipment} area='other_equipment' />
	<List table={$sheet?.Notes} area='notes' />
</div>

<style>
	.lists {
		display: grid;
		gap: var(--section-gap);
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}
</style>
