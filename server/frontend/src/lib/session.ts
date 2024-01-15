/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

import { writable } from 'svelte/store';

const sessionKey = 'session';

type Session = {
	ID: string;
	User: string;
};

const currentValue = localStorage.getItem(sessionKey);

export const session = writable<Session | null>(currentValue ? JSON.parse(currentValue) as Session : null);

session.subscribe((value) => localStorage.setItem(sessionKey, JSON.stringify(value)));

window.onstorage = (event) => {
	if (event.key === sessionKey) {
		session.set(event.newValue ? JSON.parse(event.newValue) as Session : null);
	}
}

export function updateSessionFromResponse(rsp: Response) {
	const id = rsp.headers.get('X-Session')
	const user = rsp.headers.get('X-User')
	if (id && user) {
		session.set({ ID: id, User: user });
	} else {
		session.set(null);
	}
}
