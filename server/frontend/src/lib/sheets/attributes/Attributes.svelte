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
	import PrimaryAttributes from '$lib/sheets/attributes/PrimaryAttributes.svelte';
	import SecondaryAttributes from '$lib/sheets/attributes/SecondaryAttributes.svelte';
	import PointPools from '$lib/sheets/attributes/PointPools.svelte';
	import BasicDamage from '$lib/sheets/attributes/BasicDamage.svelte';
	import BodyType from '$lib/sheets/attributes/BodyType.svelte';
	import Encumbrance from '$lib/sheets/attributes/Encumbrance.svelte';
	import Lifting from '$lib/sheets/attributes/Lifting.svelte';
</script>

<div class="content">
	<PrimaryAttributes />
	<SecondaryAttributes />
	<BasicDamage />
	<PointPools />
	<BodyType />
	<div class="enc-lift">
		<Encumbrance />
		<Lifting />
	</div>
</div>

<style>
	.content {
		display: grid;
		gap: var(--section-gap);
		grid-template-columns: 1fr 1fr;
		grid-template:
			'primary-attributes secondary-attributes body-type enc-lift'
			'basic-damage secondary-attributes body-type enc-lift'
			'pool-attributes pool-attributes body-type enc-lift';
	}

	.enc-lift {
		grid-area: enc-lift;
		flex-grow: 1;
		display: flex;
		gap: var(--section-gap);
		flex-flow: column nowrap;
	}
</style>
