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
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/rjeczalik/notify"
)

// TestLibraryConcurrentAccess verifies that the mutable state of a Library can be read from background goroutines (as
// the update checks do) while the UI thread modifies it. Prior to the accessors being added, these fields were plain
// exported fields with no synchronization at all, which the race detector flags.
func TestLibraryConcurrentAccess(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "token", "repo", filepath.Join(dir, "p0"))
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
				_ = lib.Data()
				_ = lib.Config()
				_ = lib.Favorites()
				_ = lib.Key()
				_ = lib.Valid()
				_ = lib.IsMaster()
				_ = lib.IsUser()
				_ = lib.Path()
				_ = lib.VersionOnDisk()
				_, _ = lib.AvailableReleases()
			}
		})
	}
	wg.Go(func() {
		defer close(stop)
		for i := range 100 {
			lib.Configure(LibraryConfig{
				Title:             fmt.Sprintf("Test %d", i),
				GitHubAccountName: "someone",
				AccessToken:       "token",
				RepoName:          "repo",
				UseLatest:         i%2 == 0,
			})
			lib.ToggleFavorite(fmt.Sprintf("f%d.gcs", i))
			c.NoError(lib.SetPath(filepath.Join(dir, fmt.Sprintf("p%d", i%3))))
		}
	})
	wg.Wait()
	data := lib.Data()
	c.Equal("Test 99", data.Title)
	c.Equal(filepath.Join(dir, "p0"), data.PathOnDisk)
	c.Equal(100, len(lib.Favorites()))
}

// TestLibraryDataIsACopy verifies that the snapshot returned by Data() cannot be used to reach back into the library
// and that Configure() touches only the fields the user is permitted to edit.
func TestLibraryDataIsACopy(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Original", "someone", "secret", "repo", dir)
	lib.ToggleFavorite("a.gcs")
	originalID := lib.Data().ID

	// Mutating a snapshot must not reach the library.
	snapshot := lib.Data()
	snapshot.Title = "Mutated"
	snapshot.PathOnDisk = "/mutated"
	c.Equal("Original", lib.Data().Title)
	c.Equal(dir, lib.Data().PathOnDisk)

	// Likewise for the favorites list, which is handed out as a copy of its own.
	favorites := lib.Favorites()
	favorites[0] = "mutated.gcs"
	c.Equal([]string{"a.gcs"}, lib.Favorites())

	// Configure() has no way to reach the ID, the path on disk or the favorites.
	lib.Configure(LibraryConfig{
		Title:             "Renamed",
		GitHubAccountName: "someone-else",
		AccessToken:       "new-secret",
		RepoName:          "other-repo",
		UseLatest:         true,
	})
	data := lib.Data()
	c.Equal(originalID, data.ID)
	c.Equal(dir, data.PathOnDisk)
	c.Equal([]string{"a.gcs"}, lib.Favorites())
	c.Equal(LibraryConfig{
		Title:             "Renamed",
		GitHubAccountName: "someone-else",
		AccessToken:       "new-secret",
		RepoName:          "other-repo",
		UseLatest:         true,
	}, lib.Config())
	c.Equal(lib.Config(), data.LibraryConfig)
}

func TestLibraryFavorites(t *testing.T) {
	c := check.New(t)
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.Equal(0, len(lib.Favorites()))

	lib.ToggleFavorite("a.gcs")
	lib.ToggleFavorite("b.gcs")
	c.Equal([]string{"a.gcs", "b.gcs"}, lib.Favorites())

	lib.ToggleFavorite("a.gcs")
	c.Equal([]string{"b.gcs"}, lib.Favorites())

	lib.RenameFavorite("b.gcs", "c.gcs")
	c.Equal([]string{"c.gcs"}, lib.Favorites())

	// Renaming something that isn't a favorite must not add it.
	lib.RenameFavorite("missing.gcs", "d.gcs")
	c.Equal([]string{"c.gcs"}, lib.Favorites())
}

// TestLibrarySetPathWhenWatchFails verifies that a path change to a location that cannot be watched still notifies the
// existing watchers rather than panicking on the monitor's discarded task queue.
func TestLibrarySetPathWhenWatchFails(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "", "repo", filepath.Join(dir, "watchable"))
	var syncs atomic.Int64
	token := lib.Watch(func(_ *Library, _ string, what notify.Event) {
		if what == EventRootSync {
			syncs.Add(1)
		}
	}, false)
	defer token.Stop()

	// A regular file where a directory is needed means the new path can neither be created nor watched.
	blocker := filepath.Join(dir, "blocker")
	c.NoError(os.WriteFile(blocker, []byte("not a directory"), 0o600))
	unwatchable := filepath.Join(blocker, "sub")
	c.NotPanics(func() { c.NoError(lib.SetPath(unwatchable)) })
	c.Equal(unwatchable, lib.Data().PathOnDisk)

	waitForMonitorQueue(lib)
	c.Equal(int64(1), syncs.Load())
}

