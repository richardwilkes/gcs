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
	"cmp"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

// EventRootSync is the event code used when the root path being monitored has been changed to a new path. Only the
// watches carried over to the new root receive it; a newly established watch does not, since its owner does its own
// initial scan, and a callback that rescans (re-establishing its watch) on every event would otherwise never settle.
const EventRootSync = 0xFFFFFFFF

// watchedEvents are the filesystem changes a library watch reports. Writes are included so that a file edited in place
// by another program is reported: the navigator's deep search caches file contents and relies on the watches to tell
// it which entries to drop, and an in-place edit changes neither the tree nor the file's name.
const watchedEvents = notify.Create | notify.Remove | notify.Rename | notify.Write

// keepsStoppedWatches is true where rjeczalik/notify keeps its handle on a folder it has stopped watching until
// something changes within it: on Windows.
const keepsStoppedWatches = runtime.GOOS == xos.WindowsOS

// nestsWatches is true where rjeczalik/notify watches the folders within one it watches through that folder's watch:
// on Windows and macOS.
const nestsWatches = keepsStoppedWatches || runtime.GOOS == xos.MacOS

// eventBufferSize is the capacity of the channel notify delivers events on. notify drops an event when the channel is
// full, and a dropped Create, Remove or Rename is a silently missed reload, so the buffer has to absorb bursts: an
// in-place edit produces a Write per write syscall, and a bulk copy, unzip or git pull into a library produces a Create
// and a Write per file. listenForEvents drains the channel into an unbounded queue, so the buffer only has to cover the
// time that goroutine spends descheduled, but a generous one costs only a few kilobytes.
const eventBufferSize = 1024

// notifyWatch is notify.Watch, held in a variable so that tests can hold up a hand-over (see handOver) or have it leave
// folders unwatched (see watchTree).
var notifyWatch = notify.Watch

type monitor struct {
	library    *Library
	lock       sync.RWMutex
	events     chan notify.EventInfo
	done       chan bool
	queue      *xos.TaskQueue
	tokensLock sync.RWMutex
	tokens     []*MonitorToken
}

func newMonitor(library *Library) *monitor {
	return &monitor{library: library}
}

func (m *monitor) newWatch(callback func(lib *Library, fullPath string, what notify.Event), callbackOnUIThread bool) *MonitorToken {
	token := &MonitorToken{
		monitor:    m,
		callback:   callback,
		onUIThread: callbackOnUIThread,
	}
	// No root sync is sent here. See EventRootSync for why a new watch must not receive one.
	m.startWatch(token, false)
	return token
}

func (m *monitor) startWatch(token *MonitorToken, sendSync bool) {
	unlockWatches := lockWatches()
	m.lock.Lock()
	_, statErr := os.Lstat(m.library.Path(false))
	// A missing directory cannot be watched, so it is created here. Library.Watch documents that it does this.
	root := m.library.Path(true)
	form := reportedForm(root)
	token.root = root
	m.tokensLock.Lock()
	token.watched = map[string]string{root: form}
	m.tokens = append(m.tokens, token)
	m.tokensLock.Unlock()
	releaseHold := func() {}
	if m.events == nil {
		platformWatches.releaseRemade(form, errors.Is(statErr, fs.ErrNotExist))
		m.queue = xos.NewTaskQueue(&xos.TaskQueueConfig{Workers: 1})
		m.done = make(chan bool)
		m.events = make(chan notify.EventInfo, eventBufferSize)
		var err error
		if releaseHold, err = platformWatches.handOver(m, form); err != nil {
			errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", token.root)
			m.events = nil
			m.done = nil
			m.queue.Shutdown()
			m.queue = nil
		} else {
			go m.listenForEvents()
		}
	}
	queue := m.queue
	m.lock.Unlock()
	unlockWatches()
	releaseHold()
	if sendSync {
		// Only this token is being (re)started, so only it is told to sync. Using send() here would deliver the sync to
		// every registered token, resulting in N² notifications when a monitor with N tokens is restarted.
		m.sendTo(queue, token, root, EventRootSync)
	}
}

// stop stops the monitor's watch and hands back its tokens. Unless whileStopped is nil, it is called once the watch is
// stopped, before any other goroutine can establish it again.
func (m *monitor) stop(whileStopped func()) []*MonitorToken {
	m.lock.Lock()
	// The tokens are registered by startWatch regardless of whether the filesystem watch could be established, so they
	// must be handed back and cleared here regardless as well. Otherwise a monitor whose watch failed would report no
	// tokens, leaving the caller (Library.SetPath) with nothing to restart once the path becomes watchable again.
	m.tokensLock.Lock()
	tokens := slices.Clone(m.tokens)
	m.tokens = nil
	m.tokensLock.Unlock()
	var holders []*monitor
	if m.events != nil {
		unlockHandOver := lockHandOver() // So that no folder is handed over between the stop and the notes made of it
		notify.Stop(m.events)
		holders = platformWatches.holders(m)
		platformWatches.stopped(m)
		unlockHandOver()
		close(m.events)
		<-m.done
		m.queue.Shutdown()
		m.queue = nil
		m.events = nil
		m.done = nil
	}
	if whileStopped != nil {
		whileStopped()
	}
	m.lock.Unlock()
	rewatch(holders) // Without the monitor's lock, since rewatch takes watchLock
	return tokens
}

