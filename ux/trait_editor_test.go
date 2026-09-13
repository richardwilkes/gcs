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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"testing"
)

func TestTraitEditorContainerTypes(t *testing.T) {
	for _, tc := range []struct {
		name  string
		owner gurps.DataOwner
		fixed bool
	}{
		{name: "sheet", owner: gurps.NewEntity()},
		{name: "template", owner: gurps.NewTemplate(), fixed: true},
		{name: "library", fixed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			trait := gurps.NewTrait(tc.owner, nil, true)
			e, content := buildEditorContent(nil, trait, initTraitEditor)
			popups := panelsOfType[*unison.PopupMenu[container.Type]](content)
			if len(popups) != 1 {
				t.Fatalf("got %d container popups", len(popups))
			}
			popup := popups[0]
			c.Equal(tc.fixed, popup.IndexOfItem(container.FixedCost) >= 0)
			for _, kind := range container.Types {
				if kind != container.FixedCost {
					c.True(popup.IndexOfItem(kind) >= 0)
				}
			}
			if tc.fixed {
				popup.Select(container.FixedCost)
				c.Equal(container.FixedCost, e.editorData.ContainerType)
			}
		})
	}
}
