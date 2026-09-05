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
	"bytes"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/go-git/go-billy/v6/memfs"
	"github.com/go-git/go-billy/v6/util"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xio"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/rjeczalik/notify"
)

const releaseFile = "release.txt"

// SettingsDirName is the directory within a library that holds its settings files, AncestriesDirName is the
// sub-directory of that which, by convention, holds ancestries and their name generators, and OutputTemplatesDirName is
// the directory within a library that holds its output templates.
const (
	SettingsDirName        = "Settings"
	AncestriesDirName      = "Ancestries"
	OutputTemplatesDirName = "Output Templates"
)

// defaultDownloadSizeEstimate scales the download portion of the progress bar the first time a library is updated,
// before a real size has been recorded for it. A guess that is too small shows as a bar that pauses just short of the
// end of the download; one that is too large, as a bar that jumps to it. Being a little over the size of the Master
// Library, which is by far the largest and most commonly updated of them, trades the pause for the jump. Every update
// records what it actually transferred, so this only has to be close enough to be useful once.
const defaultDownloadSizeEstimate = 32 * 1024 * 1024

var (
	libraryChangeLock     sync.Mutex
	notifyOfLibraryChange func()
	pendingLibraryChange  bool
)

// SetNotifyOfLibraryChangeFunc sets the function that will be called to notify of library changes. Passing nil removes
// the current function. If a change was reported before a function was installed, the newly installed function is
// called immediately, so that changes detected during startup aren't lost. The function may be called from any
// goroutine.
func SetNotifyOfLibraryChangeFunc(f func()) {
	libraryChangeLock.Lock()
	notifyOfLibraryChange = f
	deliver := f != nil && pendingLibraryChange
	if deliver {
		pendingLibraryChange = false
	}
	libraryChangeLock.Unlock()
	if deliver {
		xos.SafeCall(f, nil)
	}
}

// NotifyOfLibraryChange notifies of a library change. If no notification function has been installed yet, the
// notification is retained and delivered to the next function installed by SetNotifyOfLibraryChangeFunc. May be called
// from any goroutine.
func NotifyOfLibraryChange() {
	libraryChangeLock.Lock()
	f := notifyOfLibraryChange
	if f == nil {
		pendingLibraryChange = true
	}
	libraryChangeLock.Unlock()
	if f != nil {
		xos.SafeCall(f, nil)
	}
}

// libraryPersistentData holds the data that will be serialized for a Library. It is deliberately not part of the
// public API, so that the on-disk format can evolve without dragging the accessors along with it. Note that the GitHub
// account name and repository name are absent: they are carried by the key the library is filed under within a
// Libraries set, and are restored from it by ConfigureForKey().
type libraryPersistentData struct {
	ID          tid.TID  `json:"id"`
	Title       string   `json:"title,omitzero"`
	AccessToken string   `json:"access_token,omitzero"`
	PathOnDisk  string   `json:"path,omitzero"`
	Favorites   []string `json:"favorites,omitempty"`
	UseLatest   bool     `json:"use_latest,omitzero"`
}

// LibraryConfig holds the Library fields the user is permitted to edit. The ID, the path on disk and the favorites are
// deliberately absent: an ID never changes once assigned, the path must go through SetPath() so that the filesystem
// watches can be restarted, and the favorites have their own methods.
type LibraryConfig struct {
	Title             string
	GitHubAccountName string
	AccessToken       string
	RepoName          string
	UseLatest         bool
}

// LibraryData is a snapshot of a Library's state, as returned by Data(). The favorites are not included, since they are
// a list that is manipulated independently; use Favorites() for those.
type LibraryData struct {
	LibraryConfig
	ID         tid.TID
	PathOnDisk string
}

// Library holds information about a library of data files. Every field is private and reachable only through the
// accessors, since a library is read by background goroutines (update checks) while the UI thread may be modifying it.
type Library struct {
	monitor           *monitor
	data              libraryPersistentData
	gitHubAccountName string
	repoName          string
	releases          []Release
	current           string
	check             *libraryCheck // the update check in flight; nil while none is
	lock              sync.RWMutex
	checkSeq          int  // see libraryCheck.seq
	checked           bool // an update check has completed since the repository was last changed
}

