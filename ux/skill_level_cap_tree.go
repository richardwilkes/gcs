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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// asLevelCapGroup returns the node as a group, or nil if it isn't one.
func asLevelCapGroup(node gurps.LevelCap) *gurps.LevelCapGroup {
	if group, ok := node.(*gurps.LevelCapGroup); ok {
		return group
	}
	return nil
}

// levelCapGroupName returns the name of a group that takes the highest of its children, or else the lowest.
func levelCapGroupName(highest bool) string {
	if highest {
		return i18n.Text("Highest of")
	}
	return i18n.Text("Lowest of")
}

func (p *skillLevelCapPanel) treeRoot() gurps.LevelCap {
	return p.levelCap.Cap
}

func (p *skillLevelCapPanel) treeChildren(node gurps.LevelCap) (*[]gurps.LevelCap, bool) {
	if group := asLevelCapGroup(node); group != nil {
		return (*[]gurps.LevelCap)(&group.Children), true
	}
	return nil, false
}

func (p *skillLevelCapPanel) treeSetParent(node, group gurps.LevelCap) {
	node.SetParentGroup(asLevelCapGroup(group))
}

func (p *skillLevelCapPanel) treeClone(node, parent gurps.LevelCap) gurps.LevelCap {
	return node.Clone(asLevelCapGroup(parent))
}

// treeAll implements treeNodes. A group that takes the lowest of its children is the first kind.
func (p *skillLevelCapPanel) treeAll(group gurps.LevelCap) bool {
	g := asLevelCapGroup(group)
	return g != nil && !g.Highest
}

func (p *skillLevelCapPanel) treeNewGroup(parent gurps.LevelCap, all bool) gurps.LevelCap {
	group := gurps.NewLevelCapGroup(asLevelCapGroup(parent))
	group.Highest = !all
	return group
}

// treeCanUngroup implements treeNodes. Ungrouping loses nothing but the group's kind, which the sentenceTree checks.
func (p *skillLevelCapPanel) treeCanUngroup(_ gurps.LevelCap) bool {
	return true
}

func (p *skillLevelCapPanel) treeTitles(node gurps.LevelCap) treeEditTitles {
	if asLevelCapGroup(node) != nil {
		return treeEditTitles{
			duplicate: i18n.Text("Duplicate Group"),
			move:      i18n.Text("Move Group"),
			delete:    i18n.Text("Delete Group"),
		}
	}
	return treeEditTitles{
		duplicate: i18n.Text("Duplicate Cap Value"),
		move:      i18n.Text("Move Cap Value"),
		delete:    i18n.Text("Delete Cap Value"),
	}
}

func (p *skillLevelCapPanel) treeHasLead() bool {
	return false
}

func (p *skillLevelCapPanel) treeGroupName(group gurps.LevelCap) string {
	g := asLevelCapGroup(group)
	return levelCapGroupName(g != nil && g.Highest)
}

// treeGroupContents implements treeNodes: the sentences of the group's children, a group among them named for whether
// it takes the lowest or the highest of its own.
func (p *skillLevelCapPanel) treeGroupContents(group gurps.LevelCap) string {
	children := asLevelCapGroup(group).Children
	parts := make([]string, 0, len(children))
	for _, child := range children {
		if g := asLevelCapGroup(child); g != nil {
			parts = append(parts, levelCapGroupName(g.Highest))
		} else {
			parts = append(parts, stripEm.Replace(child.Describe(p.entity, emphasize)))
		}
	}
	return strings.Join(parts, i18n.Text("; "))
}

// treeGroupHead implements treeNodes: the group's pill, which says whether it takes the lowest or the highest of its
// children. Choosing one for an empty root keeps its head, as choosing a group type for it does.
func (p *skillLevelCapPanel) treeGroupHead(group gurps.LevelCap, path string, head, _ *unison.Panel) *unison.ThemeColor {
	g := asLevelCapGroup(group)
	color := groupColor(!g.Highest)
	pill := compactPopup(&p.sentenceRows, path+keyPill, i18n.Text("Group Type"), []bool{false, true}, g.Highest,
		levelCapGroupName, func(highest bool) {
			g.Highest = highest
			p.headed = p.headed || path == treeRootPath
		})
	pill.Tooltip = newWrappedTooltip(i18n.Text("Lowest of: the cap is the lowest of the values in the group. Highest of: it is the highest of them. A value resting on something the sheet doesn't have counts that as 0, and an empty group caps nothing."))
	stylePill(pill, color)
	addCentered(head, pill)
	return color
}

// treeEmpty implements treeNodes. An empty root shows its placeholder alone, until a group type has been chosen for it
// or unless it already takes the highest of its children.
func (p *skillLevelCapPanel) treeEmpty(group gurps.LevelCap, path string) (text string, bare bool) {
	if g := asLevelCapGroup(group); path == treeRootPath && !p.headed && !g.Highest {
		return i18n.Text("No cap values. Click here to add one."), true
	}
	return i18n.Text("Empty group. Add a cap value or drag one here."), false
}

// treeEmptyClick implements treeNodes. A value comes in several kinds, so the placeholder opens the menu that offers
// them.
func (p *skillLevelCapPanel) treeEmptyClick(_ gurps.LevelCap, _ string) func() {
	return nil
}

func (p *skillLevelCapPanel) treeRow(node gurps.LevelCap, path string) *unison.Panel {
	return p.row(node, path)
}

func (p *skillLevelCapPanel) treeAddEntries(group gurps.LevelCap, path string) []menuEntry {
	return p.addEntries(asLevelCapGroup(group), path, i18n.Text("Cap Value"), i18n.Text("Structure"))
}

func (p *skillLevelCapPanel) treeGroupAdds(group gurps.LevelCap, path string) []menuEntry {
	return p.addEntries(asLevelCapGroup(group), path, i18n.Text("Add Cap Value"), i18n.Text("Add Structure"))
}

// treeGroupLabels implements treeGroupLabeler.
func (p *skillLevelCapPanel) treeGroupLabels() (all, other string) {
	return i18n.Text("Lowest of Group"), i18n.Text("Highest of Group")
}
