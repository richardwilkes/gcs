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

import { get, writable } from 'svelte/store';
import { apiPrefix } from '$lib/dev.ts';
import { page } from '$lib/page.ts';

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
};

export function updateSessionFromResponse(rsp: Response) {
	const id = rsp.headers.get('X-Session');
	const user = rsp.headers.get('X-User');
	if (id && user) {
		session.set({ ID: id, User: user });
	} else {
		session.set(null);
	}
}

export async function checkSession() {
	const sess = get(session);
	if (sess) {
		const rsp = await fetch(apiPrefix('/session'), {
			method: 'GET',
			headers: { 'X-Session': sess.ID },
			cache: 'no-store'
		});
		if (rsp.ok) {
			updateSessionFromResponse(rsp);
			if (get(session)) {
				return;
			}
		}
	}
	session.set(null);
	page.set({ID:'login', NextID: get(page).NextID});
}