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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

// TestNewToolbar verifies the shared toolbar shell every dockable builds on: the standard insets plus the one-pixel
// surface-edge line along the bottom, and layout data that stretches it across its parent.
func TestNewToolbar(t *testing.T) {
	c := check.New(t)
	toolbar := newToolbar()
	c.NotNil(toolbar.Border(), "toolbar has a border")
	want := unison.StdInsets()
	want.Bottom++
	c.Equal(want, toolbar.Border().Insets(), "border is the standard insets plus the bottom edge line")
	data, ok := toolbar.LayoutData().(*unison.FlexLayoutData)
	c.True(ok, "layout data is flex layout data")
	c.Equal(align.Fill, data.HAlign, "toolbar fills horizontally")
	c.True(data.HGrab, "toolbar grabs horizontal space")
	c.Equal(0, len(toolbar.Children()), "toolbar starts empty")
}

// TestFinishToolbarLayout verifies the single-row layout takes its column count from the children present when it is
// installed, so every child lands on the one row.
func TestFinishToolbarLayout(t *testing.T) {
	c := check.New(t)
	toolbar := newToolbar()
	for range 3 {
		toolbar.AddChild(unison.NewPanel())
	}
	finishToolbarLayout(toolbar)
	layout, ok := toolbar.Layout().(*unison.FlexLayout)
	c.True(ok, "layout is a flex layout")
	c.Equal(3, layout.Columns, "one column per child")
	c.Equal(float32(unison.StdHSpacing), layout.HSpacing, "standard horizontal spacing")
	c.Equal(float32(0), layout.VSpacing, "no vertical spacing on a single row")
}

// TestAddUIScaleField verifies the scale field is appended to the toolbar, is bound to the caller's accessors and the
// standard UI scale bounds, and scales the scroller's content when edited, honoring the display PPI adjustment that the
// sheet-like views ask for.
func TestAddUIScaleField(t *testing.T) {
	c := check.New(t)
	general := gurps.GlobalSettings().General
	savedResolution := general.MonitorResolution
	general.MonitorResolution = 144 // twice the 72 PPI baseline, so the adjusted scale is easy to predict
	t.Cleanup(func() { general.MonitorResolution = savedResolution })

	for _, one := range []struct {
		name          string
		adjustForPPI  bool
		wantPPIFactor float32
	}{
		{name: "unadjusted", adjustForPPI: false, wantPPIFactor: 1},
		{name: "adjusted for display PPI", adjustForPPI: true, wantPPIFactor: 2},
	} {
		t.Run(one.name, func(_ *testing.T) {
			scroller := unison.NewScrollPanel()
			content := unison.NewPanel()
			scroller.SetContent(content, behavior.Unmodified, behavior.Unmodified)
			toolbar := newToolbar()
			toolbar.AddChild(NewDefaultInfoPop())
			scale := 100
			var setTo []int
			field := addUIScaleField(toolbar, func() int { return 75 }, func() int { return scale },
				func(v int) { scale = v; setTo = append(setTo, v) }, one.adjustForPPI, scroller)
			children := toolbar.Children()
			c.Equal(2, len(children), "scale field was added to the toolbar")
			c.True(children[1] == field.AsPanel(), "scale field is the last child")
			c.Equal(gurps.InitialUIScaleMin, field.Min(), "standard minimum scale")
			c.Equal(gurps.InitialUIScaleMax, field.Max(), "standard maximum scale")
			c.Equal("100%", field.Text(), "field shows the current scale")

			field.SetText("200")
			c.Equal([]int{200}, setTo, "edit was handed to the setter")
			c.Equal(geom.NewPoint(2*one.wantPPIFactor, 2*one.wantPPIFactor), content.Scale(),
				"scroller content was scaled")

			field.SetText("1000")
			c.Equal([]int{200, gurps.InitialUIScaleMax}, setTo, "out-of-range edit was clamped to the maximum")
		})
	}
}
