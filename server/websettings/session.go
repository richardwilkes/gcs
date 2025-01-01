// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package websettings

import (
	"time"

	"github.com/richardwilkes/toolbox/tid"
)

// Constants
const (
	MaxSessionDuration = time.Hour * 24 * 30
	SessionGracePeriod = time.Minute * 10
)

// Session holds a session's information.
type Session struct {
	ID       tid.TID   `json:"id"`
	UserKey  string    `json:"user"`
	Issued   time.Time `json:"issued"`
	LastUsed time.Time `json:"last_used"`
}

// Clone creates a copy of this session.
func (s *Session) Clone() *Session {
	other := *s
	return &other
}

// Expired returns true if this session has expired.
func (s *Session) Expired() bool {
	return time.Since(s.Issued) > MaxSessionDuration && time.Since(s.LastUsed) > SessionGracePeriod
}
