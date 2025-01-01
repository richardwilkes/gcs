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
	interface Props {
		style?: string;
		tip?: string;
		right?: boolean;
		center?: boolean;
		wrap?: boolean;
		noBottomBorder?: boolean;
		editable?: boolean;
		multiLine?: boolean;
		strikeout?: boolean;
		onblur?: (event: FocusEvent) => void;
		children?: import('svelte').Snippet;
	}

	let {
		style = '',
		tip = '',
		right = false,
		center = false,
		wrap = false,
		noBottomBorder = false,
		editable = false,
		multiLine = false,
		strikeout = false,
		onblur = (_: FocusEvent) => {},
		children
	}: Props = $props();

	interface FieldElement extends HTMLDivElement {
		suppressSelectAll?: boolean;
	}

	let field: FieldElement | undefined = $state();
	let inDrop: boolean;
	let contentPriorToDrop: string;

	function filterKey(event: KeyboardEvent) {
		if (editable && !multiLine) {
			if (event.key === 'Enter') {
				event.preventDefault();
			}
		}
		if (field) {
			field.suppressSelectAll = false;
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
			if (field) {
				contentPriorToDrop = field.innerText;
			}
		}
	}

	function filterInput(_: Event) {
		if (inDrop) {
			inDrop = false;
			if (editable && !multiLine && field) {
				let text = field.innerText;
				if (text.includes('\n')) {
					text = text.replaceAll('\n', '');
					field.innerText = text;
					let pos = text.length - Math.min(text.length, contentPriorToDrop.length);
					if (text !== contentPriorToDrop) {
						for (let i = 0; i < text.length && i < contentPriorToDrop.length; i++) {
							if (
								text[text.length - (i + 1)] !=
								contentPriorToDrop[contentPriorToDrop.length - (i + 1)]
							) {
								pos = text.length - i;
								break;
							}
						}
					}
					selectRange(pos, pos);
				}
			}
		}
	}

	function selectRange(from: number, to: number) {
		let sel = window.getSelection();
		if (sel) {
			sel.removeAllRanges();
			if (field) {
				let first = field.childNodes[0];
				if (!first) {
					first = field;
				}
				if (first) {
					const range = document.createRange();
					range.setStart(first, from);
					range.setEnd(first, to);
					sel.addRange(range);
				}
			}
		}
	}

	function mouseDown(_: MouseEvent) {
		if (field) {
			field.suppressSelectAll = true;
		}
	}

	function focusIn(_: FocusEvent) {
		if (field) {
			if (field.suppressSelectAll) {
				field.suppressSelectAll = false;
			} else {
				selectRange(0, field.innerText.length);
			}
		}
	}

	function focusOut(_: FocusEvent) {
		if (field) {
			field.suppressSelectAll = false;
		}
	}
</script>

<!-- svelte-ignore a11y_interactive_supports_focus -->
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
	onkeydown={filterKey}
	onpaste={filterPaste}
	ondrop={filterDrag}
	oninput={filterInput}
	{onblur}
	onmousedown={mouseDown}
	onfocusin={focusIn}
	onfocusout={focusOut}
>
	{@render children?.()}
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
