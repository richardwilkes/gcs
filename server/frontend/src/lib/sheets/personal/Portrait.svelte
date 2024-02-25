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
	import Header from '$lib/sheets/widget/Header.svelte';
	import Silhouette from '$lib/svg/Silhouette.svelte';

	export let imageURL: string | undefined;

	let clientHeight: number;
	let pictureHeight: number;
	let naturalWidth: number;
	let naturalHeight: number;
	let width: string;
	let height: string;

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
						width = "100%";
						height = naturalHeight * (pictureHeight / naturalWidth) + 'px';
					} else {
						width = naturalWidth * (pictureHeight / naturalHeight) + 'px';
						height = "100%";
					}
				}
			} else {
				width = pictureHeight + 'px';
				height = pictureHeight + 'px';
			}
		}
	}
</script>

<div class='block'>
	<Header>Portrait</Header>
	<div class='portrait' bind:clientHeight style='width:{pictureHeight}px;'>
		{#if clientHeight}
			{#if imageURL}
				<img src={imageURL} bind:naturalWidth bind:naturalHeight {width} {height} alt='Portrait' />
			{:else}
				<Silhouette style='width: {width}; height: {height};' />
			{/if}
		{/if}
	</div>
</div>

<style>
	.block {
		border: var(--standard-border);
		display: flex;
		flex-direction: column;
	}

	.portrait {
		background-color: var(--color-background);
		display: flex;
		justify-content: center;
		align-items: center;
		overflow: clip;
		height: 100%;
	}
</style>
