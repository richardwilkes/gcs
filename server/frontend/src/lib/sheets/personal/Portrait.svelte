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
	import Header from '$lib/sheets/widget/Header.svelte';
	import SilhouetteSVG from '$svg/Silhouette.svg?raw';
	import { sheet, updateSheetField } from '$lib/sheet.ts';
	import { sheetPath } from '$lib/url.ts';

	export let imageURL: string | undefined;

	let clientHeight: number;
	let pictureHeight: number;
	let naturalWidth: number;
	let naturalHeight: number;
	let width: string;
	let height: string;
	let inputElement: HTMLInputElement | undefined;
	let inDrag = false;

	function onInputElementClick(event: MouseEvent) {
		event.stopPropagation();
	}

	function onClickCallback(event: MouseEvent) {
		if (event.button === 0 && event.detail === 2) {
			// Double-click
			openFileDialog();
		}
	}

	function onDragEnterCallback(event: DragEvent) {
		if (event.dataTransfer && event.dataTransfer.items) {
			let count = 0;
			for (const item of event.dataTransfer.items) {
				if (item.kind === 'file' && item.type.startsWith('image/')) {
					count++;
					if (count > 1) {
						return;
					}
				}
			}
			if (count === 1) {
				event.preventDefault();
				event.dataTransfer.dropEffect = 'copy';
				inDrag = true;
			}
		}
	}

	function ignoreDragEnterCallback(_: DragEvent) {}

	function onDragOverCallback(event: DragEvent) {
		if (inDrag && event.dataTransfer) {
			event.preventDefault();
			event.dataTransfer.dropEffect = 'copy';
		}
	}

	function onDragLeaveCallback(event: DragEvent) {
		if (inDrag) {
			inDrag = false;
			event.preventDefault();
		}
	}

	function onDropCallback(event: DragEvent) {
		if (inDrag) {
			inDrag = false;
			if (event.dataTransfer && event.dataTransfer.items) {
				for (const item of event.dataTransfer.items) {
					if (item.kind === 'file' && item.type.startsWith('image/')) {
						const file = item.getAsFile();
						if (file) {
							event.preventDefault();
							updatePortrait(file);
							return;
						}
					}
				}
			}
		}
	}

	function onChangeCallback(_: Event) {
		if (inputElement && inputElement.files) {
			updatePortrait(inputElement.files[0]);
			inputElement.files = null;
		}
	}

	function openFileDialog() {
		if (inputElement) {
			inputElement.click();
		}
	}

	function updatePortrait(file: File) {
		if ($sheetPath) {
			const reader = new FileReader();
			reader.onload = () => {
				let s = reader.result as string;
				const i = s.indexOf(',');
				if (i > 0) {
					s = s.substring(i + 1);
				}
				updateSheetField($sheetPath, 'field.binary', 'Portrait', s).then((updatedSheet) => {
					sheet.update((_) => updatedSheet);
				});
			};
			reader.readAsDataURL(file);
		}
	}

	$: {
		if (clientHeight) {
			if (!pictureHeight) {
				// The picture height is only set the first time the clientHeight is set to avoid it growing to fit the image
				// height rather than the outer container height.
				pictureHeight = clientHeight;
			}
			if (imageURL) {
				if (naturalWidth && naturalHeight) {
					if (naturalWidth > naturalHeight) {
						width = '100%';
						height = naturalHeight * (pictureHeight / naturalWidth) + 'px';
					} else {
						width = naturalWidth * (pictureHeight / naturalHeight) + 'px';
						height = '100%';
					}
				}
			} else {
				width = pictureHeight + 'px';
				height = pictureHeight + 'px';
			}
		}
	}
</script>

<div class="block">
	<Header>Portrait</Header>
	<!-- svelte-ignore a11y-click-events-have-key-events -->
	<div
		class="portrait"
		bind:clientHeight
		style="width:{pictureHeight}px;"
		tabindex="0"
		role="button"
		on:click={onClickCallback}
		on:dragenter={onDragEnterCallback}
		on:dragover={onDragOverCallback}
		on:dragleave={onDragLeaveCallback}
		on:drop={onDropCallback}>
		<input
			bind:this={inputElement}
			type="file"
			autocomplete="off"
			tabindex="-1"
			style="display:none"
			accept="image/*"
			on:change={onChangeCallback}
			on:click={onInputElementClick} />
		{#if clientHeight}
			{#if imageURL}
				<img
					class="noDrag"
					src={imageURL}
					bind:naturalWidth
					bind:naturalHeight
					{width}
					{height}
					alt="Portrait" />
			{:else}
				<div style="width: {width}; height: {height}; pointer-events: none;">{@html SilhouetteSVG}</div>
			{/if}
		{/if}
		<!-- svelte-ignore a11y-no-static-element-interactions -->
		<div class="tip-container" class:inDrag on:dragenter={ignoreDragEnterCallback}>
			<div class="tip noDrag" class:hide={inDrag}>Drop an image here or double-click to change the portrait</div>
		</div>
	</div>
</div>

<style>
	.block {
		border: var(--standard-border);
		display: flex;
		flex-direction: column;
	}

	.portrait {
		background-color: var(--color-below-surface);
		display: flex;
		justify-content: center;
		align-items: center;
		overflow: clip;
		height: 100%;
		outline: none;
	}

	.tip-container {
		position: absolute;
		top: 0;
		left: 0;
		bottom: 0;
		right: 0;
		display: flex;
		justify-items: center;
		align-items: center;
		opacity: 0;
	}

	.tip-container:hover {
		background-color: rgba(0 0 0 / 30%);
		opacity: 100%;
	}

	.tip {
		color: white;
		font-weight: bold;
		padding: 1em;
		text-align: center;
		user-select: none;
	}

	.inDrag {
		border: 2px dashed var(--color-focus);
		opacity: 100%;
	}

	.noDrag {
		pointer-events: none;
	}

	.hide {
		display: none;
	}
</style>
