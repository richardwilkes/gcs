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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

func (s *Server) colorsHandler(w http.ResponseWriter, _ *http.Request) {
	m := make(map[string]unison.ThemeColor)
	for _, c := range gurps.CurrentColors() {
		m[strings.ReplaceAll(c.ID, "_", "-")] = *c.Color
	}
	JSONResponse(w, http.StatusOK, m)
}
