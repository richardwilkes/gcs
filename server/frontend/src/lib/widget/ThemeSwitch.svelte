<script lang="ts">
	import { currentTheme } from '$lib/theme.ts';
	import Label from '$lib/sheets/widget/Label.svelte';
</script>

<Label title="Theme" />
<select bind:value={$currentTheme.kind}>
	<option value="system">System</option>
	<option value="light">Light</option>
	<option value="dark">Dark</option>
</select>

<style>
	select {
		font: var(--font-page-primary-labels);
	}
</style>
