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
	import { sheet, updateSheetField } from '$lib/sheet.js';
	import { sheetPath } from '$lib/url.js';
	import Field from '$lib/sheets/widget/Field.svelte';

	interface Props {
		key: string;
		right?: boolean;
	}

	let { key, right = false }: Props = $props();

	async function updateField(event: FocusEvent) {
		if ($sheetPath) {
			const target = event.target as HTMLElement;
			try {
				let updatedSheet = await updateSheetField($sheetPath, 'field.text', key, target.innerText);
				target.innerText = extractField(updatedSheet, key, key);
				sheet.update((_) => updatedSheet);
			} catch {
				target.innerText = extractField($sheet, key, key);
			}
		}
	}

	function extractField(obj: unknown, prop: string, originalProp: string): string {
		if (typeof obj !== 'object') {
			return '';
		}
		if (Array.isArray(obj)) {
			for (const item of obj as unknown[]) {
				if (typeof item !== 'object') {
					continue;
				}
				const o = item as { [key: string]: unknown };
				if (originalProp.startsWith('PointPools.')) {
					if (prop.endsWith('.Current')) {
						if (o['Key'] !== prop.substring(0, prop.length - 8)) {
							continue;
						}
						return extractField(o, 'Value', originalProp);
					}
					if (o['Key'] !== prop) {
						continue;
					}
					return extractField(o, 'Max', originalProp);
				}
				if (o['Key'] !== prop) {
					continue;
				}
				return extractField(o, 'Value', originalProp);
			}
			return '';
		}
		const o = obj as { [key: string]: unknown };
		const i = prop.indexOf('.');
		if (i > -1) {
			return extractField(o[prop.substring(0, i)], prop.substring(i + 1), originalProp);
		}
		const value = o[prop];
		if (typeof value === 'string') {
			return value as string;
		}
		return '';
	}
</script>

<Field editable {right} style="width:100%;" onblur={(target) => updateField(target)}
	>{extractField($sheet, key, key)}</Field
>
