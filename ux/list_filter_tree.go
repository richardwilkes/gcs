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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// filterGroupMode is how a filter group combines its children, which its pill offers: a group's All and Not taken
// together.
type filterGroupMode byte

const (
	filterAllOf filterGroupMode = iota
	filterAnyOf
	filterNoneOf
	filterNotAllOf
)

var filterGroupModes = []filterGroupMode{filterAllOf, filterAnyOf, filterNoneOf, filterNotAllOf}

func groupModeOf(group *gurps.FilterGroup) filterGroupMode {
	switch {
	case group.All && group.Not:
		return filterNotAllOf
	case group.Not:
		return filterNoneOf
	case group.All:
		return filterAllOf
	default:
		return filterAnyOf
	}
}

func (m filterGroupMode) String() string {
	switch m {
	case filterAnyOf:
		return i18n.Text("Any of")
	case filterNoneOf:
		return i18n.Text("None of")
	case filterNotAllOf:
		return i18n.Text("Not all of")
	default:
		return i18n.Text("All of")
	}
}

// apply sets the group's All and Not to the mode's.
func (m filterGroupMode) apply(group *gurps.FilterGroup) {
	group.All = m == filterAllOf || m == filterNotAllOf
	group.Not = m == filterNoneOf || m == filterNotAllOf
}

// asFilterGroup returns the node as a group, or nil if it isn't one.
func asFilterGroup(node gurps.FilterNode) *gurps.FilterGroup {
	if group, ok := node.(*gurps.FilterGroup); ok {
		return group
	}
	return nil
}

func (p *listFilterPanel) treeRoot() gurps.FilterNode {
	return p.filter.Root
}

func (p *listFilterPanel) treeChildren(node gurps.FilterNode) (*[]gurps.FilterNode, bool) {
	if group, ok := node.(*gurps.FilterGroup); ok {
		return (*[]gurps.FilterNode)(&group.Children), true
	}
	return nil, false
}

func (p *listFilterPanel) treeSetParent(node, group gurps.FilterNode) {
	node.SetParentGroup(asFilterGroup(group))
}

func (p *listFilterPanel) treeClone(node, parent gurps.FilterNode) gurps.FilterNode {
	return node.Clone(asFilterGroup(parent))
}

func (p *listFilterPanel) treeAll(group gurps.FilterNode) bool {
	g, ok := group.(*gurps.FilterGroup)
	return ok && g.All
}

func (p *listFilterPanel) treeNewGroup(parent gurps.FilterNode, all bool) gurps.FilterNode {
	group := gurps.NewFilterGroup(asFilterGroup(parent))
	group.All = all
	return group
}

// treeCanUngroup implements treeNodes. A negated group can't be ungrouped, since its children would lose the negation.
func (p *listFilterPanel) treeCanUngroup(group gurps.FilterNode) bool {
	g, ok := group.(*gurps.FilterGroup)
	return ok && !g.Not
}

func (p *listFilterPanel) treeTitles(node gurps.FilterNode) treeEditTitles {
	if _, ok := node.(*gurps.FilterGroup); ok {
		return treeEditTitles{
			duplicate: i18n.Text("Duplicate Group"),
			move:      i18n.Text("Move Group"),
			delete:    i18n.Text("Delete Group"),
		}
	}
	return treeEditTitles{
		duplicate: i18n.Text("Duplicate Condition"),
		move:      i18n.Text("Move Condition"),
		delete:    i18n.Text("Delete Condition"),
	}
}

func (p *listFilterPanel) treeHasLead() bool {
	return false
}

func (p *listFilterPanel) treeGroupName(group gurps.FilterNode) string {
	return groupModeOf(asFilterGroup(group)).String()
}

// treeGroupHead implements treeNodes: the group's pill, which says how it combines its children. Choosing a mode for an
// empty root keeps its head, as choosing a group type for it does.
func (p *listFilterPanel) treeGroupHead(group gurps.FilterNode, path string, head, _ *unison.Panel) *unison.ThemeColor {
	g := asFilterGroup(group)
	color := groupColor(g.All)
	pill := compactPopup(&p.sentenceRows, path+keyPill, i18n.Text("Match"), filterGroupModes, groupModeOf(g),
		filterGroupMode.String, func(m filterGroupMode) {
			m.apply(g)
			p.headed = p.headed || path == treeRootPath
		})
	pill.Tooltip = newWrappedTooltip(i18n.Text("All of: everything in the group must match. Any of: at least one thing in it must match. None of: nothing in it may match. Not all of: at least one thing in it must not match. An empty group is left out, whichever it is."))
	stylePill(pill, color)
	addCentered(head, pill)
	return color
}

// treeEmpty implements treeNodes. An empty root shows its placeholder alone, until a group type has been chosen for it
// or unless it already says something other than "All of".
func (p *listFilterPanel) treeEmpty(group gurps.FilterNode, path string) (text string, bare bool) {
	if g := asFilterGroup(group); path == treeRootPath && !p.headed && g.All && !g.Not {
		return i18n.Text("No conditions. Click here to add one."), true
	}
	return i18n.Text("Empty group. Add a condition or drag one here."), false
}

func (p *listFilterPanel) treeRow(node gurps.FilterNode, path string) *unison.Panel {
	return p.row(node, path)
}

func (p *listFilterPanel) treeAddEntries(group gurps.FilterNode, path string) []menuEntry {
	return p.addEntries(asFilterGroup(group), path, i18n.Text("Condition"), i18n.Text("Structure"))
}

func (p *listFilterPanel) treeGroupAdds(group gurps.FilterNode, path string) []menuEntry {
	return p.addEntries(asFilterGroup(group), path, i18n.Text("Add Condition"), i18n.Text("Add Structure"))
}