// libraryCheck describes an update check in flight, so that a caller arriving while it runs can wait on it and then
// tell whether what it produced serves them or they need to make a check of their own.
type libraryCheck struct {
	done chan struct{} // closed once the check has finished and answered has been set
	// seq is the checkSeq the check was made for. The check is discarded on landing if that has moved on, which it does
	// whenever the repository or the content on disk is replaced.
	seq      int
	answered bool // set under the library lock before done is closed; see CheckForAvailableUpgrade
}

// NewLibrary creates a new library.
func NewLibrary(title, githubAccountName, accessToken, repoName, pathOnDisk string) *Library {
	lib := &Library{
		data: libraryPersistentData{
			ID:          tid.MustNewTID(kinds.NavigatorLibrary),
			Title:       title,
			AccessToken: accessToken,
			PathOnDisk:  pathOnDisk,
		},
		gitHubAccountName: githubAccountName,
		repoName:          repoName,
	}
	lib.refreshVersionOnDisk()
	return lib
}

// MarshalJSONTo implements json.MarshalerTo.
func (l *Library) MarshalJSONTo(enc *jsontext.Encoder) error {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return json.MarshalEncode(enc, &l.data)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (l *Library) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.data = libraryPersistentData{}
	return json.UnmarshalDecode(dec, &l.data)
}

// Data returns a snapshot of this library's state. The returned value is a copy, so modifying it has no effect on the
// library; use Configure(), SetID() or SetPath() to make changes.
func (l *Library) Data() LibraryData {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.snapshot()
}

// snapshot builds the LibraryData for the current state. The library lock must be held when calling this.
func (l *Library) snapshot() LibraryData {
	return LibraryData{
		LibraryConfig: l.config(),
		ID:            l.data.ID,
		PathOnDisk:    l.data.PathOnDisk,
	}
}

// Config returns the portion of this library's state that the user is permitted to edit.
func (l *Library) Config() LibraryConfig {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.config()
}

// config extracts the user-editable portion of this library's state. The library lock must be held when calling this.
func (l *Library) config() LibraryConfig {
	return LibraryConfig{
		Title:             l.data.Title,
		GitHubAccountName: l.gitHubAccountName,
		AccessToken:       l.data.AccessToken,
		RepoName:          l.repoName,
		UseLatest:         l.data.UseLatest,
	}
}

// Configure applies the given configuration as a single atomic operation. Note that this may change the value returned
// by Key(). A change to the repository being followed -- the account, the repository name, or whether to follow its
// latest commit rather than its releases -- discards what any earlier check found, since that described the old one:
// the releases are dropped, the library is left needing a check (see NeedsUpgradeCheck), a check still in flight for
// the old repository is discarded when it finishes, and the version on disk is re-read, since whether the library is
// the User Library can change along with the key.
func (l *Library) Configure(config LibraryConfig) {
	l.lock.Lock()
	l.data.Title = config.Title
	l.data.AccessToken = config.AccessToken
	repoChanged := l.setRepository(config.GitHubAccountName, config.RepoName, config.UseLatest)
	l.lock.Unlock()
	if repoChanged {
		l.refreshVersionOnDisk()
	}
}

// setRepository points the library at the given repository and reports whether that differs from the one it was
// following. A change discards what any earlier check found, as described for Configure(), apart from the re-read of
// the version on disk, which the caller must do once the lock has been released, since it reads a file. Dropping the
// releases and bumping the sequence number in the same critical section as the change of repository is what keeps a
// check from landing in between with the old repository's answer. The library lock must be held when calling this.
func (l *Library) setRepository(gitHubAccountName, repoName string, useLatest bool) bool {
	if l.gitHubAccountName == gitHubAccountName && l.repoName == repoName && l.data.UseLatest == useLatest {
		return false
	}
	l.gitHubAccountName = gitHubAccountName
	l.repoName = repoName
	l.data.UseLatest = useLatest
	l.releases = nil
	l.checked = false
	l.checkSeq++
	return true
}

// SetID sets the unique ID of this library. Should only be used for a library that doesn't yet have one.
func (l *Library) SetID(id tid.TID) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.data.ID = id
}

// Valid returns true if the library has a path on disk and a title. A file may hold a null in place of a library, so
// this must be checked before dereferencing one that came from a file.
func (l *Library) Valid() bool {
	if l == nil {
		return false
	}
	l.lock.RLock()
	defer l.lock.RUnlock()
	return strings.TrimSpace(l.data.PathOnDisk) != "" && strings.TrimSpace(l.data.Title) != ""
}

