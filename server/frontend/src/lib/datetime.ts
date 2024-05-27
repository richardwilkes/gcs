// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

const stdDateTimeFmt = new Intl.DateTimeFormat('default', {
	month: 'short',
	day: 'numeric',
	year: 'numeric',
	hour: 'numeric',
	minute: '2-digit',
	hour12: true
});

export function formatDateStamp(date: Date | string | undefined) {
	switch (typeof date) {
		case 'string':
			date = new Date(date);
			break;
		case 'undefined':
			date = new Date();
			break;
	}
	return stdDateTimeFmt.format(date);
}