// TestLibrarySetPathAfterWatchFailsRestartsWatch verifies the mirror image of the case above: a library whose watch
// could not be established at all still restarts its watchers when the path changes to a location that can be watched.
// The tokens are registered whether or not the filesystem watch succeeded, so stop() must hand them back for the
// restart even when there is no event channel to shut down. Without that, the library silently stops reporting changes
// for the rest of the session even though it now sits somewhere perfectly watchable.
func TestLibrarySetPathAfterWatchFailsRestartsWatch(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()

	// A regular file where a directory is needed means the initial path can neither be created nor watched.
	blocker := filepath.Join(dir, "blocker")
	c.NoError(os.WriteFile(blocker, []byte("not a directory"), 0o600))
	lib := NewLibrary("Test", "someone", "", "repo", filepath.Join(blocker, "sub"))

	var syncs atomic.Int64
	events := make(chan string, 16)
	token := lib.Watch(func(_ *Library, fullPath string, what notify.Event) {
		if what == EventRootSync {
			syncs.Add(1)
			return
		}
		select {
		case events <- fullPath:
		default:
		}
	}, false)
	defer token.Stop()

	watchable := filepath.Join(dir, "watchable")
	c.NoError(lib.SetPath(watchable))
	waitForMonitorQueue(lib)
	c.Equal(int64(1), syncs.Load(), "the watch is told the root moved")
	c.Equal(watchable, token.root, "the watch is now rooted at the new path")

	// Everything above this point is what the restart actually changed, and it holds on every platform. The live-event
	// check below cannot run on Windows under -race: notify's readdcw backend reads an event's name by viewing its
	// 4096-byte ReadDirectoryChangesW buffer through a *[syscall.MAX_LONG_PATH]uint16, i.e. a 64KB array type
	// (watcher_readdcw.go:406, still present in v0.9.3, the newest release). The bytes it goes on to read stay inside
	// the buffer -- the kernel bounds each record to the size handed to ReadDirectoryChangesW -- but the conversion
	// itself nominally spans past the allocation, so checkptr, which -race turns on, aborts the process the first time
	// an event is delivered. That takes the whole test binary with it rather than failing a single test. Release builds
	// never enable -race, so no shipped binary carries the check.
	if raceEnabled && runtime.GOOS == "windows" {
		t.Skip("rjeczalik/notify v0.9.3 trips checkptr on Windows when a watch event is delivered under -race")
	}

	// The watch must actually be live at the new location, not merely re-registered.
	created := filepath.Join(watchable, "new.gcs")
	c.NoError(os.WriteFile(created, []byte("{}"), 0o600))
	// The platform watchers may also report the creation of the new root itself, so look past anything else.
	deadline := time.After(10 * time.Second)
	for {
		select {
		case p := <-events:
			if filepath.Base(p) == "new.gcs" {
				return
			}
		case <-deadline:
			t.Fatalf("no filesystem event was delivered for %s", created)
		}
	}
}

// TestLibraryWatchDeliversNothingUntilSomethingHappens verifies that establishing a watch delivers no event of its own,
// root sync or otherwise. The caller has just scanned the library itself -- that is why it is now watching it -- so an
// event at this point is at best redundant. For a callback that rescans whatever it is told about (the library
// navigator does exactly that, and re-establishes its watches as part of the rescan), it would never settle: each
// rescan would establish a watch that hands it another event.
func TestLibraryWatchDeliversNothingUntilSomethingHappens(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "", "repo", filepath.Join(dir, "root"))
	var events atomic.Int64
	token := lib.Watch(func(_ *Library, _ string, _ notify.Event) { events.Add(1) }, false)
	defer token.Stop()

	waitForMonitorQueue(lib)
	c.Equal(int64(0), events.Load(), "a newly established watch reports nothing")

	// A second watch on the same library must not stir up the first one either.
	other := lib.Watch(func(_ *Library, _ string, _ notify.Event) { events.Add(1) }, false)
	defer other.Stop()
	waitForMonitorQueue(lib)
	c.Equal(int64(0), events.Load(), "adding another watch reports nothing")

	// Moving the root is what a root sync is for, and now both watches get exactly one.
	c.NoError(lib.SetPath(filepath.Join(dir, "moved")))
	waitForMonitorQueue(lib)
	c.Equal(int64(2), events.Load(), "each watch is told the root moved, once")
}

// TestLibrarySetPathSyncsEachWatcherOnce verifies that restarting the watches after a path change delivers exactly one
// root sync to each watcher, rather than one per watcher to every watcher.
func TestLibrarySetPathSyncsEachWatcherOnce(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	lib := NewLibrary("Test", "someone", "", "repo", filepath.Join(dir, "one"))
	const watchers = 3
	syncs := make([]atomic.Int64, watchers)
	tokens := make([]*MonitorToken, 0, watchers)
	for i := range watchers {
		tokens = append(tokens, lib.Watch(func(_ *Library, _ string, what notify.Event) {
			if what == EventRootSync {
				syncs[i].Add(1)
			}
		}, false))
	}
	defer func() {
		for _, token := range tokens {
			token.Stop()
		}
	}()

	c.NoError(lib.SetPath(filepath.Join(dir, "two")))
	waitForMonitorQueue(lib)
	for i := range watchers {
		c.Equal(int64(1), syncs[i].Load(), "watcher %d", i)
	}
}

func TestLibraryJSONRoundTrip(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, releaseFile), []byte("4\n"), 0o600))
	lib := NewLibrary("Extra", "someone", "secret", "repo", dir)
	lib.ToggleFavorite("b.gcs")
	lib.ToggleFavorite("a.gcs")
	libs := Libraries(map[string]*Library{lib.Key(): lib})
	data, err := jio.Marshal(&libs)
	c.NoError(err)

	// Pin the serialized shape. In particular, the GitHub account and repository names must not appear in the library
	// object itself, since they travel in the key it is filed under.
	var raw map[string]map[string]any
	c.NoError(jio.Unmarshal(data, &raw))
	obj, found := raw["someone/repo"]
	c.True(found)
	c.Equal([]string{"access_token", "favorites", "id", "path", "title"}, slices.Sorted(maps.Keys(obj)))

	var loaded Libraries
	c.NoError(jio.Unmarshal(data, &loaded))
	restored, ok := loaded["someone/repo"]
	c.True(ok)
	c.NotNil(restored)
	c.Equal(lib.Data(), restored.Data())
	c.Equal(lib.Favorites(), restored.Favorites())
	current, _ := restored.AvailableReleases()
	c.Equal("4", current, "the version on disk isn't saved, so it must be read when the library is loaded")
}