func (m *monitor) listenForEvents() {
	for evt := range m.events {
		if !strings.HasPrefix(filepath.Base(evt.Path()), releaseFilePrefix) {
			m.send(evt.Path(), evt.Event())
		}
	}
	m.done <- true
}

// send an event to every registered token. Only called from listenForEvents and from reportMove, under platformWatches'
// lock while it has the monitor among those that have handed folders over, both of which stop ends before it shuts the
// queue down.
func (m *monitor) send(fullPath string, what notify.Event) {
	m.queue.Submit(func() {
		m.tokensLock.RLock()
		tokens := slices.Clone(m.tokens)
		m.tokensLock.RUnlock()
		for _, token := range tokens {
			m.deliver(token, fullPath, what)
		}
	})
}

// sendTo sends an event to just the given token. The queue will be nil if the filesystem watch could not be
// established, in which case the event is delivered directly, since there is no worker to hand it off to.
func (m *monitor) sendTo(queue *xos.TaskQueue, token *MonitorToken, fullPath string, what notify.Event) {
	if queue == nil {
		m.deliver(token, fullPath, what)
		return
	}
	queue.Submit(func() { m.deliver(token, fullPath, what) })
}

// deliver hands an event to the token's callback, once for each path the library knows the reported location by. A
// root sync already names the root as the library knows it, so it is passed through as is.
func (m *monitor) deliver(token *MonitorToken, fullPath string, what notify.Event) {
	paths := []string{fullPath}
	if what != EventRootSync {
		paths = token.libraryPaths(fullPath)
	}
	for _, p := range paths {
		if token.onUIThread {
			unison.InvokeTask(func() { token.callback(m.library, p, what) })
		} else {
			token.callback(m.library, p, what)
		}
	}
}

// MonitorToken holds a token that can be used to stop a library watch.
type MonitorToken struct {
	monitor  *monitor
	callback func(*Library, string, notify.Event)
	root     string
	// watched maps each path this token watches, as the library knows it, to the form the platform watcher was handed
	// and reports changes in (see reportedForm). The root is always present; AddSubPath adds the rest. Guarded by the
	// monitor's tokensLock, since it is read while events are delivered, which happens on the monitor's queue -- and
	// the monitor's own lock is held while that queue is shut down and drained, so waiting on it there would deadlock.
	watched    map[string]string
	onUIThread bool
}

// reportedForm returns the form in which the platform watcher reports changes beneath the given path, which is also the
// form the watcher must be handed. rjeczalik/notify resolves the symlinks in each watched path (on macOS, turning /var
// into /private/var) and reports changes beneath the resolved path. On macOS, FSEvents also reports paths in their
// on-disk case and notify drops any change whose path does not begin with the watched path, so a watch on a path typed
// in the wrong case would report nothing; resolvePath produces the form the watcher agrees with. A path that cannot be
// resolved is returned as is, since the watcher can't watch it anyway.
func reportedForm(p string) string {
	if resolved, err := resolvePath(p); err == nil {
		return resolved
	}
	return p
}

// libraryPaths returns the paths the library knows the given reported path by: for each watched path whose reported
// form covers it, the reported prefix is swapped for the path as the library knows it. There is usually exactly one.
// There can be more when a symlinked directory inside the library (see AddSubPath) resolves to somewhere else inside
// it, since the library then holds the same file under two names and a change to it must be reported under both. A
// path beneath none of the watched paths is returned as is.
func (m *MonitorToken) libraryPaths(reported string) []string {
	m.monitor.tokensLock.RLock()
	defer m.monitor.tokensLock.RUnlock()
	var paths []string
	for known, form := range m.watched {
		if reported == form {
			paths = append(paths, known)
		} else if strings.HasPrefix(reported, strings.TrimSuffix(form, string(filepath.Separator))+string(filepath.Separator)) {
			paths = append(paths, filepath.Join(known, reported[len(form):]))
		}
	}
	if len(paths) == 0 {
		return []string{reported}
	}
	slices.Sort(paths)
	return paths
}

// Library returns the library this token is attached to.
func (m *MonitorToken) Library() *Library {
	return m.monitor.library
}

// AddSubPath adds a sub-path within the library to watch. Should only be called for symlinks, since the native OS
// monitoring typically does not traverse those on its own. As for the library's own folder (see Library.Watch), another
// program moving the folder the symlink leads to elsewhere is reported as a change to the sub-path.
func (m *MonitorToken) AddSubPath(relativePath string) {
	var releaseHold func()
	defer func() { // Last, once the locks are let go of
		if releaseHold != nil {
			releaseHold()
		}
	}()
	unlockWatches := lockWatches()
	defer unlockWatches()
	m.monitor.lock.Lock()
	defer m.monitor.lock.Unlock()
	if m.monitor.events != nil {
		if fullPath, err := filepath.Abs(filepath.Join(m.root, relativePath)); err != nil {
			errs.Log(err)
		} else if _, watched := m.watched[fullPath]; !watched {
			form := reportedForm(fullPath)
			platformWatches.releaseRemade(form, false)
			m.monitor.tokensLock.Lock()
			m.watched[fullPath] = form // Ahead of the hand-over, so that a move reported once it is noted maps back
			m.monitor.tokensLock.Unlock()
			if releaseHold, err = platformWatches.handOver(m.monitor, form); err != nil {
				m.monitor.tokensLock.Lock()
				delete(m.watched, fullPath)
				m.monitor.tokensLock.Unlock()
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", fullPath)
			}
		}
	}
}

