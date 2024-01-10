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
	export let title: string | number = '';
	export let left = false;
	export let center = false;
	export let borderLeft = false;
	export let borderRight = false;
</script>

<div class="label" class:left class:center class:borderLeft class:borderRight>{title}</div>

<style>
	.label {
		font: var(--font-page-primary-labels);
		padding: var(--padding-standard);
		text-align: right;
		white-space: nowrap;
	}

	.left {
		text-align: left;
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
</style>
