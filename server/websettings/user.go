// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package websettings

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// User holds information for a web user.
type User struct {
	Name           string            `json:"name"`
	HashedPassword string            `json:"hash"`
	AccessList     map[string]Access `json:"access_list"` // Key is the name the user sees and uses for the directory.
}

// Clone creates a copy of this user.
func (u *User) Clone() *User {
	accessList := make(map[string]Access, len(u.AccessList))
	for k, v := range u.AccessList {
		accessList[k] = v
	}
	return &User{
		Name:           u.Name,
		HashedPassword: u.HashedPassword,
		AccessList:     accessList,
	}
}

// Key returns the key for this user.
func (u *User) Key() string {
	return UserNameToKey(u.Name)
}

// AccessListWithKeys returns the access list with keys.
func (u *User) AccessListWithKeys() []*AccessWithKey {
	list := make([]*AccessWithKey, 0, len(u.AccessList))
	for k, v := range u.AccessList {
		list = append(list, &AccessWithKey{Key: k, Access: v})
	}
	slices.SortStableFunc(list, func(a, b *AccessWithKey) int {
		return txt.NaturalCmp(a.Key, b.Key, true)
	})
	return list
}

func (u *User) String() string {
	if len(u.AccessList) == 1 {
		return fmt.Sprintf(i18n.Text("%s [1 access point]"), u.Name)
	}
	return fmt.Sprintf(i18n.Text("%s [%d access points]"), u.Name, len(u.AccessList))
}

// UserNameToKey converts a user name to a key.
func UserNameToKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// HashPassword hashes passwords.
func HashPassword(in string) string {
	h := sha256.Sum256([]byte(in + "@gcs"))
	return base64.RawStdEncoding.EncodeToString(h[:])
}
