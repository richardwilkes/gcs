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
	"slices"
	"strconv"
	"strings"

	"github.com/richardwilkes/gcs/v5/ux/colors"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/weight"
)

// A node of a tree is found by its path: treeRootPath for the root, then the index of each child on the way down,
// separated by dots. A widget's reference key is the path of its node followed by one of the suffixes below, one of
// those of sentenceRows, or a colon and a name of its own.
const (
	treeRootPath = "r"
	keyAdd       = ":add"
	keyPill      = ":pill"
)

// treeDropKey marks what a dragged node can be dropped on, holding a treeDropSpot: the path of the node it shows and
// which part of the node it is, one of the dropOn values.
const treeDropKey = "tree.drop"

type treeDropSpot struct {
	path string
	part int
}

const (
	dropOnRow = iota
	dropOnHead
	dropOnGroup
	dropOnEmpty
)

// treeNodes is what a sentenceTree needs from the panel that embeds it: the shape of its tree of nodes of type N, some
// of which are groups holding others, and the parts of a group and its rows that differ from one tree to the next.
type treeNodes[N comparable] interface {
	// treeRoot returns the root group.
	treeRoot() N
	// treeChildren returns the children of a group, or false for a node that isn't one.
	treeChildren(node N) (*[]N, bool)
	// treeSetParent makes the group the parent of the node.
	treeSetParent(node, group N)
	// treeClone returns a deep copy of the node, owned by the parent.
	treeClone(node, parent N) N
	// treeAll reports whether the group requires all of its children rather than any one of them.
	treeAll(group N) bool
	// treeNewGroup returns a new, empty group owned by the parent, which requires all of its children when all is set.
	treeNewGroup(parent N, all bool) N
	// treeCanUngroup reports whether a group may be replaced by its children, as far as the tree is concerned; the
	// sentenceTree also requires that doing so keeps what the group means.
	treeCanUngroup(group N) bool
	// treeTitles returns the titles of the edits that restructure the tree.
	treeTitles() treeEditTitles
	// treeHasLead reports whether the head of a group starts with something ahead of its pill, such as the status icon
	// of a prerequisite.
	treeHasLead() bool
	// treeGroupName returns the name a screen reader gives the group.
	treeGroupName(group N) string
	// treeGroupHead adds what goes between a group head's grip and its more button, its pill keyed by the path and
	// keyPill among it, and returns the color of the group. box is the panel of the whole group.
	treeGroupHead(group N, path string, head, box *unison.Panel) *unison.ThemeColor
	// treeEmpty returns the text of the placeholder of an empty group, and whether the group shows the placeholder
	// alone, with no head, as an untouched empty root does.
	treeEmpty(group N, path string) (text string, bare bool)
	// treeRow returns the panel of a node that isn't a group.
	treeRow(node N, path string) *unison.Panel
	// treeAddEntries returns what an add button offers to add to the group.
	treeAddEntries(group N, path string) []menuEntry
	// treeGroupAdds returns what the more menu of a group offers to add to it, ahead of the rest of the menu.
	treeGroupAdds(group N, path string) []menuEntry
}

// treeEditTitles holds the titles of the edits that duplicate, move and delete a node.
type treeEditTitles struct {
	duplicate, move, delete string
}

// sentenceTree is a panel of sentence rows for a tree of nodes of type N. Each group is a head, saying how it combines
// its children, over its children hanging from a rail in the head's color. A panel embeds it, calls initRows and
// initTree, and adds group(root, treeRootPath) to its content.
type sentenceTree[T any, N comparable] struct {
	sentenceRows[T]
	nodes treeNodes[N]
}

// initTree sets up the tree, whose shape nodes describes, and lets its rows be dragged.
func (p *sentenceTree[T, N]) initTree(nodes treeNodes[N]) {
	p.nodes = nodes
	p.initDrop(func(where geom.Point, data any) (*unison.Panel, int) {
		target, _, at := p.dropAt(where, data)
		return target, at
	}, p.drop)
}

