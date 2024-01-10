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
	import SubHeader from '$lib/sheets/widget/SubHeader.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import { pc } from '$lib/entity.ts';
	import { Fixed } from '$lib/fixed.ts';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Weight from '$lib/svg/Weight.svelte';

	type EncData = {
		current: boolean;
		mod: number;
		level: string;
		maxLoad: Fixed;
		move: Fixed;
		dodge: Fixed;
	};

	let encList: EncData[] = [
		{ current: false, mod: 0, level: 'None', maxLoad: Fixed.Zero, move: Fixed.Zero, dodge: Fixed.Zero },
		{ current: false, mod: 1, level: 'Light', maxLoad: Fixed.Zero, move: Fixed.Zero, dodge: Fixed.Zero },
		{ current: false, mod: 2, level: 'Medium', maxLoad: Fixed.Zero, move: Fixed.Zero, dodge: Fixed.Zero },
		{ current: false, mod: 3, level: 'Heavy', maxLoad: Fixed.Zero, move: Fixed.Zero, dodge: Fixed.Zero },
		{ current: false, mod: 4, level: 'X-Heavy', maxLoad: Fixed.Zero, move: Fixed.Zero, dodge: Fixed.Zero }
	];

	$: {
		for (let i = 0; i < encList.length; i++) {
			encList[i].move = new Fixed($pc.calc.move[i]);
			encList[i].dodge = new Fixed($pc.calc.dodge[i]);
		}
		let basicLift = Fixed.extract($pc.calc.basic_lift).value;
		encList[0].maxLoad = basicLift;
		encList[1].maxLoad = basicLift.mul(Fixed.Two);
		encList[2].maxLoad = basicLift.mul(Fixed.Three);
		encList[3].maxLoad = basicLift.mul(Fixed.Six);
		encList[4].maxLoad = basicLift.mul(Fixed.Ten);
		let total = new Fixed(0);
		for (let equipment of $pc.equipment ?? []) {
			total = total.add(Fixed.extract(equipment.calc.extended_weight).value);
		}
		let found = false;
		for (let one of encList) {
			one.current = false;
			if (!found && one.maxLoad.gte(total)) {
				one.current = true;
				found = true;
			}
		}
	}
</script>

<div class="content">
	<Header>Encumbrance, Move & Dodge</Header>
	<div class="fields">
		<SubHeader title="Level" span={3} />
		<SubHeader title="Max Load" />
		<SubHeader title="Move" />
		<SubHeader title="Dodge" />
		{#each encList as one, i}
			{@const banding = i % 2 === 1}
			{@const current = one.current}
			<div class="marker" class:current class:banding>
				{#if one.current}
					<Weight />
				{/if}
			</div>
			<div class:current class:banding><Label title={one.mod} /></div>
			<div class:current class:banding><Label title={one.level} left={true} borderRight={true} /></div>
			<div class:current class:banding>
				<Field right={true} borderRight={true} noBottomBorder={true}>{one.maxLoad.comma() + ' lb'}</Field>
			</div>
			<div class:current class:banding>
				<Field right={true} borderRight={true} noBottomBorder={true}>{one.move.toString()}</Field>
			</div>
			<div class:current class:banding>
				<Field right={true} noBottomBorder={true}>{one.dodge.toString()}</Field>
			</div>
		{/each}
	</div>
</div>

<style>
	.content {
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		grid-template-columns: 0fr 0fr 1fr 0fr 0fr 0fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
	}

	.marker {
		padding-left: 4px;
		display: flex;
		align-items: center;
	}

	.banding {
		background-color: var(--color-banding);
		color: var(--color-on-banding);
	}

	.current {
		background-color: var(--color-marker);
		color: var(--color-on-marker);
		fill: var(--color-on-marker);
	}
</style>
