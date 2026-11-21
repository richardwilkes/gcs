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
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/rjeczalik/notify"
	"golang.org/x/sys/unix"
)

// treeLock is held throughout notifyWatchTree and notifyWatchFolder, so that no other watch takes as its own the
// watches a failed one of notifyWatchTree leaves behind before they are dropped.
var treeLock sync.Mutex

// watchTree has rjeczalik/notify watch the folder at form and every folder within it that can be read, reporting their
// changes on events (see watchReadable). notify discards the errors of the inotify watches it makes for these folders,
// so how many of them it then does not watch, as when the user is at fs.inotify.max_user_watches, leaving out those
// whose failure watchReadable logs, is logged here, as is a failure to tell, but only once (see
// folderMoves.limitLogged). Neither is returned, since the folders it does watch go on being watched.
func watchTree(form string, events chan<- notify.EventInfo) error {
	folderMoves.lock.Lock()
	logged := folderMoves.limitLogged
	folderMoves.lock.Unlock()
	if logged {
		_, err := watchReadable(form, events)
		return err
	}
	folders := foldersWithin(form) // Before the watch, leaving out any made since, which notify may not watch yet
	failed, err := watchReadable(form, events)
	if err != nil {
		return err
	}
	watched, err := notifyWatchedInodes()
	if err != nil {
		logLimitOnce(errs.NewWithCause("unable to tell which folders are watched", err), "path", form)
		return nil
	}
	unwatched := 0
	for p, ino := range folders {
		if !watched[ino] && folderInode(p) == ino && // Still there, so notify was handed it
			!slices.ContainsFunc(failed, func(f string) bool { return f == p || isWithin(f, p) }) {
			unwatched++
		}
	}
	if unwatched != 0 {
		logLimitOnce(errs.New("unable to watch every folder at or within filesystem path; the user may be at "+
			"fs.inotify.max_user_watches"), "path", form, "unwatched", unwatched)
	}
	return nil
}

// watchReadable has rjeczalik/notify watch the folder at form and every folder within it that can be read. It returns
// the error that keeps it from watching form, having watched nothing on events, or else the folders within form whose
// failure it has logged (see watchWithin). notify fails the watch of a whole tree when it can't read a folder in it
// (see cannotRead), so such a tree is watched as its folder on its own and each folder within it in turn. A folder made
// later within one watched on its own is reported, but is watched only once form is watched afresh. As with notify's
// watch of a tree, the user being at fs.inotify.max_user_watches fails no watch, since watchTree counts what it leaves
// unwatched.
func watchReadable(form string, events chan<- notify.EventInfo) (failed []string, err error) {
	if err = notifyWatchTree(form, events); err == nil || !cannotRead(err) {
		return nil, err
	}
	if err = notifyWatchFolder(form, events); err != nil && !errors.Is(err, unix.ENOSPC) {
		return nil, err
	}
	return watchWithin(form, events), nil
}