func childPath(path string, i int) string {
	return path + "." + strconv.Itoa(i)
}

// node returns the node at the path, or the zero value if there is none.
func (p *sentenceTree[T, N]) node(path string) N {
	var none N
	parts := strings.Split(path, ".")
	if parts[0] != treeRootPath {
		return none
	}
	node := p.nodes.treeRoot()
	for _, part := range parts[1:] {
		children, ok := p.nodes.treeChildren(node)
		i, err := strconv.Atoi(part)
		if !ok || err != nil || i < 0 || i >= len(*children) {
			return none
		}
		node = (*children)[i]
	}
	return node
}

// locate returns the group holding the node at the path, other than the root, and the node's index within it.
func (p *sentenceTree[T, N]) locate(path string) (parent N, index int) {
	parentPath, last, found := strings.CutLast(path, ".")
	if !found {
		return parent, -1
	}
	index, err := strconv.Atoi(last)
	group := p.node(parentPath)
	if _, ok := p.nodes.treeChildren(group); ok && err == nil {
		return group, index
	}
	return parent, -1
}

// pathOf returns the path of the node, or an empty string if it is not in the tree.
func (p *sentenceTree[T, N]) pathOf(target N) string {
	var none N
	if target == none {
		return ""
	}
	var find func(node N, path string) string
	find = func(node N, path string) string {
		if node == target {
			return path
		}
		if children, ok := p.nodes.treeChildren(node); ok {
			for i, child := range *children {
				if found := find(child, childPath(path, i)); found != "" {
					return found
				}
			}
		}
		return ""
	}
	return find(p.nodes.treeRoot(), treeRootPath)
}

// restructure changes the shape of the tree through the widget with the reference key from, then rebuilds. change
// returns the node that is the result, whose more button takes the focus; with none, the widget with the reference key
// fallback does. The open row stays open wherever it ends up, and closes if it is removed.
func (p *sentenceTree[T, N]) restructure(title, from, fallback string, change func() (dst N)) {
	open := p.node(p.open)
	p.sentenceRows.restructure(title, from, fallback, func() string {
		dst := change()
		p.open = p.pathOf(open)
		var none N
		if dst == none {
			return ""
		}
		return p.pathOf(dst) + keyMore
	})
}

// groupColor returns the color of a group that requires all of its children, or any one of them.
func groupColor(all bool) *unison.ThemeColor {
	if all {
		return colors.Grouping1
	}
	return colors.Grouping2
}

// stylePill has the popup that heads a group look like a pill in the group's color.
func stylePill[V comparable](pill *unison.PopupMenu[V], color *unison.ThemeColor) {
	pill.HMargin = 10
	pill.CornerRadius = geom.NewUniformSize(pillCornerRadius)
	pill.BackgroundInk = color
	pill.OnBackgroundInk = color.DeriveOn()
	pill.EdgeInk = unison.Transparent
	desc := pill.Font.Descriptor()
	desc.Weight = weight.Bold
	pill.Font = desc.Font()
}

