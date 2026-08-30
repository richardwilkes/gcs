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
	"bytes"
	"encoding/json/jsontext"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLibrariesNilReadsBehaveLikeAnEmptySet verifies that the read accessors are safe on a nil *Libraries, matching the
// nil-map semantics of the type this replaced. A settings file without a "libraries" key leaves Settings.Libraries nil
// until EnsureValidity replaces it, and code like ScanForNamedFileSets reads whatever set it is handed, so reads on a
// nil set must see an empty set rather than panicking.
func TestLibrariesNilReadsBehaveLikeAnEmptySet(t *testing.T) {
	c := check.New(t)
	var libs *Libraries
	c.Equal(0, libs.Len())
	c.Nil(libs.Lookup("someone/repo"))
	c.Equal(0, len(libs.List()))
	var buf bytes.Buffer
	c.NoError(libs.MarshalJSONTo(jsontext.NewEncoder(&buf)))
	c.Equal("null", strings.TrimSpace(buf.String()))
}

// TestLibrariesAccessIsSafeForConcurrentMutation verifies the contract the navigator's background content loaders rely
// on: every access to the library set goes through accessors that share the set's lock, so a loader reading the set
// while the UI thread re-keys or removes a library is neither a data race nor a fatal "concurrent map read and map
// write", and the list List returns is a snapshot that mutations made after it was taken do not affect. Run under the
// race detector, this test fails if any of those accesses stop going through the lock.
func TestLibrariesAccessIsSafeForConcurrentMutation(t *testing.T) {
	c := check.New(t)
	s := &Settings{Libraries: NewLibraries()}
	stable := NewLibrary("stable", "", "", "stable", t.TempDir())
	s.Libraries.Store(stable.Key(), stable)
	churn := NewLibrary("churn", "", "", "churn", t.TempDir())

	// Readers stand in for the deep search content loaders; the mutation loop below stands in for the UI thread
	// re-keying a library in the library settings dialog or removing one in the navigator.
	var missedStable atomic.Bool
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if s.Libraries.Lookup(stable.Key()) == nil {
					missedStable.Store(true)
				}
			}
		})
	}
	for range 1000 {
		s.Libraries.Store(churn.Key(), churn)
		s.Libraries.Remove(churn.Key())
	}
	close(stop)
	wg.Wait()
	c.False(missedStable.Load(), "every lookup must find the library that was never removed")

	// A snapshot is unaffected by mutations made after it was taken.
	snapshot := s.Libraries.List()
	s.Libraries.Store(churn.Key(), churn)
	c.False(slices.Contains(snapshot, churn), "a library stored after the snapshot was taken must not appear in it")
	c.NotNil(s.Libraries.Lookup(churn.Key()), "a fresh lookup must see the stored library")
}

// TestLibrariesRekeyNeverLeavesTheLibraryAbsent verifies that re-keying a library, as the library settings dialog does
// when its repository changes, is a single operation on the set: a Remove followed by a Store would leave the library
// absent in between, which a deep search content loader reading the set at that moment would observe (see
// Libraries.Rekey). Each reader checks the size of the set under a single lock acquisition, so any window in which the
// library is missing is observable rather than being masked by the reader's own timing.
func TestLibrariesRekeyNeverLeavesTheLibraryAbsent(t *testing.T) {
	c := check.New(t)
	libs := NewLibraries()
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "", "first", dir)
	libs.Store(lib.Key(), lib)
	size := libs.Len()
	configs := []LibraryConfig{
		{Title: "Test", GitHubAccountName: "someone", RepoName: "first"},
		{Title: "Test", GitHubAccountName: "someone", RepoName: "second"},
	}

	var sawAbsent atomic.Bool
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for {
				select {
				case <-stop:
					return
				default:
				}
				if libs.Len() != size {
					sawAbsent.Store(true)
				}
			}
		})
	}
	for i := range 1000 {
		oldKey := lib.Key()
		lib.Configure(configs[(i+1)%len(configs)])
		libs.Rekey(oldKey, lib)
	}
	close(stop)
	wg.Wait()
	c.False(sawAbsent.Load(), "the set must never be seen without the library being re-keyed")

	// The library ends up under its new key only.
	c.Equal(size, libs.Len())
	c.Equal(lib, libs.Lookup("someone/first"))
	c.Nil(libs.Lookup("someone/second"))

	// A re-key to the same key is a no-op rather than a removal.
	libs.Rekey(lib.Key(), lib)
	c.Equal(lib, libs.Lookup("someone/first"))
	c.Equal(size, libs.Len())
}

