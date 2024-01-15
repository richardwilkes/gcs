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
	import '$lib/global.css';
	import { goto } from '$app/navigation';
	import { apiPrefix } from '$lib/dev.ts';
	import { session, updateSessionFromResponse } from '$lib/session.ts';

	if (!['/login', '/login/', '/login/index.html'].includes(window.location.pathname)) {
		if ($session) {
			fetch(apiPrefix('/session'), {
				method: 'GET',
				headers: {'X-Session': $session.ID},
				cache: 'no-store'
			})
				.then(rsp => {
					if (!rsp.ok) {
						session.set(null);
						goto('/login');
					} else {
						updateSessionFromResponse(rsp);
						if (!$session) {
							goto('/login');
						}
					}
				})
				.catch(error => {
					console.error(error);
					session.set(null);
					goto('/login');
				});
		} else {
			goto('/login');
		}
	}
</script>

<slot />
