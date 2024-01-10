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
	import Shell from '$lib/shell/Shell.svelte';
	import Personal from '$lib/sheets/personal/Personal.svelte';
	import Attributes from '$lib/sheets/attributes/Attributes.svelte';
	import Lists from '$lib/sheets/lists/Lists.svelte';
	import { pc } from '$lib/entity.ts';
	import PCLoaderButton from '$lib/sheets/widget/PCLoaderButton.svelte';
</script>

<Shell>
	<PCLoaderButton slot='toolbar'/>
	<div slot='content' class='content'>
	{#if $pc}
		<Personal />
		<Attributes />
		<Lists />
	{/if}
	</div>
</Shell>

<style>
	.content {
		background-color: var(--color-page);
		color: var(--color-on-page);
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		align-items: stretch;
		align-content: stretch;
		gap: var(--section-gap);
		padding: 5px;
		overflow: auto;
		flex-grow: 1;
	}
</style>
