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

	"github.com/richardwilkes/toolbox/cmdline"
)

func (s *Server) versionHandler(w http.ResponseWriter, _ *http.Request) {
	type versionResponse struct {
		Name      string
		Copyright string
		Version   string
		Build     string
		Git       string
		Modified  bool
	}
	JSONResponse(w, http.StatusOK, versionResponse{
		Name:      cmdline.AppName,
		Copyright: cmdline.Copyright(),
		Version:   cmdline.AppVersion,
		Build:     cmdline.BuildNumber,
		Git:       cmdline.GitVersion,
		Modified:  cmdline.VCSModified,
	})
}
