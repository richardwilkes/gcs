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
	import EquipmentList from '$lib/sheets/lists/EquipmentList.svelte';
</script>

<div class="content">
	<EquipmentList />
	<EquipmentList carried={false} />
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
