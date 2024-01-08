<script lang="ts">
	export let name = '';
	export let value = '';
	export let right = false;
</script>

<input class="field" class:right type="text" {name} {value} />

<style>
	.field {
		font: var(--font-page-primary-fields);
		padding: var(--padding-standard);
		border: none;
		border-bottom: 1px solid var(--color-control-edge);
		color: var(--color-on-editable);
		background-color: var(--color-editable);
		min-width: 20px;
		outline-style: none;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}

	.right {
		text-align: right;
	}
</style>