// group returns the panel for a group: its head, over its children hanging from a rail in the head's color. Its more
// menu adds to it, as do the root's add button and the add button beside the placeholder of an empty group.
func (p *sentenceTree[T, N]) group(group N, path string) *unison.Panel {
	box := newColumn()
	if path != treeRootPath {
		box.RefKey = path + ":group"
		box.ClientData()[treeDropKey] = treeDropSpot{path: path, part: dropOnGroup}
	}
	head := unison.NewPanel()
	// The same insets on the sides as a row's, so that the grips and the buttons on the right line up down the panel.
	insets := geom.Insets{Top: 2, Left: 4, Bottom: 2, Right: 8}
	if path == treeRootPath && !p.nodes.treeHasLead() {
		// With nothing before it, the root's pill keeps the room a lead would take from the edge.
		insets.Left += 4
	}
	head.SetBorder(unison.NewEmptyBorder(insets))
	head.ClientData()[treeDropKey] = treeDropSpot{path: path, part: dropOnHead}
	if path != treeRootPath {
		grip := p.grip(head, path)
		putOnLine(grip.AsPanel(), controlHeight(head), grip.svg.Size.Height)
	}
	children, _ := p.nodes.treeChildren(group)
	var emptyLine *unison.Panel
	if len(*children) == 0 {
		text, bare := p.nodes.treeEmpty(group, path)
		empty := newEmptyPlaceholder(path+":empty", text, nil)
		// A click opens the menu where it lands, as a right-click does; a key opens it at the placeholder.
		empty.ClickCallback = func() { showMenu(empty.AsPanel(), p.nodes.treeAddEntries(group, path)) }
		empty.ContextMenuCallback = func(geom.Point) unison.Menu {
			return newEntriesMenu(p.nodes.treeAddEntries(group, path))
		}
		empty.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
			empty.Pressed = false
			empty.MarkForRedraw()
			if where.In(empty.ContentRect(false)) {
				if unison.IsAccessibilityActive() {
					empty.RequestFocus()
				}
				empty.ShowContextMenu(where)
			}
			return true
		}
		empty.ClientData()[treeDropKey] = treeDropSpot{path: path, part: dropOnEmpty}
		// The placeholder shares its line with an add button, in line with the more buttons.
		add := newIconButton(path+keyAdd, unison.CircledAddSVG, i18n.Text("Add to this group"))
		add.ClickCallback = func() { showMenu(add.AsPanel(), p.nodes.treeAddEntries(group, path)) }
		fitLine(add)
		add.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle})
		// An untouched empty root has nothing for its pill, status or rail to speak of, so its placeholder takes their
		// place, and it isn't a group to a screen reader.
		if bare {
			head.AddChild(empty)
			head.AddChild(add)
			box.AddChild(hbox(head, unison.StdHSpacing))
			return box
		}
		emptyLine = unison.NewPanel()
		emptyLine.SetBorder(unison.NewEmptyBorder(geom.Insets{Right: 8}))
		emptyLine.AddChild(empty)
		emptyLine.AddChild(add)
		hbox(emptyLine, unison.StdHSpacing)
	}
	box.Accessibility.Role = role.Group
	box.Accessibility.Name = p.nodes.treeGroupName(group)
	color := p.nodes.treeGroupHead(group, path, head, box)
	if path != treeRootPath {
		p.moreButton(head, group, path)
	} else {
		// The root can't be moved or deleted, so in place of a more button it has an add button, keyed as one.
		add := newIconButton(path+keyMore, unison.CircledAddSVG, i18n.Text("Add to this group"))
		add.ClickCallback = func() { showMenu(add.AsPanel(), p.nodes.treeAddEntries(group, path)) }
		addCentered(head, add)
	}
	more := head.Children()[len(head.Children())-1]
	more.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End, VAlign: align.Middle, HGrab: true})
	// Fitted to its line, since centering a shorter button could put its icon on a half pixel, blurring it.
	fitLine(more)
	box.AddChild(hbox(head, unison.StdHSpacing))

	rail := newColumn()
	// Lighter than the pill in dark mode, so that the rail has a contrast of at least 3:1 with the surface.
	rail.SetBorder(unison.NewCompoundBorder(unison.NewEmptyBorder(geom.Insets{Left: 14}),
		unison.NewLineBorder(color.DeriveLightness(0, 0.08), geom.Size{}, geom.Insets{Left: 3}, false),
		unison.NewEmptyBorder(geom.Insets{Left: 6, Bottom: 2})))
	for i, child := range *children {
		if _, ok := p.nodes.treeChildren(child); ok {
			rail.AddChild(p.group(child, childPath(path, i)))
		} else {
			row := p.nodes.treeRow(child, childPath(path, i))
			row.ClientData()[treeDropKey] = treeDropSpot{path: childPath(path, i), part: dropOnRow}
			rail.AddChild(row)
		}
	}
	if emptyLine != nil {
		rail.AddChild(emptyLine)
	}
	box.AddChild(rail)
	return box
}