// releaseFilePrefix begins the names of the files and folders that touchFolder and awaitPlatformWatcher make and
// remove. The watches do not report them.
const releaseFilePrefix = ".gcs-release-"

// releaseGrace is how long after a folder is let go of that StopWatchesWithin goes on reporting it, since the platform
// watcher closes its handle on a goroutine of its own soon afterward.
const releaseGrace = time.Second

// platformWatches records the folders every monitor has handed to the platform watcher, so that the ones it may still
// hold open can be released; see ReleaseStoppedWatches.
var platformWatches = &watchedFolders{
	byMonitor: make(map[*monitor]map[string]bool),
	kept:      make(map[string]bool),
	toRelease: make(map[string]bool),
	ids:       make(map[string]folderID),
	letGo:     make(map[string]time.Time),
	entries:   make(map[string]func()),
}

// watchLock is held, on Windows only (see lockWatches), while the platform watcher is handed a folder to watch, along
// with the release that precedes it (see releaseRemade), throughout rewatchLocked, so that the folders it stops
// watching stay unwatched until it watches them again (see releaseMovedFolder), and throughout release, so that no
// folder is watched again before the platform watcher has let go of a handle it held there. Locks are taken in this
// order: watchLock, a monitor's lock, handOverLock, platformWatches.lock, which the watches for moves take to report
// one (see reportMove), so they are stopped without it.
var watchLock sync.Mutex

// handOverLock is held, where the platform watcher nests watches (see lockHandOver), while it is handed a folder along
// with the notes made of that (see handOver), while a monitor's watch is stopped along with the notes made of that (see
// monitor.stop), and throughout rewatchLocked, so that stopped counts as handed over just the folders the platform
// watcher watches at the stop, the ones it goes on watching through the stopped monitor's.
var handOverLock sync.Mutex

// lockHandOver locks handOverLock where the platform watcher nests watches, and returns the function that unlocks it.
// Elsewhere handOverLock guards nothing.
func lockHandOver() (unlock func()) {
	if !nestsWatches {
		return func() {}
	}
	handOverLock.Lock()
	return handOverLock.Unlock
}

// lockWatches locks watchLock where releaseMovedFolder runs, on Windows, and returns the function that unlocks it.
// Elsewhere watchLock guards nothing.
func lockWatches() (unlock func()) {
	if !keepsStoppedWatches {
		return func() {}
	}
	watchLock.Lock()
	return watchLock.Unlock
}

// watchedFolders is the type of platformWatches.
type watchedFolders struct {
	lock sync.Mutex
	// byMonitor holds the folders each monitor has handed to the platform watcher, in the form it handed them over in
	// (see reportedForm).
	byMonitor map[*monitor]map[string]bool
	// kept holds the folders, where the platform watcher nests watches, that a stopped monitor handed it and that it
	// goes on watching through a handle of its own for the monitors that have handed it folders within them, until it
	// stops watching for all of those (see stopped and rewatchLocked).
	kept map[string]bool
	// toRelease holds the folders the platform watcher may have stopped watching, each true while a release of it is
	// under way.
	toRelease map[string]bool
	// ids holds, on Windows, the id of the folder at each path in byMonitor, kept or toRelease as of when it was last
	// handed to the platform watcher, which is the folder it holds at that path even once another program has moved it.
	ids map[string]folderID
	// letGo holds when each folder released within the last releaseGrace was released.
	letGo map[string]time.Time
	// entries holds the function that stops the watch for moves (see watchFolderEntry) of each folder whose moves are
	// watched for (see movesWatchedLocked), or nil while that watch is being started (see syncEntries).
	entries map[string]func()
}

// handOver hands the folder to the platform watcher to watch on the monitor's events channel, noting that (see handing
// and added) while it holds handOverLock. The monitor's lock must be held. rjeczalik/notify fails to watch a folder
// holding a kept one, which it watches only for the folders handed over within it, so the kept folders within this one
// are first handed over for a monitor of their own (see holdKeptLocked), until release is called, without any of the
// locks (see watchLock), since that has the monitors whose folders hold them watched afresh.
func (w *watchedFolders) handOver(m *monitor, form string) (release func(), err error) {
	unlock := lockHandOver()
	defer unlock()
	release = w.holdKeptLocked(form)
	return release, w.handOverLocked(m, form)
}

// handOverLocked is handOver without the hold, for a caller that holds handOverLock where lockHandOver takes it.
func (w *watchedFolders) handOverLocked(m *monitor, form string) error {
	failed := w.handing(m, form)
	if err := watchTree(form, m.events); err != nil {
		failed()
		return err
	}
	w.added(m, form)
	return nil
}

