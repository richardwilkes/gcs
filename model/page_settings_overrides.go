/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

import (
	"strings"
)

// PageSettingsOverrides holds page setting overrides.
type PageSettingsOverrides struct {
	Size         *PaperSize
	Orientation  *PaperOrientation
	TopMargin    *PaperLength
	LeftMargin   *PaperLength
	BottomMargin *PaperLength
	RightMargin  *PaperLength
}

// ParseSize and set the override, if applicable.
func (p *PageSettingsOverrides) ParseSize(in string) {
	in = strings.TrimSpace(in)
	if in != "" {
		size := ExtractPaperSize(in)
		p.Size = &size
	}
}

// ParseOrientation and set the override, if applicable.
func (p *PageSettingsOverrides) ParseOrientation(in string) {
	in = strings.TrimSpace(in)
	if in != "" {
		orientation := ExtractPaperOrientation(in)
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

func parseLengthString(in string) *PaperLength {
	in = strings.TrimSpace(in)
	if in == "" {
		return nil
	}
	length := PaperLengthFromString(in)
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
