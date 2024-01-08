<script lang="ts">
	import Header from '$lib/sheets/widget/Header.svelte';
	import Silhouette from '$lib/svg/Silhouette.svelte';

	export let imageURL: string | undefined;

	let naturalWidth: number;
	let naturalHeight: number;
	let width: string;
	let height: string;

	$: {
		if (naturalWidth > naturalHeight) {
			width = '150px';
			height = naturalHeight * (150 / naturalWidth) + 'px';
		} else {
			width = naturalWidth * (150 / naturalHeight) + 'px';
			height = '150px';
		}
	}
</script>

<div class="block">
	<Header>Portrait</Header>
	<div class="portrait">
		{#if imageURL}
			<img src={imageURL} bind:naturalWidth bind:naturalHeight {width} {height} alt="Portrait" />
		{:else}
			<Silhouette />
		{/if}
	</div>
</div>

<style>
	.block {
		border: var(--standard-border);
		display: flex;
		flex-direction: column;
		justify-content: flex-start;
	}

	.portrait {
		background-position: center;
		background-repeat: no-repeat;
		background-size: contain;
		background-color: var(--color-background);
		width: 150px;
		height: 150px;
		display: flex;
		justify-content: center;
		align-items: center;
		overflow: clip;
	}
</style>