// TestLibrariesMasterAndUserOnlyReadWhenPresent verifies that Master() and User() take only the read lock when the
// library they return is already in the set, which is every call after NewLibraries(). Taking the write lock would hold
// up the background readers of the set, and for the create path it would do so across NewLibrary's read of the version
// file on disk. The read lock is held across the calls here, so either of them taking the write lock deadlocks and is
// reported by the timeout rather than passing.
func TestLibrariesMasterAndUserOnlyReadWhenPresent(t *testing.T) {
	c := check.New(t)
	libs := NewLibraries()
	master := libs.Lookup(masterGitHubAccountName + "/" + masterRepoName)
	user := libs.Lookup("/" + userRepoName)
	c.NotNil(master)
	c.NotNil(user)

	libs.lock.RLock()
	defer libs.lock.RUnlock()
	type result struct{ master, user *Library }
	done := make(chan result, 1)
	go func() { done <- result{master: libs.Master(), user: libs.User()} }()
	select {
	case got := <-done:
		c.Equal(master, got.master)
		c.Equal(user, got.user)
	case <-time.After(5 * time.Second):
		t.Fatal("Master() or User() took the write lock for a library that is already present")
	}
}

// TestLibrariesMasterAndUserCreateOnce verifies the get-or-create contract behind Master() and User() on a set that
// lacks them: callers that miss at the same time share one build, with the first to miss building the library while
// the rest wait for it, rather than each building one and discarding all but the first stored. The build is held open
// until it is known to be in flight, so the overlap is forced rather than left to scheduling, and readers must still
// get through while it is.
func TestLibrariesMasterAndUserCreateOnce(t *testing.T) {
	c := check.New(t)
	const key = "someone/stuff"
	libs := &Libraries{m: make(map[string]*Library)}
	dir := t.TempDir()
	var builds atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	create := func() *Library {
		builds.Add(1)
		startedOnce.Do(func() { close(started) })
		<-release
		return NewLibrary("Test", "someone", "", "stuff", dir)
	}
	got := make([]*Library, 8)
	var wg sync.WaitGroup
	for i := range got {
		wg.Go(func() { got[i] = libs.getOrCreate(key, create) })
	}
	<-started
	c.Nil(libs.Lookup(key), "nothing is stored while the build is in flight")
	c.Equal(0, libs.Len(), "readers must not be blocked by a build in flight")
	close(release)
	wg.Wait()
	c.Equal(int32(1), builds.Load(), "callers that miss concurrently must share one build")
	c.Equal(1, libs.Len(), "exactly one library must be stored")
	stored := libs.Lookup(key)
	c.NotNil(stored)
	for i, lib := range got {
		c.Equal(stored, lib, "caller %d must get the stored library", i)
	}
	c.Equal(int32(1), builds.Load(), "a later call must not build again")
	c.Equal(stored, libs.getOrCreate(key, create))

	// Master() and User() are that path with their own keys and builders.
	for _, tc := range []struct {
		name string
		key  string
		get  func(*Libraries) *Library
	}{
		{name: "master", key: masterGitHubAccountName + "/" + masterRepoName, get: (*Libraries).Master},
		{name: "user", key: "/" + userRepoName, get: (*Libraries).User},
	} {
		libs = &Libraries{m: make(map[string]*Library)}
		first := tc.get(libs)
		c.Equal(tc.key, first.Key(), "%s: the library's own key must match the key it is stored under", tc.name)
		c.Equal(first, libs.Lookup(tc.key), "%s: the library must be stored under its key", tc.name)
		c.Equal(first, tc.get(libs), "%s: a later call must find the stored library", tc.name)
		c.Equal(1, libs.Len(), "%s: a later call must not add another library", tc.name)
	}
}

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
	libs := &Libraries{m: make(map[string]*Library)}
	for _, name := range []string{"alpha", "beta"} {
		lib := NewLibrary(name, "", "", name, t.TempDir())
		c.Equal("0", cachedVersion(lib), "nothing has been recorded for %s yet", name)
		// Written after the library was built, so its cached version starts out as "0" and becoming "7" is proof that
		// the background check actually ran against it.
		c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("7\n"), 0o600))
		libs.Store(lib.Key(), lib)
	}

	libs.PerformUpdateChecks()

	// Modify the map while the checks are in flight, as the UI thread may.
	extra := NewLibrary("extra", "", "", "extra", t.TempDir())
	for range 1000 {
		libs.Store(extra.Key(), extra)
		libs.Remove(extra.Key())
	}

	// Every library in the set at the time of the call is checked, whatever happened to the map afterwards.
	for _, lib := range libs.List() {
		waitForCachedVersion(t, lib, "7")
	}
	c.Equal(2, libs.Len(), "the additions and removals canceled out")
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
	lib := libs.Lookup("someone/repo")
	c.NotNil(lib)
	current, releases := lib.AvailableReleases()
	c.Equal("4", current, "the version on disk is known without a check having run")
	c.Equal(0, len(releases), "no update check has run")

	// The User Library reports no version at all, since it isn't something that gets released.
	user := libs.Lookup("/" + userRepoName)
	c.NotNil(user)
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
