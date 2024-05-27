// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

export function navTo(target: string, queryKey?: string, queryValue?: string, replace: boolean = false) {
	const url = new URL(window.location.origin);
	url.hash = target;
	if (queryKey && queryValue !== undefined) {
		url.searchParams.set(queryKey, queryValue);
	}
	if (replace) {
		window.history.replaceState(null, '', url);
	} else {
		window.history.pushState(null, '', url);
	}
}
