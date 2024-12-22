// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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

var stdPaperSizes = []paperSize{
	{
		name:   "letter",
		width:  paper.Length{Length: 8.5, Units: paper.Inch},
		height: paper.Length{Length: 11, Units: paper.Inch},
	},
	{
		name:   "legal",
		width:  paper.Length{Length: 8.5, Units: paper.Inch},
		height: paper.Length{Length: 14, Units: paper.Inch},
	},
	{
		name:   "tabloid",
		width:  paper.Length{Length: 11, Units: paper.Inch},
		height: paper.Length{Length: 17, Units: paper.Inch},
	},
	{
		name:   "a0",
		width:  paper.Length{Length: 841, Units: paper.Millimeter},
		height: paper.Length{Length: 1189, Units: paper.Millimeter},
	},
	{
		name:   "a1",
		width:  paper.Length{Length: 594, Units: paper.Millimeter},
		height: paper.Length{Length: 841, Units: paper.Millimeter},
	},
	{
		name:   "a2",
		width:  paper.Length{Length: 420, Units: paper.Millimeter},
		height: paper.Length{Length: 594, Units: paper.Millimeter},
	},
	{
		name:   "a3",
		width:  paper.Length{Length: 297, Units: paper.Millimeter},
		height: paper.Length{Length: 420, Units: paper.Millimeter},
	},
	{
		name:   "a4",
		width:  paper.Length{Length: 210, Units: paper.Millimeter},
		height: paper.Length{Length: 297, Units: paper.Millimeter},
	},
	{
		name:   "a5",
		width:  paper.Length{Length: 148, Units: paper.Millimeter},
		height: paper.Length{Length: 210, Units: paper.Millimeter},
	},
	{
		name:   "a6",
		width:  paper.Length{Length: 105, Units: paper.Millimeter},
		height: paper.Length{Length: 148, Units: paper.Millimeter},
	},
}

type paperSize struct {
	name   string
	width  paper.Length
	height paper.Length
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
		Size:         stdPaperSizes[0].name,
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
		return stdPaperSizes[0].name
	}
	return ToPageSize(w, h)
}

// MustParsePageSize parses a page size string and returns the width and height in inches.
func MustParsePageSize(size string) (width, height paper.Length) {
	var valid bool
	if width, height, valid = ParsePageSize(size); valid {
		return width, height
	}
	return stdPaperSizes[0].width, stdPaperSizes[0].height
}

// ParsePageSize parses a page size string and returns the width and height in inches.
func ParsePageSize(size string) (width, height paper.Length, valid bool) {
	size = strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(strings.TrimSpace(size)), "na-"), "iso-")
	for _, one := range stdPaperSizes {
		if size == one.name {
			return one.width, one.height, true
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
	for _, one := range stdPaperSizes {
		if width == one.width && height == one.height {
			return one.name
		}
	}
	if outOfPaperRange(width) || outOfPaperRange(height) {
		return stdPaperSizes[0].name
	}
	return width.String() + " x " + height.String()
}
