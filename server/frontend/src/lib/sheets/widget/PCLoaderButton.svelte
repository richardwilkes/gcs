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
	import { pc, loadPC } from '$lib/entity.ts';

	let files: FileList;

	$: {
		if (files && files.length) {
			loadPC(files[0])
				.then((newPC) => ($pc = newPC))
				.catch((err) => {
					console.error(err);
					alert(`Error loading ${files[0].name}`);
				});
		}
	}
</script>

<input bind:files type="file" accept=".gcs" />

<style>
	input {
		flex-grow: 1;
		font-weight: bold;
	}
</style>
