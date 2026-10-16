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

// groupMode is how a group of filters or prerequisites combines its children, which its pill offers: a group's All and
// Not taken together.
type groupMode byte

const (
	groupAllOf groupMode = iota
	groupAnyOf
	groupNoneOf
	groupNotAllOf
)

var groupModes = []groupMode{groupAllOf, groupAnyOf, groupNoneOf, groupNotAllOf}

func groupModeOf(group *gurps.FilterGroup) groupMode {
	return groupModeFor(group.All, group.Not)
}

// groupModeFor returns the mode of a group with the given All and Not.
func groupModeFor(all, not bool) groupMode {
	switch {
	case all && not:
		return groupNotAllOf
	case not:
		return groupNoneOf
	case all:
		return groupAllOf
	default:
		return groupAnyOf
	}
}

func (m groupMode) String() string {
	switch m {
	case groupAnyOf:
		return i18n.Text("Any of")
	case groupNoneOf:
		return i18n.Text("None of")
	case groupNotAllOf:
		return i18n.Text("Not all of")
	default:
		return i18n.Text("All of")
	}
}

// flags returns the All and Not of a group with the mode.
func (m groupMode) flags() (all, not bool) {
	return m == groupAllOf || m == groupNotAllOf, m == groupNoneOf || m == groupNotAllOf
}

// apply sets the group's All and Not to the mode's.
func (m groupMode) apply(group *gurps.FilterGroup) {
	group.All, group.Not = m.flags()
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
		duplicate: i18n.Text("Duplicate Filter"),
		move:      i18n.Text("Move Filter"),
		delete:    i18n.Text("Delete Filter"),
	}
}

func (p *listFilterPanel) treeHasLead() bool {
	return false
}

func (p *listFilterPanel) treeGroupName(group gurps.FilterNode) string {
	return groupModeOf(asFilterGroup(group)).String()
}

// treeGroupContents implements treeNodes: the sentences of the group's children, a group among them named for how it
// combines its own.
func (p *listFilterPanel) treeGroupContents(group gurps.FilterNode) string {
	children := asFilterGroup(group).Children
	parts := make([]string, 0, len(children))
	for _, child := range children {
		if g, ok := child.(*gurps.FilterGroup); ok {
			parts = append(parts, groupModeOf(g).String())
		} else {
			parts = append(parts, stripEm.Replace(p.describe(child)))
		}
	}
	return strings.Join(parts, i18n.Text("; "))
}

// treeGroupHead implements treeNodes: the group's pill, which says how it combines its children. Choosing a mode for an
// empty root keeps its head, as choosing a group type for it does.
func (p *listFilterPanel) treeGroupHead(group gurps.FilterNode, path string, head, _ *unison.Panel) *unison.ThemeColor {
	g := asFilterGroup(group)
	color := groupColor(g.All)
	pill := compactPopup(&p.sentenceRows, path+keyPill, i18n.Text("Match"), groupModes, groupModeOf(g),
		groupMode.String, func(m groupMode) {
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
		return i18n.Text("No filters. Click here to add one."), true
	}
	return i18n.Text("Empty group. Add a filter or drag one here."), false
}

// treeEmptyClick implements treeNodes. A condition is the only kind of row a filter has, so the placeholder adds one
// rather than offering a menu.
func (p *listFilterPanel) treeEmptyClick(group gurps.FilterNode, path string) func() {
	return func() { p.addCondition(asFilterGroup(group), path) }
}

func (p *listFilterPanel) treeRow(node gurps.FilterNode, path string) *unison.Panel {
	return p.row(node, path)
}

func (p *listFilterPanel) treeAddEntries(group gurps.FilterNode, path string) []menuEntry {
	return p.addEntries(asFilterGroup(group), path, i18n.Text("Filter"), i18n.Text("Structure"))
}

func (p *listFilterPanel) treeGroupAdds(group gurps.FilterNode, path string) []menuEntry {
	return p.addEntries(asFilterGroup(group), path, i18n.Text("Add Filter"), i18n.Text("Add Structure"))
}
