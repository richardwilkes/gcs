/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package library

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/rjeczalik/notify"
)

const releaseFile = "release.txt"

// Library holds information about a library of data files.
type Library struct {
	Title             string `json:"title,omitempty"`
	GitHubAccountName string `json:"-"`
	RepoName          string `json:"-"`
	PathOnDisk        string `json:"path,omitempty"`
	LastSeen          string `json:"last_seen,omitempty"`
	monitor           *monitor
	lock              sync.RWMutex
	upgrade           *Release
}

// NewLibrary creates a new library.
func NewLibrary(title, githubAccountName, repoName, pathOnDisk string) *Library {
	lib := &Library{
		Title:             title,
		GitHubAccountName: githubAccountName,
		RepoName:          repoName,
		PathOnDisk:        pathOnDisk,
	}
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
		jot.Error(errs.Wrap(err))
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
		tokens := l.monitor.stop()
		l.PathOnDisk = p
		l.LastSeen = l.VersionOnDisk()
		for _, token := range tokens {
			l.monitor.startWatch(token)
		}
	}
	return nil
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
	return l.GitHubAccountName == userGitHubAccountName && l.RepoName == userRepoName
}

// CheckForAvailableUpgrade returns releases that can be upgraded to.
func (l *Library) CheckForAvailableUpgrade(ctx context.Context, client *http.Client) {
	l.lock.Lock()
	l.upgrade = nil
	l.lock.Unlock()
	incompatibleFutureLibraryVersion := strconv.Itoa(gid.CurrentDataVersion + 1)
	minimumLibraryVersion := strconv.Itoa(gid.MinimumLibraryVersion)
	available, err := LoadReleases(ctx, client, l.GitHubAccountName, l.RepoName, l.VersionOnDisk(),
		func(version, notes string) bool {
			return incompatibleFutureLibraryVersion == version ||
				txt.NaturalLess(version, minimumLibraryVersion, true) ||
				txt.NaturalLess(incompatibleFutureLibraryVersion, version, true)
		})
	var upgrade *Release
	if err != nil {
		jot.Error(err)
		upgrade = &Release{CheckFailed: true}
	} else {
		switch len(available) {
		case 0:
			upgrade = &Release{}
		case 1:
			upgrade = &available[0]
		default:
			for _, one := range available[1:] {
				available[0].Notes += "\n\n## Version " + one.Version + "\n" + one.Notes
			}
			upgrade = &available[0]
		}
	}
	l.lock.Lock()
	l.upgrade = upgrade
	l.lock.Unlock()
}

// AvailableUpdate returns the available release that can be updated to.
func (l *Library) AvailableUpdate() *Release {
	l.lock.RLock()
	defer l.lock.RUnlock()
	if l.upgrade == nil {
		return nil
	}
	r := *l.upgrade
	return &r
}

// Less returns true if this Library should be placed before the other Library.
func (l *Library) Less(other *Library) bool {
	if l.IsUser() {
		return true
	}
	if l.IsMaster() {
		return !other.IsUser()
	}
	if txt.NaturalLess(l.GitHubAccountName, other.GitHubAccountName, true) {
		return true
	}
	if l.GitHubAccountName != other.GitHubAccountName {
		return false
	}
	return txt.NaturalLess(l.RepoName, other.RepoName, true)
}

// VersionOnDisk returns the version of the data on disk, if it can be determined.
func (l *Library) VersionOnDisk() string {
	data, err := os.ReadFile(filepath.Join(l.PathOnDisk, releaseFile))
	if err != nil {
		if !os.IsNotExist(err) {
			jot.Warn(errs.NewWithCause("unable to load "+releaseFile+" from library: "+l.Title, err))
		}
		return "0"
	}
	return strings.TrimSpace(string(bytes.SplitN(data, []byte{'\n'}, 2)[0]))
}

// Download the release onto the local disk.
func (l *Library) Download(ctx context.Context, client *http.Client, release Release) error {
	p := l.Path()
	if err := os.MkdirAll(p, 0o750); err != nil {
		return errs.NewWithCause("unable to create "+p, err)
	}
	data, err := l.downloadRelease(ctx, client, release)
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
				return errs.NewWithCause("unable to create "+parent, err)
			}
			if err = extractFile(f, fullPath); err != nil {
				return errs.NewWithCause("unable to create "+fullPath, err)
			}
		}
	}
	f := filepath.Join(root, releaseFile)
	if err = os.WriteFile(f, []byte(release.Version+"\n"), 0o640); err != nil {
		return errs.NewWithCause("unable to create "+f, err)
	}
	return nil
}

func extractFile(f *zip.File, dst string) (err error) {
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