// TestLibrariesUnmarshalSkipsNullEntries verifies that a null in place of a library -- which a hand-edited or damaged
// settings file may hold, and which decodes without error as a nil *Library -- is skipped rather than dereferenced. A
// panic here would bypass the recovery loadSettingsOrDefaults provides and take GCS down at startup.
func TestLibrariesUnmarshalSkipsNullEntries(t *testing.T) {
	c := check.New(t)
	var libs Libraries
	c.NotPanics(func() { c.NoError(jio.Unmarshal([]byte(`{"a/b":null}`), &libs)) })
	c.Equal(0, len(libs))

	// A null alongside a usable entry must cost only the null.
	c.NotPanics(func() {
		c.NoError(jio.Unmarshal([]byte(`{"a/b":null,"someone/repo":{"title":"Good","path":"/libs/good"}}`), &libs))
	})
	c.Equal(1, len(libs))
	lib, ok := libs["someone/repo"]
	c.True(ok)
	c.NotNil(lib)
	c.Equal("Good", lib.Data().Title)

	// The guard belongs to Valid() itself, so anything else holding a library that came from a file is covered too.
	var missing *Library
	c.NotPanics(func() { c.False(missing.Valid()) })
}

// TestLibraryDownloadReleaseChecksStatusCode verifies that an HTTP error response is reported as an HTTP failure rather
// than being handed to the zip reader as if it were archive content, which yielded a misleading "unable to open
// archive" error.
func TestLibraryDownloadReleaseChecksStatusCode(t *testing.T) {
	c := check.New(t)
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, "not a zip file")
	}))
	defer srv.Close()
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	release := Release{Version: "1.0.0", ZipFileURL: srv.URL}

	for _, code := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusTooManyRequests} {
		status = code
		data, err := lib.downloadRelease(t.Context(), srv.Client(), &release, nil)
		c.HasError(err, "status %d", code)
		c.Nil(data, "status %d", code)
		c.Contains(err.Error(), strconv.Itoa(code), "status %d", code)
		c.Contains(err.Error(), srv.URL, "status %d", code)
	}

	// A successful response must still hand back the body unchanged, and must account for every byte of it, since the
	// progress bar is scaled by that count.
	status = http.StatusOK
	var counted int64
	data, err := lib.downloadRelease(t.Context(), srv.Client(), &release, func(n int64) { counted += n })
	c.NoError(err)
	c.Equal([]byte("not a zip file"), data)
	c.Equal(int64(len("not a zip file")), counted)
}

// TestCheckForAvailableUpgradeNotifiesLateInstalledFunc verifies that an update check completing before the UI has
// installed a notification function still delivers that notification once one is installed. The update checks are
// started by Start() before the workspace (and therefore the navigator) exists, so a fast check would otherwise be
// silently dropped, leaving no "update available" indicator until some unrelated reload happened to occur.
func TestCheckForAvailableUpgradeNotifiesLateInstalledFunc(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	defer resetLibraryChangeNotification()

	client, setReleases := newReleasesServer(t, "5")

	// The library has no release.txt, so its version on disk is "0" while the newest release is "5": an upgrade is
	// available and a notification is due.
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	lib.CheckForAvailableUpgrade(t.Context(), client)
	current, releases := lib.AvailableReleases()
	c.Equal("0", current)
	c.Equal(1, len(releases))

	// Installing the function afterwards must still deliver the notification...
	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })
	c.Equal(int64(1), calls.Load())

	// ...but only once, so a later navigator doesn't get a stale notification.
	var later atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { later.Add(1) })
	c.Equal(int64(0), later.Load())

	// With a function installed, a check that turns up a release the previous one didn't notifies directly.
	setReleases("5.1", "5")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), later.Load())

	// The library being brought up to date outside of the app takes the update that was on offer away, which the
	// navigator has to hear about so that the indicator comes down; a check after that has nothing new to say.
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("5.1\n"), 0o600))
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), later.Load(), "an update that is no longer on offer must be announced")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), later.Load(), "a check that finds nothing newer than what is on disk must not notify")
}

// TestCheckForAvailableUpgradeNotifiesOnceWhenNothingChanges verifies that repeating a check whose answer hasn't
// changed notifies only the first time. A notification reloads the entire library tree, so a check that ran every hour
// and notified every time would restart the filesystem watches, drop the caches and disturb whatever the user had in
// progress, over and over, for as long as an update went uninstalled.
func TestCheckForAvailableUpgradeNotifiesOnceWhenNothingChanges(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	// The function is installed before the first check, so nothing can be latched as a pending notification instead of
	// being counted here.
	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, _ := newReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), calls.Load(), "the first check surfaces the pending update")
	for range 3 {
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}
	c.Equal(int64(1), calls.Load(), "repeating the same check announces nothing new")

	// A library whose content was replaced on disk outside of the app is a change the next check must announce, since
	// the version the navigator is showing is no longer the one that is there.
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("4\n"), 0o600))
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), calls.Load(), "the version on disk changed")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), calls.Load(), "...and once is enough for that, too")
}

// TestCheckForAvailableUpgradeStaysQuietForLocalLibrary verifies that a library that isn't backed by a GitHub repo
// never notifies. There are no releases for it, so the newest release reads as "" while the version on disk reads as
// "0" for want of a release.txt, and comparing those two directly made every check of such a library announce an
// update that doesn't exist -- which, with periodic checks, would reload the library tree on every one of them.
func TestCheckForAvailableUpgradeStaysQuietForLocalLibrary(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	// An empty GitHub account name means there is nothing to ask about, so the nil client is never touched. Should that
	// ever cease to be true, the test fails loudly rather than reaching out to the network.
	lib := NewLibrary("Local", "", "", "local", t.TempDir())
	for range 3 {
		lib.CheckForAvailableUpgrade(t.Context(), nil)
	}
	c.Equal(int64(0), calls.Load())
	current, releases := lib.AvailableReleases()
	c.Equal("0", current)
	c.Equal(0, len(releases))
}

