// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox/xio/network/xhttp"
)

func (s *Server) installPageRefHandlers() {
	s.mux.HandleFunc("GET /ref/{key}/{name}", s.pageRefHandler)
}

func (s *Server) pageRefHandler(w http.ResponseWriter, r *http.Request) {
	ref := gurps.GlobalSettings().PageRefs.Lookup(r.PathValue("key"))
	if ref == nil {
		xhttp.ErrorStatus(w, http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, ref.Path)
}