// ConfigureForKey configures the GitHub account name and repository name from the given key, which is the form they
// take when a Libraries set is saved. A change of repository is treated exactly as Configure() treats one. A freshly
// loaded library follows no repository until this is called, so for it the key is always a change, and the re-read of
// the version on disk that comes with the change is what fills the version in: it isn't part of what was saved, and an
// update check is the only other thing that sets it -- which, with the periodic checks turned off, may never run.
func (l *Library) ConfigureForKey(key string) error {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 {
		return errs.Newf("invalid library key: %s", key)
	}
	repoName := strings.TrimSpace(parts[1])
	if repoName == "" {
		return errs.Newf("invalid library key: %s", key)
	}
	l.lock.Lock()
	repoChanged := l.setRepository(strings.TrimSpace(parts[0]), repoName, l.data.UseLatest)
	l.lock.Unlock()
	if repoChanged {
		l.refreshVersionOnDisk()
	}
	return nil
}

// Key returns a key representing this Library.
func (l *Library) Key() string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.gitHubAccountName + "/" + l.repoName
}

// Path returns the path on disk to this Library, creating any necessary directories.
func (l *Library) Path() string {
	l.lock.RLock()
	p := l.data.PathOnDisk
	l.lock.RUnlock()
	if err := os.MkdirAll(p, 0o750); err != nil {
		errs.Log(err, "path", p)
	}
	return p
}

// AncestriesPath returns the path on disk to this Library's ancestries directory, creating it if necessary.
func (l *Library) AncestriesPath() string {
	p := filepath.Join(l.Path(), SettingsDirName, AncestriesDirName)
	if err := os.MkdirAll(p, 0o750); err != nil {
		errs.Log(err, "path", p)
	}
	return p
}

// SetPath updates the path to the Library as well as the version.
func (l *Library) SetPath(newPath string) error {
	p, err := filepath.Abs(newPath)
	if err != nil {
		return errs.NewWithCause("unable to update library path to "+newPath, err)
	}
	l.lock.RLock()
	unchanged := l.data.PathOnDisk == p
	l.lock.RUnlock()
	if unchanged {
		return nil
	}
	// The monitor must not be touched while the library lock is held, since establishing a watch calls back into
	// Path(), which needs the library lock itself.
	m := l.obtainMonitor()
	tokens := m.stop()
	l.lock.Lock()
	l.data.PathOnDisk = p
	l.lock.Unlock()
	l.refreshVersionOnDisk()
	for _, token := range tokens {
		m.startWatch(token, true)
	}
	return nil
}

// Favorites returns the paths, relative to the library's path on disk, that have been marked as favorites.
func (l *Library) Favorites() []string {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return slices.Clone(l.data.Favorites)
}

// ToggleFavorite marks the given path, relative to the library's path on disk, as a favorite if it isn't already one,
// or removes it from the favorites if it is.
func (l *Library) ToggleFavorite(relativePath string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if i := slices.Index(l.data.Favorites, relativePath); i != -1 {
		l.data.Favorites = slices.Delete(l.data.Favorites, i, i+1)
	} else {
		l.data.Favorites = append(l.data.Favorites, relativePath)
	}
}

// RenameFavorite replaces the favorite with the old path with one using the new path. Does nothing if the old path
// isn't a favorite. Both paths are relative to the library's path on disk.
func (l *Library) RenameFavorite(oldRelativePath, newRelativePath string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	if i := slices.Index(l.data.Favorites, oldRelativePath); i != -1 {
		l.data.Favorites = slices.Delete(l.data.Favorites, i, i+1)
		l.data.Favorites = append(l.data.Favorites, newRelativePath)
	}
}

// CleanupFavorites prunes out any favorites that can no longer be read.
func (l *Library) CleanupFavorites() {
	l.lock.Lock()
	defer l.lock.Unlock()
	var favs []string
	for _, one := range l.data.Favorites {
		p := filepath.Join(l.data.PathOnDisk, one)
		if fi, err := os.Stat(p); err == nil {
			if mode := fi.Mode(); (mode.IsDir() || mode.IsRegular()) && mode.Perm()&0o400 != 0 {
				favs = append(favs, one)
			}
		}
	}
	slices.Sort(favs)
	l.data.Favorites = favs
}

