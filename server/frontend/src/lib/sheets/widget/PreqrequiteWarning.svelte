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
	import Warning from '$lib/svg/Warning.svelte';

	export let reason = '';
</script>

<div class="content" title={reason}>
	<Warning />
	Unsatisfied prerequisite(s)!
</div>

<style>
	.content {
		font-size: 85%;
		padding: var(--padding-standard);
		white-space: nowrap;
		background-color: var(--color-error);
		color: var(--color-on-error);
		fill: var(--color-on-error);
		align-self: flex-start;
		border-radius: 4px;
		margin-left: 4px;
	}
</style>
