// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package websettings

import "fmt"

// Access holds the configuration for a user's access to a directory (and all of its sub-paths) on the server.
type Access struct {
	Dir      string `json:"dir"`
	ReadOnly bool   `json:"read_only"`
}

// AccessWithKey holds the configuration for a user's access to a directory (and all of its sub-paths) on the server and
// also includes the key by which it is referenced.
type AccessWithKey struct {
	Key string `json:"key"`
	Access
}

func (a *AccessWithKey) String() string {
	ro := ""
	if a.ReadOnly {
		ro = "(read-only) "
	}
	return fmt.Sprintf("%s: %s%s", a.Key, ro, a.Dir)
}
