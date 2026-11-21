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
	"encoding/binary"
	"errors"
	"path/filepath"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/unix"
)

// inotifyInit1 and inotifyAddWatch are unix.InotifyInit1 and unix.InotifyAddWatch, held in variables so that tests can
// have them fail.
var (
	inotifyInit1    = unix.InotifyInit1
	inotifyAddWatch = unix.InotifyAddWatch
)

// folderMoves is the inotify instance through which watchFolderEntry watches folders for their moves, shared by every
// such watch, since a user may have only a few instances.
var folderMoves = struct {
	lock sync.Mutex
	fd   int // -1 until a folder is first watched
	// initLogged is true once a failure to make the instance has been logged. Each watch tries again, but the failure
	// is logged only once, since the watches are established afresh on every reload of the navigator.
	initLogged bool
	// limitLogged is true once a failure to watch a folder because the user is at fs.inotify.max_user_watches has
	// been logged, which, for the same reason, is logged only once.
	limitLogged bool
	// watches holds the watches told of the moves of the folder each inotify watch descriptor watches.
	watches map[int32]map[*movesWatch]bool
}{fd: -1, watches: make(map[int32]map[*movesWatch]bool)}

// movesWatch is a watch made by watchFolderEntry.
type movesWatch struct {
	changed func(what notify.Event)
	wds     []int32 // The inotify watch descriptors of the folders holding the watched one
}

// watchFolderEntry calls changed, on a goroutine of its own, for each move of a folder holding the folder at p, until
// the returned function is called, which returns once changed is no longer being called. rjeczalik/notify watches each
// folder through an inotify watch, which follows the folder wherever another program moves it, and reports the move of
// only the folder it was asked to watch, so its watch goes on reporting the changes made where the folders within a
// moved one went, and none in folders put in their places. A folder that can't be read, as a /home with mode 711 can't,
// can't be watched, so its moves go unreported.
func watchFolderEntry(p string, changed func(what notify.Event)) (stop func()) {
	w := &movesWatch{changed: changed}
	folderMoves.lock.Lock()
	defer folderMoves.lock.Unlock()
	if folderMoves.fd == -1 {
		fd, err := inotifyInit1(unix.IN_CLOEXEC)
		if err != nil {
			if !folderMoves.initLogged {
				folderMoves.initLogged = true
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", p)
			}
			return func() {}
		}
		folderMoves.fd = fd
		go readFolderMoves(fd)
	}
	for dir := filepath.Dir(p); dir != filepath.Dir(dir); dir = filepath.Dir(dir) { // The root can't be moved
		wd, err := inotifyAddWatch(folderMoves.fd, dir, unix.IN_MOVE_SELF|unix.IN_ONLYDIR|unix.IN_DONT_FOLLOW)
		if err != nil {
			switch {
			case errors.Is(err, unix.EACCES): // What a folder that can't be read brings
			case errors.Is(err, unix.ENOSPC): // What the user at fs.inotify.max_user_watches brings
				if !folderMoves.limitLogged {
					folderMoves.limitLogged = true
					errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", dir)
				}
			default:
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", dir)
			}
			continue
		}
		watches := folderMoves.watches[int32(wd)]
		if watches == nil {
			watches = make(map[*movesWatch]bool)
			folderMoves.watches[int32(wd)] = watches
		}
		watches[w] = true
		w.wds = append(w.wds, int32(wd))
	}
	return func() {
		folderMoves.lock.Lock()
		defer folderMoves.lock.Unlock()
		for _, wd := range w.wds {
			if watches := folderMoves.watches[wd]; watches[w] {
				delete(watches, w)
				if len(watches) == 0 {
					delete(folderMoves.watches, wd)
					// EINVAL comes for a folder that is gone, whose watch the system has removed already
					_, err := unix.InotifyRmWatch(folderMoves.fd, uint32(wd))
					if err != nil && !errors.Is(err, unix.EINVAL) {
						errs.Log(err)
					}
				}
			}
		}
	}
}

// readFolderMoves tells the watches of each move fd reports, for as long as fd can be read. The watches are told under
// folderMoves' lock, so that one stopped is told nothing more.
func readFolderMoves(fd int) {
	buf := make([]byte, 4096)
	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}
			errs.Log(errs.NewWithCause("unable to watch filesystem paths", err))
			return
		}
		folderMoves.lock.Lock()
		for offset := 0; offset+unix.SizeofInotifyEvent <= n; {
			wd := int32(binary.NativeEndian.Uint32(buf[offset:]))
			mask := binary.NativeEndian.Uint32(buf[offset+4:])
			offset += unix.SizeofInotifyEvent + int(binary.NativeEndian.Uint32(buf[offset+12:]))
			switch {
			case mask&unix.IN_Q_OVERFLOW != 0: // Moves may have gone unreported
				told := make(map[*movesWatch]bool)
				for _, watches := range folderMoves.watches {
					for w := range watches {
						if !told[w] {
							told[w] = true
							w.changed(notify.Rename)
						}
					}
				}
			case mask&unix.IN_IGNORED != 0: // The watch was removed, by the system for a folder that is gone
				delete(folderMoves.watches, wd)
			case mask&unix.IN_MOVE_SELF != 0:
				for w := range folderMoves.watches[wd] {
					w.changed(notify.Rename)
				}
			}
		}
		folderMoves.lock.Unlock()
	}
}
