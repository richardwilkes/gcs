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
	import SubHeader from '$lib/sheets/widget/SubHeader.svelte';
	import Field from '$lib/sheets/widget/Field.svelte';
	import Label from '$lib/sheets/widget/Label.svelte';
	import { pc } from '$lib/entity.ts';

	function calcDRString(drMap?: { [key: string]: number }) {
		if (!drMap) {
			return '';
		}
		const all = drMap['all'] || 0;
		let keys = Array.from(Object.keys(drMap))
			.filter((key) => key !== 'all')
			.sort();
		keys.unshift('all');
		let s = '';
		for (let key of keys) {
			let dr = drMap[key] || 0;
			if (key != 'all') {
				dr += all;
			}
			if (s.length !== 0) {
				s += '/';
			}
			s += dr;
		}
		return s;
	}
</script>

<div class="content">
	<Header>{$pc?.settings?.body_type?.name ?? 'Unknown'}</Header>
	<div class="fields">
		<SubHeader title="Roll" />
		<SubHeader title="Location" />
		<SubHeader title="DR" />
		{#each $pc?.settings?.body_type?.locations ?? [] as location, i}
			{@const banding = i % 2 === 1}
			<div class:banding><Label title={location.calc.roll_range} center={true} /></div>
			<div class="name" class:banding>
				<Label title={location.choice_name} left={true} />
				<Label title={location.hit_penalty ?? 0} />
			</div>
			<div class:banding>
				<Field center={true}>{calcDRString(location.calc.dr)}</Field>
			</div>
		{/each}
	</div>
</div>

<style>
	.content {
		grid-area: body-type;
		flex-grow: 1;
		display: flex;
		flex-direction: column;
		border: var(--standard-border);
	}

	.fields {
		display: grid;
		flex-grow: 1;
		grid-template-columns: 0fr 1fr 0fr;
		align-items: stretch;
		align-content: stretch;
		white-space: nowrap;
		background-color: var(--color-content);
		color: var(--color-on-content);
	}

	.name {
		display: flex;
		justify-content: space-between;
		border-left: 1px solid var(--color-header);
		border-right: 1px solid var(--color-header);
	}

	.banding {
		background-color: var(--color-banding);
		color: var(--color-on-banding);
	}
</style>
