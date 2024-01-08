<script lang="ts">
	import Portrait from '$lib/sheets/personal/Portrait.svelte';
	import Identity from '$lib/sheets/personal/Identity.svelte';
	import Misc from '$lib/sheets/personal/Misc.svelte';
	import Points from '$lib/sheets/personal/Points.svelte';
	import Description from '$lib/sheets/personal/Description.svelte';
	import { pc } from '$lib/entity.ts';

	$: imageURL = $pc.profile?.portrait ? 'data:;base64,' + $pc.profile?.portrait : undefined;
</script>

<div class="content">
	<Portrait {imageURL} />
	<div class="middle">
		<div class="top">
			<Identity />
			<Misc />
		</div>
		<Description />
	</div>
	<Points />
</div>

<style>
	.content {
		display: flex;
		gap: var(--section-gap);
		align-items: stretch;
	}

	.top {
		display: flex;
		gap: var(--section-gap);
	}

	.middle {
		flex-grow: 1;
		gap: var(--section-gap);
		display: flex;
		flex-direction: column;
		justify-content: stretch;
	}
</style>
