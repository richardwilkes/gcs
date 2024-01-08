<script lang="ts">
	export let title = '';
	export let span = 1;
</script>

<div class="header" style="grid-column-end: span {span}">{title}</div>

<style>
	.header {
		background-color: var(--color-header);
		color: var(--color-on-header);
		text-align: center;
		font: var(--font-page-primary-labels);
		padding: var(--padding-standard);
	}
</style>
