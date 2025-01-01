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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/cell"
	"github.com/richardwilkes/unison/enums/align"
)

// PageRefCellAlias is used an alias to request the page reference cell, if any.
const PageRefCellAlias = -10

// These constants are used to specify images to use in column headers.
const (
	HeaderCheckmark     = "checkmark"
	HeaderCoins         = "coins"
	HeaderWeight        = "weight"
	HeaderBookmark      = "bookmark"
	HeaderDatabase      = "database"
	HeaderStackedCoins  = "stacked-coins"
	HeaderStackedWeight = "stacked-weight"
)

// HeaderData holds data for creating a column header's visual representation.
type HeaderData struct {
	Title           string
	Detail          string
	Less            func(a, b string) bool
	TitleIsImageKey bool
	Primary         bool
}

// CellData holds data for creating a cell's visual representation.
type CellData struct {
	Type              cell.Type
	Disabled          bool
	Dim               bool
	Checked           bool
	Alignment         align.Enum
	Primary           string
	Secondary         string
	Tooltip           string
	UnsatisfiedReason string
	TemplateInfo      string
	InlineTag         string
}

// ForSort returns a string that can be used to sort or search against for this data.
func (c *CellData) ForSort() string {
	switch c.Type {
	case cell.Text:
		if c.Secondary != "" {
			return c.Primary + "\n" + c.Secondary
		}
		return c.Primary
	case cell.Toggle:
		if c.Checked {
			return "√"
		}
	case cell.PageRef, cell.Tags, cell.Markdown:
		return c.Primary
	}
	return ""
}
