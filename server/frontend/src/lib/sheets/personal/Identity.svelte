<script lang="ts">
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import EditableStringField from '$lib/sheets/widget/EditableStringField.svelte';
	import { pc } from '$lib/entity.ts';
</script>

<div class="content">
	<Header>Identity</Header>
	<div class="fields">
		<Label title="Name" />
		<EditableStringField name="profile.name" value={$pc.profile?.name ?? ''} />
		<Label title="Title" />
		<EditableStringField name="profile.title" value={$pc.profile?.title ?? ''} />
		<Label title="Organization" />
		<EditableStringField name="profile.org" value={$pc.profile?.organization ?? ''} />
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
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
		padding-right: 4px;
		padding-bottom: 2px;
	}
</style>
