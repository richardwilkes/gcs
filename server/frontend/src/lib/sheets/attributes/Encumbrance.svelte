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
	import { sheet } from '$lib/sheet.ts';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Weight from '$lib/svg/Weight.svelte';
</script>

<div class="content">
	<Header>Encumbrance, Move & Dodge</Header>
	<div class="fields">
		<SubHeader span={3}>Level</SubHeader>
		<SubHeader>Max Load</SubHeader>
		<SubHeader>Move</SubHeader>
		<SubHeader>Dodge</SubHeader>
		{#each ['None', 'Light', 'Medium', 'Heavy', 'X-Heavy'] as label, i}
			{@const banding = i % 2 === 1}
			{@const current = $sheet?.Encumbrance.Current === i}
			{@const overloaded = current && $sheet?.Encumbrance.Overloaded}
			<div class="marker" class:current class:overloaded class:banding>
				{#if current}<Weight />{/if}
			</div>
			<div class='right' class:current class:overloaded class:banding><Label>{i}</Label></div>
			<div class='border' class:current class:overloaded class:banding><Label left>{label}</Label></div>
			<div class='right border' class:current class:overloaded class:banding>
				<Field noBottomBorder>{$sheet?.Encumbrance.MaxLoad[i]}</Field>
			</div>
			<div class='right border' class:current class:overloaded class:banding>
				<Field noBottomBorder right>{$sheet?.Encumbrance.Move[i]}</Field>
			</div>
			<div class='right' class:current class:overloaded class:banding>
				<Field noBottomBorder right>{$sheet?.Encumbrance.Dodge[i]}</Field>
			</div>
		{/each}
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
		grid-template-columns: 0fr 0fr 1fr 0fr 0fr 0fr;
		grid-template-rows: 0fr;
		grid-auto-rows: 1fr;
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

	.right {
		justify-content: end;
	}

	.border {
		border-right: var(--standard-border);
	}

	.marker {
		padding-left: 4px;
		display: flex;
		align-items: center;
	}

	.current {
		background-color: var(--color-secondary);
		color: var(--color-on-secondary);
		fill: var(--color-on-secondary);
	}

	.overloaded {
		background-color: var(--color-error);
		color: var(--color-on-error);
		fill: var(--color-on-error);
	}
</style>
