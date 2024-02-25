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
	import Header from '$lib/sheets/widget/Header.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import { sheet } from '$lib/sheet.ts';
</script>

<div class="content">
	<Header>Basic Damage</Header>
	<div class="fields">
		<div class='right'><Field noBottomBorder>{$sheet?.BasicDamage.Thrust ?? ""}</Field></div>
		<div class='left'><Label title="Basic Thrust" /></div>
		<div class='right banding'><Field noBottomBorder>{$sheet?.BasicDamage.Swing ?? ""}</Field></div>
		<div class='left banding'><Label title="Basic Swing" /></div>
	</div>
</div>

<style>
	.content {
		grid-area: basic-damage;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 1fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-surface);
		color: var(--color-on-surface);
	}

	.left, .right {
		display: flex;
		align-items: center;
	}

	.right {
		justify-content: end;
	}
</style>
