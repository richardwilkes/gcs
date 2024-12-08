import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { resolve } from 'path';

// https://vite.dev/config/
export default defineConfig({
	plugins: [svelte()],
	resolve: {
		alias: {
			$lib: resolve('./src/lib'),
			$svg: resolve('./src/svg'),
			$page: resolve('./src/page'),
		},
	},
});
