<script lang="ts">
	export let value = 0;
</script>

<div class="field">[{value}]</div>

<style>
	.field {
		font: var(--font-page-secondary-fields);
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid rgba(0, 0, 0, 0);
		color: var(--color-on-editable);
		text-align: right;
	}
</style>
