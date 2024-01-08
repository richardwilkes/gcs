<script lang="ts">
	import PCLoaderButton from '$lib/sheets/widget/PCLoaderButton.svelte';
	import ThemeSwitch from '$lib/widget/ThemeSwitch.svelte';
</script>

<div class="toolbar">
	<PCLoaderButton />
	<ThemeSwitch />
</div>

<style>
	.toolbar {
		background-color: var(--color-background);
		border-bottom: 1px solid var(--color-divider);
		padding: 4px 8px;
		display: flex;
		justify-content: flex-start;
		align-items: baseline;
		align-content: center;
		column-gap: 8px;
		row-gap: 4px;
	}
</style>