// holdKeptLocked hands the kept folders within form to the platform watcher for a monitor of their own, which no
// library has, until the returned function is called. Its lock is not taken, since nothing can stop or rewatch it
// before handOverLock, which must be held where lockHandOver takes it, is let go of.
func (w *watchedFolders) holdKeptLocked(form string) (release func()) {
	w.lock.Lock()
	var held []string
	for f := range w.kept {
		if isWithin(form, f) {
			held = append(held, f)
		}
	}
	w.lock.Unlock()
	if len(held) == 0 {
		return func() {}
	}
	hold := &monitor{
		queue:  xos.NewTaskQueue(&xos.TaskQueueConfig{Workers: 1}),
		done:   make(chan bool),
		events: make(chan notify.EventInfo),
	}
	go hold.listenForEvents() // Which, with no tokens, reports nothing
	for _, f := range held {
		if err := w.handOverLocked(hold, f); err != nil {
			errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", f)
		}
	}
	return func() { hold.stop(nil) }
}

// handing notes, ahead of added, that the monitor is handing the folder to the platform watcher, so that
// StopWatchesWithin finds the monitor meanwhile. It returns the function that undoes this, should the platform watcher
// fail to take the folder.
func (w *watchedFolders) handing(m *monitor, form string) (failed func()) {
	w.lock.Lock()
	defer w.lock.Unlock()
	forms := w.byMonitor[m]
	if forms == nil {
		forms = make(map[string]bool)
		w.byMonitor[m] = forms
	}
	if forms[form] {
		return func() {}
	}
	forms[form] = true
	return func() {
		w.lock.Lock()
		delete(w.byMonitor[m], form)
		if len(w.byMonitor[m]) == 0 {
			delete(w.byMonitor, m)
		}
		w.lock.Unlock()
		w.syncEntries() // In case one was started for the folder meanwhile
	}
}

// added notes that the monitor has handed the folder to the platform watcher, which then stops watching the folders
// handed to it, or kept, that lie within it, and watches them through this one, and has the moves of the folders it
// watches watched for (see syncEntries).
func (w *watchedFolders) added(m *monitor, form string) {
	var id folderID
	identified := false
	if keepsStoppedWatches {
		var err error
		if id, err = idOfFolder(form); err != nil {
			errs.Log(errs.NewWithCause("unable to identify watched folder", err), "path", form)
		} else {
			identified = true
		}
	}
	w.lock.Lock()
	if identified {
		w.ids[form] = id
	} else {
		delete(w.ids, form)
	}
	for _, forms := range w.byMonitor {
		for f := range forms {
			if isWithin(form, f) {
				w.toRelease[f] = false
			}
		}
	}
	for f := range w.kept {
		if f == form {
			delete(w.kept, f) // Handed over again, so watched for the monitor
		} else if isWithin(form, f) {
			delete(w.kept, f)
			w.toRelease[f] = false
		}
	}
	forms := w.byMonitor[m]
	if forms == nil {
		forms = make(map[string]bool)
		w.byMonitor[m] = forms
	}
	forms[form] = true
	w.lock.Unlock()
	w.syncEntries()
}

// stopped notes that the monitor's watch has been stopped, so the platform watcher may have stopped watching the
// folders it handed over. One it watched through a handle of its own is kept, if other monitors have handed it folders
// within that one, since rjeczalik/notify goes on watching it for their sake, and a kept one is no longer watched once
// none of those remains. The watches for moves of the folders no longer watched are stopped (see syncEntries).
// handOverLock must be held where lockHandOver takes it.
func (w *watchedFolders) stopped(m *monitor) {
	w.lock.Lock()
	forms := w.byMonitor[m]
	own := make(map[string]bool, len(forms))
	for f := range forms {
		own[f] = nestsWatches && w.watchedLocked(f)
	}
	delete(w.byMonitor, m)
	for f := range forms {
		if own[f] && !w.handedLocked(f) && w.holdsHandedLocked(f, nil) {
			w.kept[f] = true
		} else {
			w.toRelease[f] = false
		}
	}
	w.unkeepLocked(nil)
	w.lock.Unlock()
	w.syncEntries()
}

// unkeepLocked drops the kept folders that no monitor other than those given has handed a folder within, which the
// platform watcher has stopped watching once it stopped watching for the monitors, and notes that it may still hold
// them. The lock must be held.
func (w *watchedFolders) unkeepLocked(monitors []*monitor) {
	for f := range w.kept {
		if !w.holdsHandedLocked(f, monitors) {
			delete(w.kept, f)
			w.toRelease[f] = false
		}
	}
}

// watchedLocked returns true if the platform watcher watches the folder through a handle of its own: some monitor has
// handed it over, or it is kept, and no folder handed over or kept holds it. The lock must be held.
func (w *watchedFolders) watchedLocked(form string) bool {
	return (w.kept[form] || w.handedLocked(form)) && !w.heldLocked(form)
}

