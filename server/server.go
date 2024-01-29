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
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

var _ http.Handler = &Server{}

// State is the state of the server.
type State int

// Possible states for the server.
const (
	Stopped State = iota
	Starting
	Running
	Stopping
)

var (
	//go:embed frontend/build frontend/build/_app
	siteFS embed.FS

	state    atomic.Int32
	siteLock sync.Mutex
	site     *Server
)

// Server holds the embedded web server.
type Server struct {
	server      *xhttp.Server
	siteHandler http.Handler
}

// CurrentState returns the current state of the server.
func CurrentState() State {
	return State(state.Load())
}

// WaitUntilState waits until the server is in one of the specified states.
func WaitUntilState(state ...State) {
	for {
		current := CurrentState()
		for _, s := range state {
			if current == s {
				return
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// Start the server in the background. If the server is already running, nothing happens.
func Start() {
	siteLock.Lock()
	defer siteLock.Unlock()
	if site != nil {
		return
	}
	state.Store(int32(Starting))
	settings := gurps.GlobalSettings().WebServer
	settings.Validate()
	siteContentFS, err := fs.Sub(siteFS, "frontend/build")
	fatal.IfErr(err)
	s := &Server{
		server: &xhttp.Server{
			CertFile:            settings.CertFile,
			KeyFile:             settings.KeyFile,
			ShutdownGracePeriod: fxp.SecondsToDuration(settings.ShutdownGracePeriod),
			WebServer: &http.Server{
				Addr:         settings.Address,
				ReadTimeout:  fxp.SecondsToDuration(settings.ReadTimeout),
				WriteTimeout: fxp.SecondsToDuration(settings.ReadTimeout),
				IdleTimeout:  fxp.SecondsToDuration(settings.ReadTimeout),
			},
		},
		siteHandler: http.FileServer(http.FS(siteContentFS)),
	}
	site = s
	s.server.WebServer.Handler = s
	s.server.StartedChan = make(chan any, 1)
	go func() {
		<-s.server.StartedChan
		state.Store(int32(Running))
	}()
	var once sync.Once
	s.server.ShutdownCallback = func(_ *slog.Logger) {
		once.Do(func() {
			state.Store(int32(Stopped))
			go func() {
				// Has to be done in a goroutine to avoid a deadlock
				siteLock.Lock()
				if site == s { // only if it still points to our server
					site = nil
				}
				siteLock.Unlock()
			}()
		})
	}
	go func() {
		if err = s.server.Run(); err != nil {
			errs.Log(err)
			// In case we errored out before the server finished starting up, call the shutdown callback.
			if s.server.ShutdownCallback != nil {
				s.server.ShutdownCallback(s.server.Logger)
			}
		}
	}()
}

// Stop stops the server if it is running.
func Stop() {
	siteLock.Lock()
	defer siteLock.Unlock()
	if site != nil {
		site.Shutdown()
		site = nil
	}
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Deal with CORS first
	origin := r.Header.Get("Origin")
	header := w.Header()
	header.Set("Access-Control-Allow-Origin", origin)
	header.Set("Access-Control-Expose-Headers", "*")
	header.Set("Access-Control-Allow-Methods", "*")
	header.Set("Access-Control-Allow-Headers", "*")
	header.Set("Access-Control-Allow-Credentials", "true")
	header.Add("Vary", "Origin")

	switch r.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			s.siteHandler.ServeHTTP(w, r)
		} else {
			switch r.URL.Path {
			case "/api/session":
				s.sessionHandler(w, r)
			case "/api/version":
				s.versionHandler(w, r)
			case "/api/colors":
				s.colorsHandler(w, r)
			case "/api/sheets":
				s.sheetsHandler(w, r)
			default:
				if !strings.HasPrefix(r.URL.Path, "/api/sheet/") {
					xhttp.ErrorStatus(w, http.StatusNotFound)
					return
				}
				s.sheetHandler(w, r)
			}
		}
	case http.MethodPost:
		switch r.URL.Path {
		case "/api/login":
			s.loginHandler(w, r)
		case "/api/logout":
			s.logoutHandler(w, r)
		default:
			xhttp.ErrorStatus(w, http.StatusNotFound)
		}
	default:
		xhttp.ErrorStatus(w, http.StatusMethodNotAllowed)
	}
}

// Shutdown shuts down the server.
func (s *Server) Shutdown() {
	state.Store(int32(Stopping))
	s.server.Shutdown()
}

// JSONFromRequest extracts JSON content from the body of the request.
func JSONFromRequest(r *http.Request, data any) error {
	defer xio.DiscardAndCloseIgnoringErrors(r.Body)
	return jio.Load(r.Context(), r.Body, data)
}

// JSONResponse writes a JSON response with a status code.
func JSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		errs.Log(errs.Wrap(err))
	}
}
