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
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/network/xhttp/web"
)

var _ http.Handler = &Server{}

// Monitor is an interface that can be implemented to be notified when the server starts and stops.
type Monitor interface {
	WebServerStarted(*Server)
	WebServerStopped(*Server)
}

// Server holds the embedded web server.
type Server struct {
	web.Server
}

// StartServerInBackground starts the server in the background. Both parameters may be nil.
func StartServerInBackground(monitor Monitor) *Server {
	settings := &gurps.GlobalSettings().WebServer
	host, portStr, err := net.SplitHostPort(settings.Address)
	if err != nil {
		errs.Log(errs.NewWithCause("invalid address; using defaults instead", err))
		host = "localhost"
		portStr = "0"
	}
	var port int
	if port, err = strconv.Atoi(portStr); err != nil {
		errs.Log(errs.NewWithCause("invalid port; selecting a random port instead", err))
	}
	s := &Server{
		Server: web.Server{
			CertFile:            settings.CertFile,
			KeyFile:             settings.KeyFile,
			ShutdownGracePeriod: fxp.SecondsToDuration(settings.ShutdownGracePeriod),
			Ports:               []int{port},
			WebServer: &http.Server{
				Addr:         host,
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
			},
		},
	}
	s.WebServer.Handler = s
	if !toolbox.IsNil(monitor) {
		s.StartedChan = make(chan any, 1)
		go func() {
			<-s.StartedChan
			monitor.WebServerStarted(s)
		}()
		var once sync.Once
		s.ShutdownCallback = func() { once.Do(func() { monitor.WebServerStopped(s) }) }
	}
	go func() {
		if err := s.Run(); err != nil {
			errs.Log(err)
			if s.ShutdownCallback != nil {
				s.ShutdownCallback() // In case we errored out before the server finished starting up
			}
		}
	}()
	return s
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		switch req.URL.Path {
		case "/":
			fmt.Fprintln(w, "Hello, world!")
		default:
			http.Error(w, i18n.Text("Not Found"), http.StatusNotFound)
		}
	default:
		http.Error(w, i18n.Text("Method Not Allowed"), http.StatusMethodNotAllowed)
	}
}