// TestNeedsUpgradeCheck verifies what the Library Explorer relies on to keep its update buttons usable when the periodic
// checks are turned off: a library with a repository behind it reports that it needs a check until one completes, a
// check that couldn't reach the repository leaves it needing one so that it can be tried again, and a library with no
// repository never needs one, since there is nothing to ask.
func TestNeedsUpgradeCheck(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	local := NewLibrary("Local", "", "", "local", t.TempDir())
	c.False(local.NeedsUpgradeCheck(), "a library without a repository has nothing to check")
	local.CheckForAvailableUpgrade(t.Context(), nil)
	c.False(local.NeedsUpgradeCheck())

	client, _ := newReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.True(lib.NeedsUpgradeCheck(), "a library that has never been checked needs a check")

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	lib.CheckForAvailableUpgrade(ctx, client)
	c.True(lib.NeedsUpgradeCheck(), "a check that failed must leave the library needing one")
	_, releases := lib.AvailableReleases()
	c.Equal(0, len(releases))

	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.False(lib.NeedsUpgradeCheck(), "a check that completed satisfies the need")
	_, releases = lib.AvailableReleases()
	c.Equal(1, len(releases))
}

// TestCheckForAvailableUpgradeNotifiesWhenReleaseAppears verifies that a check turning up a release the previous check
// didn't have does notify, which is the entire point of checking more than once.
func TestCheckForAvailableUpgradeNotifiesWhenReleaseAppears(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, setReleases := newReleasesServer(t)
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(0), calls.Load(), "a repo with no releases yet has nothing to announce")

	setReleases("5")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), calls.Load(), "the release that has just appeared is announced")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), calls.Load(), "...and only once")

	// A newer release turning up later is a change of its own.
	setReleases("5.1", "5")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), calls.Load())
	current, releases := lib.AvailableReleases()
	c.Equal("0", current)
	c.Equal(2, len(releases))
}

// TestCheckForAvailableUpgradeNotifiesWhenUpdateDisappears verifies that a check which finds an update that was on offer
// no longer is -- the release having been withdrawn, or the library having been brought up to date outside of the app
// -- notifies, so that the Library Explorer's indicator and buttons don't go on offering it. The rule that keeps a
// repeated check quiet used to keep this one quiet too, leaving the stale indicator up until something else happened
// to reload the tree.
func TestCheckForAvailableUpgradeNotifiesWhenUpdateDisappears(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, setReleases := newReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), calls.Load(), "the update is announced")

	// Withdrawn: the repo no longer has a release to offer.
	setReleases()
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), calls.Load(), "the update being withdrawn must be announced")
	_, releases := lib.AvailableReleases()
	c.Equal(0, len(releases))
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(2), calls.Load(), "...but only once")

	// Installed outside of the app: the release is back, and the disk already has it.
	setReleases("5")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(3), calls.Load(), "the update reappearing must be announced")
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("5\n"), 0o600))
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(4), calls.Load(), "the library being brought up to date must be announced")
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(4), calls.Load(), "...but only once")

	// A library that never had an update on offer has nothing to take away.
	fresh := NewLibrary("Fresh", "someone", "", "repo", t.TempDir())
	c.NoError(os.WriteFile(filepath.Join(fresh.Data().PathOnDisk, releaseFile), []byte("5\n"), 0o600))
	for range 2 {
		fresh.CheckForAvailableUpgrade(t.Context(), client)
	}
	c.Equal(int64(4), calls.Load(), "a library that is already up to date has nothing to announce")
}

// TestConfigureDiscardsChecksOfTheOldRepository verifies that pointing a library at a different repository -- or at the
// same one's latest commit rather than its releases -- discards what an earlier check found and leaves the library
// needing a check, since those releases belong to the old repository. Without this, the Library Explorer went on
// offering the old repository's releases, and its on-click check never engaged because the library still counted as
// checked. Changes to the other settings leave the check state alone: there is nothing about the answer they would
// change.
func TestConfigureDiscardsChecksOfTheOldRepository(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	client, _ := newReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.False(lib.NeedsUpgradeCheck())
	_, releases := lib.AvailableReleases()
	c.Equal(1, len(releases))

	// A new title or token doesn't change which repository is being asked, so the answer stands.
	config := lib.Config()
	config.Title = "Renamed"
	config.AccessToken = "token"
	lib.Configure(config)
	c.False(lib.NeedsUpgradeCheck(), "renaming a library must not discard its check")
	_, releases = lib.AvailableReleases()
	c.Equal(1, len(releases))

	// Re-applying the same configuration, which the settings dialog does, stands too.
	lib.Configure(lib.Config())
	c.False(lib.NeedsUpgradeCheck(), "re-applying the same configuration must not discard the check")

	for _, one := range []struct {
		name   string
		change func(config *LibraryConfig)
	}{
		{name: "account", change: func(config *LibraryConfig) { config.GitHubAccountName += "x" }},
		{name: "repo", change: func(config *LibraryConfig) { config.RepoName += "x" }},
		{name: "use latest", change: func(config *LibraryConfig) { config.UseLatest = !config.UseLatest }},
	} {
		lib = NewLibrary("Test", "someone", "", "repo", t.TempDir())
		lib.CheckForAvailableUpgrade(t.Context(), client)
		c.False(lib.NeedsUpgradeCheck())
		config = lib.Config()
		one.change(&config)
		lib.Configure(config)
		c.True(lib.NeedsUpgradeCheck(), "changing the %s must leave the library needing a check", one.name)
		_, releases = lib.AvailableReleases()
		c.Equal(0, len(releases), "changing the %s must discard the old repository's releases", one.name)
		c.Equal(config, lib.Config(), "the new configuration must have been applied")
	}

	// The version on disk follows the key, since whether the library is the User Library -- which has no version --
	// is a property of the key.
	lib = NewLibrary("Test", "someone", "", "repo", t.TempDir())
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("4\n"), 0o600))
	lib.refreshVersionOnDisk()
	current, _ := lib.AvailableReleases()
	c.Equal("4", current)
	config = lib.Config()
	config.GitHubAccountName = ""
	config.RepoName = userRepoName
	lib.Configure(config)
	c.True(lib.IsUser())
	current, _ = lib.AvailableReleases()
	c.Equal("", current, "the User Library has no version on disk")
	c.False(lib.NeedsUpgradeCheck(), "the User Library has no repository to check")
}

