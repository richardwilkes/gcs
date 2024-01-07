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

package gurps

import "github.com/richardwilkes/gcs/v5/model/fxp"

// WebServerSettings holds the settings for the embedded web server.
type WebServerSettings struct {
	Enabled             bool    `json:"enabled"`
	Address             string  `json:"address,omitempty"`
	CertFile            string  `json:"cert_file,omitempty"`
	KeyFile             string  `json:"key_file,omitempty"`
	ShutdownGracePeriod fxp.Int `json:"shutdown_grace_period,omitempty"`
}
