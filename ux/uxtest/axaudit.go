// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package uxtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/role"
)

// AXNameAudit collects the controls in a headless workspace that an assistive technology would be handed with no name.
// A field, popup or color well is named by the label laid out before it, by the label it points at with
// Accessibility.LabeledBy, or by an Accessibility.Name of its own; an icon-only button by its tooltip. A control with
// none of those is announced as "edit text" or "pop up button" and nothing more.
type AXNameAudit struct {
	t        *testing.T
	screen   *unison.HeadlessScreen
	wnd      *unison.Window
	failures []string
}

// NewAXNameAudit returns an audit of the workspace window wnd that the screen drives.
func NewAXNameAudit(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window) *AXNameAudit {
	return &AXNameAudit{t: t, screen: screen, wnd: wnd}
}

// Open runs fn on the UI thread and returns the dockable it opened, or nil -- and a test error -- if it opened some
// other number of them.
func (a *AXNameAudit) Open(fn func()) unison.Dockable {
	a.t.Helper()
	var opened []unison.Dockable
	a.screen.Do(func() {
		before := make(map[unison.Dockable]bool)
		for _, d := range workspace.AllDockables() {
			before[d] = true
		}
		fn()
		for _, d := range workspace.AllDockables() {
			if !before[d] {
				opened = append(opened, d)
			}
		}
	})
	if len(opened) != 1 {
		a.t.Errorf("expected exactly one dockable to open; %d did", len(opened))
		return nil
	}
	return opened[0]
}

// CheckOpened opens a dockable with fn and checks the controls in it.
func (a *AXNameAudit) CheckOpened(view string, fn func()) {
	a.t.Helper()
	if d := a.Open(fn); d != nil {
		a.Check(view, d)
	}
}

// Check describes the window afresh and records every control under root that has no name. The tree is walked rather
// than the panels, so that what a table describes without a panel apiece -- its column headers and the cells of its
// rows -- is checked along with everything else.
func (a *AXNameAudit) Check(view string, root unison.Paneler) {
	a.t.Helper()
	tree := a.screen.AccessibilityTree(a.wnd)
	if tree == nil {
		a.t.Fatalf("%s: the window was not described", view)
	}
	rootNode := a.screen.AccessibilityNodeFor(root)
	if rootNode == nil {
		a.t.Fatalf("%s: the view was not described", view)
	}
	// The panels that were described, by node, so that a failure can say what the control is and where it sits. What
	// a widget describes without a panel apiece -- a table's rows and cells -- has no entry and is placed by its
	// ancestors instead.
	panels := make(map[accessibility.NodeID]*unison.Panel)
	var visible []*unison.Panel
	a.screen.Do(func() {
		var walk func(p *unison.Panel)
		walk = func(p *unison.Panel) {
			if p.Hidden {
				return
			}
			visible = append(visible, p)
			for _, child := range p.Children() {
				walk(child)
			}
		}
		walk(root.AsPanel())
	})
	for _, p := range visible {
		if node := a.screen.AccessibilityNodeFor(p); node != nil {
			panels[node.ID] = p
		}
	}
	seen := make(map[accessibility.NodeID]bool)
	var walk func(id accessibility.NodeID)
	walk = func(id accessibility.NodeID) {
		if seen[id] {
			return
		}
		seen[id] = true
		node := tree.Node(id)
		if node == nil {
			return
		}
		if !node.Ignored && node.Name == "" && axNodeNeedsAName(node) {
			a.failures = append(a.failures, a.describe(view, tree, node, panels))
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(rootNode.ID)
}

// Report fails the test with every unnamed control the checks found, if there were any.
func (a *AXNameAudit) Report() {
	a.t.Helper()
	if len(a.failures) != 0 {
		a.t.Errorf("%d controls would be announced with no name:\n%s", len(a.failures),
			strings.Join(a.failures, "\n"))
	}
}

// axNodeNeedsAName reports whether a node is something a person is expected to read, change or act on, and so must be
// announced by name. A table's rows and cells are where its keyboard focus is reported, so they are focusable without
// being controls: a cell with nothing in it is rightly announced as blank, and one with a control in it is checked
// through that control.
//
// A document is listed outright rather than left to the focusable fallback, because whether it is focusable depends on
// the platform: on Windows and Linux a markdown view takes the keyboard focus while a screen reader is running, so that
// the reader's cursor can be moved into it, and it is then announced by name on landing there; on macOS it never does.
// Naming it here keeps the check the same everywhere, so that a markdown view left unnamed fails on the machine the
// change was made on rather than only in CI.
func axNodeNeedsAName(node *accessibility.Node) bool {
	switch node.Role {
	case role.TextField, role.TextArea, role.SpinButton, role.ComboBox, role.PopupButton, role.Slider,
		role.ProgressBar, role.ColorWell, role.List, role.Table, role.Tree, role.Button, role.CheckBox,
		role.RadioButton, role.ToggleButton, role.Link, role.ColumnHeader, role.Document:
		return true
	case role.Row, role.Cell:
		return false
	default:
		return node.Focusable
	}
}

// describe says where an unnamed control is, in terms that point at the code that built it.
func (a *AXNameAudit) describe(view string, tree *accessibility.Tree, node *accessibility.Node, panels map[accessibility.NodeID]*unison.Panel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "  %s: %v", view, node.Role)
	if node.Placeholder != "" {
		fmt.Fprintf(&b, " placeholder=%q", node.Placeholder)
	}
	if node.Description != "" {
		fmt.Fprintf(&b, " description=%q", axTruncate(node.Description))
	}
	if p := panels[node.ID]; p != nil {
		a.screen.Do(func() {
			fmt.Fprintf(&b, " %s", axPanelTypeName(p))
			if parent := p.Parent(); parent != nil {
				if i := parent.IndexOfChild(p); i > 0 {
					fmt.Fprintf(&b, " after %s", axPanelTypeName(parent.Children()[i-1]))
				} else {
					b.WriteString(" first in its parent")
				}
			}
			b.WriteString(" in ")
			for q, depth := p.Parent(), 0; q != nil && depth < 4; q, depth = q.Parent(), depth+1 {
				if depth > 0 {
					b.WriteString(" < ")
				}
				b.WriteString(axPanelTypeName(q))
			}
		})
		return b.String()
	}
	if node.Role == role.ColumnHeader || node.Role == role.Cell {
		fmt.Fprintf(&b, " column=%d", node.ColumnIndex)
	}
	b.WriteString(" (no panel of its own) in")
	for _, id := range tree.Path(node.ID) {
		if id == node.ID {
			continue
		}
		if n := tree.Node(id); n != nil && n.Name != "" {
			fmt.Fprintf(&b, " %v %q /", n.Role, axTruncate(n.Name))
		}
	}
	return b.String()
}

// axPanelTypeName names a panel by its type, with a label's text alongside since that is what tells labels apart.
func axPanelTypeName(p *unison.Panel) string {
	name := fmt.Sprintf("%T", p.Self)
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimPrefix(name, "github.com/richardwilkes/gcs/v5/")
	name = strings.TrimPrefix(name, "github.com/richardwilkes/unison.")
	if label, ok := p.Self.(*unison.Label); ok {
		name += fmt.Sprintf("(%q)", label.String())
	}
	return name
}

// axTruncate keeps a description short enough to read in a failure.
func axTruncate(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 60 {
		return s[:60] + "…"
	}
	return s
}
