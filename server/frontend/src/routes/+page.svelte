<script>
	import PCSheet from '$lib/sheets/PCSheet.svelte';
	import { pc } from '$lib/entity.ts';
</script>

<svelte:head>
	<title>{$pc?.profile?.name ?? 'GURPS Character Sheet'}</title>
</svelte:head>

<PCSheet />
