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

import {writable} from 'svelte/store';

type Page = {
	ID: string;
	NextID: string;
	Sheet?: string;
};

export const page = writable<Page>({ID: 'home', NextID: 'home'});
export const previousPage = writable<Page>({ID: 'home', NextID: 'home'});