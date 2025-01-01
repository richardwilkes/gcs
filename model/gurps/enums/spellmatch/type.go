// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package spellmatch

import "github.com/richardwilkes/toolbox/errs"

// Matcher defines the methods that a spell matcher must implement.
type Matcher interface {
	Matches(map[string]string, string) bool
	MatchesList(map[string]string, ...string) bool
}

// MatchForType applies the matcher and returns the result.
func (enum Type) MatchForType(matcher Matcher, replacements map[string]string, name, powerSource string, colleges []string) bool {
	switch enum {
	case AllColleges:
		return true
	case CollegeName:
		return matcher.MatchesList(replacements, colleges...)
	case PowerSource:
		return matcher.Matches(replacements, powerSource)
	case Name:
		return matcher.Matches(replacements, name)
	default:
		errs.Log(errs.New("unhandled spell match type"), "type", int(enum))
		return false
	}
}
