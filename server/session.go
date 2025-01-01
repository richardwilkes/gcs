// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package server

import (
	"net/http"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/server/websettings"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

const sessionIDHeader = "X-Session"

func (s *Server) installSessionHandlers() {
	s.mux.HandleFunc("GET /api/session", s.sessionHandler)
	s.mux.HandleFunc("POST /api/login", s.loginHandler)
	s.mux.HandleFunc("POST /api/logout", s.logoutHandler)
}

func (s *Server) sessionHandler(w http.ResponseWriter, r *http.Request) {
	if id, userName, ok := sessionFromRequest(r); !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
	} else {
		setSessionHeaders(w, id, userName)
	}
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	password := r.FormValue("password")
	settings := gurps.GlobalSettings().WebServer
	actualName, hashedPassword, ok := settings.LookupUserNameAndPassword(name)
	if !ok || hashedPassword != websettings.HashPassword(password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	setSessionHeaders(w, settings.CreateSession(actualName), actualName)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if id, _, ok := sessionFromRequest(r); !ok {
		xhttp.ErrorStatus(w, http.StatusUnauthorized)
	} else {
		gurps.GlobalSettings().WebServer.RemoveSession(id)
	}
}

func sessionFromRequest(r *http.Request) (sessionID tid.TID, userName string, ok bool) {
	rawID := r.Header.Get(sessionIDHeader)
	if rawID == "" {
		return "", "", false
	}
	var err error
	if sessionID, err = tid.FromStringOfKind(rawID, kinds.Session); err != nil {
		return "", "", false
	}
	userName, ok = gurps.GlobalSettings().WebServer.LookupSession(sessionID)
	return sessionID, userName, ok
}

func setSessionHeaders(w http.ResponseWriter, id tid.TID, userName string) {
	h := w.Header()
	h.Set(sessionIDHeader, string(id))
	h.Set("X-User", userName)
}
