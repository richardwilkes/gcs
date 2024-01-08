<script lang="ts">
	import Header from '$lib/sheets/widget/Header.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import { pc } from '$lib/entity.ts';
	import { Fixed } from '$lib/fixed.ts';

	let basicLift = new Fixed(0);

	$: {
		basicLift = basicLift.add(Fixed.extract($pc.calc.basic_lift).value);
	}
</script>

<div class="content">
	<Header>Encumbrance, Move & Dodge</Header>
	<div class="fields">
		<div>
			<Field right={true}>{basicLift.comma() + ' lb'}</Field>
		</div>
		<div><Label title="Basic Lift" left={true} /></div>
		<div class="banding">
			<Field right={true}>{basicLift.mul(Fixed.Two).comma() + ' lb'}</Field>
		</div>
		<div class="banding"><Label title="One-Handed Lift" left={true} /></div>
		<div>
			<Field right={true}>{basicLift.mul(Fixed.Eight).comma() + ' lb'}</Field>
		</div>
		<div><Label title="Two-Handed Lift" left={true} /></div>
		<div class="banding">
			<Field right={true}>{basicLift.mul(Fixed.Twelve).comma() + ' lb'}</Field>
		</div>
		<div class="banding"><Label title="Shove & Knock Over" left={true} /></div>
		<div>
			<Field right={true}>{basicLift.mul(Fixed.TwentyFour).comma() + ' lb'}</Field>
		</div>
		<div><Label title="Running Shove & Knock Over" left={true} /></div>
		<div class="banding">
			<Field right={true}>{basicLift.mul(Fixed.Fifteen).comma() + ' lb'}</Field>
		</div>
		<div class="banding"><Label title="Carry On Back" left={true} /></div>
		<div>
			<Field right={true}>{basicLift.mul(Fixed.Fifty).comma() + ' lb'}</Field>
		</div>
		<div><Label title="Shift Slightly" left={true} /></div>
	</div>
</div>

<style>
	.content {
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0.3fr 0.7fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
	}

	.banding {
		background-color: var(--color-banding);
		color: var(--color-on-banding);
	}
</style>
