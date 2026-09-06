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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// newToolbar creates the standard dockable toolbar shell: a panel with the surface-edge line along its bottom inside
// the standard insets, sized to fill the width of its parent. Callers add their children and then either call
// finishToolbarLayout for a single row or install their own layout for multi-row toolbars.
func newToolbar() *unison.Panel {
	toolbar := unison.NewPanel()
	toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	return toolbar
}

// finishToolbarLayout lays out the toolbar's current children in a single row. Call it after the last child has been
// added, since the column count is taken from the children present at the time of the call.
func finishToolbarLayout(toolbar *unison.Panel) {
	toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
}

// addUIScaleField adds the standard UI scale field, bounded by gurps.InitialUIScaleMin and gurps.InitialUIScaleMax,
// to the toolbar. initial supplies the default scale, get and set access the current scale, and scroller is the panel
// whose content is scaled. adjustForDisplayPPI should be set for sheet-like content that is laid out in points.
func addUIScaleField(toolbar *unison.Panel, initial, get func() int, set func(int), adjustForDisplayPPI bool, scroller *unison.ScrollPanel) *PercentageField {
	field := NewScaleField(gurps.InitialUIScaleMin, gurps.InitialUIScaleMax, initial, get, set, nil, false,
		adjustForDisplayPPI, scroller)
	toolbar.AddChild(field)
	return field
}
