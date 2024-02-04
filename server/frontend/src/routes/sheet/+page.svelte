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
	import { goto } from '$app/navigation';
	import FileSelectionModal from '$lib/filetree/FileSelectionModal.svelte';
	import Shell from '$lib/shell/Shell.svelte';
</script>

<Shell>
	<div slot='content' class='content'>
		<FileSelectionModal showModal={true} path='/sheets' title='Select a Sheet'
												callback={(path) => goto('/sheet/' + path)} />
	</div>
</Shell>

<style>
	.content {
		background-color: var(--color-page);
		flex-grow: 1;
	}
</style>