// movesWatchedLocked returns true if the moves of the folder are watched for (see watchFolderEntry): where the platform
// watcher nests watches, if it watches the folder through a handle of its own, since it watches the folders within
// such a folder by their paths, and on Windows a watch for the moves of one of those would keep folders from being
// renamed that the platform watcher's handle does not; elsewhere, if some monitor has handed it over. The lock must be
// held.
func (w *watchedFolders) movesWatchedLocked(form string) bool {
	if nestsWatches {
		return w.watchedLocked(form)
	}
	return w.handedLocked(form)
}

// handedLocked returns true if some monitor has handed the folder to the platform watcher. The lock must be held.
func (w *watchedFolders) handedLocked(form string) bool {
	for _, forms := range w.byMonitor {
		if forms[form] {
			return true
		}
	}
	return false
}

// heldLocked returns true if a folder some monitor has handed to the platform watcher, or a kept one, holds the
// folder. The lock must be held.
func (w *watchedFolders) heldLocked(form string) bool {
	for f := range w.kept {
		if isWithin(f, form) {
			return true
		}
	}
	for _, forms := range w.byMonitor {
		for f := range forms {
			if isWithin(f, form) {
				return true
			}
		}
	}
	return false
}

// holdsHandedLocked returns true if a monitor other than those given has handed the platform watcher a folder within
// the folder. The lock must be held.
func (w *watchedFolders) holdsHandedLocked(form string, except []*monitor) bool {
	for m, forms := range w.byMonitor {
		if !slices.Contains(except, m) {
			for f := range forms {
				if isWithin(form, f) {
					return true
				}
			}
		}
	}
	return false
}

// forgetLocked drops the folder from toRelease, along with its id unless it is still handed over or kept. The lock
// must be held.
func (w *watchedFolders) forgetLocked(form string) {
	delete(w.toRelease, form)
	if !w.kept[form] && !w.handedLocked(form) {
		delete(w.ids, form)
	}
}

// syncEntries starts a watch for the moves of each folder whose moves are watched for (see movesWatchedLocked) and has
// none, and stops each one whose folder no longer is. It is called after each change to those folders, and a watch
// started meanwhile for a folder that one has made no longer watched for is stopped again, by that call or by this
// one. None of the locks other than watchLock, a monitor's and handOverLock may be held.
func (w *watchedFolders) syncEntries() {
	var stops []func()
	var starts []string
	w.lock.Lock()
	for f, stop := range w.entries {
		if !w.movesWatchedLocked(f) {
			delete(w.entries, f)
			if stop != nil {
				stops = append(stops, stop)
			}
		}
	}
	consider := func(f string) {
		if _, ok := w.entries[f]; !ok && w.movesWatchedLocked(f) {
			w.entries[f] = nil
			starts = append(starts, f)
		}
	}
	for f := range w.kept {
		consider(f)
	}
	for _, forms := range w.byMonitor {
		for f := range forms {
			consider(f)
		}
	}
	w.lock.Unlock()
	for _, stop := range stops {
		stop()
	}
	for _, f := range starts {
		// Without the lock, since this opens folders
		stop := watchFolderEntry(f, func(what notify.Event) { w.reportMove(f, what) })
		w.lock.Lock()
		if current, ok := w.entries[f]; ok && current == nil { // Still wanted, and not started by another call
			w.entries[f] = stop
			stop = nil
		}
		w.lock.Unlock()
		if stop != nil {
			stop()
		}
	}
}

// reportMove has each monitor that has handed the platform watcher the folder at form, or one within it, report a
// change of the given kind to the folder it handed over, which moved along with the one at form, so that it is watched
// afresh.
func (w *watchedFolders) reportMove(form string, what notify.Event) {
	w.lock.Lock()
	defer w.lock.Unlock()
	for m, forms := range w.byMonitor {
		for f := range forms {
			if f == form || isWithin(form, f) {
				m.send(f, what)
			}
		}
	}
}

