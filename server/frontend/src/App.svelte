<script lang='ts'>
	import { page } from '$lib/page.ts';
	import { checkSession } from '$lib/session.ts';
	import Login from '$page/Login.svelte';
	import Shell from '$lib/shell/Shell.svelte';
	import { sheet } from '$lib/sheet.ts';
	import LoadSheet from '$page/LoadSheet.svelte';
	import Sheet from '$page/Sheet.svelte';

	checkSession();
</script>

<svelte:head>
	<title>{$sheet?.Identity.Name ?? 'GURPS Character Sheet'}</title>
</svelte:head>

<Shell>
	<svelte:fragment slot='content'>
		{#if $page.ID === 'login'}
			<Login />
		{:else if $page.ID === 'sheet'}
			<Sheet />
		{:else}
			<LoadSheet />
		{/if}
	</svelte:fragment>
</Shell>