// TestCheckForAvailableUpgradeJoinsACheckInFlight verifies that a check made while another is under way waits for that
// one rather than making a request of its own. The Library Explorer asks on the user's behalf while the launch-time or
// periodic check may still be running, and without this each library cost two requests -- which count against GitHub's
// unauthenticated rate limit -- for one answer.
func TestCheckForAvailableUpgradeJoinsACheckInFlight(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	first := make(chan struct{})
	go func() {
		defer close(first)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	second := make(chan struct{})
	go func() {
		defer close(second)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	select {
	case <-second:
		t.Fatal("the second check must wait for the first rather than return at once")
	case <-time.After(50 * time.Millisecond):
	}
	c.True(lib.NeedsUpgradeCheck(), "a check in flight doesn't yet satisfy the need for one")
	srv.release()
	<-first
	<-second
	c.Equal(int64(1), srv.requests.Load(), "the second check must not make a request of its own")
	c.False(lib.NeedsUpgradeCheck())
	_, releases := lib.AvailableReleases()
	c.Equal(1, len(releases))

	// Once the check is over, the next one is a check of its own.
	third := make(chan struct{})
	go func() {
		defer close(third)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	srv.release()
	<-third
	c.Equal(int64(2), srv.requests.Load())
}

// TestCheckForAvailableUpgradeWaiterHonorsItsContext verifies that a check waiting on another gives up when its own
// context is done. The check made from the Library Explorer allows itself a minute and may be canceled by the user,
// and neither must be held up by the five minutes the background check allows itself.
func TestCheckForAvailableUpgradeWaiterHonorsItsContext(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	first := make(chan struct{})
	go func() {
		defer close(first)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	lib.CheckForAvailableUpgrade(ctx, client) // Must return without waiting for the first check to finish
	c.True(lib.NeedsUpgradeCheck(), "giving up on the wait must not count as a completed check")
	srv.release()
	<-first
	c.False(lib.NeedsUpgradeCheck(), "the first check must complete as usual")
	c.Equal(int64(1), srv.requests.Load())
}

// TestConfigureDiscardsACheckInFlight verifies that a check which was under way when the library was pointed at a
// different repository is discarded when it finishes, since its answer describes the old repository. The settings
// dialog can be applied while the launch-time check is still running, and the old answer landing after the new
// configuration would put the old repository's releases back on display and mark the library as checked.
func TestConfigureDiscardsACheckInFlight(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	done := make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	config := lib.Config()
	config.RepoName = "other"
	lib.Configure(config)
	srv.release()
	<-done
	c.True(lib.NeedsUpgradeCheck(), "the stale check must not count as a check of the new repository")
	_, releases := lib.AvailableReleases()
	c.Equal(0, len(releases), "the stale check's releases must be discarded")
	c.Equal(int64(0), calls.Load(), "a discarded check has nothing to announce")

	// The next check is of the new repository, and lands as usual.
	done = make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	srv.release()
	<-done
	c.False(lib.NeedsUpgradeCheck(), "the check of the new repository must complete")
	c.Equal(int64(1), calls.Load())
}

// TestConfigureForKeyDiscardsChecksOfTheOldRepository verifies that ConfigureForKey() treats a change of repository as
// Configure() does: the old repository's releases are dropped, the library is left needing a check, a check in flight
// is discarded when it finishes, and the version on disk is re-read. Loading a saved set is its only use today, on
// libraries that have never been checked, but it is exported, and nothing keeps it from being called on one that has.
func TestConfigureForKeyDiscardsChecksOfTheOldRepository(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	done := make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	srv.release()
	<-done
	c.False(lib.NeedsUpgradeCheck())
	_, releases := lib.AvailableReleases()
	c.Equal(1, len(releases))

	// Re-applying the same key is not a change, so the answer stands.
	c.NoError(lib.ConfigureForKey("someone/repo"))
	c.False(lib.NeedsUpgradeCheck(), "re-applying the same key must not discard the check")
	_, releases = lib.AvailableReleases()
	c.Equal(1, len(releases))

	c.NoError(lib.ConfigureForKey("someone/other"))
	c.Equal("someone/other", lib.Key())
	c.True(lib.NeedsUpgradeCheck(), "changing the key must leave the library needing a check")
	_, releases = lib.AvailableReleases()
	c.Equal(0, len(releases), "changing the key must discard the old repository's releases")

	// A check that was under way when the key changed is discarded when it finishes.
	done = make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	c.NoError(lib.ConfigureForKey("someone/another"))
	srv.release()
	<-done
	c.True(lib.NeedsUpgradeCheck(), "the stale check must not count as a check of the new repository")
	_, releases = lib.AvailableReleases()
	c.Equal(0, len(releases), "the stale check's releases must be discarded")

	// The version on disk follows the key, since whether the library is the User Library -- which has no version --
	// is a property of the key.
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("4\n"), 0o600))
	c.NoError(lib.ConfigureForKey("someone/repo"))
	current, _ := lib.AvailableReleases()
	c.Equal("4", current)
	c.NoError(lib.ConfigureForKey("/" + userRepoName))
	c.True(lib.IsUser())
	current, _ = lib.AvailableReleases()
	c.Equal("", current, "the User Library has no version on disk")
	c.False(lib.NeedsUpgradeCheck(), "the User Library has no repository to check")

	// A key that doesn't name a repository is rejected without changing anything.
	c.NoError(lib.ConfigureForKey("someone/repo"))
	for _, key := range []string{"", "someone", "someone/", "someone/ "} {
		c.HasError(lib.ConfigureForKey(key), "key %q", key)
		c.Equal("someone/repo", lib.Key(), "key %q", key)
	}
}

// TestSetPathDiscardsACheckInFlight verifies that a check which was under way when the library was pointed at a
// different directory is discarded when it finishes. A check reads the version on disk before it takes the library
// lock, so one that read it from the old directory could otherwise land after the move and put the old directory's
// version back in place of the new one, showing an update for a library that is current until the next scheduled check
// put it right.
func TestSetPathDiscardsACheckInFlight(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, srv := newBlockingReleasesServer(t, "5")
	oldDir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(oldDir, releaseFile), []byte("4\n"), 0o600))
	newDir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(newDir, releaseFile), []byte("5\n"), 0o600))
	lib := NewLibrary("Test", "someone", "", "repo", oldDir)
	current, _ := lib.AvailableReleases()
	c.Equal("4", current)

	done := make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	c.NoError(lib.SetPath(newDir))
	current, _ = lib.AvailableReleases()
	c.Equal("5", current, "the version must follow the path")
	srv.release()
	<-done
	c.True(lib.NeedsUpgradeCheck(), "a check that began before the path changed must be discarded")
	current, releases := lib.AvailableReleases()
	c.Equal("5", current, "the discarded check must not put back the version it read")
	c.Equal(0, len(releases), "the discarded check's releases must not be recorded")
	c.Equal(int64(0), calls.Load(), "a discarded check has nothing to announce")

	// Re-applying the same path is not a change, so a check that spans it lands as usual.
	done = make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	c.NoError(lib.SetPath(newDir))
	srv.release()
	<-done
	c.False(lib.NeedsUpgradeCheck(), "re-applying the same path must not discard the check")
	current, releases = lib.AvailableReleases()
	c.Equal("5", current)
	c.Equal(1, len(releases))
	c.Equal(int64(0), calls.Load(), "a library that is current has nothing to announce")
}

