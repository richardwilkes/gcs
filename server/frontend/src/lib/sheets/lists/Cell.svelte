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
	import Field from '$lib/sheets/widget/Field.svelte';
	import type { Cell, Column } from '$lib/sheet.ts';
	import Icon from '$lib/sheets/lists/Icon.svelte';
	import Tag from '$lib/sheets/lists/Tag.svelte';
	import PageRef from '$lib/sheets/widget/PageRef.svelte';
	import SvelteMarkdown from 'svelte-markdown';

	interface Props {
		cell: Cell;
		column: Column; // TODO: Add support for the other fields in the Cell data
	}

	let { cell, column }: Props = $props();
</script>

{#if cell.Type === 'page_ref'}
	<PageRef pageRef={cell.Primary} />
{:else}
	<Field
		tip={cell.Tooltip}
		right={cell.Alignment === 'end'}
		center={cell.Alignment === 'middle'}
		noBottomBorder
		wrap={column.Primary}
		strikeout={cell.Disabled}>
		{#if cell.Type === 'toggle'}
			{#if cell.Checked}
				<div class="icon">
					<Icon key="checkmark" />
				</div>
			{:else}
				&nbsp;
			{/if}
		{:else if cell.Type === 'markdown'}
			<!-- TODO: Render markdown -->
			<div class="markdown">
				<SvelteMarkdown source={cell.Primary} />
				{#if cell.InlineTag}
					<Tag>{cell.InlineTag}</Tag>
				{/if}
				{#if cell.Secondary}
					{#each cell.Secondary.split('\n') as line}
						<br /><span class="secondary"><SvelteMarkdown source={line} /></span>
					{/each}
				{/if}
			</div>
		{:else}
			{cell.Primary}
			{#if cell.InlineTag}
				<Tag>{cell.InlineTag}</Tag>
			{/if}
			{#if cell.Secondary}
				{#each cell.Secondary.split('\n') as line}
					<br /><span class="secondary">{line}</span>
				{/each}
			{/if}
		{/if}
		{#if cell.UnsatisfiedReason}
			<Tag warning tip={cell.UnsatisfiedReason}>Unsatisfied prerequisite(s)</Tag>
		{/if}
	</Field>
{/if}

<style>
	.secondary {
		font-size: 85%;
	}

	.icon {
		padding-top: 4px;
		display: flex;
		justify-content: center;
	}

	.markdown > :global(pre) {
		background-color: var(--color-above-surface);
		color: var(--color-on-above-surface);
	}

	.markdown :global(table),
	.markdown :global(th),
	.markdown :global(td) {
		border: 2px solid var(--color-surface-edge);
		border-collapse: collapse;
		padding: 1px 6px;
	}
</style>
