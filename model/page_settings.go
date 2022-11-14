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

// PageSettings holds page settings.
type PageSettings struct {
	Size         PaperSize        `json:"paper_size"`
	Orientation  PaperOrientation `json:"orientation"`
	TopMargin    PaperLength      `json:"top_margin"`
	LeftMargin   PaperLength      `json:"left_margin"`
	BottomMargin PaperLength      `json:"bottom_margin"`
	RightMargin  PaperLength      `json:"right_margin"`
}

// NewPageSettings returns new settings with factory defaults.
func NewPageSettings() *PageSettings {
	return &PageSettings{
		Size:         LetterPaperSize,
		Orientation:  Portrait,
		TopMargin:    PaperLength{Length: 0.25, Units: InchPaperUnits},
		LeftMargin:   PaperLength{Length: 0.25, Units: InchPaperUnits},
		BottomMargin: PaperLength{Length: 0.25, Units: InchPaperUnits},
		RightMargin:  PaperLength{Length: 0.25, Units: InchPaperUnits},
	}
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (p *PageSettings) EnsureValidity() {
	p.Size = p.Size.EnsureValid()
	p.Orientation = p.Orientation.EnsureValid()
	p.TopMargin.EnsureValidity()
	p.LeftMargin.EnsureValidity()
	p.BottomMargin.EnsureValidity()
	p.RightMargin.EnsureValidity()
}

// Clone a copy of this.
func (p *PageSettings) Clone() *PageSettings {
	clone := *p
	return &clone
}
