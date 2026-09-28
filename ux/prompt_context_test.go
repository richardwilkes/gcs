// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPromptOperationTitle verifies that a prompt is titled with its operation's name and its step, and with whichever
// of them it has when it lacks the other, and that setting the step leaves the operation it was set on alone.
func TestPromptOperationTitle(t *testing.T) {
	c := check.New(t)
	op := promptOperation{name: "Apply Template", description: "Applying template Knight to Sir Bob"}
	c.Equal("Apply Template: Modifiers", op.at("Modifiers").title())
	c.Equal("Apply Template", op.title(), "setting the step must not change the operation it was set on")
	c.Equal("Substitutions", promptOperation{}.at("Substitutions").title())
	c.Equal("", promptOperation{}.title())
}
