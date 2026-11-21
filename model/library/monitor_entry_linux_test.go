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
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/unix"
)

// TestWatchFolderEntryLogsALastingFailureAtMostOnce verifies that the watches for folders' moves log a failure that
// lasts at most once, since the watches are established afresh on every reload of the navigator: nothing for a folder
// holding the watched one that can't be read, as a /home with mode 711 can't, since such a folder can't be watched, and
// a failure to make an inotify instance, as when the user is at fs.inotify.max_user_instances, or to watch a folder,
// as when the user is at fs.inotify.max_user_watches, once.
func TestWatchFolderEntryLogsALastingFailureAtMostOnce(t *testing.T) {
	for _, one := range []struct {
		name   string
		logged int
		// setup readies the failure and returns the path of the folder to watch.
		setup func(t *testing.T) string
	}{
		{
			name: "a folder holding it can't be read",
			setup: func(t *testing.T) string {
				if os.Geteuid() == 0 {
					t.Skip("root can read every folder")
				}
				c := check.New(t)
				locked := filepath.Join(t.TempDir(), "locked")
				p := filepath.Join(locked, "lib")
				c.NoError(os.MkdirAll(p, 0o750))
				c.NoError(os.Chmod(locked, 0o311))
				t.Cleanup(func() { c.NoError(os.Chmod(locked, 0o750)) }) // Ahead of the removal of the temporary folder
				return p
			},
		},
		{
			name:   "no inotify instance can be made",
			logged: 1,
			setup: func(t *testing.T) string {
				original := inotifyInit1
				inotifyInit1 = func(int) (int, error) { return -1, unix.EMFILE }
				folderMoves.lock.Lock()
				fd, initLogged := folderMoves.fd, folderMoves.initLogged
				folderMoves.fd, folderMoves.initLogged = -1, false
				folderMoves.lock.Unlock()
				t.Cleanup(func() {
					inotifyInit1 = original
					folderMoves.lock.Lock()
					folderMoves.fd, folderMoves.initLogged = fd, initLogged
					folderMoves.lock.Unlock()
				})
				return filepath.Join(t.TempDir(), "lib")
			},
		},
		{
			name:   "the user is at fs.inotify.max_user_watches",
			logged: 1,
			setup: func(t *testing.T) string {
				original := inotifyAddWatch
				inotifyAddWatch = func(int, string, uint32) (int, error) { return -1, unix.ENOSPC }
				folderMoves.lock.Lock()
				limitLogged := folderMoves.limitLogged
				folderMoves.limitLogged = false
				folderMoves.lock.Unlock()
				t.Cleanup(func() {
					inotifyAddWatch = original
					folderMoves.lock.Lock()
					folderMoves.limitLogged = limitLogged
					folderMoves.lock.Unlock()
				})
				return filepath.Join(t.TempDir(), "lib")
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			p := one.setup(t)
			var logged bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logged, nil)))
			t.Cleanup(func() { slog.SetDefault(prev) })
			for range 2 { // As for two folders, or two reloads
				watchFolderEntry(p, func(notify.Event) {})()
			}
			check.New(t).Equal(one.logged, strings.Count(logged.String(), "\n"), "records logged:\n%s",
				logged.String())
		})
	}
}
