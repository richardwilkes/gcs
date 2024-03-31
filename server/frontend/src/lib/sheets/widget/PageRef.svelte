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
	import { refPrefix } from '$lib/dev';
	import { sheet } from '$lib/sheet.ts';

	export let pageRef = '';

	let tip = '';
	let single = '';
	let prefix = '';
	let uri = '';

	$: {
		const parts = pageRef.split(/,|;/);
		tip = parts.join('\n');
		single = parts.length < 2 ? pageRef : parts[0] + '+';
		prefix = parts[0];
		let i = prefix.length - 1;
		while (i >= 0) {
			const ch = prefix[i];
			if (ch >= '0' && ch <= '9') {
				i--;
			} else {
				i++;
				break;
			}
		}
		if (i > 0) {
			const page = prefix.substring(i, prefix.length);
			prefix = prefix.substring(0, i);
			uri = refPrefix(prefix);
			if (page) {
				let pageNum = parseInt(page, 10);
				if (pageNum) {
					const offset = $sheet?.PageRefOffsets[prefix];
					if (offset) {
						pageNum += offset;
					}
					uri += `#page=${pageNum}`;
				}
			}
		} else {
			uri = '';
			prefix = '';
			single = '';
			tip = '';
		}
	}
</script>

<div class="ref" title={tip}>
	{#if single}
		<a href={uri} target={'pageref_' + prefix}>{single}</a>
	{/if}
</div>

<style>
	.ref {
		padding: var(--padding-standard);
		border: none;
		border-radius: 0;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
		min-width: 20px;
		outline-style: none;
		margin-left: 2px;
		margin-right: 2px;
	}
</style>
