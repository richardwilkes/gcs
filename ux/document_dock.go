// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var (
	_ unison.Dockable = &DocumentDock{}
	_ KeyedDockable   = &DocumentDock{}
)

// DocumentsDockKey is the key used to store the DocumentDock in the top Dock.
const DocumentsDockKey = "documents"

// DocumentDock holds the document dock.
type DocumentDock struct {
	*unison.Dock
}

// NewDocumentDock creates a new DocumentDock.
func NewDocumentDock() *DocumentDock {
	d := &DocumentDock{
		Dock: unison.NewDock(),
	}
	d.Self = d
	return d
}

// DockKey implements KeyedDockable.
func (d *DocumentDock) DockKey() string {
	return DocumentsDockKey
}

// TitleIcon implements unison.Dockable
func (d *DocumentDock) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  unison.DocumentSVG,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (d *DocumentDock) Title() string {
	return i18n.Text("Document Workspace")
}

// Tooltip implements unison.Dockable
func (d *DocumentDock) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (d *DocumentDock) Modified() bool {
	return false
}
