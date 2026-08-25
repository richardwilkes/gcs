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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// sheetLayoutNodeKey is the client data key the layout node a panel was built for is recorded under. Every panel the
// builder returns carries one, so that the layout editor and the page exporter can map a panel back into the tree.
const sheetLayoutNodeKey = "sheetLayoutNode"

// sheetLayoutContainerKey is the client data key the Row or Column node a container panel the builder made was built
// from is recorded under. Only the container panels the builder creates itself carry one, which is what tells them
// apart from the block panels the leaf function hands back: the node recorded under sheetLayoutNodeKey can't be used
// for that, since a container left with a single child is dropped and its node recorded on the child that took its
// place. Nor is that node always the container's own: when an outer container was dropped onto this one, the node
// recorded here is still the one whose children this panel's children were built from, which is what anything that
// deals with the pairs of them, like the seams the layout editor collects, has to be told.
const sheetLayoutContainerKey = "sheetLayoutContainer"

// sheetLayoutSquareKey is the client data key the Block node whose Square flag governs a panel's slot is recorded
// under. Only a panel whose slot takes its width from the height of the row it is in carries one, and the node it names
// is always a Block, even where the node recorded under sheetLayoutNodeKey is the container that block was left alone
// in: the flag lives on the block, so that is the node the editor has to clear it on.
const sheetLayoutSquareKey = "sheetLayoutSquare"

// layoutLeafFunc returns the panel to show for the block with the given key, or nil if that block is not to be shown
// at all -- because it has nothing to say, or because this kind of sheet has no such block.
type layoutLeafFunc func(key string) unison.Paneler

// buildLayoutBands builds one panel per band of the given layout root, in order. Bands that come back empty, because
// none of the blocks in them had anything to show, are omitted, so the result may be shorter than the root's list of
// children.
func buildLayoutBands(root *gurps.SheetLayoutNode, leaf layoutLeafFunc) []*unison.Panel {
	if root == nil {
		return nil
	}
	bands := make([]*unison.Panel, 0, len(root.Children))
	for _, child := range root.Children {
		if band := buildLayoutNode(child, leaf); !xreflect.IsNil(band) {
			bands = append(bands, band.AsPanel())
		}
	}
	return bands
}

// buildLayoutNode builds the panel for one node of a sheet layout and everything below it, returning nil if nothing in
// it had anything to show. A Column becomes a panel that stacks its children, a Row one that places them side by side
// with their widths in proportion to their weights, and a Block whatever the leaf function hands back for its key.
//
// The builder is the sole authority on the layout data of the panels it returns: each of them, leaves included, is
// given a fresh unison.FlexLayoutData that fills its slot and carries the minimum height its node calls for, replacing
// whatever the panel had. Each of them also records, under sheetLayoutNodeKey, the node that governs the slot it
// occupies. That is normally the node it was built for, but a container left with a single child once the rest of them
// came back nil is dropped and its child returned in its place, and the child then records the container's node, since
// it is the container's weight and minimum height that decide how the slot it took over is sized. A minimum height of
// the child's own still wins over the container's, matching the way the model collapses single-child containers.
//
// A Block's square flag travels with its panel in the same way. A container is never square, but the child that takes a
// dropped container's place keeps the flag of the block it was built from rather than picking up the container's, since
// it is that block's content the row would be squaring.
func buildLayoutNode(node *gurps.SheetLayoutNode, leaf layoutLeafFunc) unison.Paneler {
	if node == nil {
		return nil
	}
	if node.Type == layoutnode.Block {
		panel := leaf(node.Key)
		if xreflect.IsNil(panel) {
			return nil
		}
		// The blocks that aren't lists are built once and used again by every rebuild, so a flag that is no longer set
		// has to be taken off the panel rather than merely not being put on it.
		if node.Square {
			panel.AsPanel().ClientData()[sheetLayoutSquareKey] = node
		} else {
			delete(panel.AsPanel().ClientData(), sheetLayoutSquareKey)
		}
		return applyLayoutNode(panel, node, node.MinHeight.Pixels())
	}
	panels := make([]unison.Paneler, 0, len(node.Children))
	weights := make([]fxp.Int, 0, len(node.Children))
	squares := make([]bool, 0, len(node.Children))
	for _, child := range node.Children {
		if built := buildLayoutNode(child, leaf); !xreflect.IsNil(built) {
			panels = append(panels, built)
			weights = append(weights, child.Weight)
			squares = append(squares, squareLayoutNodeOf(built.AsPanel()) != nil)
		}
	}
	switch len(panels) {
	case 0:
		return nil
	case 1:
		only := panels[0].AsPanel()
		minHeight := node.MinHeight.Pixels()
		if data, ok := only.LayoutData().(*unison.FlexLayoutData); ok && data.MinSize.Height > 0 {
			minHeight = data.MinSize.Height
		}
		return applyLayoutNode(panels[0], node, minHeight)
	default:
		container := unison.NewPanel()
		if node.Type == layoutnode.Row {
			container.SetLayout(&weightedRowLayout{
				weights:   weights,
				square:    squares,
				hSpacing:  1,
				minHeight: node.MinHeight.Pixels(),
			})
		} else {
			container.SetLayout(&unison.FlexLayout{
				Columns:  1,
				VSpacing: 1,
			})
			for _, panel := range panels {
				// A column standing beside something taller is given the full height of the row it is in, and its
				// children share out the extra so that the bottom of the last of them is the bottom of the row. Without
				// this they would keep their natural heights and leave the page showing below them. Only the children
				// of a column are given this: a root band that grabbed height would be stretched to the bottom of the
				// paper on an exported page, since those pages are forced to the full height of the sheet.
				if data, ok := panel.AsPanel().LayoutData().(*unison.FlexLayoutData); ok {
					data.VGrab = true
				}
			}
		}
		for _, panel := range panels {
			container.AddChild(panel)
		}
		container.ClientData()[sheetLayoutContainerKey] = node
		return applyLayoutNode(container, node, node.MinHeight.Pixels())
	}
}

// squareLayoutNodeOf returns the Block node whose Square flag governs the slot the given panel occupies, or nil if the
// panel's slot takes its width from its weight like any other.
func squareLayoutNodeOf(p *unison.Panel) *gurps.SheetLayoutNode {
	node, ok := p.ClientData()[sheetLayoutSquareKey].(*gurps.SheetLayoutNode)
	if !ok {
		return nil
	}
	return node
}

// sheetLayoutContainerNodeOf returns the Row or Column node the given container panel was built from, or nil if the
// panel is one the leaf function handed back rather than a container the builder made.
func sheetLayoutContainerNodeOf(p *unison.Panel) *gurps.SheetLayoutNode {
	node, ok := p.ClientData()[sheetLayoutContainerKey].(*gurps.SheetLayoutNode)
	if !ok {
		return nil
	}
	return node
}

// applyLayoutNode gives the panel the layout data its slot in the tree calls for and records the node that governs
// that slot on it.
func applyLayoutNode(p unison.Paneler, node *gurps.SheetLayoutNode, minHeight float32) unison.Paneler {
	panel := p.AsPanel()
	panel.SetLayoutData(&unison.FlexLayoutData{
		HAlign:  align.Fill,
		VAlign:  align.Fill,
		HGrab:   true,
		MinSize: geom.Size{Height: minHeight},
	})
	panel.ClientData()[sheetLayoutNodeKey] = node
	return p
}
