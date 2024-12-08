<!--
  - Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang="ts">
	import FileTree from '$lib/filetree/FileTree.svelte';
	import { ShowAs } from '$lib/show.ts';
	import { navTo } from '$lib/nav';
</script>

<div class="content">
	<FileTree
		showAs={ShowAs.Dialog}
		path="/sheets"
		title="Select a Sheet"
		onSuccess={(path) => navTo('sheet/' + path, undefined, undefined, true)}
		onCancel={() => window.history.back()}
	/>
</div>

<style>
	.content {
		background-color: var(--color-below-surface);
		flex-grow: 1;
		display: flex;
		align-items: center;
	}
</style>
