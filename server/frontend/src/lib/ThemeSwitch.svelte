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
	import { currentTheme, Theme } from '$lib/theme.ts';
	import SunSVG from '$svg/Sun.svg?raw';
	import LightningSVG from '$svg/Lightning.svg?raw';
	import MoonSVG from '$svg/Moon.svg?raw';
</script>

<div class="content">
	<div
		class="highlight"
		class:left={$currentTheme === Theme.Light}
		class:middle={$currentTheme === Theme.System}
		class:right={$currentTheme === Theme.Dark} />
	<button
		class="button"
		class:active={$currentTheme === Theme.Light}
		on:click={() => ($currentTheme = Theme.Light)}
		title="Light Mode">
		{@html SunSVG}
	</button>
	<button
		class="button"
		class:active={$currentTheme === Theme.System}
		on:click={() => ($currentTheme = Theme.System)}
		title="System Mode">
		{@html LightningSVG}
	</button>
	<button
		class="button"
		class:active={$currentTheme === Theme.Dark}
		on:click={() => ($currentTheme = Theme.Dark)}
		title="Dark Mode">
		{@html MoonSVG}
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
		background-color: var(--color-above-surface);
		border: 1px solid var(--color-surface-edge);
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
		color: var(--color-on-below-surface);
		cursor: pointer;
		z-index: 1;
	}

	.button > :global(svg) {
		color: var(--color-on-above-surface);
		width: var(--button-icon-size);
		height: var(--button-icon-size);
	}

	.active > :global(svg) {
		color: var(--color-on-focus);
	}

	.highlight {
		position: absolute;
		top: 0;
		left: 0;
		width: var(--button-width);
		height: var(--button-height);
		border-radius: var(--button-height);
		margin: var(--button-margin);
		background-color: var(--color-focus);
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

	.button:focus {
		outline: none;
	}
</style>
