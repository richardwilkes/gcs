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
	"sync"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/fatal"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

var _ http.Handler = &Server{}

//go:embed frontend/build frontend/build/_app
var siteFS embed.FS

// Monitor is an interface that can be implemented to be notified when the server starts and stops.
type Monitor interface {
	WebServerStarted(*Server)
	WebServerStopped(*Server)
}

// Server holds the embedded web server.
type Server struct {
	server      *xhttp.Server
	siteHandler http.Handler
	prefix      string
}

// StartServerInBackground starts the server in the background. Both parameters may be nil.
func StartServerInBackground(useDevMode bool, monitor Monitor) *Server {
	settings := &gurps.GlobalSettings().WebServer
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
	if useDevMode {
		s.prefix = "http://localhost:5173"
	}
	s.server.WebServer.Handler = s
	if !toolbox.IsNil(monitor) {
		s.server.StartedChan = make(chan any, 1)
		go func() {
			<-s.server.StartedChan
			monitor.WebServerStarted(s)
		}()
		var once sync.Once
		s.server.ShutdownCallback = func(_ *slog.Logger) { once.Do(func() { monitor.WebServerStopped(s) }) }
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
	return s
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.siteHandler.ServeHTTP(w, r)
	case http.MethodPost:
		switch r.URL.Path {
		case "/api/login":
			s.loginHandler(w, r)
		case "/api/logout":
			s.logoutHandler(w, r)
		default:
			http.Error(w, i18n.Text("Not Found"), http.StatusNotFound)
		}
	default:
		http.Error(w, i18n.Text("Method Not Allowed"), http.StatusMethodNotAllowed)
	}
}

// Shutdown shuts down the server.
func (s *Server) Shutdown() {
	s.server.Shutdown()
}
