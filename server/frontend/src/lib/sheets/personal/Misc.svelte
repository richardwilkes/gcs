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
	import { formatDateStamp } from '$lib/datetime.ts';
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import EditableStringField from '$lib/sheets/widget/EditableStringField.svelte';
	import { pc } from '$lib/entity.ts';
</script>

<div class="content">
	<Header>Miscellaneous</Header>
	<div class="fields">
		<Label title="Created" />
		<Field>{formatDateStamp($pc.created_date)}</Field>
		<Label title="Modified" />
		<Field>{formatDateStamp($pc.modified_date)}</Field>
		<Label title="Player" />
		<EditableStringField name="profile.player" value={$pc.profile?.player_name ?? ''} />
	</div>
</div>

<style>
	.content {
		grid-area: miscellaneous;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		grid-template-columns: 0fr 1fr;
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
		padding-right: 4px;
		padding-bottom: 2px;
	}
</style>
