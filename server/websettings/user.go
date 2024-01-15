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

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// User holds information for a web user.
type User struct {
	Name           string            `json:"name"`
	HashedPassword string            `json:"hash"`
	AccessList     map[string]Access `json:"access_list"` // Key is the name the user sees and uses for the directory.
}

func userNameToKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// HashPassword hashes passwords.
func HashPassword(in string) string {
	h := sha256.Sum256([]byte(in + "@gcs"))
	return base64.RawStdEncoding.EncodeToString(h[:])
}
