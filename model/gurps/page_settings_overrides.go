// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/paper"
)

// PageSettingsOverrides holds page setting overrides.
type PageSettingsOverrides struct {
	Size         *string
	Orientation  *paper.Orientation
	TopMargin    *paper.Length
	LeftMargin   *paper.Length
	BottomMargin *paper.Length
	RightMargin  *paper.Length
}

// ParseSize and set the override, if applicable.
func (p *PageSettingsOverrides) ParseSize(in string) {
	if w, h, valid := ParsePageSize(in); valid {
		size := ToPageSize(w, h)
		p.Size = &size
	}
}

// ParseOrientation and set the override, if applicable.
func (p *PageSettingsOverrides) ParseOrientation(in string) {
	in = strings.TrimSpace(in)
	if in != "" {
		orientation := paper.ExtractOrientation(in)
		p.Orientation = &orientation
	}
}

// ParseTopMargin and set the override, if applicable.
func (p *PageSettingsOverrides) ParseTopMargin(in string) {
	p.TopMargin = parseLengthString(in)
}

// ParseLeftMargin and set the override, if applicable.
func (p *PageSettingsOverrides) ParseLeftMargin(in string) {
	p.LeftMargin = parseLengthString(in)
}

// ParseBottomMargin and set the override, if applicable.
func (p *PageSettingsOverrides) ParseBottomMargin(in string) {
	p.BottomMargin = parseLengthString(in)
}

// ParseRightMargin and set the override, if applicable.
func (p *PageSettingsOverrides) ParseRightMargin(in string) {
	p.RightMargin = parseLengthString(in)
}

func parseLengthString(in string) *paper.Length {
	in = strings.TrimSpace(in)
	if in == "" {
		return nil
	}
	length := paper.LengthFromString(in)
	return &length
}

// Apply the overrides to a Page.
func (p *PageSettingsOverrides) Apply(page *PageSettings) {
	if p.Size != nil {
		page.Size = *p.Size
	}
	if p.Orientation != nil {
		page.Orientation = *p.Orientation
	}
	if p.TopMargin != nil {
		page.TopMargin = *p.TopMargin
	}
	if p.LeftMargin != nil {
		page.LeftMargin = *p.LeftMargin
	}
	if p.BottomMargin != nil {
		page.BottomMargin = *p.BottomMargin
	}
	if p.RightMargin != nil {
		page.RightMargin = *p.RightMargin
	}
}
