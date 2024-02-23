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
	import Bookmark from '$lib/svg/Bookmark.svelte';
	import CheckMark from '$lib/svg/CheckMark.svelte';
	import Coins from '$lib/svg/Coins.svelte';
	import Stack from '$lib/svg/Stack.svelte';
	import Weight from '$lib/svg/Weight.svelte';

	export let key: string;
	export let tip = '';
</script>

<div class='icon' title={tip}>
{#if key.startsWith('stacked-')}
	<Stack />
	<svelte:self key={key.slice(8)} />
{:else if key === 'bookmark'}
	<Bookmark />
{:else if key === 'checkmark'}
	<CheckMark />
{:else if key === 'coins'}
	<Coins />
{:else if key === 'weight'}
	<Weight />
{:else}
	{key}
{/if}
</div>

<style>
	.icon {
		display: flex;
		gap: 3px;
	}
</style>