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
	import ThemeSwitch from '$lib/widget/ThemeSwitch.svelte';
</script>

<div class='toolbar'>
	<div class='fill'><slot></slot></div>
	<ThemeSwitch />
</div>

<style>
	.toolbar {
		border-bottom: 1px solid var(--color-divider);
		padding: 4px 8px;
		display: flex;
		justify-content: flex-start;
		align-items: baseline;
		align-content: center;
		column-gap: 8px;
		row-gap: 4px;
	}

	.fill {
		display: flex;
		justify-content: flex-start;
		align-items: baseline;
		align-content: center;
		column-gap: 8px;
		row-gap: 4px;
		flex-grow: 1;
	}
</style>
