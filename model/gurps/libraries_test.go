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
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPerformUpdateChecksSnapshotsTheSet verifies that the set of libraries to check is captured on the calling
// goroutine rather than being ranged over from the background one. The UI thread adds and removes libraries in place --
// the library settings dialog re-keys one when its repository changes, and the navigator removes one outright -- so a
// background goroutine ranging the live map is a "concurrent map iteration and map write", which is fatal to the
// process rather than merely being flagged by the race detector.
func TestPerformUpdateChecksSnapshotsTheSet(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	// An empty GitHub account name short-circuits LoadReleases, so the checks below never reach the network.
	libs := Libraries(make(map[string]*Library))
	for _, name := range []string{"alpha", "beta"} {
		lib := NewLibrary(name, "", "", name, t.TempDir())
		c.Equal("0", cachedVersion(lib), "nothing has been recorded for %s yet", name)
		// Written after the library was built, so its cached version starts out as "0" and becoming "7" is proof that
		// the background check actually ran against it.
		c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("7\n"), 0o600))
		libs[lib.Key()] = lib
	}

	libs.PerformUpdateChecks()

	// Modify the map while the checks are in flight, as the UI thread may.
	extra := NewLibrary("extra", "", "", "extra", t.TempDir())
	for range 1000 {
		libs[extra.Key()] = extra
		delete(libs, extra.Key())
	}

	// Every library in the set at the time of the call is checked, whatever happened to the map afterwards.
	for _, lib := range libs.List() {
		waitForCachedVersion(t, lib, "7")
	}
	c.Equal(2, len(libs), "the additions and removals canceled out")
}

// TestLibrariesUnmarshalRecordsVersionOnDisk verifies that libraries restored from a settings file know what version is
// on disk without an update check having run. Only an update check filled that in previously, so turning the periodic
// checks off would have left the Library Explorer showing no version at all.
func TestLibrariesUnmarshalRecordsVersionOnDisk(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, releaseFile), []byte("4\n1234\n"), 0o600))

	var libs Libraries
	p := strconv.Quote(dir)
	c.NoError(jio.Unmarshal([]byte(`{"someone/repo":{"title":"Test","path":`+p+`},`+
		`"/`+userRepoName+`":{"title":"User Library","path":`+p+`}}`), &libs))
	lib, ok := libs["someone/repo"]
	c.True(ok)
	current, releases := lib.AvailableReleases()
	c.Equal("4", current, "the version on disk is known without a check having run")
	c.Equal(0, len(releases), "no update check has run")

	// The User Library reports no version at all, since it isn't something that gets released.
	user, ok := libs["/"+userRepoName]
	c.True(ok)
	c.Equal("", cachedVersion(user))
}

// cachedVersion returns the version the library has recorded for what is on disk.
func cachedVersion(lib *Library) string {
	current, _ := lib.AvailableReleases()
	return current
}

// waitForCachedVersion blocks until the library has recorded the given version for what is on disk, failing the test if
// that doesn't happen within a generous deadline. The update checks run on a goroutine of the caller's own making, so
// there is nothing to join.
func waitForCachedVersion(t *testing.T, lib *Library, want string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for cachedVersion(lib) != want {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s to record version %s", lib.Data().Title, want)
		}
		time.Sleep(time.Millisecond)
	}
}
