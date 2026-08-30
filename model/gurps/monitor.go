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
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

// EventRootSync the event code used when the root path being monitored has been changed to a new path. Only the
// watches that were carried over to the new root receive it; establishing a watch does not deliver one, since whoever
// establishes it does its own initial scan. Sending one at that point would put a callback that rescans on every event
// into an endless cycle, as each rescan re-establishes the watch and would be handed another sync in turn.
const EventRootSync = 0xFFFFFFFF

// watchedEvents are the filesystem changes a library watch reports. Writes are included so that a file edited in place
// by another program is reported: the navigator's deep search caches file contents and relies on the watches to tell
// it which entries to drop, and an in-place edit changes neither the tree nor the file's name.
const watchedEvents = notify.Create | notify.Remove | notify.Rename | notify.Write

// eventBufferSize is the capacity of the channel notify delivers events on. notify drops an event outright when the
// channel is full rather than blocking, and a dropped Create, Remove or Rename is a silently missed reload, so the
// buffer has to absorb bursts: watching Writes means an in-place edit produces one event per write syscall, and a bulk
// copy, unzip or git pull into a library produces a Create and a Write per file, all competing with the tree changes
// for slots. listenForEvents drains the channel into an unbounded queue as fast as it can, so the buffer only has to
// cover the time that goroutine spends descheduled, but a generous one costs a few kilobytes and makes a drop remote
// rather than merely unlikely.
const eventBufferSize = 1024

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
	m.lock.Lock()
	token.root = m.library.Path()
	form := reportedForm(token.root)
	m.tokensLock.Lock()
	token.watched = map[string]string{token.root: form}
	m.tokens = append(m.tokens, token)
	m.tokensLock.Unlock()
	if m.events == nil {
		m.queue = xos.NewTaskQueue(&xos.TaskQueueConfig{Workers: 1})
		m.done = make(chan bool)
		m.events = make(chan notify.EventInfo, eventBufferSize)
		if err := notify.Watch(form+"/...", m.events, watchedEvents); err != nil {
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
	root := token.root
	m.lock.Unlock()
	if sendSync {
		// Only this token is being (re)started, so only it is told to sync. Using send() here would deliver the sync to
		// every registered token, resulting in N² notifications when a monitor with N tokens is restarted.
		m.sendTo(queue, token, root, EventRootSync)
	}
}

func (m *monitor) stop() []*MonitorToken {
	m.lock.Lock()
	defer m.lock.Unlock()
	// The tokens are registered by startWatch regardless of whether the filesystem watch could be established, so they
	// must be handed back and cleared here regardless as well. Otherwise a monitor whose watch failed would report no
	// tokens, leaving the caller (Library.SetPath) with nothing to restart once the path becomes watchable again.
	m.tokensLock.Lock()
	tokens := slices.Clone(m.tokens)
	m.tokens = nil
	m.tokensLock.Unlock()
	if m.events != nil {
		notify.Stop(m.events)
		close(m.events)
		<-m.done
		m.queue.Shutdown()
		m.queue = nil
		m.events = nil
		m.done = nil
	}
	return tokens
}

func (m *monitor) listenForEvents() {
	for evt := range m.events {
		m.send(evt.Path(), evt.Event())
	}
	m.done <- true
}

// send an event to every registered token. Only called from listenForEvents, which cannot be running unless the queue
// exists.
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
	// and reports changes in (see reportedForm). The root is always present; AddSubPath adds the rest. Guarded by the monitor's tokensLock, since it is read
	// while events are delivered, which happens on the monitor's queue -- and the monitor's own lock is held while that
	// queue is shut down and drained, so waiting on it there would deadlock.
	watched    map[string]string
	onUIThread bool
}

// reportedForm returns the form in which the platform watcher reports changes beneath the given path, which is also the
// form the path must be handed to the watcher in. rjeczalik/notify resolves the symlinks in each path it is asked to
// watch (on macOS, that also turns /var into /private/var) and reports every change beneath the resolved path rather
// than the one it was given. On macOS, FSEvents additionally reports each path in the case it has on disk, and notify
// drops any change whose reported path does not begin with the path it was asked to watch, so a watch established on a
// path typed in the wrong case would report nothing at all; resolvePath is the platform's way of arriving at the form
// the watcher will agree with. A path that cannot be resolved is returned as is; the watcher fails to establish a watch
// on such a path, so nothing is ever reported beneath it anyway.
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
// monitoring typically does not traverse those on its own.
func (m *MonitorToken) AddSubPath(relativePath string) {
	m.monitor.lock.Lock()
	defer m.monitor.lock.Unlock()
	if m.monitor.events != nil {
		if fullPath, err := filepath.Abs(filepath.Join(m.root, relativePath)); err != nil {
			errs.Log(err)
		} else if _, watched := m.watched[fullPath]; !watched {
			form := reportedForm(fullPath)
			if err = notify.Watch(form+"/...", m.monitor.events, watchedEvents); err != nil {
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", fullPath)
			} else {
				m.monitor.tokensLock.Lock()
				m.watched[fullPath] = form
				m.monitor.tokensLock.Unlock()
			}
		}
	}
}

// Stop this watch.
func (m *MonitorToken) Stop() {
	m.monitor.tokensLock.Lock()
	if i := slices.Index(m.monitor.tokens, m); i != -1 {
		m.monitor.tokens = slices.Delete(m.monitor.tokens, i, i+1)
		if len(m.monitor.tokens) == 0 {
			m.monitor.tokensLock.Unlock()
			m.monitor.stop()
			return
		}
	}
	m.monitor.tokensLock.Unlock()
}
