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
	export let name = '';
	export let value = 0;
	export let right = false;
</script>

<input class="field" class:right type="text" {name} {value} />

<style>
	.field {
		font-weight: bold;
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid var(--color-control-edge);
		color: var(--color-on-editable);
		background-color: var(--color-editable);
		min-width: 20px;
		outline-style: none;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}

	.right {
		text-align: right;
	}
</style>
