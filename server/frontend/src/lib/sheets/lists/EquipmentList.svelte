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
	import Field from '$lib/sheets/widget/Field.svelte';
	import { type Equipment, pc } from '$lib/entity.ts';
	import { Fixed } from '$lib/fixed.ts';
	import Bookmark from '$lib/svg/Bookmark.svelte';
	import CheckMark from '$lib/svg/CheckMark.svelte';
	import Coins from '$lib/svg/Coins.svelte';
	import Stack from '$lib/svg/Stack.svelte';
	import Weight from '$lib/svg/Weight.svelte';
	import PreqrequiteWarning from '$lib/sheets/widget/PreqrequiteWarning.svelte';

	export let carried = true;

	let totalWeight = new Fixed(0);
	let totalValue = new Fixed(0);
	let flat: FlatEquipment[] = [];

	type FlatEquipment = {
		item: Equipment;
		depth: number;
	};

	function flatten(depth: number, equipment?: Equipment[]) {
		if (!equipment) return;
		for (const item of equipment) {
			flat.push({ item: item, depth: depth });
			if (item.children) flatten(depth + 1, item.children);
		}
	}

	$: {
		const list = carried ? $pc.equipment : $pc.other_equipment;
		totalWeight = new Fixed(0);
		totalValue = new Fixed(0);
		for (const item of list ?? []) {
			totalWeight = totalWeight.add(Fixed.extract(item.calc.extended_weight).value);
			totalValue = totalValue.add(new Fixed(item.calc.extended_value));
		}
		flat = [];
		flatten(0, list);
	}
</script>

<div class='content' class:carried>
	{#if carried}
		<Header
			tip='Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.'>
			<CheckMark />
		</Header>
	{/if}
	<Header tip='Quantity'>#</Header>
	<Header>
		{#if carried}
			Carried Equipment ({totalWeight.toString()} lb; ${totalValue.comma()})
		{:else}
			Other Equipment (${totalValue.comma()})
		{/if}
	</Header>
	<Header tip='The number of uses remaining'>Uses</Header>
	<Header tip='Tech Level'>TL</Header>
	<Header tip='Legality Class'>LC</Header>
	<Header tip='The value of one of these pieces of equipment'>
		<Coins />
	</Header>
	<Header tip='The weight of one of these pieces of equipment'>
		<Weight />
	</Header>
	<Header tip='The value of all of these pieces of equipment, plus the value of any contained equipment'>
		<Stack />
		<Coins />
	</Header>
	<Header tip='The weight of all of these pieces of equipment, plus the weight of any contained equipment'>
		<Stack />
		<Weight />
	</Header>
	<Header tip='A reference to the book and page the item appears on e.g. B22 would refer to "Basic Set", page 22'>
		<Bookmark />
	</Header>
	{#each flat as item, i}
		{@const banding = i % 2 === 1}
		{@const notes = item.item.calc.resolved_notes ? item.item.calc.resolved_notes : item.item.notes}
		{#if carried}
			<div class:banding>
				<Field center={true}>
					{#if item.item.equipped}
						<CheckMark />
					{/if}
				</Field>
			</div>
		{/if}
		<div class:column={carried} class:banding>
			<Field right={true}>{item.item.quantity ?? 0}</Field>
		</div>
		<div class:banding class='description column' style='padding-left: {item.depth * 16}px'>
			<Field wrap={true}>{item.item.description ?? ''}</Field>
			{#if notes}
				<Field smaller={true} wrap={true}>{notes}</Field>
			{/if}
			{#if item.item.calc.unsatisfied_reason}
				<PreqrequiteWarning reason={item.item.calc.unsatisfied_reason.replace(/●/g, '•')} />
			{/if}
		</div>
		<div class='column' class:banding>
			<Field right={true}>{item.item.max_uses ? item.item.uses ?? '0' : ''}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{item.item.tech_level ?? ''}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{item.item.legality_class ?? ''}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{new Fixed(item.item.value).comma()}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{Fixed.extract(item.item.weight ?? '0').value.comma() + ' lb'}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{new Fixed(item.item.calc.extended_value).comma()}</Field>
		</div>
		<div class='column' class:banding>
			<Field right={true}>{Fixed.extract(item.item.calc.extended_weight).value.comma() + ' lb'}</Field>
		</div>
		<div class='column' class:banding>
			<Field>{item.item.reference ?? ''}</Field>
		</div>
	{/each}
</div>

<style>
	.content {
		grid-area: other_equipment;
		display: grid;
		grid-template-columns: auto 1fr repeat(8, auto);
		border: var(--standard-border);
	}

	.column {
		border-left: 1px solid var(--color-header);
	}

	.carried {
		grid-area: equipment;
		grid-template-columns: auto auto 1fr repeat(8, auto);
	}

	.description {
		display: flex;
		flex-direction: column;
	}

	.banding {
		background-color: var(--color-banding);
		color: var(--color-on-banding);
	}
</style>
