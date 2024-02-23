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
</script>

<div class='content'>
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
	.content {
		display: grid;
		gap: var(--section-gap);
		grid-template-columns: 1fr 1fr;
		grid-template:
			'reactions conditional_modifiers'
			'melee melee'
			'ranged ranged'
			'traits skills'
			'spells spells'
			'equipment equipment'
			'other_equipment other_equipment'
			'notes notes';
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}
</style>