// addKey returns the reference key of what adds to the group at the path: the add button beside its placeholder while
// it is empty, or else its more button, or the root's add button in its place.
func (p *sentenceTree[T, N]) addKey(group N, path string) string {
	if children, ok := p.nodes.treeChildren(group); ok && len(*children) == 0 {
		return path + keyAdd
	}
	return path + keyMore
}

// moreButton adds the button for the node's more menu.
func (p *sentenceTree[T, N]) moreButton(parent *unison.Panel, node N, path string) {
	addMoreButton(parent, path, func() []menuEntry { return p.moreEntries(node, path) })
}

// moreEntries returns the entries of the node's more menu: for a group, what can be added to it, then Duplicate, Move
// up and down, which step into and out of groups, Wrap in group, Ungroup when that keeps what the group means, and
// Delete.
func (p *sentenceTree[T, N]) moreEntries(node N, path string) []menuEntry {
	parent, i := p.locate(path)
	siblings, _ := p.nodes.treeChildren(parent)
	from := path + keyMore
	titles := p.nodes.treeTitles()
	var entries []menuEntry
	children, isGroup := p.nodes.treeChildren(node)
	if isGroup {
		entries = append(p.nodes.treeGroupAdds(node, path), menuEntry{})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Duplicate"), Act: func() {
		p.restructure(titles.duplicate, from, "", func() N {
			dst := p.nodes.treeClone(node, parent)
			*siblings = slices.Insert(*siblings, i+1, dst)
			return dst
		})
	}})
	moveTitles := []string{i18n.Text("Move Up"), i18n.Text("Move Down")}
	for k, dir := range []int{-1, 1} {
		to, at, ok := p.moveTarget(path, dir)
		if !ok {
			continue
		}
		title := moveTitles[k]
		entries = append(entries, menuEntry{Label: title, Act: func() {
			p.restructure(title, from, "", func() N { return p.relocate(parent, i, to, at) })
		}})
	}
	entries = append(entries, menuEntry{Label: i18n.Text("Wrap in Group"), Act: func() {
		p.restructure(i18n.Text("Wrap in Group"), from, "", func() N {
			wrapper := p.nodes.treeNewGroup(parent, !p.nodes.treeAll(parent))
			held, _ := p.nodes.treeChildren(wrapper)
			*held = append(*held, node)
			p.nodes.treeSetParent(node, wrapper)
			(*siblings)[i] = wrapper
			return node
		})
	}})
	if isGroup && len(*children) != 0 && p.nodes.treeCanUngroup(node) &&
		(p.nodes.treeAll(node) == p.nodes.treeAll(parent) || len(*children) == 1) {
		entries = append(entries, menuEntry{Label: i18n.Text("Ungroup"), Act: func() {
			p.restructure(i18n.Text("Ungroup"), from, "", func() N {
				for _, child := range *children {
					p.nodes.treeSetParent(child, parent)
				}
				*siblings = slices.Replace(*siblings, i, i+1, *children...)
				return (*children)[0]
			})
		}})
	}
	entries = append(entries, menuEntry{}, menuEntry{Label: i18n.Text("Delete"), Act: func() {
		parentPath := path[:strings.LastIndexByte(path, '.')]
		// What adds to the parent takes the focus when nothing comes after, which is its placeholder's add button once
		// it is empty.
		fallback := parentPath + keyMore
		if len(*siblings) == 1 {
			fallback = parentPath + keyAdd
		}
		p.restructure(titles.delete, from, fallback, func() N {
			*siblings = slices.Delete(*siblings, i, i+1)
			if i < len(*siblings) {
				return (*siblings)[i]
			}
			var none N
			return none
		})
	}})
	return entries
}

