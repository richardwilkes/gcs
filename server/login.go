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
	"net/http"
)

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	name := r.FormValue("name")
	password := r.FormValue("password")
	fmt.Println("name:", name)
	fmt.Println("password:", password)

	http.Redirect(w, r, s.prefix+"/", http.StatusFound)
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement
	http.Redirect(w, r, s.prefix+"/login", http.StatusFound)
}
