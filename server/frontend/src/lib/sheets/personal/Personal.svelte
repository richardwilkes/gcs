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
	import Portrait from '$lib/sheets/personal/Portrait.svelte';
	import Identity from '$lib/sheets/personal/Identity.svelte';
	import Misc from '$lib/sheets/personal/Misc.svelte';
	import Points from '$lib/sheets/personal/Points.svelte';
	import Description from '$lib/sheets/personal/Description.svelte';
	import { sheet } from '$lib/sheet.ts';

	$: imageURL = $sheet?.Portrait ? 'data:;base64,' + $sheet.Portrait : undefined;
</script>

<div class="content">
	<Portrait {imageURL} />
	<div class="middle">
		<div class="top">
			<Identity />
			<Misc />
		</div>
		<Description />
	</div>
	<Points />
</div>

<style>
	.content {
		display: flex;
		gap: var(--section-gap);
		align-items: stretch;
	}

	.top {
		display: flex;
		gap: var(--section-gap);
	}

	.middle {
		flex-grow: 1;
		gap: var(--section-gap);
		display: flex;
		flex-direction: column;
		justify-content: stretch;
	}
</style>