// dropAt returns the panel a node dragged to where would be dropped on, the path of its node and where it would go:
// before or after a row, before a group over the top of its head and into it below that, after a group beside or below
// its last child, or into an empty group. The panel is nil where nothing can go, such as into itself.
func (p *sentenceTree[T, N]) dropAt(where geom.Point, data any) (target *unison.Panel, path string, at int) {
	from := p.dragPath(data)
	var none N
	if p.node(from) == none {
		return nil, "", 0
	}
	for target = p.PanelAt(where); target != nil && target != p.AsPanel(); target = target.Parent() {
		spot, isTarget := target.ClientData()[treeDropKey].(treeDropSpot)
		if !isTarget {
			continue
		}
		if spot.path == from || strings.HasPrefix(spot.path, from+".") {
			return nil, "", 0
		}
		y := target.PointFromRoot(p.PointToRoot(where)).Y
		height := target.FrameRect().Height
		switch spot.part {
		case dropOnEmpty:
			return target, spot.path, dropInto
		case dropOnGroup:
			// Only the rail's margins reach the group itself; beside or below its last child is after it.
			rows := target.Children()[len(target.Children())-1].Children()
			last := rows[len(rows)-1]
			if where.Y > p.RectFromRoot(last.RectToRoot(last.ContentRect(true))).CenterY() {
				return target, spot.path, dropAfter
			}
			return nil, "", 0
		case dropOnHead:
			if spot.path != treeRootPath && y < height*0.3 {
				return target, spot.path, dropBefore
			}
			return target, spot.path, dropInto
		default:
			if y < height/2 {
				return target, spot.path, dropBefore
			}
			return target, spot.path, dropAfter
		}
	}
	return nil, "", 0
}

// drop moves the dragged node to where it would go.
func (p *sentenceTree[T, N]) drop(where geom.Point, data any) {
	target, path, at := p.dropAt(where, data)
	p.dragExit()
	if target == nil {
		return
	}
	to, index := p.locate(path)
	if group := p.node(path); at == dropInto {
		if children, ok := p.nodes.treeChildren(group); ok {
			to, index = group, len(*children)
		}
	} else if at == dropAfter {
		index++
	}
	dragged := p.dragPath(data)
	from, i := p.locate(dragged)
	if from == to && i < index {
		// Taking the node out moves what comes after it up by one.
		index--
	}
	p.restructure(p.nodes.treeTitles().move, dragged+keyMore, "", func() N {
		if from == to && i == index {
			var none N
			return none
		}
		return p.relocate(from, i, to, index)
	})
}

// moveTarget returns where Move up (dir -1) or Move down (dir 1) takes the node at the path: into an adjacent group,
// at its near end, past an adjacent sibling, or out of its group at either end. at is an index into to once the node
// is out of its group. ok is false at either end of the root.
func (p *sentenceTree[T, N]) moveTarget(path string, dir int) (to N, at int, ok bool) {
	parent, i := p.locate(path)
	siblings, _ := p.nodes.treeChildren(parent)
	if j := i + dir; j >= 0 && j < len(*siblings) {
		if children, isGroup := p.nodes.treeChildren((*siblings)[j]); isGroup {
			if dir < 0 {
				return (*siblings)[j], len(*children), true
			}
			return (*siblings)[j], 0, true
		}
		return parent, j, true
	}
	if parentPath := path[:strings.LastIndexByte(path, '.')]; parentPath != treeRootPath {
		to, at = p.locate(parentPath)
		if dir > 0 {
			at++
		}
		return to, at, true
	}
	return to, 0, false
}

// relocate moves the node at index i of the group from to index at of the group to, as that group stands once the node
// is out, returning the node.
func (p *sentenceTree[T, N]) relocate(from N, i int, to N, at int) N {
	src, _ := p.nodes.treeChildren(from)
	dst, _ := p.nodes.treeChildren(to)
	node := (*src)[i]
	*src = slices.Delete(*src, i, i+1)
	p.nodes.treeSetParent(node, to)
	*dst = slices.Insert(*dst, at, node)
	return node
}
