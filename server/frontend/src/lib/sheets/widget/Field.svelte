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
	export let right = false;
	export let center = false;
	export let borderLeft = false;
	export let borderRight = false;
	export let noBottomBorder = false;
	export let smaller = false;
	export let wrap = false;
</script>

<div
	class="field"
	class:right
	class:center
	class:borderLeft
	class:borderRight
	class:noBottomBorder
	class:smaller
	class:wrap>
	<slot>&nbsp;</slot>
</div>

<style>
	.field {
		font: var(--font-page-primary-fields);
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid rgba(0, 0, 0, 0);
		color: var(--color-on-editable);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.right {
		text-align: right;
	}

	.center {
		text-align: center;
	}

	.borderLeft {
		border-left: 1px solid var(--color-header);
	}

	.borderRight {
		border-right: 1px solid var(--color-header);
	}

	.noBottomBorder {
		border-bottom: none;
	}

	.smaller {
		font: var(--font-page-secondary-fields);
	}

	.wrap {
		white-space: normal;
	}
</style>
