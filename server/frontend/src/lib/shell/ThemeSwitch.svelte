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
	import { currentTheme, Theme } from '$lib/theme.ts';
</script>

<div class='content'>
	<div class='highlight'
			 class:left={$currentTheme === Theme.Light}
			 class:middle={$currentTheme === Theme.System}
			 class:right={$currentTheme === Theme.Dark} />
	<button class='button' on:click={() => $currentTheme = Theme.Light} title='Light Mode'>
		<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 512 512' class='svg'
				 class:active={$currentTheme === Theme.Light}>
			<path
				d='M256 0c-13.3 0-24 10.7-24 24v64c0 13.3 10.7 24 24 24s24-10.7 24-24V24c0-13.3-10.7-24-24-24zm0 400c-13.3 0-24 10.7-24 24v64c0 13.3 10.7 24 24 24s24-10.7 24-24v-64c0-13.3-10.7-24-24-24zm232-120c13.3 0 24-10.7 24-24s-10.7-24-24-24h-64c-13.3 0-24 10.7-24 24s10.7 24 24 24h64zm-376-24c0-13.3-10.7-24-24-24H24c-13.3 0-24 10.7-24 24s10.7 24 24 24h64c13.3 0 24-10.7 24-24zm325-147.1c9.4-9.4 9.4-24.6 0-33.9s-24.6-9.4-33.9 0l-45.3 45.3c-9.4 9.4-9.4 24.6 0 33.9s24.6 9.4 33.9 0l45.3-45.3zM154.2 357.8c-9.4-9.4-24.6-9.4-33.9 0L75 403.1c-9.4 9.4-9.4 24.6 0 33.9s24.6 9.4 33.9 0l45.3-45.3c9.4-9.4 9.4-24.6 0-33.9zM403.1 437c9.4 9.4 24.6 9.4 33.9 0s9.4-24.6 0-33.9l-45.3-45.3c-9.4-9.4-24.6-9.4-33.9 0s-9.4 24.6 0 33.9l45.3 45.3zM154.2 154.2c9.4-9.4 9.4-24.6 0-33.9L108.9 75c-9.4-9.4-24.6-9.4-33.9 0s-9.4 24.6 0 33.9l45.3 45.3c9.4 9.4 24.6 9.4 33.9 0zM256 368a112 112 0 1 0 0-224 112 112 0 1 0 0 224z' />
		</svg>
	</button>
	<button class='button' on:click={() => $currentTheme = Theme.System} title='System Mode'>
		<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 384 512' class='svg'
				 class:active={$currentTheme === Theme.System}>
			<path
				d='M0 256 28.5 28c2-16 15.6-28 31.8-28h168.6c15 0 27.1 12.1 27.1 27.1 0 3.2-.6 6.5-1.7 9.5L208 160h139.3c20.2 0 36.7 16.4 36.7 36.7 0 7.4-2.2 14.6-6.4 20.7l-192.2 281c-5.9 8.6-15.6 13.7-25.9 13.7h-2.9c-15.7 0-28.5-12.8-28.5-28.5 0-2.3.3-4.6.9-6.9L176 288H32c-17.7 0-32-14.3-32-32z' />
		</svg>
	</button>
	<button class='button' on:click={() => $currentTheme = Theme.Dark} title='Dark Mode'>
		<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 384 512' class='svg'
				 class:active={$currentTheme === Theme.Dark}>
			<path
				d='M223.5 32C100 32 0 132.3 0 256s100 224 223.5 224c60.6 0 115.5-24.2 155.8-63.4 5-4.9 6.3-12.5 3.1-18.7s-10.1-9.7-17-8.5c-9.8 1.7-19.8 2.6-30.1 2.6-96.9 0-175.5-78.8-175.5-176 0-65.8 36-123.1 89.3-153.3 6.1-3.5 9.2-10.5 7.7-17.3s-7.3-11.9-14.3-12.5c-6.3-.5-12.6-.8-19-.8z' />
		</svg>
	</button>
</div>


<style>
	.content {
		position: relative;
		overflow: hidden;
		display: inline-flex;
		flex-direction: row;
		--button-width: 32px;
		--button-margin: 2px;
		--button-icon-size: 1em;
		--button-height: calc(var(--button-icon-size) + var(--button-margin) * 2);
		background: var(--color-surface-variant);
		border: 1px solid var(--color-outline);
		border-radius: var(--button-height);
	}

	.button {
		border-radius: var(--button-height);
		height: var(--button-height);
		display: flex;
		align-items: center;
		justify-content: center;
		width: var(--button-width);
		background: none;
		border: none;
		margin: var(--button-margin);
		color: var(--color-on-surface-variant);
		cursor: pointer;
		z-index: 1;
	}

	.svg {
		fill: var(--color-on-surface-variant);
		width: var(--button-icon-size);
		height: var(--button-icon-size);
	}

	.highlight {
		position: absolute;
		top: 0;
		left: 0;
		width: var(--button-width);
		height: var(--button-height);
		border-radius: var(--button-height);
		margin: var(--button-margin);
		background-color: var(--color-tertiary);
		transition: left 200ms ease-out;
	}

	.left {
		left: 0;
	}

	.middle {
		left: calc(var(--button-width) + var(--button-margin) * 2);
	}

	.right {
		left: calc(var(--button-width) * 2 + var(--button-margin) * 4);
	}

	.active {
		fill: var(--color-on-tertiary);
	}

	.button:focus {
		outline: none;
	}
</style>