// holders returns the other monitors that have handed the platform watcher a folder at or holding one that m has
// handed it, along with those that have handed it a folder at or holding one of theirs, and so on. On Windows and
// macOS, rjeczalik/notify stops watching a folder when it stops the last watch it keeps through that folder for one
// within it, though the folder itself is still to be watched, so these are watched again (see rewatch) once m's watch
// is stopped.
func (w *watchedFolders) holders(m *monitor) []*monitor {
	if !nestsWatches {
		return nil
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	set := map[*monitor]bool{m: true}
	w.addHoldersLocked(set)
	delete(set, m)
	return slices.Collect(maps.Keys(set))
}

// watchingOver returns the monitors that have handed the platform watcher a folder at or within the outermost folder
// handed to it, or kept, that holds dir, or at or within dir if none does, along with their holders (see holders).
// rjeczalik/notify keeps a watch at dir to itself while it watches a folder holding dir, and hands it over to a folder
// it watches within dir, so the watch reaches the platform watcher at dir only while none of these is watched.
func (w *watchedFolders) watchingOver(dir string) []*monitor {
	w.lock.Lock()
	defer w.lock.Unlock()
	outermost := w.outermostLocked(dir)
	set := make(map[*monitor]bool)
	for m, forms := range w.byMonitor {
		for f := range forms {
			if f == outermost || isWithin(outermost, f) {
				set[m] = true
				break
			}
		}
	}
	w.addHoldersLocked(set)
	return slices.Collect(maps.Keys(set))
}

// addHoldersLocked adds to the set each monitor that has handed the platform watcher a folder at or holding one that a
// monitor in the set has handed it, until there are no more. The lock must be held.
func (w *watchedFolders) addHoldersLocked(set map[*monitor]bool) {
	for added := true; added; {
		added = false
		for other, forms := range w.byMonitor {
			if set[other] {
				continue
			}
		search:
			for f := range forms {
				for m := range set {
					for held := range w.byMonitor[m] {
						if f == held || isWithin(f, held) {
							set[other] = true
							added = true
							break search
						}
					}
				}
			}
		}
	}
}

// rewatch is rewatchLocked, with nothing to call between, for a caller that holds none of the locks (see watchLock).
// Without monitors it has nothing to do, so does not wait for watchLock.
func rewatch(monitors []*monitor) {
	if len(monitors) == 0 {
		return
	}
	unlockWatches := lockWatches()
	defer unlockWatches()
	rewatchLocked(monitors, nil)
}

// rewatchLocked has the platform watcher stop watching for the monitors, calls between unless it is nil, and then has
// it watch the folders they have handed it again, those holding others first, so that it watches each as before, save
// that the kept folders it watched only for these are watched no longer. Unless there are no monitors, it holds
// handOverLock throughout, between included, which keeps the folders handed over as they are and the events channels
// of their monitors open. It is called only where the platform watcher nests watches (see holders and
// releaseMovedFolder), where lockHandOver takes that lock. watchLock must be held where lockWatches takes it, and of the
// other locks at most a monitor's.
func rewatchLocked(monitors []*monitor, between func()) {
	if len(monitors) != 0 { // Otherwise a stop elsewhere need not wait for between
		unlockHandOver := lockHandOver()
		defer unlockHandOver()
	}
	type handed struct {
		m    *monitor
		form string
	}
	var watching []*monitor
	var watches []handed
	platformWatches.lock.Lock()
	for _, m := range monitors {
		if forms := platformWatches.byMonitor[m]; len(forms) != 0 { // A monitor with none is not watching
			watching = append(watching, m)
			for form := range forms {
				watches = append(watches, handed{m: m, form: form})
			}
		}
	}
	platformWatches.lock.Unlock()
	for _, m := range watching {
		notify.Stop(m.events)
	}
	platformWatches.lock.Lock()
	platformWatches.unkeepLocked(monitors)
	platformWatches.lock.Unlock()
	if between != nil {
		between()
	}
	slices.SortFunc(watches, func(a, b handed) int { return cmp.Compare(len(a.form), len(b.form)) })
	for _, one := range watches {
		if err := notify.Watch(one.form+"/...", one.m.events, watchedEvents); err != nil {
			errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", one.form)
		}
	}
	platformWatches.syncEntries()
}

// release releases the folders the platform watcher may have stopped watching and does not watch now, only those at
// or within dir unless dir is empty. It returns true if it released any there, now or within the last releaseGrace.
// None of the locks may be held (see watchLock).
func (w *watchedFolders) release(dir string) bool {
	unlockWatches := lockWatches()
	defer unlockWatches()
	w.lock.Lock()
	var dirs []string
	for f := range w.toRelease {
		switch {
		case dir != "" && f != dir && !isWithin(dir, f):
		case w.watchedLocked(f):
			w.forgetLocked(f)
		default:
			w.toRelease[f] = true
			dirs = append(dirs, f)
		}
	}
	w.lock.Unlock()
	for _, f := range dirs {
		releaseFolder(f)
	}
	now := time.Now()
	w.lock.Lock()
	defer w.lock.Unlock()
	for _, f := range dirs {
		if w.toRelease[f] { // Not noted again while it was being released
			w.forgetLocked(f)
		}
		w.letGo[f] = now
	}
	recent := false
	for f, at := range w.letGo {
		if now.Sub(at) >= releaseGrace {
			delete(w.letGo, f)
		} else if dir == "" || f == dir || isWithin(dir, f) {
			recent = true
		}
	}
	return recent
}

// releaseRemade has the platform watcher let go of the folder it stopped watching at form, if it may still hold it and
// the folder at form is another: one just made because form was missing, as made reports, or one that another program
// put in place of one it moved. The handle moved along with the folder, and rjeczalik/notify would take it back once
// form is watched again, so that the changes reported for form would be those made where the folder went. A folder just
// made is removed for the release (see releaseMovedFolder), then made again; otherwise the change that releases the
// handle is made where the folder went (see heldFolder) and waited for, unless the folder has been removed, which has
// notify close the handle itself. A folder that a watched folder holds is left for ReleaseStoppedWatches, since notify
// then watches it through that folder. Called with watchLock held just before form is handed over: by startWatch, which
// holds the lock of its monitor, not yet watching, and by AddSubPath, which makes no folder, so passes made false. Does
// nothing on platforms other than Windows.
func (w *watchedFolders) releaseRemade(form string, made bool) {
	if !keepsStoppedWatches {
		return
	}
	w.lock.Lock()
	_, stopped := w.toRelease[form]
	stale := stopped && !w.watchedLocked(form) && !w.heldLocked(form)
	if stale {
		w.toRelease[form] = true
	}
	w.lock.Unlock()
	if !stale {
		return
	}
	released := false
	if made && os.Remove(form) == nil { // Fails if something has been put there meanwhile
		released = true
		releaseMovedFolder(form)
		if err := os.Mkdir(form, 0o750); err != nil && !errors.Is(err, fs.ErrExist) {
			errs.Log(err, "path", form)
		}
	} else if at, err := w.heldFolder(form); err != nil {
		errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", form)
	} else if at == "" {
		released = true
	} else if at != form {
		// No watched folder holds form, so none is or holds the folder holding it, as awaitPlatformWatcher needs.
		if err = touchFolder(at, func() { awaitPlatformWatcher(filepath.Dir(form)) }); err != nil {
			errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", at)
		} else {
			released = true
		}
	}
	w.lock.Lock()
	if w.toRelease[form] { // Not noted again while it was being released
		if released {
			w.forgetLocked(form)
		} else {
			w.toRelease[form] = false
		}
	}
	w.lock.Unlock()
}

// heldFolder returns the path of the folder the platform watcher opened at form: form itself, unless another folder has
// been put there since form was last handed over (see ids), in which case the one handed over is looked for by its id,
// and an empty string if it is gone. rjeczalik/notify closes its handle on a folder it has stopped watching once the
// folder is removed. An error identifying the folder at form comes with form, and one looking for the other with an
// empty string.
func (w *watchedFolders) heldFolder(form string) (string, error) {
	w.lock.Lock()
	id, known := w.ids[form]
	w.lock.Unlock()
	if !known {
		return form, nil
	}
	current, err := idOfFolder(form)
	if err != nil || current == id {
		return form, err
	}
	return findFolder(id, form)
}

// outsideWatches returns the folder holding the outermost folder handed to the platform watcher, or kept, that holds
// dir, or holding dir if none does: one that no folder handed over or kept is or holds, as awaitPlatformWatcher needs,
// for as long as watchLock is held.
func (w *watchedFolders) outsideWatches(dir string) string {
	w.lock.Lock()
	defer w.lock.Unlock()
	return filepath.Dir(w.outermostLocked(dir))
}

// outermostLocked returns the outermost folder handed to the platform watcher, or kept, that holds dir, or dir if none
// does. The lock must be held.
func (w *watchedFolders) outermostLocked(dir string) string {
	outermost := dir
	consider := func(f string) {
		if isWithin(f, dir) && len(f) < len(outermost) {
			outermost = f
		}
	}
	for f := range w.kept {
		consider(f)
	}
	for _, forms := range w.byMonitor {
		for f := range forms {
			consider(f)
		}
	}
	return outermost
}

// isWithin returns true if p lies within dir.
func isWithin(dir, p string) bool {
	rel, err := filepath.Rel(dir, p)
	return err == nil && rel != "." && filepath.IsLocal(rel)
}

// releaseFolder has the platform watcher close its handle on the folder it opened at dir, and on any folder holding
// that one, if it no longer watches them, by making and removing a file within it, wherever it is now, if anywhere (see
// heldFolder), or through releaseMovedFolder if dir is gone. It closes them on a goroutine of its own soon afterward,
// or, if the folder has moved, before this returns, since it would take the handle back were dir watched first.
// watchLock must be held, and none of the other locks. Does nothing on platforms other than Windows.
func releaseFolder(dir string) {
	if !keepsStoppedWatches {
		return
	}
	at, err := platformWatches.heldFolder(dir)
	if err == nil && at != "" {
		var await func()
		if at != dir {
			parent := platformWatches.outsideWatches(dir)
			await = func() { awaitPlatformWatcher(parent) }
		}
		err = touchFolder(at, await)
	}
	if err != nil {
		if at == dir && errors.Is(err, fs.ErrNotExist) {
			releaseMovedFolder(dir)
		} else {
			errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", dir)
		}
	}
}

// touchFolder makes a file within dir and removes it, the two changes after which rjeczalik/notify closes its handles
// on a folder it has stopped watching. Unless await is nil, it is called after each, so that notify has read both by
// the time this returns, since it reads the second only once it has handled the first.
func touchFolder(dir string, await func()) error {
	f, err := os.CreateTemp(dir, releaseFilePrefix+"*")
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		errs.Log(err, "path", f.Name())
	}
	if await != nil {
		await()
	}
	if err = os.Remove(f.Name()); err == nil && await != nil {
		await()
	}
	return err
}

