<script lang="ts">
	import Warning from '$lib/svg/Warning.svelte';

	export let reason = '';
</script>

<div class="content" title={reason}>
	<Warning />
	Unsatisfied prerequisite(s)!
</div>

<style>
	.content {
		font: var(--font-page-secondary-fields);
		padding: var(--padding-standard);
		white-space: nowrap;
		background-color: var(--color-error);
		color: var(--color-on-error);
		fill: var(--color-on-error);
		align-self: flex-start;
		border-radius: 4px;
		margin-left: 4px;
	}
</style>
