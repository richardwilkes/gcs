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
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import EditableTextField from '$lib/sheets/widget/EditableTextField.svelte';
</script>

<div class="content">
	<Header>Identity</Header>
	<div class="fields">
		<div class="banding"><Label>Name</Label></div>
		<div class="banding">
			<EditableTextField key="Identity.Name" />
		</div>
		<div><Label>Title</Label></div>
		<div>
			<EditableTextField key="Identity.Title" />
		</div>
		<div class="banding"><Label>Organization</Label></div>
		<div class="banding">
			<EditableTextField key="Identity.Organization" />
		</div>
	</div>
</div>

<style>
	.content {
		grid-area: identity;
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
		background-color: var(--color-below-surface);
		color: var(--color-on-below-surface);
		padding-bottom: 2px;
	}

	.fields > div {
		display: flex;
		align-items: center;
		justify-content: end;
	}
</style>
