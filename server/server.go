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
	"compress/flate"
	"embed"
	"log/slog"
	"net/http"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/server/state"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
	"github.com/vearutop/statigz"
)

var _ http.Handler = &Server{}

var (
	//go:embed frontend/dist
	siteFS embed.FS

	siteLock sync.Mutex
	site     *Server
)

// Server holds the embedded web server.
type Server struct {
	server         *xhttp.Server
	mux            *http.ServeMux
	sheetsLock     sync.RWMutex
	entitiesByPath map[string]webEntity
}

// Start the server in the background. If the server is already running, nothing happens.
func Start(errorCallback func(error)) {
	siteLock.Lock()
	defer siteLock.Unlock()
	if site != nil {
		return
	}
	state.Set(state.Starting)
	settings := gurps.GlobalSettings().WebServer
	settings.Validate()
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
		mux:            http.NewServeMux(),
		entitiesByPath: make(map[string]webEntity),
	}
	s.installConfigurationHandlers()
	s.installSessionHandlers()
	s.installSheetHandlers()
	s.mux.Handle("GET /", statigz.FileServer(siteFS, statigz.FSPrefix("frontend/dist"), statigz.EncodeOnInit))
	site = s
	s.server.WebServer.Handler = s
	s.server.StartedChan = make(chan any, 1)
	go func() {
		<-s.server.StartedChan
		state.Set(state.Running)
	}()
	var once sync.Once
	s.server.ShutdownCallback = func(_ *slog.Logger) {
		once.Do(func() {
			state.Set(state.Stopped)
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
		if err := s.server.Run(); err != nil {
			errs.Log(err)
			if errorCallback != nil {
				errorCallback(err)
			}
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

	// Handle preflight requests
	if r.Method == http.MethodOptions {
		return
	}

	// Add the current session and user to the logger
	if id, userName, ok := sessionFromRequest(r); ok {
		if md := xhttp.MetadataFromRequest(r); md != nil {
			md.Logger = md.Logger.With("session", id, "user", userName)
		}
	}

	s.mux.ServeHTTP(w, r)
}

// Shutdown shuts down the server.
func (s *Server) Shutdown() {
	state.Set(state.Stopping)
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

// CompressedJSONResponse writes a compressed JSON response with a status code.
func CompressedJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	f, err := flate.NewWriter(w, flate.BestCompression)
	if err != nil {
		errs.Log(errs.NewWithCause("unable to create compressor", err))
		JSONResponse(w, statusCode, data)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Encoding", "deflate")
	w.WriteHeader(statusCode)
	if err = json.NewEncoder(f).Encode(data); err != nil {
		errs.Log(errs.Wrap(err))
	}
	if err = f.Close(); err != nil {
		errs.Log(errs.NewWithCause("unable to close compressor", err))
	}
}
