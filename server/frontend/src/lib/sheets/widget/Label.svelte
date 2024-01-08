<script lang="ts">
	export let title: string | number = '';
	export let left = false;
	export let center = false;
	export let borderLeft = false;
	export let borderRight = false;
</script>

<div class="label" class:left class:center class:borderLeft class:borderRight>{title}</div>

<style>
	.label {
		font: var(--font-page-primary-labels);
		padding: var(--padding-standard);
		text-align: right;
		white-space: nowrap;
	}

	.left {
		text-align: left;
	}

	.center {
		text-align: center;
	}

	.borderLeft {
		border-left: 1px solid var(--color-header);
	}

	.borderRight {
		border-right: 1px solid var(--color-header);
	}
</style>
