<script lang="ts">
	export let tip = '';
</script>

<div class="header" title={tip}>
	<slot />
</div>

<style>
	.header {
		background-color: var(--color-header);
		color: var(--color-on-header);
		fill: var(--color-on-header);
		text-align: center;
		padding: var(--padding-standard);
		font: var(--font-page-primary-labels);
	}
</style>
