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
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.BasicLift}</Field>
		</div>
		<div><Label>Basic Lift</Label></div>
		<div class="banding">
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.OneHandedLift}</Field>
		</div>
		<div class="banding"><Label>One-Handed Lift</Label></div>
		<div>
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.TwoHandedLift}</Field>
		</div>
		<div><Label>Two-Handed Lift</Label></div>
		<div class="banding">
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.ShoveAndKnockOver}</Field>
		</div>
		<div class="banding"><Label title="Shove & Knock Over" /></div>
		<div>
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.RunningShoveAndKnockOver}</Field>
		</div>
		<div><Label>Running Shove & Knock Over</Label></div>
		<div class="banding">
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.CarryOnBack}</Field>
		</div>
		<div class="banding"><Label>Carry On Back</Label></div>
		<div>
			<Field noBottomBorder>{$sheet?.LiftingAndMovingThings.ShiftSlightly}</Field>
		</div>
		<div><Label>Shift Slightly</Label></div>
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
		grid-template-columns: 0fr 1fr;
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