// Watch for changes in the directory tree of this library. Each change is reported with the full path of what changed,
// named the way this library names it: beneath Path(), and beneath any symlinked directory registered with
// MonitorToken.AddSubPath, rather than wherever those resolve to on disk. A path built from Path() and a directory
// entry within it can therefore be compared directly with what the callback receives.
func (l *Library) Watch(callback func(lib *Library, fullPath string, what notify.Event), callbackOnUIThread bool) *MonitorToken {
	return l.obtainMonitor().newWatch(callback, callbackOnUIThread)
}

// StopAllWatches that were previously established.
func (l *Library) StopAllWatches() {
	l.obtainMonitor().stop()
}

// obtainMonitor returns the monitor for this library, creating it if needed. The library lock must not be held when
// calling this, nor may it be held while calling into the returned monitor.
func (l *Library) obtainMonitor() *monitor {
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.monitor == nil {
		l.monitor = newMonitor(l)
	}
	return l.monitor
}

// IsMasterLibraryKey returns true if the given GitHub account name and repository name identify the Master Library.
func IsMasterLibraryKey(githubAccountName, repoName string) bool {
	return githubAccountName == masterGitHubAccountName && repoName == masterRepoName
}

// IsUserLibraryKey returns true if the given GitHub account name and repository name identify the User Library.
func IsUserLibraryKey(githubAccountName, repoName string) bool {
	return githubAccountName == "" && repoName == userRepoName
}

// IsMaster returns true if this is the Master Library.
func (l *Library) IsMaster() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return IsMasterLibraryKey(l.gitHubAccountName, l.repoName)
}

// IsUser returns true if this is the User Library.
func (l *Library) IsUser() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return IsUserLibraryKey(l.gitHubAccountName, l.repoName)
}

// CheckForAvailableUpgrade retrieves the releases that can be upgraded to, recording them for AvailableReleases(). Only
// one check of a library runs at a time: a call made while another is in flight -- the Library Explorer asking on the
// user's behalf while the launch-time or periodic check is still under way, say -- waits for that one to finish rather
// than making a second request for the same answer, and returns once it has, or once ctx is done. The check that was
// waited on serves the caller only if it answered for the repository the library follows now, whether with its
// releases or with a failure to reach it, which asking again at once would only repeat. One that was discarded because
// the repository was reconfigured while it ran -- its answer describing a repository the library no longer follows --
// or that was cut short by its own caller's context rather than by this one's, leaves the need it was waited on for
// unmet, so the caller then makes a check of its own. Without that, a check made from the settings dialog or the
// Library Explorer right after a repository change would wait on the doomed launch-time check and come back with
// nothing, leaving the library unchecked until the next scheduled check.
func (l *Library) CheckForAvailableUpgrade(ctx context.Context, client *http.Client) {
	for ctx.Err() == nil {
		l.lock.Lock()
		if check := l.check; check != nil {
			l.lock.Unlock()
			select {
			case <-check.done:
			case <-ctx.Done():
				return
			}
			l.lock.RLock()
			served := check.answered && check.seq == l.checkSeq
			l.lock.RUnlock()
			if served {
				return
			}
			continue
		}
		check := &libraryCheck{done: make(chan struct{}), seq: l.checkSeq}
		l.check = check
		// The configuration is captured along with the sequence number so that a check always lands, or is discarded,
		// as a check of the repository it actually asked. Taken separately, a reconfiguration between the two would
		// have the check ask the new repository and then throw the answer away as stale.
		data := l.snapshot()
		l.lock.Unlock()
		l.performCheck(ctx, client, check, &data)
		return
	}
}

