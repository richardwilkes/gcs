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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/unison"
)

var _ unison.TableRowData[*tocNode] = &tocNode{}

type tocNode struct {
	id         tid.TID
	owner      *PDFDockable
	parent     *tocNode
	title      string
	pageNumber int
	children   []*tocNode
}

func newTOC(owner *PDFDockable, parent *tocNode, toc []*PDFTableOfContents) []*tocNode {
	if len(toc) == 0 {
		return nil
	}
	nodes := make([]*tocNode, len(toc))
	for i, one := range toc {
		nodes[i] = &tocNode{
			id:         gurps.IDForPDFTOC(owner.path, one.Title, one.PageNumber),
			owner:      owner,
			parent:     parent,
			title:      one.Title,
			pageNumber: one.PageNumber,
			children:   newTOC(owner, parent, one.Children),
		}
	}
	return nodes
}

func (n *tocNode) CloneForTarget(_ unison.Paneler, _ *tocNode) *tocNode {
	return nil // Not used
}

func (n *tocNode) ID() tid.TID {
	return n.id
}

func (n *tocNode) Parent() *tocNode {
	return n.parent
}

func (n *tocNode) SetParent(_ *tocNode) {
	// Not used
}

func (n *tocNode) Container() bool {
	return n.CanHaveChildren()
}

func (n *tocNode) CanHaveChildren() bool {
	return len(n.children) != 0
}

func (n *tocNode) Children() []*tocNode {
	return n.children
}

func (n *tocNode) SetChildren(_ []*tocNode) {
	// Not used
}

func (n *tocNode) CellDataForSort(_ int) string {
	return n.title
}

func (n *tocNode) ColumnCell(_, _ int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	var img *unison.SVG
	if n.CanHaveChildren() {
		if n.IsOpen() {
			img = svg.OpenFolder
		} else {
			img = svg.ClosedFolder
		}
	} else {
		img = svg.Bookmark
	}
	size := unison.LabelFont.Size() + 5
	label := unison.NewLabel()
	label.OnBackgroundInk = foreground
	label.SetTitle(n.title)
	label.Drawable = &unison.DrawableSVG{
		SVG:  img,
		Size: unison.NewSize(size, size),
	}
	return label
}

func (n *tocNode) IsOpen() bool {
	return gurps.IsNodeOpen(n)
}

func (n *tocNode) SetOpen(open bool) {
	if gurps.SetNodeOpen(n, open) {
		n.owner.adjustTableSizeEventually()
	}
}
