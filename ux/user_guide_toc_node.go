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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/userguide"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// userGuideClosedKeyPrefix starts the keys the closed groups of the User Guide's table of contents are recorded under
// (see gurps.IsClosed), which go on with the group's title, preceded by those of the groups holding it and a slash.
const userGuideClosedKeyPrefix = "userguide:"

var _ unison.TableRowData[*userGuideTOCNode] = &userGuideTOCNode{}

// userGuideTOCNode is a row of the User Guide's table of contents: a page, or a group of rows, which starts out open.
type userGuideTOCNode struct {
	owner     *userGuideDockable
	parent    *userGuideTOCNode
	id        tid.TID
	title     string
	page      string // "" for a group
	closedKey string // "" for a page
	children  []*userGuideTOCNode
}

// newUserGuideTOC returns the rows for the entries, beneath parent. keyPrefix is what the closed keys of the groups
// among them are formed from: "" at the top, else the closed key of parent, less userGuideClosedKeyPrefix, and a slash.
func newUserGuideTOC(owner *userGuideDockable, parent *userGuideTOCNode, entries []*userguide.TOCEntry, keyPrefix string) []*userGuideTOCNode {
	if len(entries) == 0 {
		return nil
	}
	nodes := make([]*userGuideTOCNode, len(entries))
	for i, entry := range entries {
		node := &userGuideTOCNode{
			owner:  owner,
			parent: parent,
			id:     tid.MustNewTID(kinds.TableOfContents),
			title:  entry.Title,
			page:   entry.Page,
		}
		if len(entry.Children) != 0 {
			key := keyPrefix + entry.Title
			node.closedKey = userGuideClosedKeyPrefix + key
			node.children = newUserGuideTOC(owner, node, entry.Children, key+"/")
		}
		nodes[i] = node
	}
	return nodes
}

func (n *userGuideTOCNode) CloneForTarget(_ unison.Paneler, _ *userGuideTOCNode) *userGuideTOCNode {
	return nil // Not used
}

func (n *userGuideTOCNode) ID() tid.TID {
	return n.id
}

func (n *userGuideTOCNode) Parent() *userGuideTOCNode {
	return n.parent
}

func (n *userGuideTOCNode) SetParent(_ *userGuideTOCNode) {
	// Not used
}

func (n *userGuideTOCNode) CanHaveChildren() bool {
	return len(n.children) != 0
}

func (n *userGuideTOCNode) Children() []*userGuideTOCNode {
	return n.children
}

func (n *userGuideTOCNode) SetChildren(_ []*userGuideTOCNode) {
	// Not used
}

func (n *userGuideTOCNode) CellDataForSort(_ int) string {
	return n.title
}

func (n *userGuideTOCNode) ColumnCell(_, _ int, foreground, _ unison.Ink, _, _, _ bool) unison.Paneler {
	return tocCell(n.title, n.CanHaveChildren(), n.IsOpen(), svg.MarkdownFile, foreground)
}

func (n *userGuideTOCNode) IsOpen() bool {
	return n.CanHaveChildren() && !gurps.IsClosed(n.closedKey)
}

func (n *userGuideTOCNode) SetOpen(open bool) {
	if n.CanHaveChildren() && gurps.SetClosedState(n.closedKey, !open) {
		n.owner.adjustTableSizeEventually()
	}
}