// TestDownloadDiscardsACheckInFlight verifies that a check which was under way while the library's content was being
// replaced by a download is discarded when it finishes, whether the download succeeded or, having moved the old content
// aside and then put it back, failed. A check reads the version on disk before it takes the library lock, so one that
// read it before the download was done could otherwise land afterwards with the version it read -- or with the "0"
// that stands in for the directory while it is aside -- and show an update for a library that had just been brought
// up to date. The update flow waits on that check once the download is over, and would have been told it was served.
func TestDownloadDiscardsACheckInFlight(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, srv := newBlockingReleasesServer(t, "5")
	dir := t.TempDir()
	c.NoError(os.WriteFile(filepath.Join(dir, releaseFile), []byte("4\n"), 0o600))
	lib := NewLibrary("Test", "someone", "", "repo", dir)
	current, _ := lib.AvailableReleases()
	c.Equal("4", current)

	archive := newLibraryArchiveServer(t)
	for _, one := range []struct {
		name    string
		status  int
		version string // what must be on disk once the download is over
	}{
		{name: "failed download", status: http.StatusInternalServerError, version: "4"},
		{name: "successful download", status: http.StatusOK, version: "5"},
	} {
		done := make(chan struct{})
		go func() {
			defer close(done)
			lib.CheckForAvailableUpgrade(t.Context(), client)
		}()
		<-srv.started
		archive.status.Store(int64(one.status))
		err := lib.Download(t.Context(), archive.client, &Release{Version: "5", ZipFileURL: archive.url}, nil)
		if one.status == http.StatusOK {
			c.NoError(err, one.name)
		} else {
			c.HasError(err, one.name)
		}
		current, _ = lib.AvailableReleases()
		c.Equal(one.version, current, "%s: the version must be re-read once the download is over", one.name)
		srv.release()
		<-done
		c.True(lib.NeedsUpgradeCheck(), "%s: a check that began before the download must be discarded", one.name)
		var releases []Release
		current, releases = lib.AvailableReleases()
		c.Equal(one.version, current, "%s: the discarded check must not put back the version it read", one.name)
		c.Equal(0, len(releases), "%s: the discarded check's releases must not be recorded", one.name)
		c.Equal(int64(0), calls.Load(), "%s: a discarded check has nothing to announce", one.name)
	}
	content, err := os.ReadFile(filepath.Join(dir, "a.txt"))
	c.NoError(err)
	c.Equal(libraryArchiveFileContent, string(content), "the successful download must have installed the content")

	// With the download over, a check lands as usual and finds the library current.
	done := make(chan struct{})
	go func() {
		defer close(done)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	srv.release()
	<-done
	c.False(lib.NeedsUpgradeCheck())
	current, releases := lib.AvailableReleases()
	c.Equal("5", current)
	c.Equal(1, len(releases))
	c.Equal(int64(0), calls.Load(), "a library that is current has nothing to announce")
}

// TestCheckForAvailableUpgradeAsksAgainWhenTheJoinedCheckIsDiscarded verifies that a check which waited on one that was
// then discarded -- the library having been pointed at a different repository while it ran -- goes on to make a check
// of its own, of the new repository. The settings dialog starts a check as soon as a new repository is applied, and
// the Library Explorer's buttons make one on demand; with the launch-time check still in flight, both used to wait on
// it, see it discarded, and return with nothing, leaving the library unchecked until the next scheduled check.
func TestCheckForAvailableUpgradeAsksAgainWhenTheJoinedCheckIsDiscarded(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	first := make(chan struct{})
	go func() {
		defer close(first)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	<-srv.started
	config := lib.Config()
	config.RepoName = "other"
	lib.Configure(config)
	second := make(chan struct{})
	go func() {
		defer close(second)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	select {
	case <-second:
		t.Fatal("the second check must wait for the first rather than return at once")
	case <-time.After(50 * time.Millisecond):
	}
	c.Equal(int64(1), srv.requests.Load(), "the second check must not ask while the first is still in flight")
	srv.release()
	<-first
	c.True(lib.NeedsUpgradeCheck(), "the first check was of the old repository and must be discarded")
	<-srv.started
	select {
	case <-second:
		t.Fatal("the second check must not return before its own request has been answered")
	default:
	}
	srv.release()
	<-second
	c.Equal(int64(2), srv.requests.Load(), "the second check must make a request of its own")
	c.Equal([]string{"/repos/someone/repo/releases", "/repos/someone/other/releases"}, srv.requestPaths(),
		"the second check must ask the new repository")
	c.False(lib.NeedsUpgradeCheck())
	_, releases := lib.AvailableReleases()
	c.Equal(1, len(releases))
	c.Equal(int64(1), calls.Load(), "the new repository's update must be announced")
}

// TestCheckForAvailableUpgradeAsksAgainWhenTheJoinedCheckIsCanceled verifies that a check which waited on one that was
// cut short by its own caller's context -- the user canceling the Library Explorer's check, say, while the periodic
// check was waiting on it -- goes on to make a check of its own, since its context is still good and the library is
// still unchecked. A check that failed to reach the repository is not asked again, since asking at once would only fail
// again.
func TestCheckForAvailableUpgradeAsksAgainWhenTheJoinedCheckIsCanceled(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	t.Cleanup(resetLibraryChangeNotification)

	client, srv := newBlockingReleasesServer(t, "5")
	lib := NewLibrary("Test", "someone", "", "repo", t.TempDir())
	ctx, cancel := context.WithCancel(t.Context())
	first := make(chan struct{})
	go func() {
		defer close(first)
		lib.CheckForAvailableUpgrade(ctx, client)
	}()
	<-srv.started
	second := make(chan struct{})
	go func() {
		defer close(second)
		lib.CheckForAvailableUpgrade(t.Context(), client)
	}()
	select {
	case <-second:
		t.Fatal("the second check must wait for the first rather than return at once")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	<-first
	c.True(lib.NeedsUpgradeCheck(), "a canceled check doesn't count as a completed one")
	<-srv.started  // The second check's own request
	<-srv.finished // The canceled request's handler has let go, so the release below reaches the second request
	srv.release()
	<-second
	c.Equal(int64(2), srv.requests.Load(), "the second check must make a request of its own")
	c.False(lib.NeedsUpgradeCheck())
	_, releases := lib.AvailableReleases()
	c.Equal(1, len(releases))

	// A failure to reach the repository is passed on to whoever was waiting, rather than being tried again at once.
	unreachable := &blockingFailingTransport{started: make(chan struct{}, 16), releases: make(chan struct{}, 16)}
	failing := &http.Client{Transport: unreachable}
	other := NewLibrary("Other", "someone", "", "repo", t.TempDir())
	first = make(chan struct{})
	go func() {
		defer close(first)
		other.CheckForAvailableUpgrade(t.Context(), failing)
	}()
	<-unreachable.started
	second = make(chan struct{})
	go func() {
		defer close(second)
		other.CheckForAvailableUpgrade(t.Context(), failing)
	}()
	select {
	case <-second:
		t.Fatal("the second check must wait for the first rather than return at once")
	case <-time.After(50 * time.Millisecond):
	}
	unreachable.releases <- struct{}{}
	<-first
	<-second
	c.Equal(int64(1), unreachable.attempts.Load(), "a check that failed to reach the repository must not be repeated at once")
	c.True(other.NeedsUpgradeCheck(), "a check that couldn't reach the repository doesn't count as a completed one")
}

// blockingFailingTransport holds each request until told to let it go, then fails it as though the repository couldn't
// be reached.
type blockingFailingTransport struct {
	started  chan struct{} // receives one value per request, as it arrives
	releases chan struct{} // each value sent lets one request fail
	attempts atomic.Int64
}

func (t *blockingFailingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.attempts.Add(1)
	t.started <- struct{}{}
	select {
	case <-t.releases:
	case <-req.Context().Done():
	}
	return nil, errors.New("no route to host")
}

// TestNotifyOfLibraryChangeConcurrent verifies that the notification function may be installed by the UI thread while
// background goroutines report library changes. It was previously a plain package variable written by the navigator
// and read by the update-check goroutines, which the race detector flags.
func TestNotifyOfLibraryChangeConcurrent(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	defer resetLibraryChangeNotification()

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
				NotifyOfLibraryChange()
			}
		})
	}
	var calls atomic.Int64
	wg.Go(func() {
		defer close(stop)
		for range 100 {
			SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })
			SetNotifyOfLibraryChangeFunc(nil)
		}
	})
	wg.Wait()

	// Whatever interleaving occurred, the notification state must still be usable afterwards.
	resetLibraryChangeNotification()
	var after atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { after.Add(1) })
	c.Equal(int64(0), after.Load())
	NotifyOfLibraryChange()
	c.Equal(int64(1), after.Load())
}

