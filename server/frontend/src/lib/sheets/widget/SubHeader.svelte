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
	export let span = 1;
	export let tip = '';
</script>

<div class="header" title={tip} style="grid-column-end: span {span}">
	<slot>&nbsp;</slot>
</div>

<style>
	.header {
		background-color: var(--color-header);
		color: var(--color-on-header);
		text-align: center;
		font-variant: small-caps;
		padding: var(--padding-standard);
	}
</style>
