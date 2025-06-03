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
	"archive/zip"
	"bytes"
	"context"
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

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/rjeczalik/notify"
)

const releaseFile = "release.txt"

// NotifyOfLibraryChangeFunc will be called to notify of library changes.
var NotifyOfLibraryChangeFunc func()

// Library holds information about a library of data files.
type Library struct {
	ID                tid.TID  `json:"id"`
	Title             string   `json:"title,omitempty"`
	GitHubAccountName string   `json:"-"`
	AccessToken       string   `json:"access_token,omitempty"`
	RepoName          string   `json:"-"`
	PathOnDisk        string   `json:"path,omitempty"`
	Favorites         []string `json:"favorites,omitempty"`
	monitor           *monitor
	lock              sync.RWMutex
	releases          []Release
	current           string
}

// NewLibrary creates a new library.
func NewLibrary(title, githubAccountName, accessToken, repoName, pathOnDisk string) *Library {
	lib := &Library{
		ID:                tid.MustNewTID(kinds.NavigatorLibrary),
		Title:             title,
		GitHubAccountName: githubAccountName,
		AccessToken:       accessToken,
		RepoName:          repoName,
		PathOnDisk:        pathOnDisk,
	}
	lib.current = lib.VersionOnDisk()
	lib.monitor = newMonitor(lib)
	return lib
}

// Valid returns true if the library has a path on disk and a title.
func (l *Library) Valid() bool {
	return strings.TrimSpace(l.PathOnDisk) != "" && strings.TrimSpace(l.Title) != ""
}

// ConfigureForKey configures the GitHubAccountName and RepoName from the given key.
func (l *Library) ConfigureForKey(key string) {
	parts := strings.SplitN(key, "/", 2)
	l.GitHubAccountName = strings.TrimSpace(parts[0])
	l.RepoName = strings.TrimSpace(parts[1])
}

// Key returns a key representing this Library.
func (l *Library) Key() string {
	return l.GitHubAccountName + "/" + l.RepoName
}

// Path returns the path on disk to this Library, creating any necessary directories.
func (l *Library) Path() string {
	if err := os.MkdirAll(l.PathOnDisk, 0o750); err != nil {
		errs.Log(err, "path", l.PathOnDisk)
	}
	return l.PathOnDisk
}

// SetPath updates the path to the Library as well as the version.
func (l *Library) SetPath(newPath string) error {
	p, err := filepath.Abs(newPath)
	if err != nil {
		return errs.NewWithCause("unable to update library path to "+newPath, err)
	}
	if l.PathOnDisk != p {
		if l.monitor == nil {
			l.PathOnDisk = p
			l.monitor = newMonitor(l)
			l.lock.Lock()
			l.current = l.VersionOnDisk()
			l.lock.Unlock()
		} else {
			tokens := l.monitor.stop()
			l.PathOnDisk = p
			l.lock.Lock()
			l.current = l.VersionOnDisk()
			l.lock.Unlock()
			for _, token := range tokens {
				l.monitor.startWatch(token, true)
			}
		}
	}
	return nil
}

// CleanupFavorites prunes out any favorites that can no longer be read.
func (l *Library) CleanupFavorites() {
	var favs []string
	for _, one := range l.Favorites {
		path := filepath.Join(l.PathOnDisk, one)
		if fi, err := os.Stat(path); err == nil {
			if mode := fi.Mode(); (mode.IsDir() || mode.IsRegular()) && mode.Perm()&0o400 != 0 {
				favs = append(favs, one)
			}
		}
	}
	slices.Sort(favs)
	l.Favorites = favs
}

// Watch for changes in the directory tree of this library.
func (l *Library) Watch(callback func(lib *Library, fullPath string, what notify.Event), callbackOnUIThread bool) *MonitorToken {
	return l.monitor.newWatch(callback, callbackOnUIThread)
}

// StopAllWatches that were previously established.
func (l *Library) StopAllWatches() {
	l.monitor.stop()
}

// IsMaster returns true if this is the Master Library.
func (l *Library) IsMaster() bool {
	return l.GitHubAccountName == masterGitHubAccountName && l.RepoName == masterRepoName
}

// IsUser returns true if this is the User Library.
func (l *Library) IsUser() bool {
	return l.GitHubAccountName == "" && l.RepoName == userRepoName
}

