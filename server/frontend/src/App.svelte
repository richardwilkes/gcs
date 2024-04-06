<script lang="ts">
	import { checkSession } from '$lib/session.ts';
	import { saveSheet, sheet } from '$lib/sheet.ts';
	import Toolbar from '$lib/Toolbar.svelte';
	import Login from '$page/Login.svelte';
	import LoadSheet from '$page/LoadSheet.svelte';
	import Sheet from '$page/Sheet.svelte';
	import Footer from '$lib/Footer.svelte';
	import SheetFile from '$lib/svg/SheetFile.svelte';
	import { url, sheetPath } from '$lib/url.ts';
	import { navTo } from '$lib/nav.ts';

	function open() {
		navTo('#');
	}

	async function save() {
		if ($sheetPath && $sheet && $sheet.Modified) {
			const updatedSheet = await saveSheet($sheetPath);
			sheet.update((_) => updatedSheet);
		}
	}

	checkSession();
</script>

<svelte:head>
	<title>{$sheet?.Identity.Name ?? 'GURPS Character Sheet'}</title>
</svelte:head>

<div class="shell">
	<Toolbar>
		{#if $sheetPath}
			{#if $sheet && !$sheet.ReadOnly}
				<button class="save" disabled={!$sheet.Modified} on:click={save}>Save</button>
			{/if}
			<button class="open" title="Open…" on:click={open}>
				<SheetFile style="width: 1.2em; height: 1.2em; fill: var(--color-on-surface);" />
				{$sheetPath}
				{#if $sheet && $sheet.ReadOnly}
					<span class="ro">(read only)</span>
				{/if}
			</button>
		{/if}
	</Toolbar>
	<div class="content">
		{#if $url.hash === '#login'}
			<Login />
		{:else if $sheetPath}
			<Sheet path={$sheetPath} />
		{:else}
			<LoadSheet />
		{/if}
	</div>
	<Footer />
</div>

<style>
	.shell {
		position: absolute;
		top: 0;
		left: 0;
		bottom: 0;
		right: 0;
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		gap: 0;
		background-color: var(--color-background);
	}

	.content {
		display: flex;
		flex-grow: 1;
		overflow: auto;
	}

	.open {
		border: none;
		background: none;
		display: flex;
		align-items: center;
		gap: 2px;
		cursor: pointer;
		color: var(--color-on-surface-variant);
		font-weight: normal;
	}

	.save {
		padding: var(--padding-standard);
	}

	.ro {
		padding-left: 1em;
		font-size: 0.7em;
	}
</style>
