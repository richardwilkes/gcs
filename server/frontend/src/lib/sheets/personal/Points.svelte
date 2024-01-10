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
	import { pc } from '$lib/entity.ts';

	let total: number;
	let unspent: number;
	let race: number;
	let attributes: number;
	let advantages: number;
	let disadvantages: number;
	let quirks: number;
	let skills: number;
	let spells: number;

	$: {
		race = 0;
		attributes = 0;
		advantages = 0;
		disadvantages = 0;
		quirks = 0;
		skills = 0;
		spells = 0;
		total = $pc.total_points;
		unspent = 0;
		// TODO: Calculate points for each category
		// TODO: Need to account for pc.settings?.exclude_unspent_points_from_total
	}
</script>

<div class="content">
	<Header>{total} Points</Header>
	<div class="fields">
		<Field right={true}>{unspent}</Field>
		<Label title="Unspent" left={true} />
		<div class="banding">
			<Field right={true}>{race}</Field>
		</div>
		<div class="banding"><Label title="Race" left={true} /></div>
		<Field right={true}>{attributes}</Field>
		<Label title="Attributes" left={true} />
		<div class="banding">
			<Field right={true}>{advantages}</Field>
		</div>
		<div class="banding"><Label title="Advantages" left={true} /></div>
		<Field right={true}>{disadvantages}</Field>
		<Label title="Disadvantages" left={true} />
		<div class="banding">
			<Field right={true}>{quirks}</Field>
		</div>
		<div class="banding"><Label title="Quirks" left={true} /></div>
		<Field right={true}>{skills}</Field>
		<Label title="Skills" left={true} />
		<div class="banding">
			<Field right={true}>{spells}</Field>
		</div>
		<div class="banding"><Label title="Spells" left={true} /></div>
	</div>
</div>

<style>
	.content {
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		flex-grow: 1;
		display: grid;
		grid-template-columns: 0fr 1fr;
		align-content: space-between;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
	}

	.banding {
		background-color: var(--color-banding);
		color: var(--color-on-banding);
	}
</style>
