<script lang='ts'>
	import { page, previousPage } from '$lib/page.ts';
	import { checkSession } from '$lib/session.ts';
	import { sheet } from '$lib/sheet.ts';
	import Toolbar from '$lib/Toolbar.svelte';
	import Login from '$page/Login.svelte';
	import LoadSheet from '$page/LoadSheet.svelte';
	import Sheet from '$page/Sheet.svelte';
	import Footer from '$lib/Footer.svelte';
	import SheetFile from '$lib/svg/SheetFile.svelte';

	function open() {
		$previousPage = $page;
		$page = { ID: 'home', NextID: 'home' };
	}

	checkSession();
</script>

<svelte:head>
	<title>{$sheet?.Identity.Name ?? 'GURPS Character Sheet'}</title>
</svelte:head>

<div class='shell'>
	<Toolbar>
		{#if $page.ID === 'sheet'}
			<button class='open' title='Open…' on:click={open}>
				<SheetFile style='width: 1.2em; height: 1.2em; fill: var(--color-on-surface);' />
				{$page.Sheet}
			</button>
		{/if}
	</Toolbar>
	<div class='content'>
		{#if $page.ID === 'login'}
			<Login />
		{:else if $page.ID === 'sheet'}
			<Sheet />
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
	}
</style>
