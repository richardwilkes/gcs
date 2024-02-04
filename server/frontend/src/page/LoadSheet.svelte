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
	import FileSelectionModal from '$lib/filetree/FileSelectionModal.svelte';
	import { page } from '$lib/page.ts';
</script>

<div class='content'>
	<FileSelectionModal showModal={true} path='/sheets' title='Select a Sheet'
											callback={(path) => $page = {ID:'sheet', NextID: 'sheet', Sheet: path}} />
</div>

<style>
	.content {
		background-color: var(--color-page);
		flex-grow: 1;
	}
</style>