// watchWithin has rjeczalik/notify, which already watches dir on its own, watch each folder within dir as watchReadable
// watches form, passing over those that can't be read. Any other failure is logged, as watchFolderEntry does, rather
// than returned, since dir and the other folders go on being watched, and the folders it leaves unwatched are returned.
func watchWithin(dir string, events chan<- notify.EventInfo) (failed []string) {
	entries, err := os.ReadDir(dir) // Once dir is watched, so that a folder made since is reported
	if err != nil {
		return logUnwatched(dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() { // Leaving out symlinks, as notify does
			p := filepath.Join(dir, entry.Name())
			within, watchErr := watchReadable(p, events)
			failed = append(failed, within...)
			if watchErr != nil {
				failed = append(failed, logUnwatched(p, watchErr)...)
			}
		}
	}
	return failed
}

// logUnwatched logs err, which leaves the folders within the one at p unwatched, or that one too, and returns p, unless
// err is what a folder that can't be read brings.
func logUnwatched(p string, err error) []string {
	if cannotRead(err) {
		return nil
	}
	errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", p)
	return []string{p}
}

// cannotRead returns true if err is what a folder that can't be read brings: one the user may not read, as one with
// mode 000 or another user's private one, or one gone since the folder holding it was read.
func cannotRead(err error) bool {
	return errors.Is(err, unix.EACCES) || errors.Is(err, unix.ENOENT)
}

// notifyWatchFolder has rjeczalik/notify watch the folder at p on its own, holding treeLock (see notifyWatchTree).
func notifyWatchFolder(p string, events chan<- notify.EventInfo) error {
	treeLock.Lock()
	defer treeLock.Unlock()
	return notifyWatch(p, events, watchedEvents)
}

// notifyWatchTree has rjeczalik/notify watch the folder at form and every folder within it. When that fails, as for a
// folder within that can't be read, notify keeps the watches it made until then, which no channel holds, and a later
// watch of a folder they cover takes them as its own without looking for the rest, so they are dropped, as stopping
// any channel does.
func notifyWatchTree(form string, events chan<- notify.EventInfo) error {
	treeLock.Lock()
	defer treeLock.Unlock()
	err := notifyWatch(form+"/...", events, watchedEvents)
	if err != nil {
		notify.Stop(make(chan notify.EventInfo))
	}
	return err
}

// logLimitOnce logs the error with the given arguments unless folderMoves.limitLogged is set, and sets it.
func logLimitOnce(err error, args ...any) {
	folderMoves.lock.Lock()
	logged := folderMoves.limitLogged
	folderMoves.limitLogged = true
	folderMoves.lock.Unlock()
	if !logged {
		errs.Log(err, args...)
	}
}

// foldersWithin returns the inode of the folder at form and of each folder within it that can be read, keyed by path,
// walking them as rjeczalik/notify does: without following symlinks. The others are left out, since inotify can't
// watch a folder it can't read (see watchReadable).
func foldersWithin(form string) map[string]uint64 {
	folders := make(map[string]uint64)
	for pending := []string{form}; len(pending) != 0; {
		dir := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		ino := folderInode(dir)
		if ino == 0 {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		folders[dir] = ino
		for _, entry := range entries {
			if entry.IsDir() {
				pending = append(pending, filepath.Join(dir, entry.Name()))
			}
		}
	}
	return folders
}

// folderInode returns the inode of the folder at p, or 0 if there is no folder there.
func folderInode(p string) uint64 {
	var st unix.Stat_t
	if unix.Lstat(p, &st) != nil || st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return 0
	}
	return st.Ino
}

// notifyWatchedInodes returns the inode numbers rjeczalik/notify watches: those that the inotify instances of this
// process other than folderMoves' watch, as /proc lists them. Their devices are left out, since the one stat reports
// for a folder can differ from the one listed, as within a btrfs subvolume.
func notifyWatchedInodes() (map[uint64]bool, error) {
	folderMoves.lock.Lock()
	own := strconv.Itoa(folderMoves.fd)
	folderMoves.lock.Unlock()
	fds, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return nil, err
	}
	inodes := make(map[uint64]bool)
	for _, fd := range fds {
		if fd.Name() == own {
			continue
		}
		// Passes over what is not an inotify instance, and what has been closed since, as what ReadDir read through has
		if target, linkErr := os.Readlink("/proc/self/fd/" + fd.Name()); linkErr != nil ||
			target != "anon_inode:inotify" {
			continue
		}
		info, readErr := os.ReadFile("/proc/self/fdinfo/" + fd.Name())
		if readErr != nil {
			continue
		}
		for line := range strings.Lines(string(info)) {
			if !strings.HasPrefix(line, "inotify ") { // Each watch is listed as "inotify wd:<hex> ino:<hex> ..."
				continue
			}
			for field := range strings.FieldsSeq(line) {
				if hex, ok := strings.CutPrefix(field, "ino:"); ok {
					if ino, parseErr := strconv.ParseUint(hex, 16, 64); parseErr == nil {
						inodes[ino] = true
					}
					break
				}
			}
		}
	}
	return inodes, nil
}
