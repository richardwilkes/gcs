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
	import BookmarkSVG from '$svg/Bookmark.svg?raw';
	import CheckmarkSVG from '$svg/Checkmark.svg?raw';
	import CoinsSVG from '$svg/Coins.svg?raw';
	import StackSVG from '$svg/Stack.svg?raw';
	import WeightSVG from '$svg/Weight.svg?raw';

	export let key: string;
	export let tip = '';
</script>

<div class="icon" title={tip}>
	{#if key.startsWith('stacked-')}
		{@html StackSVG}
		<svelte:self key={key.slice(8)} />
	{:else if key === 'bookmark'}
		{@html BookmarkSVG}
	{:else if key === 'checkmark'}
		{@html CheckmarkSVG}
	{:else if key === 'coins'}
		{@html CoinsSVG}
	{:else if key === 'weight'}
		{@html WeightSVG}
	{:else}
		{key}
	{/if}
</div>

<style>
	.icon {
		display: flex;
		gap: 3px;
		height: 0.75em;
	}
</style>
