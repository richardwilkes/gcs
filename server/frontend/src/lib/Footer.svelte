<!--
  - Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
  -
  - This Source Code Form is subject to the terms of the Mozilla Public
  - License, version 2.0. If a copy of the MPL was not distributed with
  - this file, You can obtain one at http://mozilla.org/MPL/2.0/.
  -
  - This Source Code Form is "Incompatible With Secondary Licenses", as
  - defined by the Mozilla Public License, version 2.0.
  -->

<script lang='ts'>
	import { apiPrefix } from '$lib/dev.ts';

	type VersionResponse = {
		Name: string;
		Copyright: string;
		Version: string;
		Build: string;
		Git: string;
		Modified: boolean;
	};

	let version: VersionResponse | undefined;

	fetch(apiPrefix('/version'), {
		method: 'GET',
		cache: 'no-store'
	}).then(rsp => {
		if (!rsp.ok) {
			console.log(rsp.status + ' ' + rsp.statusText);
		} else {
			rsp.json().then(json => {
				version = json as VersionResponse;
			}).catch(error => console.log(error));
		}
	}).catch(error => {
		console.log(error);
	});
</script>

<div class='footer'>
	<div>
		{#if version}
			{version.Name}
			{#if version.Modified || version.Version === '0.0'}
				Development Version
			{:else}
				Version {version.Version}
			{/if}
		{:else}
			&nbsp;
		{/if}
	</div>
	<div class='secondary'>
		{#if version}
			{version.Copyright}
			{#if !version.Modified && version.Version !== '0.0'}
				Build {version.Build}. Git {version.Git.slice(0, 7)}.
			{/if}
		{:else}
			&nbsp;
		{/if}
	</div>
</div>

<style>
	.footer {
		background-color: var(--color-surface);
		color: var(--color-on-surface);
		border-top: 1px solid var(--color-outline-variant);
		padding: 4px 8px;
		text-align: center;
		font-weight: bold;
		font-size: 75%;
	}

	.secondary {
		font-size: 85%;
		color: var(--color-outline);
	}
</style>