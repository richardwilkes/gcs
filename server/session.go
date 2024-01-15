/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package server

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

const (
	maxSessionDuration = time.Hour * 24 * 30
	sessionGracePeriod = time.Minute * 10
	sessionIDHeader    = "X-Session"
	userHeader         = "X-User"
)

var (
	sessionLock sync.Mutex
	sessions    = make(map[uuid.UUID]Session)
	userLock    sync.RWMutex
	users       = make(map[string]User)
)

// User holds a user's information.
type User struct {
	Name     string
	Password string
}

// Session holds a session's information.
type Session struct {
	ID       uuid.UUID
	User     string
	Issued   time.Time
	LastUsed time.Time
}

func init() { // TODO: Remove
	userLock.Lock()
	users["richard wilkes"] = User{
		Name:     "Richard Wilkes",
		Password: "test",
	}
	userLock.Unlock()
}

func (s *Server) sessionHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionFromRequest(r)
	if err != nil {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return
	}
	setSessionHeaders(w, session)
}

func sessionFromRequest(r *http.Request) (*Session, error) {
	rawID := r.Header.Get(sessionIDHeader)
	if rawID == "" {
		return nil, errs.New("Session ID not found")
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, errs.NewWithCause("Invalid session ID", err)
	}

	sessionLock.Lock()
	session, ok := sessions[id]
	if !ok {
		sessionLock.Unlock()
		return nil, errs.New("Session not found")
	}
	if time.Since(session.Issued) > maxSessionDuration && time.Since(session.LastUsed) > sessionGracePeriod {
		delete(sessions, id)
		sessionLock.Unlock()
		return nil, errs.New("Session expired")
	}
	session.LastUsed = time.Now()
	sessions[id] = session
	sessionLock.Unlock()

	return &session, nil
}

func setSessionHeaders(w http.ResponseWriter, session *Session) {
	h := w.Header()
	h.Set(sessionIDHeader, session.ID.String())
	h.Set(userHeader, session.User)
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.ToLower(strings.TrimSpace(r.FormValue("name")))
	password := r.FormValue("password")

	userLock.RLock()
	user, ok := users[name]
	userLock.RUnlock()

	if !ok || user.Password != password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session := Session{
		ID:       uuid.New(),
		User:     user.Name,
		Issued:   time.Now(),
		LastUsed: time.Now(),
	}
	sessionLock.Lock()
	sessions[session.ID] = session
	sessionLock.Unlock()

	setSessionHeaders(w, &session)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionFromRequest(r)
	if err != nil {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
		return
	}

	sessionLock.Lock()
	delete(sessions, session.ID)
	sessionLock.Unlock()
}