// TestNotifyOfLibraryChangePendingSurvivesRemoval verifies that removing the notification function doesn't discard a
// notification that hasn't been delivered yet.
func TestNotifyOfLibraryChangePendingSurvivesRemoval(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	defer resetLibraryChangeNotification()

	NotifyOfLibraryChange()
	SetNotifyOfLibraryChangeFunc(nil)
	var calls atomic.Int64
	SetNotifyOfLibraryChangeFunc(func() { calls.Add(1) })
	c.Equal(int64(1), calls.Load())
}

// TestNotifyOfLibraryChangeSurvivesPanic verifies that a notification function that panics doesn't take the calling
// background goroutine down with it.
func TestNotifyOfLibraryChangeSurvivesPanic(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	defer resetLibraryChangeNotification()

	SetNotifyOfLibraryChangeFunc(func() { panic("boom") })
	c.NotPanics(NotifyOfLibraryChange)
}

// resetLibraryChangeNotification restores the package-level library change notification state, so that tests touching
// it don't leak into one another.
func resetLibraryChangeNotification() {
	libraryChangeLock.Lock()
	notifyOfLibraryChange = nil
	pendingLibraryChange = false
	libraryChangeLock.Unlock()
}

// newReleasesServer starts a stand-in for the GitHub releases API. It returns a client that reaches it in place of the
// API host baked into LoadReleases, along with a function that replaces what it serves, so that a test can run more
// than one check against differing answers. Each version given becomes a release, in the order supplied.
func newReleasesServer(t *testing.T, initialVersions ...string) (client *http.Client, setReleases func(versions ...string)) {
	t.Helper()
	var body atomic.Pointer[string]
	setReleases = func(versions ...string) {
		entries := make([]string, 0, len(versions))
		for _, version := range versions {
			entries = append(entries,
				fmt.Sprintf(`{"tag_name":"v%s","body":"notes","zipball_url":"http://127.0.0.1/z.zip"}`, version))
		}
		content := "[" + strings.Join(entries, ",") + "]"
		body.Store(&content)
	}
	setReleases(initialVersions...)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, *body.Load())
	}))
	t.Cleanup(srv.Close)
	return &http.Client{Transport: &redirectingTransport{host: srv.Listener.Addr().String()}}, setReleases
}

