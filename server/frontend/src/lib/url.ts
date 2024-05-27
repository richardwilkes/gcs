// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

import { derived, writable } from 'svelte/store';

const href = writable(window.location.href);

const originalPushState = history.pushState;
const originalReplaceState = history.replaceState;

const updateHref = () => href.set(window.location.href);

history.pushState = function (data: unknown, unused: string, url?: string | URL | null) {
	originalPushState.apply(this, [data, unused, url]);
	updateHref();
};

history.replaceState = function (data: unknown, unused: string, url?: string | URL | null) {
	originalReplaceState.apply(this, [data, unused, url]);
	updateHref();
};

window.addEventListener('popstate', updateHref);
window.addEventListener('hashchange', updateHref);

export const url = derived(href, (value) => new URL(value));

export const sheetPath = derived(url, (value) => {
	if (value.hash.startsWith('#sheet/')) {
		return decodeURIComponent(value.hash.substring(7));
	}
	return '';
});
