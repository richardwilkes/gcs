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

<script lang="ts"></script>

<div class='waiting'></div>

<style>
	.waiting {
		display: inline-block;
		width: 80px;
		height: 80px;
	}

	.waiting:after {
		content: " ";
		display: block;
		width: 64px;
		height: 64px;
		margin: 8px;
		border-radius: 50%;
		border: 6px solid var(--color-tertiary);
		border-color: var(--color-tertiary) transparent var(--color-tertiary) transparent;
		animation: waiting 1.2s linear infinite;
	}

	@keyframes waiting {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}
</style>