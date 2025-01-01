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

const (
	minPaperLengthPixels = 100
	maxPaperLengthPixels = 32000
)

// StdPaperSizes holds the standard paper sizes.
var StdPaperSizes = []PaperSize{
	{
		Name:   "letter",
		Width:  paper.Length{Length: 8.5, Units: paper.Inch},
		Height: paper.Length{Length: 11, Units: paper.Inch},
	},
	{
		Name:   "legal",
		Width:  paper.Length{Length: 8.5, Units: paper.Inch},
		Height: paper.Length{Length: 14, Units: paper.Inch},
	},
	{
		Name:   "tabloid",
		Width:  paper.Length{Length: 11, Units: paper.Inch},
		Height: paper.Length{Length: 17, Units: paper.Inch},
	},
	{
		Name:   "a0",
		Width:  paper.Length{Length: 841, Units: paper.Millimeter},
		Height: paper.Length{Length: 1189, Units: paper.Millimeter},
	},
	{
		Name:   "a1",
		Width:  paper.Length{Length: 594, Units: paper.Millimeter},
		Height: paper.Length{Length: 841, Units: paper.Millimeter},
	},
	{
		Name:   "a2",
		Width:  paper.Length{Length: 420, Units: paper.Millimeter},
		Height: paper.Length{Length: 594, Units: paper.Millimeter},
	},
	{
		Name:   "a3",
		Width:  paper.Length{Length: 297, Units: paper.Millimeter},
		Height: paper.Length{Length: 420, Units: paper.Millimeter},
	},
	{
		Name:   "a4",
		Width:  paper.Length{Length: 210, Units: paper.Millimeter},
		Height: paper.Length{Length: 297, Units: paper.Millimeter},
	},
	{
		Name:   "a5",
		Width:  paper.Length{Length: 148, Units: paper.Millimeter},
		Height: paper.Length{Length: 210, Units: paper.Millimeter},
	},
	{
		Name:   "a6",
		Width:  paper.Length{Length: 105, Units: paper.Millimeter},
		Height: paper.Length{Length: 148, Units: paper.Millimeter},
	},
}

// PageInfoProvider is an interface for types that have page information.
type PageInfoProvider interface {
	PageSettings() *PageSettings
	PageTitle() string
	ModifiedOnString() string
}

// PaperSize holds details about a standard paper size.
type PaperSize struct {
	Name   string
	Width  paper.Length
	Height paper.Length
}

// PageSettings holds page settings.
type PageSettings struct {
	Size         string            `json:"paper_size"`
	Orientation  paper.Orientation `json:"orientation"`
	TopMargin    paper.Length      `json:"top_margin"`
	LeftMargin   paper.Length      `json:"left_margin"`
	BottomMargin paper.Length      `json:"bottom_margin"`
	RightMargin  paper.Length      `json:"right_margin"`
}

// NewPageSettings returns new settings with factory defaults.
func NewPageSettings() *PageSettings {
	return &PageSettings{
		Size:         StdPaperSizes[0].Name,
		Orientation:  paper.Portrait,
		TopMargin:    paper.Length{Length: 0.25, Units: paper.Inch},
		LeftMargin:   paper.Length{Length: 0.25, Units: paper.Inch},
		BottomMargin: paper.Length{Length: 0.25, Units: paper.Inch},
		RightMargin:  paper.Length{Length: 0.25, Units: paper.Inch},
	}
}

// EnsureValidity checks the current settings for validity and if they aren't valid, makes them so.
func (p *PageSettings) EnsureValidity() {
	p.Size = EnsurePageSizeIsValid(p.Size)
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

// EnsurePageSizeIsValid ensures the given page size is valid and returns the corrected value if not.
func EnsurePageSizeIsValid(in string) string {
	w, h, ok := ParsePageSize(in)
	if !ok {
		return StdPaperSizes[0].Name
	}
	return ToPageSize(w, h)
}

// MustParsePageSize parses a page size string and returns the width and height in inches.
func MustParsePageSize(size string) (width, height paper.Length) {
	var valid bool
	if width, height, valid = ParsePageSize(size); valid {
		return width, height
	}
	return StdPaperSizes[0].Width, StdPaperSizes[0].Height
}

// ParsePageSize parses a page size string and returns the width and height in inches.
func ParsePageSize(size string) (width, height paper.Length, valid bool) {
	size = strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(size)), "na-"), "iso-")
	for _, one := range StdPaperSizes {
		if size == one.Name {
			return one.Width, one.Height, true
		}
	}
	i := strings.Index(size, "x")
	if i == -1 {
		return
	}
	var err error
	if width, err = paper.ParseLengthFromString(size[:i]); err != nil || outOfPaperRange(width) {
		return
	}
	if height, err = paper.ParseLengthFromString(size[i+1:]); err != nil || outOfPaperRange(height) {
		return
	}
	return width, height, true
}

func outOfPaperRange(length paper.Length) bool {
	pixels := length.Pixels()
	return pixels < minPaperLengthPixels || pixels > maxPaperLengthPixels
}

// ToPageSize converts a width and height to a page size string.
func ToPageSize(width, height paper.Length) string {
	for _, one := range StdPaperSizes {
		if width == one.Width && height == one.Height {
			return one.Name
		}
	}
	if outOfPaperRange(width) || outOfPaperRange(height) {
		return StdPaperSizes[0].Name
	}
	return width.String() + " x " + height.String()
}
