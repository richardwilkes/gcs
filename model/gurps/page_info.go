// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"strconv"
)

// PageInfo represents a page label, which may be numeric or non-numeric, and may have an offset.
type PageInfo struct {
	Label  string
	Offset int
}

// ToExternalForm returns the PageInfo in a form suitable for external use (e.g. in a command line). Note that only
// numeric page labels are adjusted by the offset, since we have no way to resolve a non-numeric page label for an
// external viewer.
func (p PageInfo) ToExternalForm() string {
	if pageNum, err := strconv.Atoi(p.Label); err == nil {
		return strconv.Itoa(max(pageNum+p.Offset, 1))
	}
	return p.Label
}
