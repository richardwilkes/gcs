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
	"path/filepath"
	"slices"
	"sync"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/taskqueue"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

// EventRootSync the event code used when the root path being monitored has been changed to a new path. Also occurs as
// the first event received.
const EventRootSync = 0xFFFFFFFF

type monitor struct {
	library    *Library
	lock       sync.RWMutex
	events     chan notify.EventInfo
	done       chan bool
	queue      *taskqueue.Queue
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
	m.startWatch(token, false)
	return token
}

func (m *monitor) startWatch(token *MonitorToken, sendSync bool) {
	m.lock.Lock()
	defer m.lock.Unlock()
	token.root = m.library.Path()
	token.subPaths = make(map[string]bool)
	m.tokensLock.Lock()
	m.tokens = append(m.tokens, token)
	m.tokensLock.Unlock()
	if m.events == nil {
		m.queue = taskqueue.New(taskqueue.Workers(1))
		m.done = make(chan bool)
		m.events = make(chan notify.EventInfo, 16)
		if err := notify.Watch(token.root+"/...", m.events, notify.Create|notify.Remove|notify.Rename); err != nil {
			errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", token.root)
			m.events = nil
			m.done = nil
			m.queue.Shutdown()
			m.queue = nil
		} else {
			go m.listenForEvents()
		}
	}
	if sendSync {
		m.send(token.root, EventRootSync)
	}
}

func (m *monitor) stop() []*MonitorToken {
	m.lock.Lock()
	defer m.lock.Unlock()
	var tokens []*MonitorToken
	if m.events != nil {
		m.tokensLock.RLock()
		tokens = make([]*MonitorToken, len(m.tokens))
		copy(tokens, m.tokens)
		m.tokensLock.RUnlock()
		notify.Stop(m.events)
		close(m.events)
		<-m.done
		m.queue.Shutdown()
		m.queue = nil
		m.events = nil
		m.done = nil
		m.tokensLock.Lock()
		m.tokens = nil
		m.tokensLock.Unlock()
	}
	return tokens
}

func (m *monitor) listenForEvents() {
	for evt := range m.events {
		m.send(evt.Path(), evt.Event())
	}
	m.done <- true
}

func (m *monitor) send(fullPath string, what notify.Event) {
	m.queue.Submit(func() {
		m.tokensLock.RLock()
		tokens := make([]*MonitorToken, len(m.tokens))
		copy(tokens, m.tokens)
		m.tokensLock.RUnlock()
		for _, token := range tokens {
			if token.onUIThread {
				unison.InvokeTask(func() { token.callback(m.library, fullPath, what) })
			} else {
				token.callback(m.library, fullPath, what)
			}
		}
	})
}

// MonitorToken holds a token that can be used to stop a library watch.
type MonitorToken struct {
	monitor    *monitor
	callback   func(*Library, string, notify.Event)
	root       string
	subPaths   map[string]bool
	onUIThread bool
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
		} else if !m.subPaths[fullPath] {
			if err = notify.Watch(fullPath+"/...", m.monitor.events, notify.Create|notify.Remove|notify.Rename); err != nil {
				errs.Log(errs.NewWithCause("unable to watch filesystem path", err), "path", fullPath)
			} else {
				m.subPaths[fullPath] = true
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
