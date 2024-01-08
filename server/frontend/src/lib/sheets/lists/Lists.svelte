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
	}
</style>
