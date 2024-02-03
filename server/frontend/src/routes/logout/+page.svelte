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
	import { goto } from '$app/navigation';
	import { apiPrefix } from '$lib/dev.ts';
	import { session } from '$lib/session.ts';
	import { pc } from '$lib/entity.ts';

	$pc = undefined;

	fetch(apiPrefix('/logout'), {
		method: 'POST',
		headers: { 'X-Session': $session?.ID ?? '' },
		cache: 'no-store'
	})
		.then(() => {
			goto('/login');
		})
		.catch(error => {
			console.error(error);
			goto('/login');
		});
</script>