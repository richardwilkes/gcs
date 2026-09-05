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
	"errors"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// sampleNameCount is how many sample names the name generator editor shows at a time.
const sampleNameCount = 10

// nameGeneratorSamplesPanel shows names generated from the definition being edited, so that the effect of a change can
// be seen at once, or the reason the definition cannot generate names yet.
type nameGeneratorSamplesPanel struct {
	unison.Panel
	dockable *nameGeneratorEditorDockable
	label    *wrappingLabel
	// shownHash is the hash of the definition the samples were last generated from, so that a Sync prompted by
	// something other than a change to the definition -- a field committing an unchanged value, the rebuild that
	// follows every structural edit -- does not generate them again.
	shownHash uint64
}

func newNameGeneratorSamplesPanel(d *nameGeneratorEditorDockable) *nameGeneratorSamplesPanel {
	p := &nameGeneratorSamplesPanel{dockable: d}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(geom.Insets{Bottom: unison.StdVSpacing * 2}))
	p.SetLayout(&unison.FlexLayout{Columns: 1, VSpacing: unison.StdVSpacing})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	button := unison.NewSVGButton(svg.Randomize)
	button.Tooltip = newWrappedTooltip(i18n.Text("Generate new sample names"))
	button.ClickCallback = p.refresh
	p.AddChild(newEditorSectionHeader(i18n.Text("Sample Names"),
		i18n.Text("Names generated from the current definition,\nso changes can be checked as they are made"), button))
	p.label = newWrappingLabel()
	p.label.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.NewUniformInsets(1), false),
		unison.NewEmptyBorder(geom.NewSymmetricInsets(unison.StdHSpacing, unison.StdVSpacing))))
	p.label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	originalDraw := p.label.DrawCallback
	p.label.DrawCallback = func(c *unison.Canvas, r geom.Rect) {
		c.DrawRect(r, unison.ThemeAboveSurface.Paint(c, r, paintstyle.Fill))
		originalDraw(c, r)
	}
	p.AddChild(p.label)
	p.refresh()
	return p
}

// refresh generates a fresh set of sample names from the definition as it stands. When the definition cannot generate
// names, the reason is shown in its place.
func (p *nameGeneratorSamplesPanel) refresh() {
	p.shownHash = gurps.Hash64(p.dockable.model)
	names, err := p.dockable.model.SampleNames(sampleNameCount)
	if err != nil {
		p.label.setText(errorMessage(err), unison.ThemeWarning)
		return
	}
	p.label.setText(strings.Join(names, ", "), unison.DefaultLabelTheme.OnBackgroundInk)
}

// Sync implements Syncer, so that a change typed into any field shows up in the samples as soon as it is made. Samples
// generated from the definition as it now stands are left alone.
func (p *nameGeneratorSamplesPanel) Sync() {
	if p.shownHash != gurps.Hash64(p.dockable.model) {
		p.refresh()
	}
}

// errorMessage returns the message of the error without the call stack an errs.Error would otherwise include, since
// the message is meant to be read by the user.
func errorMessage(err error) string {
	var msg string
	if detailed, ok := errors.AsType[*errs.Error](err); ok {
		msg = detailed.Message()
	} else {
		msg = err.Error()
	}
	return xstrings.FirstToUpper(msg)
}
