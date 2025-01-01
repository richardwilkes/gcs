<!--
  - Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	interface Props {
		pageRef?: string;
	}

	let { pageRef = '' }: Props = $props();

	let parts = $derived(pageRef.split(/,|;/));

	function calcPageNumStartsAt(): number {
		let i = parts[0].length - 1;
		while (i >= 0) {
			const ch = parts[0][i];
			if (ch >= '0' && ch <= '9') {
				i--;
			} else {
				i++;
				break;
			}
		}
		return i;
	}

	let pageNumStartsAt = $derived(calcPageNumStartsAt());
	let single = $derived(pageNumStartsAt > 0 ? (parts.length < 2 ? pageRef : parts[0] + '+') : '');
	let tip = $derived(pageNumStartsAt > 0 ? parts.join('\n') : '');
	let prefix = $derived(pageNumStartsAt > 0 ? parts[0].substring(0, pageNumStartsAt) : '');

	function computeURI(): string {
		if (pageNumStartsAt < 1) {
			return '';
		}
		let u = refPrefix(prefix);
		const ref = $sheet?.PageRefs[prefix];
		if (ref?.Name) {
			u += '/' + encodeURIComponent(ref.Name);
		}
		const page = parts[0].substring(pageNumStartsAt, parts[0].length);
		if (page) {
			let pageNum = parseInt(page, 10);
			if (pageNum) {
				if (ref?.Offset) {
					pageNum += ref?.Offset;
				}
				u += `#page=${pageNum}`;
			}
		}
		return u;
	}

	let uri = $derived(computeURI());
</script>

<div class="ref" title={tip}>
	{#if single}
		<a href={uri} target={'gcs_ref_' + prefix}>{single}</a>
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

	.ref a {
		color: inherit;
	}
</style>