// releaseMovedFolder has the platform watcher close its handle on the folder it watched at dir, which is no longer
// there, as when another program has moved it: the handle moved with the folder, out of reach of anything made at dir.
// rjeczalik/notify takes back the handle it stopped watching when something at dir is watched again, and closes the
// handle of a watch it stops at a path where nothing is left, so a folder is made at dir, along with any folder that
// held it and is gone too, just long enough to watch it. The libraries that would keep that watch from reaching the
// platform watcher are not watched meanwhile (see watchingOver), and watchLock, which must be held, keeps any from
// being watched again until it is done. It returns once the handle is closed.
func releaseMovedFolder(dir string) {
	var gone []string // dir, then each folder holding it that is gone too
	for p := dir; ; {
		if _, err := os.Lstat(p); err == nil {
			break
		} else if !errors.Is(err, fs.ErrNotExist) {
			errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", p)
			return
		}
		gone = append(gone, p)
		parent := filepath.Dir(p)
		if parent == p {
			return
		}
		p = parent
	}
	if len(gone) == 0 {
		return // Something else is at dir now
	}
	rewatchLocked(platformWatches.watchingOver(dir), func() {
		i := len(gone) // gone[i:] are the folders made
		for ; i > 0; i-- {
			if err := os.Mkdir(gone[i-1], 0o750); err != nil {
				if !errors.Is(err, fs.ErrExist) { // Something else put there meanwhile is expected
					errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", gone[i-1])
				}
				break
			}
		}
		events := make(chan notify.EventInfo) // Never read, so notify drops whatever it would report on it
		var err error
		if i == 0 {
			err = notify.Watch(dir+"/...", events, watchedEvents)
		}
		for _, p := range gone[i:] {
			if rmErr := os.Remove(p); rmErr != nil {
				errs.Log(rmErr, "path", p)
			}
		}
		if i == 0 {
			if err == nil {
				notify.Stop(events)
				awaitPlatformWatcher(filepath.Dir(gone[len(gone)-1]))
			} else {
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", dir)
			}
		}
	})
}

