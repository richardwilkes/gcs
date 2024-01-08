<script lang="ts">
	import { pc, loadPC } from '$lib/entity.ts';

	let files: FileList;

	$: {
		if (files && files.length) {
			loadPC(files[0])
				.then((newPC) => ($pc = newPC))
				.catch((err) => {
					console.error(err);
					alert(`Error loading ${files[0].name}`);
				});
		}
	}
</script>

<input bind:files type="file" accept=".gcs" />

<style>
	input {
		flex-grow: 1;
		font: var(--font-page-primary-fields);
	}
</style>
