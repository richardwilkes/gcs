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
