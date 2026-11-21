// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package library

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/unix"
)

// TestLibraryWatchLogsFoldersLeftUnwatchedAtMostOnce verifies that a library watch logs how many of the library's
// folders rjeczalik/notify does not watch, which notify itself does not report, since it discards the errors of the
// inotify watches it makes for them: nothing when it watches them all, and a record naming the library's folder when it
// watches none, as when the user is at fs.inotify.max_user_watches, or only some, as when the user nears it. The record
// is made once, since the watches are established afresh on every reload of the navigator. A folder within that can't
// be read, or that goes meanwhile, is passed over (see TestLibraryWatchPassesOverFoldersThatCannotBeRead), so nothing
// is logged for it, but any other failure is logged every time: of the whole watch, or of a folder within one that is
// watched on its own for holding a folder that can't be read.
func TestLibraryWatchLogsFoldersLeftUnwatchedAtMostOnce(t *testing.T) {
	for _, one := range []struct {
		name       string
		unreadable bool // Whether a folder within the library can't be read
		failures   int
		unwatched  int
		// watch stands in for notify.Watch, unless nil.
		watch func(path string, ch chan<- notify.EventInfo, what ...notify.Event) error
	}{
		{name: "every folder watched"},
		{
			name:      "no folder watched",
			unwatched: 2,
			watch:     func(string, chan<- notify.EventInfo, ...notify.Event) error { return nil },
		},
		{
			name:      "only the library folder watched",
			unwatched: 1,
			watch: func(path string, ch chan<- notify.EventInfo, what ...notify.Event) error {
				return notify.Watch(strings.TrimSuffix(path, "/..."), ch, what...)
			},
		},
		{name: "a folder within can't be read", unreadable: true},
		{
			name:       "a folder within can't be read, with no folder watched",
			unreadable: true,
			unwatched:  2,
			// As notify does at fs.inotify.max_user_watches, where it watches no folder, failing only for one it can't read
			watch: func(path string, _ chan<- notify.EventInfo, _ ...notify.Event) error {
				if dir, ok := strings.CutSuffix(path, "/..."); ok {
					return filepath.WalkDir(dir, func(_ string, _ fs.DirEntry, err error) error { return err })
				}
				if err := unix.Access(path, unix.R_OK); err != nil {
					return err
				}
				return unix.ENOSPC
			},
		},
		{
			name:     "the folders can't be listed",
			failures: 2,
			watch: func(path string, ch chan<- notify.EventInfo, what ...notify.Event) error {
				if dir, ok := strings.CutSuffix(path, "/..."); ok {
					return &fs.PathError{Op: "open", Path: dir, Err: unix.EIO}
				}
				return notify.Watch(path, ch, what...)
			},
		},
		{
			name:       "a folder within can't be read, and another can't be listed",
			unreadable: true,
			failures:   2,
			watch: func(path string, ch chan<- notify.EventInfo, what ...notify.Event) error {
				if dir, ok := strings.CutSuffix(path, "/..."); ok && filepath.Base(dir) == "sub" {
					return &fs.PathError{Op: "open", Path: dir, Err: unix.EIO}
				}
				return notify.Watch(path, ch, what...)
			},
		},
		{
			name:       "a folder within can't be read, and another goes meanwhile",
			unreadable: true,
			watch: func(path string, ch chan<- notify.EventInfo, what ...notify.Event) error {
				if dir, ok := strings.CutSuffix(path, "/..."); ok && filepath.Base(dir) == "sub" {
					if err := os.Remove(dir); err != nil {
						return err
					}
				}
				return notify.Watch(path, ch, what...)
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			p := filepath.Join(t.TempDir(), "lib")
			c.NoError(os.MkdirAll(filepath.Join(p, "sub"), 0o750))
			if one.unreadable {
				if os.Geteuid() == 0 {
					t.Skip("root can read every folder")
				}
				locked := filepath.Join(p, "locked")
				c.NoError(os.Mkdir(locked, 0))
				t.Cleanup(func() { c.NoError(os.Chmod(locked, 0o750)) }) // Ahead of the removal of the temporary folder
			}
			if one.watch != nil {
				original := notifyWatch
				notifyWatch = one.watch
				t.Cleanup(func() { notifyWatch = original })
			}
			folderMoves.lock.Lock()
			limitLogged := folderMoves.limitLogged
			folderMoves.limitLogged = false
			folderMoves.lock.Unlock()
			t.Cleanup(func() {
				folderMoves.lock.Lock()
				folderMoves.limitLogged = limitLogged
				folderMoves.lock.Unlock()
			})
			var logged bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logged, nil)))
			t.Cleanup(func() { slog.SetDefault(prev) })
			lib := NewLibrary("Test", "someone", "", "repo", p)
			for range 2 { // As for two reloads
				lib.Watch(func(*Library, string, notify.Event) {}, false).Stop()
			}
			form := reportedForm(p)
			failures, records := 0, 0
			var record string
			for line := range strings.Lines(logged.String()) {
				switch {
				case !strings.Contains(line, form):
				case strings.Contains(line, "unable to watch filesystem path"):
					failures++
				default:
					records++
					record = line
				}
			}
			c.Equal(one.failures, failures, "failures logged:\n%s", logged.String())
			if one.unwatched == 0 {
				c.Equal(0, records, "records logged:\n%s", logged.String())
			} else {
				c.Equal(1, records, "records logged:\n%s", logged.String())
				c.Contains(record, fmt.Sprintf("unwatched=%d", one.unwatched))
			}
		})
	}
}

// TestLibraryWatchPassesOverFoldersThatCannotBeRead verifies that a library watch reports the changes made in every
// folder of the library other than one that can't be read, which rjeczalik/notify fails to watch along with the rest
// of its tree: in the folders holding it and in every other folder within those.
func TestLibraryWatchPassesOverFoldersThatCannotBeRead(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read every folder")
	}
	c := check.New(t)
	p := filepath.Join(t.TempDir(), "lib")
	holder := filepath.Join(p, "holder")
	deep := filepath.Join(holder, "beside", "deep")
	other := filepath.Join(p, "other")
	c.NoError(os.MkdirAll(deep, 0o750))
	c.NoError(os.Mkdir(other, 0o750))
	locked := filepath.Join(holder, "locked")
	c.NoError(os.Mkdir(locked, 0))
	t.Cleanup(func() { c.NoError(os.Chmod(locked, 0o750)) }) // Ahead of the removal of the temporary folder
	token, events := watchForChanges(NewLibrary("Test", "someone", "", "repo", p))
	defer token.Stop()
	for i, dir := range []string{p, holder, deep, other} {
		name := fmt.Sprintf("%d.gcs", i)
		created := filepath.Join(dir, name)
		c.NoError(os.WriteFile(created, []byte("{}"), 0o600))
		c.Equal(created, awaitChange(t, events, name))
	}
}
