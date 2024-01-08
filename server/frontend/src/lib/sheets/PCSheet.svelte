<script lang="ts">
	import Personal from '$lib/sheets/personal/Personal.svelte';
	import Attributes from '$lib/sheets/attributes/Attributes.svelte';
	import Lists from '$lib/sheets/lists/Lists.svelte';
	import Toolbar from '$lib/sheets/widget/Toolbar.svelte';
	import { pc } from '$lib/entity.ts';
</script>

<div class="shell">
	<Toolbar />
	<div class="content">
		{#if $pc}
			<Personal />
			<Attributes />
			<Lists />
		{/if}
	</div>
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
	}

	.content {
		background-color: var(--color-page);
		color: var(--color-on-page);
		display: flex;
		flex-flow: column nowrap;
		justify-content: flex-start;
		align-items: stretch;
		align-content: stretch;
		gap: var(--section-gap);
		padding: 5px;
		overflow: auto;
		flex-grow: 1;
	}
</style>