// blockingReleasesServer is a releases server that holds each request until told to let it go, so that a test can
// arrange for a check to be in flight.
type blockingReleasesServer struct {
	started  chan struct{} // receives one value per request, as it arrives
	finished chan struct{} // receives one value per request, as its handler returns
	releases chan struct{} // each value sent lets one request finish
	lock     sync.Mutex
	paths    []string // the path of each request, in the order received
	requests atomic.Int64
}

// release lets the request currently being held finish.
func (s *blockingReleasesServer) release() {
	s.releases <- struct{}{}
}

// requestPaths returns the path of each request received so far.
func (s *blockingReleasesServer) requestPaths() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return slices.Clone(s.paths)
}

// newBlockingReleasesServer returns a client whose requests reach a server that holds each request until release() is
// called, along with the server itself. Requests still held when the test ends are let go by the cleanup.
func newBlockingReleasesServer(t *testing.T, versions ...string) (*http.Client, *blockingReleasesServer) {
	t.Helper()
	entries := make([]string, 0, len(versions))
	for _, version := range versions {
		entries = append(entries,
			fmt.Sprintf(`{"tag_name":"v%s","body":"notes","zipball_url":"http://127.0.0.1/z.zip"}`, version))
	}
	content := "[" + strings.Join(entries, ",") + "]"
	s := &blockingReleasesServer{
		started:  make(chan struct{}, 16),
		finished: make(chan struct{}, 16),
		releases: make(chan struct{}, 16),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)
		s.lock.Lock()
		s.paths = append(s.paths, r.URL.Path)
		s.lock.Unlock()
		s.started <- struct{}{}
		select {
		case <-s.releases:
		case <-r.Context().Done():
		}
		fmt.Fprint(w, content)
		s.finished <- struct{}{}
	}))
	t.Cleanup(func() {
		close(s.releases)
		srv.Close()
	})
	return &http.Client{Transport: &redirectingTransport{host: srv.Listener.Addr().String()}}, s
}

// libraryArchiveFileContent is what the one file in the archive newLibraryArchiveServer serves holds.
const libraryArchiveFileContent = "hello\n"

// libraryArchiveServer serves a source archive of the shape GitHub produces for a release -- everything beneath a
// single top-level directory, with the library's content in its "Library" folder -- holding one file, or fails each
// request with whatever status it has been set to, so that a test can run a download that fails as well as one that
// succeeds.
type libraryArchiveServer struct {
	client *http.Client
	url    string
	status atomic.Int64
}

func newLibraryArchiveServer(t *testing.T) *libraryArchiveServer {
	t.Helper()
	c := check.New(t)
	archive := filepath.Join(t.TempDir(), "archive.zip")
	f, err := os.Create(archive)
	c.NoError(err)
	zw := zip.NewWriter(f)
	w, err := zw.Create("repo-abc123/Library/a.txt")
	c.NoError(err)
	_, err = w.Write([]byte(libraryArchiveFileContent))
	c.NoError(err)
	c.NoError(zw.Close())
	c.NoError(f.Close())
	s := &libraryArchiveServer{}
	s.status.Store(http.StatusOK)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status := int(s.status.Load()); status != http.StatusOK {
			http.Error(w, http.StatusText(status), status)
			return
		}
		http.ServeFile(w, r, archive)
	}))
	t.Cleanup(srv.Close)
	s.client = srv.Client()
	s.url = srv.URL + "/archive.zip"
	return s
}

// redirectingTransport sends requests to a test server rather than to the GitHub API host baked into LoadReleases.
type redirectingTransport struct {
	host string
}

func (t *redirectingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	redirected := req.Clone(req.Context())
	redirected.URL.Scheme = "http"
	redirected.URL.Host = t.host
	redirected.Host = ""
	return http.DefaultTransport.RoundTrip(redirected)
}

// waitForMonitorQueue blocks until everything already queued on the library's monitor has been processed. The queue
// runs a single worker, so a task submitted afterwards cannot run until all of the earlier ones have finished.
func waitForMonitorQueue(lib *Library) {
	lib.lock.RLock()
	m := lib.monitor
	lib.lock.RUnlock()
	if m == nil {
		return
	}
	m.lock.RLock()
	queue := m.queue
	m.lock.RUnlock()
	if queue == nil {
		return
	}
	done := make(chan struct{})
	if queue.Submit(func() { close(done) }) {
		<-done
	}
}
