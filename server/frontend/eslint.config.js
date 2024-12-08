import prettier from 'eslint-config-prettier';
import js from '@eslint/js';
import { includeIgnoreFile } from '@eslint/compat';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import { fileURLToPath } from 'node:url';
import ts from 'typescript-eslint';
const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

export default ts.config(
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs['flat/recommended'],
	prettier,
	...svelte.configs['flat/prettier'],
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},
	{
		files: ['**/*.svelte'],

		languageOptions: {
			parserOptions: {
				parser: ts.parser
			}
		}
	},
	{
		rules: {
			"no-mixed-spaces-and-tabs": "off",
			"svelte/no-at-html-tags": "off",
			"@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
			"@typescript-eslint/no-explicit-any": "off"
		}
	},
	{
		ignores: [
			"package-lock.json"
		]
	}
);

// /** @type { import("eslint").Linter.Config } */
// module.exports = {
// 	root: true,
// 	extends: [
// 		'eslint:recommended',
// 		'plugin:@typescript-eslint/recommended',
// 		'plugin:svelte/recommended'
// 	],
// 	rules: {
// 		"no-mixed-spaces-and-tabs": "off",
// 		"@typescript-eslint/no-unused-vars": ["error", { "argsIgnorePattern": "^_" }],
// 		"svelte/no-at-html-tags": "off"
// 	},
// 	parser: '@typescript-eslint/parser',
// 	plugins: ['@typescript-eslint'],
// 	parserOptions: {
// 		sourceType: 'module',
// 		ecmaVersion: 2020,
// 		extraFileExtensions: ['.svelte']
// 	},
// 	env: {
// 		browser: true,
// 		es2017: true,
// 		node: true
// 	},
// 	overrides: [
// 		{
// 			files: ['*.svelte'],
// 			parser: 'svelte-eslint-parser',
// 			parserOptions: {
// 				parser: '@typescript-eslint/parser'
// 			}
// 		}
// 	]
// };