// performCheck makes the request for the given check, which the caller has just registered as the one in flight, and
// records what it finds. Whatever happens, the check is unregistered and its waiters released before this returns.
func (l *Library) performCheck(ctx context.Context, client *http.Client, check *libraryCheck, data *LibraryData) {
	answered := false
	defer func() {
		l.lock.Lock()
		l.check = nil
		check.answered = answered
		l.lock.Unlock()
		close(check.done)
	}()
	incompatibleFutureLibraryVersion := strconv.Itoa(jio.CurrentDataVersion + 1)
	minimumLibraryVersion := strconv.Itoa(jio.MinimumLibraryVersion)
	releases, err := LoadReleases(ctx, client, data.GitHubAccountName, data.AccessToken, data.RepoName, "",
		func(version, _ string) bool {
			return incompatibleFutureLibraryVersion == version ||
				xstrings.NaturalLess(version, minimumLibraryVersion, true) ||
				xstrings.NaturalLess(incompatibleFutureLibraryVersion, version, true)
		}, data.UseLatest)
	if err != nil {
		// A failure to reach the repository is an answer of sorts: a caller waiting on this check gains nothing by
		// asking again at once. Being cut short by the context is not, since that context belonged to whoever started
		// the check, and a waiter with a live context of its own still wants the answer.
		answered = ctx.Err() == nil
		errs.Log(errs.NewWithCause("unable to access releases for library", err), "title", data.Title, "repo",
			data.RepoName, "account", data.GitHubAccountName)
		return
	}
	current := l.VersionOnDisk()
	lastRelease := ""
	if len(releases) != 0 {
		lastRelease = releases[0].Version
	}
	l.lock.Lock()
	if check.seq != l.checkSeq {
		l.lock.Unlock()
		return // The repository was changed while this check ran, so its answer is about the wrong one
	}
	answered = true
	prevCurrent := l.current
	prevLastRelease := ""
	if len(l.releases) != 0 {
		prevLastRelease = l.releases[0].Version
	}
	firstCheck := !l.checked
	l.checked = true
	l.releases = releases
	l.current = current
	l.lock.Unlock()
	// A notification reloads the entire library tree, which restarts the filesystem watches, drops what has been cached
	// and disturbs anything in progress, so one is sent only when this check turned up something the previous check
	// didn't: an update that has just become available, a library whose content changed on disk outside of the app, or
	// an update that was on offer and no longer is -- the release having been withdrawn, or the library having been
	// brought up to date outside of the app -- which must take the indicator down. The first check of a library also
	// announces an update that was already pending, which is what raises the indicator at startup. A library with no
	// releases to compare against -- a local one, or a repo whose releases were all rejected as incompatible -- has
	// nothing to announce, even though the "0" that stands in for an unknown version on disk differs from the empty
	// version of a release that isn't there.
	prevUpdateAvailable := prevLastRelease != "" && prevCurrent != prevLastRelease
	updateAvailable := lastRelease != "" && current != lastRelease
	if updateAvailable && (firstCheck || prevCurrent != current || prevLastRelease != lastRelease) ||
		prevUpdateAvailable && !updateAvailable {
		NotifyOfLibraryChange()
	}
}

// AvailableReleases returns the available releases.
func (l *Library) AvailableReleases() (current string, releases []Release) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.current, l.releases
}

// NeedsUpgradeCheck returns true if the library is backed by a GitHub repository and no update check of it has
// completed since it was pointed at that repository, either because none has been made -- with the periodic checks
// turned off, none is -- or because every one made so far failed to reach the repository. A library that isn't backed
// by a repository has nothing to check. A check that is still in flight doesn't yet satisfy the need; calling
// CheckForAvailableUpgrade() waits for it, and makes a check of its own if that one turns out not to have answered.
func (l *Library) NeedsUpgradeCheck() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return !l.checked && l.gitHubAccountName != "" && l.repoName != ""
}

// Compare the two libraries for sorting purposes.
func (l *Library) Compare(other *Library) int {
	if l.IsUser() {
		if other.IsUser() {
			return 0
		}
		return -1
	}
	if other.IsUser() {
		return 1
	}
	if l.IsMaster() {
		if other.IsMaster() {
			return 0
		}
		return -1
	}
	if other.IsMaster() {
		return 1
	}
	data := l.Data()
	otherData := other.Data()
	result := xstrings.NaturalCmp(data.Title, otherData.Title, true)
	if result == 0 {
		if result = xstrings.NaturalCmp(data.GitHubAccountName, otherData.GitHubAccountName, true); result == 0 {
			result = xstrings.NaturalCmp(data.RepoName, otherData.RepoName, true)
		}
	}
	return result
}

