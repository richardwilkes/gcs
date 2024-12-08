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
	import { ShowAs } from '$lib/show.ts';

	interface Props {
		showAs?: ShowAs;
		cancelButton?: string;
		cancelAutoFocus?: boolean;
		okButton?: string;
		okAutoFocus?: boolean;
		callback: (ok: boolean) => void;
		title?: import('svelte').Snippet;
		children?: import('svelte').Snippet;
		buttons?: import('svelte').Snippet<[any]>;
	}

	let {
		showAs = $bindable(ShowAs.None),
		cancelButton = 'Cancel',
		cancelAutoFocus = false,
		okButton = 'OK',
		okAutoFocus = true,
		callback,
		title,
		children,
		buttons
	}: Props = $props();
	export const close = (ok: boolean) => {
		showAs = ShowAs.None;
		dialog?.close();
		callback(ok);
	};

	let dialog: HTMLDialogElement | undefined = $state();

	$effect(() => {
		if (dialog) {
			switch (showAs) {
				case ShowAs.Modal:
					dialog.showModal();
					break;
				case ShowAs.Dialog:
					dialog.show();
					break;
				default:
					dialog.close();
					break;
			}
		}
	});
</script>

{#if showAs !== ShowAs.None}
	<dialog bind:this={dialog}>
		<div class="title">
			{@render title?.()}
		</div>
		<div class="content">
			{@render children?.()}
		</div>
		<div class="buttons">
			{@render buttons?.({ class: 'std-buttons' })}
			<div class="std-buttons">
				{#if cancelButton}
					<!-- svelte-ignore a11y_autofocus -->
					<button autofocus={cancelAutoFocus} onclick={() => close(false)}>{cancelButton}</button>
				{/if}
				{#if okButton}
					<!-- svelte-ignore a11y_autofocus -->
					<button autofocus={okAutoFocus} onclick={() => close(true)}>{okButton}</button>
				{/if}
			</div>
		</div>
	</dialog>
{/if}

<style>
	dialog {
		max-width: 80vw;
		max-height: 80vh;
		border-radius: 1em;
		padding: 1.5em;
		border: 1px solid var(--color-surface-edge);
		box-shadow: 0 0 64px 0 var(--color-shadow);
		background-color: var(--color-below-surface);
		overflow: clip;
		display: flex;
		flex-direction: column;
	}

	dialog::backdrop {
		background-color: rgba(0, 0, 0, 0.5);
	}

	dialog[open] {
		animation: zoom 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	@keyframes zoom {
		from {
			transform: scale(0.95);
		}
		to {
			transform: scale(1);
		}
	}

	dialog[open]::backdrop {
		animation: fade 0.2s ease-out;
	}

	@keyframes fade {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	.title {
		font-size: 125%;
		font-weight: bold;
		text-align: center;
	}

	.content {
		overflow: auto;
	}

	.buttons {
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 1.5em 1.5em 0;
	}

	.std-buttons {
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: 1fr;
		gap: 1em;
	}
</style>
