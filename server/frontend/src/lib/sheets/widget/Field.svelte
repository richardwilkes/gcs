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
	export let style = '';
	export let tip = '';
	export let right = false;
	export let center = false;
	export let wrap = false;
	export let noBottomBorder = false;
	export let editable = false;
	export let multiLine = false;
	export let strikeout = false;

	interface FieldElement extends HTMLDivElement {
		suppressSelectAll?: boolean;
	}

	let field: FieldElement;
	let inDrop: boolean;
	let contentPriorToDrop: string;

	function filterKey(event: KeyboardEvent) {
		if (editable && !multiLine) {
			if (event.key === 'Enter') {
				event.preventDefault();
			}
		}
	}

	function filterPaste(event: ClipboardEvent) {
		if (editable && !multiLine) {
			event.preventDefault();
			const text = event.clipboardData?.getData('text/plain').replaceAll('\n', '');
			if (text) {
				const selection = window.getSelection();
				if (selection?.rangeCount) {
					selection.deleteFromDocument();
					selection.getRangeAt(0).insertNode(document.createTextNode(text));
					selection.collapseToEnd();
				}
			}
		}
	}

	function filterDrag(_: DragEvent) {
		if (editable && !multiLine) {
			inDrop = true;
			contentPriorToDrop = field.innerText;
		}
	}

	function filterInput(_: Event) {
		if (inDrop) {
			inDrop = false;
			if (editable && !multiLine) {
				let text = field.innerText;
				if (text.includes('\n')) {
					text = text.replaceAll('\n', '');
					field.innerText = text;
					let pos = text.length - Math.min(text.length, contentPriorToDrop.length);
					if (text !== contentPriorToDrop) {
						for (let i = 0; i < text.length && i < contentPriorToDrop.length; i++) {
							if (
								text[text.length - (i + 1)] != contentPriorToDrop[contentPriorToDrop.length - (i + 1)]
							) {
								pos = text.length - i;
								break;
							}
						}
					}
					let sel = window.getSelection();
					if (sel) {
						sel.removeAllRanges();
						const range = document.createRange();
						range.setStart(field.childNodes[0], pos);
						range.setEnd(field.childNodes[0], pos);
						sel.addRange(range);
					}
				}
			}
		}
	}

	function mouseDown(_: MouseEvent) {
		field.suppressSelectAll = true;
	}

	function focusIn(_: FocusEvent) {
		if (field.suppressSelectAll) {
			field.suppressSelectAll = false;
		} else {
			let sel = window.getSelection();
			if (sel) {
				sel.removeAllRanges();
				const range = document.createRange();
				range.setStart(field.childNodes[0], 0);
				range.setEnd(field.childNodes[0], field.innerText.length);
				sel.addRange(range);
			}
		}
	}

	function focusOut(_: FocusEvent) {
		field.suppressSelectAll = false;
	}

	function keyDown(_: KeyboardEvent) {
		field.suppressSelectAll = false;
	}
</script>

<!-- svelte-ignore a11y-interactive-supports-focus -->
<div
	class="field"
	bind:this={field}
	class:right
	class:center
	class:editable
	class:noBottomBorder
	class:wrap
	class:strikeout
	{style}
	role="textbox"
	contenteditable={editable ? 'plaintext-only' : 'false'}
	title={tip}
	on:keydown={filterKey}
	on:paste={filterPaste}
	on:drop={filterDrag}
	on:input={filterInput}
	on:blur
	on:keydown={keyDown}
	on:mousedown={mouseDown}
	on:focusin={focusIn}
	on:focusout={focusOut}>
	<slot />
</div>

<style>
	.field {
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid transparent;
		border-radius: 0;
		text-overflow: ellipsis;
		overflow: hidden;
		white-space: nowrap;
		min-width: 20px;
		outline-style: none;
		margin-left: 2px;
		margin-right: 2px;
	}

	.editable {
		border-bottom: 1px solid var(--color-surface-edge);
		color: var(--color-on-deep-below-surface);
		background-color: var(--color-deep-below-surface);
	}

	.right {
		text-align: right;
	}

	.center {
		text-align: center;
	}

	.noBottomBorder {
		border-bottom: none;
	}

	.wrap {
		white-space: normal;
	}

	.strikeout {
		text-decoration: line-through;
		opacity: 33%;
	}
</style>
