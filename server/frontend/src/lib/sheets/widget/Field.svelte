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
	export let style = '';
	export let tip = '';
	export let right = false;
	export let center = false;
	export let wrap = false;
	export let noBottomBorder = false;
	export let editable = false;
</script>

<div class='field' class:right class:center class:editable class:noBottomBorder class:wrap {style} role='textbox'
		 contenteditable={editable} title={tip}>
	<slot>&nbsp;</slot>
</div>

<style>
	.field {
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid transparent;
		border-radius: 0;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
		min-width: 20px;
		outline-style: none;
		margin-left: 2px;
		margin-right: 2px;
	}

	.editable {
		border-bottom: 1px solid var(--color-outline-variant);
		color: var(--color-on-editable);
		background-color: var(--color-editable);
	}

	.right {
		text-align: right;
	}

	.center {
		text-align: center;
	}

	.noBottomBorder {
		border-bottom: none;
	}

	.wrap {
		white-space: normal;
	}
</style>
