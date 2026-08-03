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

// NotifyOfLibraryChangeFunc will be called to notify of library changes.
var NotifyOfLibraryChangeFunc func()

// libraryPersistentData holds the data that will be serialized for a Library. It is deliberately not part of the
// public API, so that the on-disk format can evolve without dragging the accessors along with it. Note that the GitHub
// account name and repository name are absent: they are carried by the key the library is filed under within a
// Libraries set, and are restored from it by ConfigureForKey().
type libraryPersistentData struct {
	ID          tid.TID  `json:"id"`
	Title       string   `json:"title,omitzero"`
	AccessToken string   `json:"access_token,omitzero"`
	PathOnDisk  string   `json:"path,omitzero"`
	Favorites   []string `json:"favorites,omitzero"`
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
	lock              sync.RWMutex
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
// by Key().
func (l *Library) Configure(config LibraryConfig) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.data.Title = config.Title
	l.data.AccessToken = config.AccessToken
	l.data.UseLatest = config.UseLatest
	l.gitHubAccountName = config.GitHubAccountName
	l.repoName = config.RepoName
}

// SetID sets the unique ID of this library. Should only be used for a library that doesn't yet have one.
func (l *Library) SetID(id tid.TID) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.data.ID = id
}

// Valid returns true if the library has a path on disk and a title.
func (l *Library) Valid() bool {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return strings.TrimSpace(l.data.PathOnDisk) != "" && strings.TrimSpace(l.data.Title) != ""
}

// ConfigureForKey configures the GitHub account name and repository name from the given key.
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
	defer l.lock.Unlock()
	l.gitHubAccountName = strings.TrimSpace(parts[0])
	l.repoName = repoName
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

// Watch for changes in the directory tree of this library.
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

// CheckForAvailableUpgrade returns releases that can be upgraded to.
func (l *Library) CheckForAvailableUpgrade(ctx context.Context, client *http.Client) {
	data := l.Data()
	incompatibleFutureLibraryVersion := strconv.Itoa(jio.CurrentDataVersion + 1)
	minimumLibraryVersion := strconv.Itoa(jio.MinimumLibraryVersion)
	releases, err := LoadReleases(ctx, client, data.GitHubAccountName, data.AccessToken, data.RepoName, "",
		func(version, _ string) bool {
			return incompatibleFutureLibraryVersion == version ||
				xstrings.NaturalLess(version, minimumLibraryVersion, true) ||
				xstrings.NaturalLess(incompatibleFutureLibraryVersion, version, true)
		}, data.UseLatest)
	if err != nil {
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
	l.releases = releases
	l.current = current
	l.lock.Unlock()
	if current != lastRelease && NotifyOfLibraryChangeFunc != nil {
		xos.SafeCall(NotifyOfLibraryChangeFunc, nil)
	}
}

// AvailableReleases returns the available releases.
func (l *Library) AvailableReleases() (current string, releases []Release) {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.current, l.releases
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

// refreshVersionOnDisk updates the cached "current" version from what is present on disk. The library lock must not be
// held when calling this.
func (l *Library) refreshVersionOnDisk() {
	current := l.VersionOnDisk()
	l.lock.Lock()
	l.current = current
	l.lock.Unlock()
}

// Download the release onto the local disk.
func (l *Library) Download(ctx context.Context, client *http.Client, release Release) error {
	libData := l.Data() // Not named "data", since the byte buffers below already use that name
	p := l.Path()
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
			libData.AccessToken, mfs); err != nil {
			return err
		}
		// use hash that was actually downloaded, in case a commit occurred between our original check and the download
		release.Version = hash
		if err = util.Walk(mfs, "Library", func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
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
			return nil
		}); err != nil {
			return err
		}
	} else {
		var data []byte
		data, err = l.downloadRelease(ctx, client, release)
		if err != nil {
			return err
		}
		var zr *zip.Reader
		if zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data))); err != nil {
			return errs.NewWithCause("unable to open archive "+release.ZipFileURL, err)
		}
		for _, f := range zr.File {
			fi := f.FileInfo()
			mode := fi.Mode()
			if mode&os.ModeType == 0 { // normal files only
				parts := strings.SplitN(filepath.ToSlash(f.Name), "/", 3)
				if len(parts) != 3 {
					continue
				}
				if !strings.EqualFold("Library", parts[1]) {
					continue
				}
				fullPath := filepath.Join(root, parts[2])
				if !strings.HasPrefix(fullPath, rootWithTrailingSep) {
					return errs.Newf("path outside of destination directory is not permitted: %s", fullPath)
				}
				parent := filepath.Dir(fullPath)
				if err = os.MkdirAll(parent, 0o750); err != nil {
					return errs.NewWithCause("unable to create directory "+parent, err)
				}
				if err = l.extractFile(f, fullPath); err != nil {
					return errs.NewWithCause("unable to create file "+fullPath, err)
				}
			}
		}
	}
	f := filepath.Join(root, releaseFile)
	if err = os.WriteFile(f, []byte(release.Version+"\n"), 0o640); err != nil {
		return errs.NewWithCause("unable to write version file "+f, err)
	}
	l.refreshVersionOnDisk()
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

func (l *Library) downloadRelease(ctx context.Context, client *http.Client, release Release) ([]byte, error) {
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
	var data []byte
	if data, err = io.ReadAll(rsp.Body); err != nil {
		return nil, errs.NewWithCause("unable to download "+release.ZipFileURL, err)
	}
	return data, nil
}
