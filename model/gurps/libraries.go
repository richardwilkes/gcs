// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
)

const (
	masterGitHubAccountName = "richardwilkes"
	masterRepoName          = "gcs_master_library"
	userRepoName            = "gcs_user_library"
)

// Libraries holds a Library set.
type Libraries map[string]*Library

// NewLibraries creates a new, empty, Libraries object.
func NewLibraries() Libraries {
	libs := Libraries(make(map[string]*Library))
	libs.Master()
	libs.User()
	return libs
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Libraries) UnmarshalJSON(data []byte) error {
	var loaded map[string]*Library
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}
	libs := make(map[string]*Library)
	for k, lib := range loaded {
		if lib.Valid() {
			if strings.HasPrefix(k, "*/") { // GCS v5.4 and earlier use * for local dirs that weren't on github
				k = k[1:]
			}
			lib.ConfigureForKey(k)
			lib.monitor = newMonitor(lib)
			libs[lib.Key()] = lib
		}
	}
	*l = libs
	return nil
}

// Master holds information about the master library.
func (l Libraries) Master() *Library {
	lib, ok := l[masterGitHubAccountName+"/"+masterRepoName]
	if !ok {
		lib = NewLibrary(i18n.Text("Master Library"), masterGitHubAccountName, "", masterRepoName, DefaultMasterLibraryPath())
		l[lib.Key()] = lib
	}
	return lib
}

// User holds information about the user library.
func (l Libraries) User() *Library {
	lib, ok := l["/"+userRepoName]
	if !ok {
		lib = NewLibrary(i18n.Text("User Library"), "", "", userRepoName, DefaultUserLibraryPath())
		l[lib.Key()] = lib
	}
	return lib
}

// List returns an ordered list of Library objects.
func (l Libraries) List() []*Library {
	libs := make([]*Library, 0, len(l))
	for _, lib := range l {
		libs = append(libs, lib)
	}
	slices.SortFunc(libs, func(a, b *Library) int { return a.Compare(b) })
	return libs
}

// PerformUpdateChecks checks each of the libraries for updates.
func (l Libraries) PerformUpdateChecks() {
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(len(l))
	for _, lib := range l {
		go func(one *Library) {
			defer wg.Done()
			one.CheckForAvailableUpgrade(ctx, client)
		}(lib)
	}
	wg.Wait()
}
