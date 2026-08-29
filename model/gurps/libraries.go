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
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

const (
	masterGitHubAccountName = "richardwilkes"
	masterRepoName          = "gcs_master_library"
	userRepoName            = "gcs_user_library"
)

// Libraries holds a Library set. Once the app is running, the global set held in Settings.Libraries is read from
// background goroutines: the navigator's deep search parses files off the UI thread, and parsing a character sheet runs
// Entity.Recalculate, whose SrcMatcher.PrepareHashes looks each source library up in the set. The lock guards the map
// against those readers observing the UI thread adding or removing a library, which would be a fatal concurrent map
// read and map write, so all access must go through the accessor methods. Individual Library objects carry their own
// lock; this one covers only the map itself. A nil *Libraries behaves like a nil map: the read accessors see an empty
// set, while the mutating ones panic.
type Libraries struct {
	m          map[string]*Library
	lock       sync.RWMutex
	createLock sync.Mutex // Serializes the create path of getOrCreate; readers never take it.
}

// NewLibraries creates a new Libraries object, populated with the master and user libraries.
func NewLibraries() *Libraries {
	libs := &Libraries{m: make(map[string]*Library)}
	libs.Master()
	libs.User()
	return libs
}

// Store adds or replaces the library under the given key.
func (l *Libraries) Store(key string, lib *Library) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.m[key] = lib
}

// Remove removes the library with the given key.
func (l *Libraries) Remove(key string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.m, key)
}

// Rekey moves the library from oldKey to its current Key() as one operation, for use after a Configure() that changed
// the key. Doing it as one operation rather than a Remove() followed by a Store() keeps the library from ever being
// absent from the set, so a background reader (see Libraries) never sees a set that is missing a library the user
// merely reconfigured. Nothing today would cache a bad result from such a window -- SrcMatcher.PrepareHashes skips a
// library it cannot find, and the hashes it builds feed only the UI's source-state columns, not the deep search text --
// but the single operation costs nothing and keeps the window from becoming a problem later. Any library already under
// the new key is replaced, so the caller must have checked for a collision beforehand.
func (l *Libraries) Rekey(oldKey string, lib *Library) {
	newKey := lib.Key()
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.m, oldKey)
	l.m[newKey] = lib
}

// Lookup returns the library with the given key, or nil if it isn't present.
func (l *Libraries) Lookup(key string) *Library {
	if l == nil {
		return nil
	}
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.m[key]
}

// Len returns the number of libraries in the set.
func (l *Libraries) Len() int {
	if l == nil {
		return 0
	}
	l.lock.RLock()
	defer l.lock.RUnlock()
	return len(l.m)
}

// MarshalJSONTo implements json.MarshalerTo.
func (l *Libraries) MarshalJSONTo(enc *jsontext.Encoder) error {
	if l == nil {
		return enc.WriteToken(jsontext.Null)
	}
	l.lock.RLock()
	defer l.lock.RUnlock()
	return json.MarshalEncode(enc, l.m)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (l *Libraries) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var loaded map[string]*Library
	if err := json.UnmarshalDecode(dec, &loaded); err != nil {
		return err
	}
	libs := make(map[string]*Library)
	for k, lib := range loaded {
		if lib.Valid() {
			if strings.HasPrefix(k, "*/") { // GCS v5.4 and earlier use * for local dirs that weren't on github
				k = k[1:]
			}
			// Applying the key is also what fills in the version on disk, which isn't part of what was saved; see
			// ConfigureForKey().
			if err := lib.ConfigureForKey(k); err != nil {
				errs.Log(err, "key", k)
				continue
			}
			libs[lib.Key()] = lib
		}
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.m = libs
	return nil
}

// Master holds information about the master library, creating it if not already present. It is safe on the live
// Settings.Libraries; see getOrCreate.
func (l *Libraries) Master() *Library {
	return l.getOrCreate(masterGitHubAccountName+"/"+masterRepoName, func() *Library {
		return NewLibrary(i18n.Text("Master Library"), masterGitHubAccountName, "", masterRepoName,
			DefaultMasterLibraryPath())
	})
}

// User holds information about the user library, creating it if not already present. It is safe on the live
// Settings.Libraries; see getOrCreate.
func (l *Libraries) User() *Library {
	return l.getOrCreate("/"+userRepoName, func() *Library {
		return NewLibrary(i18n.Text("User Library"), "", "", userRepoName, DefaultUserLibraryPath())
	})
}

// getOrCreate returns the library under key, building one with create and storing it if none is present. The common
// case is a lookup under the read lock, so it never holds up the background readers of the set for a library that
// exists. A miss builds under createLock rather than the write lock -- NewLibrary reads the version file on disk, and a
// file read under the write lock would stall every reader for its duration -- so readers proceed throughout, while any
// other caller that misses at the same time waits for the build in flight and then finds its result on the re-check
// rather than building a second library that would only be discarded.
func (l *Libraries) getOrCreate(key string, create func() *Library) *Library {
	if lib := l.Lookup(key); lib != nil {
		return lib
	}
	l.createLock.Lock()
	defer l.createLock.Unlock()
	if lib := l.Lookup(key); lib != nil {
		return lib // Built by the caller that held createLock before us.
	}
	lib := create()
	l.lock.Lock()
	l.m[key] = lib
	l.lock.Unlock()
	return lib
}

// List returns an ordered list of Library objects. The list is a snapshot: it is captured under the lock, so it is
// unaffected by later mutations of the set.
func (l *Libraries) List() []*Library {
	if l == nil {
		return nil
	}
	l.lock.RLock()
	libs := make([]*Library, 0, len(l.m))
	for _, lib := range l.m {
		libs = append(libs, lib)
	}
	l.lock.RUnlock()
	slices.SortFunc(libs, func(a, b *Library) int { return a.Compare(b) })
	return libs
}

// PerformUpdateChecks checks each of the libraries for updates in the background. The set is captured on the calling
// goroutine, since the map may be modified while the checks run.
func (l *Libraries) PerformUpdateChecks() {
	libs := l.List()
	go func() {
		client := &http.Client{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()
		var wg sync.WaitGroup
		for _, lib := range libs {
			wg.Go(func() { lib.CheckForAvailableUpgrade(ctx, client) })
		}
		wg.Wait()
	}()
}
