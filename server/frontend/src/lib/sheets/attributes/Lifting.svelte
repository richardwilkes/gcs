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
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<Header>Encumbrance, Move & Dodge</Header>
	<div class="fields">
		<div>
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.BasicLift}</Field>
		</div>
		<div><Label title="Basic Lift" /></div>
		<div class="banding">
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.OneHandedLift}</Field>
		</div>
		<div class="banding"><Label title="One-Handed Lift" /></div>
		<div>
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.TwoHandedLift}</Field>
		</div>
		<div><Label title="Two-Handed Lift" /></div>
		<div class="banding">
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.ShoveAndKnockOver}</Field>
		</div>
		<div class="banding"><Label title="Shove & Knock Over" /></div>
		<div>
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.RunningShoveAndKnockOver}</Field>
		</div>
		<div><Label title="Running Shove & Knock Over" /></div>
		<div class="banding">
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.CarryOnBack}</Field>
		</div>
		<div class="banding"><Label title="Carry On Back" /></div>
		<div>
			<Field noBottomBorder={true}>{$sheet?.LiftingAndMovingThings.ShiftSlightly}</Field>
		</div>
		<div><Label title="Shift Slightly" /></div>
	</div>
</div>

<style>
	.content {
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0.3fr 0.7fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}

	.fields > div {
		display: flex;
		align-items: center;
	}

	.fields > div:nth-child(odd) {
		justify-content: end;
	}
</style>