// VersionOnDisk returns the version of the data on disk, if it can be determined.
func (l *Library) VersionOnDisk() string {
	l.lock.RLock()
	isUser := IsUserLibraryKey(l.gitHubAccountName, l.repoName)
	pathOnDisk := l.data.PathOnDisk
	l.lock.RUnlock()
	if isUser {
		return ""
	}
	filePath := filepath.Join(pathOnDisk, releaseFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			errs.Log(errs.NewWithCause("unable to load release info from library", err), "path", filePath)
		}
		return "0"
	}
	return strings.TrimSpace(string(bytes.SplitN(data, []byte{'\n'}, 2)[0]))
}

// refreshVersionOnDisk updates the cached "current" version from what is present on disk and discards any update check
// still in flight, since the version that check reads -- which it does before taking the lock -- may be of what was
// there before. Everything that replaces what is on disk must call this once the replacement is in place: a change of
// path, a download, whether it succeeded or not, and a change of repository, which can change whether the library is
// the User Library. Without the discard, a check that read the old directory's version, or the "0" that stands in for
// a directory moved aside during a download, could land after the refresh and put that back as the current version,
// showing an update for a library that is in fact up to date until the next scheduled check corrected it. A discarded
// check leaves the library as it found it -- whether it had been checked is unchanged -- and whoever waited on it asks
// again; see CheckForAvailableUpgrade. The library lock must not be held when calling this.
func (l *Library) refreshVersionOnDisk() {
	current := l.VersionOnDisk()
	l.lock.Lock()
	l.current = current
	l.checkSeq++
	l.lock.Unlock()
}

