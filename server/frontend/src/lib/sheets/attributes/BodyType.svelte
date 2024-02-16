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
	import Field from '$lib/sheets/widget/Field.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<Header>{$sheet?.Body.Name ?? ''}</Header>
	<div class="fields">
		<SubHeader title="Roll" />
		<SubHeader title="Location" />
		<SubHeader title="DR" />
		{#each $sheet?.Body.Locations ?? [] as loc, i}
			{@const banding = i % 2 === 1}
			<div class:banding><Label title={loc.Roll} center={true} /></div>
			<div class="name" class:banding>
				<Label title={loc.Location} left={true} tip={loc.LocationDetail} />
				<Label title={loc.HitPenalty ?? 0}/>
			</div>
			<div class:banding>
				<Field center={true} tip={loc.DRDetail}>{loc.DR}</Field>
			</div>
		{/each}
	</div>
</div>

<style>
	.content {
		grid-area: body-type;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 1fr 0fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}

	.name {
		display: flex;
		justify-content: space-between;
		border-left: 1px solid var(--color-header);
		border-right: 1px solid var(--color-header);
	}
</style>