// awaitPlatformWatcher returns once the platform watcher has handled what it was told before the call, or after a
// second. On Windows, rjeczalik/notify handles the changes it is told of, and the handles it is told to close, in turn
// on a goroutine of its own. So a folder is watched, moved and its watch stopped, which with nothing left at the
// watched path has notify close its handle in turn, and the folder holding it can be renamed only once that handle is
// closed. Nothing changes within the watched folder, since notify reading a change on Windows trips checkptr under
// -race (see build.sh). The folders are made within parent, which must not be within a watched folder, since notify
// would watch them through that folder.
func awaitPlatformWatcher(parent string) {
	holder, err := os.MkdirTemp(parent, releaseFilePrefix+"*")
	if err != nil {
		errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", parent)
		return
	}
	watched := filepath.Join(holder, releaseFilePrefix+"watched")
	events := make(chan notify.EventInfo) // Never read, since nothing changes within the folder
	if err = os.Mkdir(watched, 0o750); err == nil {
		if err = notify.Watch(watched, events, notify.Create); err == nil {
			err = os.Rename(watched, filepath.Join(holder, releaseFilePrefix+"moved"))
			notify.Stop(events)
		}
	}
	if err != nil {
		errs.Log(errs.NewWithCause("unable to release filesystem watch", err), "path", watched)
	} else {
		renamed := holder + "-done"
		for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
			if os.Rename(holder, renamed) == nil {
				holder = renamed
				break
			}
		}
	}
	if err = os.RemoveAll(holder); err != nil {
		errs.Log(err, "path", holder)
	}
}

// ReleaseStoppedWatches has the platform watcher let go of the folders it has stopped watching. On Windows,
// rjeczalik/notify keeps its handle on such a folder open until something changes within it, which keeps any folder
// holding it from being renamed. It stops watching a library's folder when the library's watches are stopped and when
// a library holding that folder comes to be watched. A folder is let go of by making and removing a file within it,
// which the watches do not report. A watch is often stopped only to be established again, as a reload of the navigator
// does, so this is best called once whatever is to be watched has been watched afresh, which leaves those folders
// alone. The handles are closed on a goroutine of the watcher's own soon afterward.
func ReleaseStoppedWatches() {
	platformWatches.release("")
}

// StopWatchesWithin stops the watches of every library that watches the folder at dir or one within it, whether as its
// own folder or through a symlinked one (see MonitorToken.AddSubPath), and then has the platform watcher let go of the
// folders at or within dir that it has stopped watching (see ReleaseStoppedWatches). It returns true if any there were
// let go of, now or within the last releaseGrace, whose handles the platform watcher closes on a goroutine of its own
// soon afterward.
func StopWatchesWithin(dir string) bool {
	form := reportedForm(dir)
	platformWatches.lock.Lock()
	var monitors []*monitor
	for m, forms := range platformWatches.byMonitor {
		for f := range forms {
			if f == form || isWithin(form, f) {
				monitors = append(monitors, m)
				break
			}
		}
	}
	platformWatches.lock.Unlock()
	for _, m := range monitors {
		m.stop(nil)
	}
	return platformWatches.release(form)
}

// Stop this watch.
func (m *MonitorToken) Stop() {
	m.monitor.tokensLock.Lock()
	if i := slices.Index(m.monitor.tokens, m); i != -1 {
		m.monitor.tokens = slices.Delete(m.monitor.tokens, i, i+1)
		if len(m.monitor.tokens) == 0 {
			m.monitor.tokensLock.Unlock()
			m.monitor.stop(nil)
			return
		}
	}
	m.monitor.tokensLock.Unlock()
}
