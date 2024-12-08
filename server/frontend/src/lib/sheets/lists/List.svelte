<!--
  - Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang="ts">
	import type { Table } from '$lib/sheet.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import Icon from '$lib/sheets/lists/Icon.svelte';
	import Cell from '$lib/sheets/lists/Cell.svelte';

	interface Props {
		table: Table | null | undefined;
		area: string;
	}

	let { table, area }: Props = $props();

	function computeStyle(): string {
		let s = '';
		if (table && table.Rows.length !== 0) {
			s = `grid-area: ${area};grid-template-columns:`;
			for (let i = 0; i < table.Columns.length; i++) {
				s += table.Columns[i].Primary ? ' 1fr' : ' auto';
			}
			s += `;grid-template-rows: repeat(${table.Rows.length},0fr) 1fr;`;
		}
		return s;
	}

	let style = $derived(computeStyle());
</script>

{#if table && table.Rows.length !== 0}
	<div class="content" {style}>
		{#each table.Columns as column}
			<Header tip={column.Detail}>
				{#if column.TitleIsImageKey}
					<Icon key={column.Title} tip={column.Detail} />
				{:else}
					{column.Title}
				{/if}
			</Header>
		{/each}
		{#each table.Rows as row, rowIndex}
			{@const banding = rowIndex % 2 === 1}
			{#each row.Cells as cell, cellIndex}
				<div
					class:divider={cellIndex !== 0}
					class:banding
					style={row.Depth && table.Columns[cellIndex].Primary
						? `padding-left: ${row.Depth}em`
						: ''}
				>
					<Cell {cell} column={table.Columns[cellIndex]} />
				</div>
			{/each}
		{/each}
	</div>
{/if}

<style>
	.content {
		display: grid;
		justify-content: stretch;
		align-content: stretch;
		border: var(--standard-border);
	}

	.divider {
		border-left: var(--standard-border);
	}
</style>
