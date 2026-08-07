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
	"encoding/json/v2"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

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
	lib := NewLibrary("Extra", "someone", "secret", "repo", dir)
	lib.ToggleFavorite("b.gcs")
	lib.ToggleFavorite("a.gcs")
	libs := Libraries(map[string]*Library{lib.Key(): lib})
	data, err := json.Marshal(&libs)
	c.NoError(err)

	// Pin the serialized shape. In particular, the GitHub account and repository names must not appear in the library
	// object itself, since they travel in the key it is filed under.
	var raw map[string]map[string]any
	c.NoError(json.Unmarshal(data, &raw))
	obj, found := raw["someone/repo"]
	c.True(found)
	c.Equal([]string{"access_token", "favorites", "id", "path", "title"}, slices.Sorted(maps.Keys(obj)))

	var loaded Libraries
	c.NoError(json.Unmarshal(data, &loaded))
	restored, ok := loaded["someone/repo"]
	c.True(ok)
	c.NotNil(restored)
	c.Equal(lib.Data(), restored.Data())
	c.Equal(lib.Favorites(), restored.Favorites())
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
		data, err := lib.downloadRelease(t.Context(), srv.Client(), release)
		c.HasError(err, "status %d", code)
		c.Nil(data, "status %d", code)
		c.Contains(err.Error(), strconv.Itoa(code), "status %d", code)
		c.Contains(err.Error(), srv.URL, "status %d", code)
	}

	// A successful response must still hand back the body unchanged.
	status = http.StatusOK
	data, err := lib.downloadRelease(t.Context(), srv.Client(), release)
	c.NoError(err)
	c.Equal([]byte("not a zip file"), data)
}

// TestCheckForAvailableUpgradeNotifiesLateInstalledFunc verifies that an update check completing before the UI has
// installed a notification function still delivers that notification once one is installed. The update checks are
// started by Start() before the workspace (and therefore the navigator) exists, so a fast check would otherwise be
// silently dropped, leaving no "update available" indicator until some unrelated reload happened to occur.
func TestCheckForAvailableUpgradeNotifiesLateInstalledFunc(t *testing.T) {
	c := check.New(t)
	resetLibraryChangeNotification()
	defer resetLibraryChangeNotification()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `[{"tag_name":"v5","body":"notes","zipball_url":"http://127.0.0.1/z.zip"}]`)
	}))
	defer srv.Close()
	client := &http.Client{Transport: &redirectingTransport{host: srv.Listener.Addr().String()}}

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

	// With a function installed, a check reporting an upgrade notifies directly.
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), later.Load())

	// A check that finds nothing newer than what is on disk must not notify.
	c.NoError(os.WriteFile(filepath.Join(lib.Data().PathOnDisk, releaseFile), []byte("5\n"), 0o600))
	lib.CheckForAvailableUpgrade(t.Context(), client)
	c.Equal(int64(1), later.Load())
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
