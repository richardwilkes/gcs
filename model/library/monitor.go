package library

import (
	"sync"

	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/taskqueue"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
	"golang.org/x/exp/slices"
)

// EventRootSync the event code used when the root path being monitored has been changed to a new path. Also occurs as
// the first event received.
const EventRootSync = 0xFFFFFFFF

type monitor struct {
	library *Library
	lock    sync.RWMutex
	events  chan notify.EventInfo
	done    chan bool
	queue   *taskqueue.Queue
	tokens  []*MonitorToken
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
	m.startWatch(token)
	return token
}

func (m *monitor) startWatch(token *MonitorToken) {
	rootPath := m.library.Path()
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tokens = append(m.tokens, token)
	if m.events == nil {
		m.queue = taskqueue.New(taskqueue.Workers(1))
		m.done = make(chan bool)
		m.events = make(chan notify.EventInfo, 16)
		target := rootPath + "/..."
		if err := notify.Watch(target, m.events, notify.Create|notify.Remove|notify.Rename); err != nil {
			jot.Error(errs.NewWithCausef(err, "unable to watch filesystem path: %s", target))
			m.events = nil
			m.done = nil
			m.queue.Shutdown()
			m.queue = nil
		} else {
			go m.listenForEvents()
		}
	}
	m.send(rootPath, EventRootSync)
}

func (m *monitor) stop() []*MonitorToken {
	m.lock.Lock()
	tokens := m.internalStop()
	m.lock.Unlock()
	return tokens
}

func (m *monitor) internalStop() []*MonitorToken {
	var tokens []*MonitorToken
	if m.events != nil {
		tokens = make([]*MonitorToken, len(m.tokens))
		copy(tokens, m.tokens)
		notify.Stop(m.events)
		close(m.events)
		<-m.done
		m.queue.Shutdown()
		m.queue = nil
		m.events = nil
		m.done = nil
		m.tokens = nil
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
		m.lock.RLock()
		tokens := make([]*MonitorToken, len(m.tokens))
		copy(tokens, m.tokens)
		m.lock.RUnlock()
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
	onUIThread bool
}

// Stop this watch.
func (m *MonitorToken) Stop() {
	m.monitor.lock.Lock()
	if i := slices.Index(m.monitor.tokens, m); i != -1 {
		m.monitor.tokens = slices.Delete(m.monitor.tokens, i, i+1)
		if len(m.monitor.tokens) == 0 {
			m.monitor.internalStop()
		}
	}
	m.monitor.lock.Unlock()
}