// Download the release onto the local disk. progress, which may be nil, is called as the work proceeds; see
// LibraryUpdateProgress for what it receives and what is required of it.
func (l *Library) Download(ctx context.Context, client *http.Client, release *Release, progress LibraryUpdateProgress) error {
	libData := l.Data() // Not named "data", since the byte buffers below already use that name
	p := l.Path()
	// What the last download transferred is the only basis there is for scaling the download portion of the bar, and it
	// lives in the directory that is about to be moved aside, so it has to be read now.
	estimate := l.recordedDownloadSize()
	if estimate <= 0 {
		estimate = defaultDownloadSizeEstimate
	}
	// Everything to do with progress goes through the one lock, because the git transport reads its responses on
	// goroutines of go-git's choosing. Counting a chunk and reporting where that leaves things as a single unit is what
	// keeps two of those goroutines from delivering their progress out of order, which would show as a bar that jumps
	// backwards, and it means the caller's function is never entered twice at once.
	var lock sync.Mutex
	var received int64
	report := func(phase LibraryUpdatePhase, fraction float64) {
		lock.Lock()
		defer lock.Unlock()
		if progress != nil {
			progress(phase, fraction)
		}
	}
	countReceived := func(n int64) {
		lock.Lock()
		defer lock.Unlock()
		received += n
		if progress != nil {
			progress(LibraryUpdateDownloading, estimatedFraction(received, estimate))
		}
	}
	transferred := func() int64 {
		lock.Lock()
		defer lock.Unlock()
		return received
	}
	report(LibraryUpdateDownloading, 0)
	tmpDir, err := os.MkdirTemp(filepath.Dir(p), filepath.Base(p)+"_*")
	if err != nil {
		return errs.NewWithCause("unable to create temporary directory", err)
	}
	if err = os.Remove(tmpDir); err != nil {
		return errs.NewWithCause("unable to remove temporary directory:\n"+tmpDir, err)
	}
	if err = os.Rename(p, tmpDir); err != nil {
		return errs.NewWithCause("unable to move old directory aside:\n"+p+"\n"+tmpDir, err)
	}
	success := false
	defer func() {
		if success {
			if err = os.RemoveAll(tmpDir); err != nil {
				errs.Log(errs.NewWithCause("unable to remove the old data", err), "dir", tmpDir)
			}
		} else {
			if err = os.RemoveAll(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
				errs.Log(errs.NewWithCause("unable to remove the failed download data", err), "dir", p)
			}
			if err = os.Rename(tmpDir, p); err != nil {
				errs.Log(errs.NewWithCause("unable to move the old directory back into place", err), "old", tmpDir, "new", p)
			}
		}
		// Whether the new content is now in place or the old has been put back, what is on disk was replaced while an
		// update check may have been reading it, so the cached version is re-read here and any such check discarded.
		// That matters as much for a download that failed as for one that succeeded: the check would otherwise land
		// with the "0" it read while the directory was aside, and the update flow, which waits on the check either
		// way, would be told that it was served.
		l.refreshVersionOnDisk()
	}()
	if err = os.MkdirAll(p, 0o750); err != nil {
		return errs.NewWithCause("unable to create directory "+p, err)
	}
	root := filepath.Clean(p)
	rootWithTrailingSep := root
	if !strings.HasSuffix(rootWithTrailingSep, string(filepath.Separator)) {
		rootWithTrailingSep += string(filepath.Separator)
	}
	if libData.UseLatest {
		mfs := memfs.New()
		var hash string
		if hash, err = downloadLatestCommit(ctx,
			"https://github.com/"+libData.GitHubAccountName+"/"+libData.RepoName+".git",
			libData.AccessToken, mfs, countReceived); err != nil {
			return err
		}
		// use hash that was actually downloaded, in case a commit occurred between our original check and the download
		release.Version = hash
		// The clone lives in memory, so the pass that totals up the work costs nothing worth avoiding, and having the
		// total is what lets the second pass report real progress rather than a count of files against nothing.
		var total int64
		if err = util.Walk(mfs, "Library", func(_ string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !info.IsDir() {
				total += info.Size()
			}
			return nil
		}); err != nil {
			return err
		}
		var written int64
		report(LibraryUpdateInstalling, 0)
		if err = util.Walk(mfs, "Library", func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			// Writing the content out is the half of the update the context would otherwise have no say over, and it is
			// long enough to be worth interrupting, so each file is a chance to stop.
			if ctxErr := ctx.Err(); ctxErr != nil {
				return errs.Wrap(ctxErr)
			}
			parts := strings.SplitN(filepath.ToSlash(path), "/", 2)
			if len(parts) != 2 {
				return nil
			}
			if !strings.EqualFold("Library", parts[0]) {
				return nil
			}
			fullPath := filepath.Join(p, parts[1])
			if !strings.HasPrefix(fullPath, rootWithTrailingSep) {
				return errs.Newf("path outside of destination directory is not permitted: %s", fullPath)
			}
			if info.IsDir() {
				return os.Mkdir(fullPath, 0o750)
			}
			var data []byte
			if data, walkErr = util.ReadFile(mfs, path); walkErr != nil {
				return errs.NewWithCause("unable to read "+fullPath, walkErr)
			}
			if walkErr = os.WriteFile(fullPath, data, 0o640); walkErr != nil {
				return errs.NewWithCause("unable to write "+fullPath, walkErr)
			}
			written += info.Size()
			report(LibraryUpdateInstalling, exactFraction(written, total))
			return nil
		}); err != nil {
			return err
		}
	} else {
		var data []byte
		data, err = l.downloadRelease(ctx, client, release, countReceived)
		if err != nil {
			return err
		}
		var zr *zip.Reader
		if zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data))); err != nil {
			return errs.NewWithCause("unable to open archive "+release.ZipFileURL, err)
		}
		entries, total := libraryArchiveContent(zr)
		var written int64
		report(LibraryUpdateInstalling, 0)
		for _, entry := range entries {
			// Unpacking is the half of the update the context would otherwise have no say over, and it is long enough
			// to be worth interrupting, so each file is a chance to stop.
			if err = ctx.Err(); err != nil {
				return errs.Wrap(err)
			}
			fullPath := filepath.Join(root, entry.path)
			if !strings.HasPrefix(fullPath, rootWithTrailingSep) {
				return errs.Newf("path outside of destination directory is not permitted: %s", fullPath)
			}
			parent := filepath.Dir(fullPath)
			if err = os.MkdirAll(parent, 0o750); err != nil {
				return errs.NewWithCause("unable to create directory "+parent, err)
			}
			if err = l.extractFile(entry.file, fullPath); err != nil {
				return errs.NewWithCause("unable to create file "+fullPath, err)
			}
			written += int64(entry.file.UncompressedSize64)
			report(LibraryUpdateInstalling, exactFraction(written, total))
		}
	}
	// Both paths above report their progress as they go, but neither can be relied upon to have finished on a whole
	// number of anything -- an archive holding no library content reports nothing at all -- so the phase is closed out
	// here rather than leaving a bar that stops short of its end.
	report(LibraryUpdateInstalling, 1)
	// The size of what was transferred is recorded with the version so that the next update has something better than a
	// guess to scale its progress bar against. It goes on a second line, which both VersionOnDisk() and versions of GCS
	// that predate it ignore, since they read only the first.
	f := filepath.Join(root, releaseFile)
	if err = os.WriteFile(f, []byte(release.Version+"\n"+strconv.FormatInt(transferred(), 10)+"\n"), 0o640); err != nil {
		return errs.NewWithCause("unable to write version file "+f, err)
	}
	success = true
	return nil
}

