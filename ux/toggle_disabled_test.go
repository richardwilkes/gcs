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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestToggleDisabledSkipsChoiceContainers verifies that a choice container can't be disabled from the list, since its
// editor offers no way to enable it again, while one that is already disabled can still be enabled.
func TestToggleDisabledSkipsChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewTraitChoiceContainer(nil, nil)
	plain := gurps.NewTrait(nil, nil, false)
	template := newTestTemplateWithTraits(choice, plain)
	table := template.Traits.Table

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.False(canToggleDisabled(table), "a choice alone must not offer to be disabled")

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true, plain.ID(): true})
	c.True(canToggleDisabled(table), "the rest of the selection may still be toggled")
	toggleDisabled(template, table)
	c.False(choice.IsDisabled(), "the choice must be left enabled")
	c.True(plain.IsDisabled(), "the plain trait must be disabled")

	choice.SetDisabled(true)
	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.True(canToggleDisabled(table), "a disabled choice must offer to be enabled")
	toggleDisabled(template, table)
	c.False(choice.IsDisabled(), "the disabled choice must be enabled")
}
