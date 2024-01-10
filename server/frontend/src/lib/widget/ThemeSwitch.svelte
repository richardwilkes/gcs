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
	import { currentTheme } from '$lib/theme.ts';
	import Label from '$lib/sheets/widget/Label.svelte';
</script>

<Label title="Theme" />
<select bind:value={$currentTheme.kind}>
	<option value="system">System</option>
	<option value="light">Light</option>
	<option value="dark">Dark</option>
</select>

<style>
	select {
		font: var(--font-page-primary-labels);
	}
</style>
