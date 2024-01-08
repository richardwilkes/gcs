<script lang="ts">
	import Header from '$lib/sheets/widget/Header.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import { pc } from '$lib/entity.ts';
</script>

<div class="content">
	<Header>Basic Damage</Header>
	<div class="fields">
		<Field right={true}>{$pc.calc.thrust}</Field>
		<Label title="Basic Thrust" left={true} />
		<Field right={true}>{$pc.calc.swing}</Field>
		<Label title="Basic Swing" left={true} />
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
		align-items: baseline;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
	}
</style>