func (l *Library) extractFile(f *zip.File, dst string) (err error) {
	var r io.ReadCloser
	if r, err = f.Open(); err != nil {
		return errs.Wrap(err)
	}
	defer xio.CloseIgnoringErrors(r)
	var file *os.File
	if file, err = os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.FileInfo().Mode().Perm()&0o750); err != nil {
		return errs.Wrap(err)
	}
	if _, err = io.Copy(file, r); err != nil {
		err = errs.Wrap(err)
	}
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = errs.Wrap(closeErr)
	}
	return err
}

// libraryArchiveEntry is a file from a downloaded archive that belongs to the library, paired with its path relative to
// the library's directory.
type libraryArchiveEntry struct {
	file *zip.File
	path string
}

// libraryArchiveContent picks out the files in the archive that make up the library's content and totals the space they
// will take up once expanded. GitHub's source archives hold everything beneath a single top-level directory named for
// the repository and the commit, so what is wanted is the normal files below that directory's "Library" folder.
func libraryArchiveContent(zr *zip.Reader) (entries []libraryArchiveEntry, total int64) {
	for _, f := range zr.File {
		if f.FileInfo().Mode()&os.ModeType != 0 { // normal files only
			continue
		}
		parts := strings.SplitN(filepath.ToSlash(f.Name), "/", 3)
		if len(parts) != 3 || !strings.EqualFold("Library", parts[1]) {
			continue
		}
		entries = append(entries, libraryArchiveEntry{file: f, path: parts[2]})
		total += int64(f.UncompressedSize64)
	}
	return entries, total
}

// recordedDownloadSize returns the number of bytes the last download of this library transferred, or 0 if that isn't
// known. Download() writes it on the second line of the release file. It only ever scales a progress bar, so a file
// that doesn't hold what is expected is treated as "not known" rather than as an error.
func (l *Library) recordedDownloadSize() int64 {
	l.lock.RLock()
	pathOnDisk := l.data.PathOnDisk
	l.lock.RUnlock()
	data, err := os.ReadFile(filepath.Join(pathOnDisk, releaseFile))
	if err != nil {
		return 0
	}
	lines := bytes.SplitN(data, []byte{'\n'}, 3)
	if len(lines) < 2 {
		return 0
	}
	size, err := strconv.ParseInt(strings.TrimSpace(string(lines[1])), 10, 64)
	if err != nil || size < 0 {
		return 0
	}
	return size
}

func (l *Library) downloadRelease(ctx context.Context, client *http.Client, release *Release, received func(n int64)) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, release.ZipFileURL, http.NoBody)
	if err != nil {
		return nil, errs.NewWithCause("unable to create request for "+release.ZipFileURL, err)
	}
	if accessToken := l.Data().AccessToken; accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	var rsp *http.Response
	if rsp, err = client.Do(req); err != nil {
		return nil, errs.NewWithCause("unable to connect to "+release.ZipFileURL, err)
	}
	defer xio.DiscardAndCloseIgnoringErrors(rsp.Body)
	if rsp.StatusCode < 200 || rsp.StatusCode > 299 {
		return nil, errs.New("unexpected response code from " + release.ZipFileURL + " -> " + rsp.Status)
	}
	// The body is read through a counter rather than with io.ReadAll() so that the caller can follow the download as it
	// arrives. GitHub generates these archives on the fly and sends them without a Content-Length, so counting what has
	// turned up is all there is to go on.
	var buffer bytes.Buffer
	if _, err = buffer.ReadFrom(&countingReader{r: rsp.Body, received: received}); err != nil {
		return nil, errs.NewWithCause("unable to download "+release.ZipFileURL, err)
	}
	return buffer.Bytes(), nil
}
