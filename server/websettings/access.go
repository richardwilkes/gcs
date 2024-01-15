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

package websettings

// Access holds the configuration for a user's access to a directory (and all of its sub-paths) on the server.
type Access struct {
	Dir      string `json:"dir"`
	ReadOnly bool   `json:"read_only"`
}
