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

export interface Directory {
	name: string;
	path: string;
	files: string[];
	dirs: Directory[];
}

export function fillPathsForDir(dir: Directory, path: string) {
	dir.path = path ? path + '/' + dir.name : dir.name;
	for (const subDir of dir.dirs || []) {
		fillPathsForDir(subDir, dir.path);
	}
}