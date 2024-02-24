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
	import type { Table } from '$lib/sheet.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import Icon from '$lib/svg/Icon.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';

	export let table: Table | null | undefined;
	export let area: string;

	let style = '';
	if (table && table.Rows.length !== 0) {
		style = `grid-area: ${area};grid-template-columns:`;
		for (let i = 0; i < table.Columns.length; i++) {
			style += table.Columns[i].Primary ? ' 1fr' : ' auto';
		}
		style += `;grid-template-rows: repeat(${table.Rows.length},0fr) 1fr;`;
	}
</script>

{#if table && table.Rows.length !== 0}
	<div class='content' {style}>
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
				<div class:divider={cellIndex !== 0} class:banding>
					<!-- TODO: Add support for the other fields in the Cell data -->
					<Field tip={cell.Tooltip} right={cell.Alignment === 'end'} center={cell.Alignment === 'middle'}
								 noBottomBorder={true} wrap={table.Columns[cellIndex].Primary}>
						{#if cell.Type === 'toggle'}
							{#if cell.Checked}
								<div class='icon'>
									<Icon key='checkmark' />
								</div>
							{:else}
								&nbsp;
							{/if}
						{:else if cell.Type === 'page_ref'}
							{cell.Primary}
						{:else if cell.Type === 'markdown'}
							<!-- TODO: Render markdown -->
							{cell.Primary}
							{#if cell.Secondary}
								<br />
								<span class='secondary'>{cell.Secondary}</span>
							{/if}
						{:else}
							{cell.Primary}
							{#if cell.Secondary}
								<br />
								<span class='secondary'>{cell.Secondary}</span>
							{/if}
						{/if}
					</Field>
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

	.secondary {
		font-size: 85%;
	}

	.icon {
		padding-top: 4px;
	}
</style>