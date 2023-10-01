/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

// Document holds the raw data for a document.
type Document struct {
	Name       string `json:"name"`
	Ext        string `json:"ext"`
	Content    []byte `json:"content"`
	Compressed bool   `json:"compressed,omitempty"`
}