// CheckForAvailableUpgrade returns releases that can be upgraded to.
func (l *Library) CheckForAvailableUpgrade(ctx context.Context, client *http.Client) {
	incompatibleFutureLibraryVersion := strconv.Itoa(jio.CurrentDataVersion + 1)
	minimumLibraryVersion := strconv.Itoa(jio.MinimumLibraryVersion)
	releases, err := LoadReleases(ctx, client, l.GitHubAccountName, l.AccessToken, l.RepoName, "",
		func(version, _ string) bool {
			return incompatibleFutureLibraryVersion == version ||
				txt.NaturalLess(version, minimumLibraryVersion, true) ||
				txt.NaturalLess(incompatibleFutureLibraryVersion, version, true)
		})
	if err != nil {
		errs.Log(errs.NewWithCause("unable to access releases for library", err), "title", l.Title, "repo", l.RepoName, "account", l.GitHubAccountName)
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
		toolbox.Call(NotifyOfLibraryChangeFunc)
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
	result := txt.NaturalCmp(l.Title, other.Title, true)
	if result == 0 {
		if result = txt.NaturalCmp(l.GitHubAccountName, other.GitHubAccountName, true); result == 0 {
			result = txt.NaturalCmp(l.RepoName, other.RepoName, true)
		}
	}
	return result
}

// VersionOnDisk returns the version of the data on disk, if it can be determined.
func (l *Library) VersionOnDisk() string {
	if l.IsUser() {
		return ""
	}
	filePath := filepath.Join(l.PathOnDisk, releaseFile)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			errs.Log(errs.NewWithCause("unable to load release info from library", err), "path", filePath)
		}
		return "0"
	}
	return strings.TrimSpace(string(bytes.SplitN(data, []byte{'\n'}, 2)[0]))
}

// Download the release onto the local disk.
func (l *Library) Download(ctx context.Context, client *http.Client, release Release) error {
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

	const unableToCreatePrefix = "unable to create "
	if err = os.MkdirAll(p, 0o750); err != nil {
		return errs.NewWithCause(unableToCreatePrefix+p, err)
	}
	var data []byte
	data, err = l.downloadRelease(ctx, client, release)
	if err != nil {
		return err
	}
	var zr *zip.Reader
	if zr, err = zip.NewReader(bytes.NewReader(data), int64(len(data))); err != nil {
		return errs.NewWithCause("unable to open archive "+release.ZipFileURL, err)
	}
	root := filepath.Clean(p)
	rootWithTrailingSep := root
	if !strings.HasSuffix(rootWithTrailingSep, string(filepath.Separator)) {
		rootWithTrailingSep += string(filepath.Separator)
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
				return errs.Newf("path outside of root is not permitted: %s", fullPath)
			}
			parent := filepath.Dir(fullPath)
			if err = os.MkdirAll(parent, 0o750); err != nil {
				return errs.NewWithCause(unableToCreatePrefix+parent, err)
			}
			if err = l.extractFile(f, fullPath); err != nil {
				return errs.NewWithCause(unableToCreatePrefix+fullPath, err)
			}
		}
	}
	f := filepath.Join(root, releaseFile)
	if err = os.WriteFile(f, []byte(release.Version+"\n"), 0o640); err != nil {
		return errs.NewWithCause(unableToCreatePrefix+f, err)
	}
	current := l.VersionOnDisk()
	l.lock.Lock()
	l.current = current
	l.lock.Unlock()
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
	return
}

func (l *Library) downloadRelease(ctx context.Context, client *http.Client, release Release) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, release.ZipFileURL, http.NoBody)
	if err != nil {
		return nil, errs.NewWithCause("unable to create request for "+release.ZipFileURL, err)
	}
	if l.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+l.AccessToken)
	}
	var rsp *http.Response
	if rsp, err = client.Do(req); err != nil {
		return nil, errs.NewWithCause("unable to connect to "+release.ZipFileURL, err)
	}
	defer xio.DiscardAndCloseIgnoringErrors(rsp.Body)
	var data []byte
	if data, err = io.ReadAll(rsp.Body); err != nil {
		return nil, errs.NewWithCause("unable to download "+release.ZipFileURL, err)
	}
	return data, nil
}
