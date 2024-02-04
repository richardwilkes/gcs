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

<script lang='ts'>
	import Toolbar from '$lib/shell/Toolbar.svelte';
	import Footer from '$lib/shell/Footer.svelte';
</script>

<div class='shell'>
	<Toolbar>
		<slot name='toolbar'></slot>
	</Toolbar>
	<div class='content'>
		<slot name='content'></slot>
	</div>
	<Footer />
</div>

<style>
	.shell {
		position: absolute;
		top: 0;
		left: 0;
		bottom: 0;
		right: 0;
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		gap: 0;
		background-color: var(--color-background);
		color: var(--color-on-background);
	}

	.content {
		display: flex;
		flex-grow: 1;
	}
</style>
