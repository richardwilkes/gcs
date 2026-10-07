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
	"image"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellcmp"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// showPrereqPanel shows a prereqPanel for the root in a window, within a host, so that its rebuilds run and its edits
// can be undone. The panel is expanded, as most tests look at its rows; see TestPrereqPanelStartingState for how it
// starts out.
func showPrereqPanel(t *testing.T, screen *unison.HeadlessScreen, root **gurps.PrereqList, ownerIsSpell bool) (*prereqPanel, *sentenceUndoHost) {
	var p *prereqPanel
	host := &sentenceUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		host.KeyDownCallback = func(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
			if keyCode == unison.KeyEscape {
				host.escapes++
			}
			return true
		}
		p = newPrereqPanel(gurps.NewEntity(), root, prereq.TypesForNonEquipment, ownerIsSpell)
		// Expanded ahead of showing, so that the window is sized for the rows.
		if p.collapse.collapsed {
			p.collapse.toggle()
		}
		host.AddChild(p)
	})
	showInTestWindow(t, screen, 700, host)
	return p, host
}

// newTestPrereqTree returns an "all of" list holding a trait, an "any of" list holding a skill and a script, and an
// unknown prerequisite.
func newTestPrereqTree() *gurps.PrereqList {
	anyOf := gurps.NewPrereqList()
	anyOf.All = false
	anyOf.Prereqs = gurps.Prereqs{gurps.NewSkillPrereq(), gurps.NewScriptPrereq()}
	root := gurps.NewPrereqList()
	root.Prereqs = gurps.Prereqs{
		gurps.NewTraitPrereq(), anyOf,
		gurps.NewUnknownPrereq("future", []byte(`{"type":"future"}`)),
	}
	return root.CloneAsPrereqList(nil)
}

// prereqShape returns the types of the root's children, each list followed by the types of its children.
func prereqShape(root *gurps.PrereqList) []prereq.Type {
	var types []prereq.Type
	for _, one := range root.Prereqs {
		types = append(types, one.PrereqType())
		if list, ok := one.(*gurps.PrereqList); ok {
			for _, child := range list.Prereqs {
				types = append(types, child.PrereqType())
			}
		}
	}
	return types
}

// TestPrereqPanelBuildingChangesNothing opens every row in turn, and shows a missing list, checking that neither
// changes the data, since opening an editor must not mark it modified.
func TestPrereqPanelBuildingChangesNothing(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	sp := gurps.NewSpellPrereq()
	sp.SamePowerSource = true // Only a spell's own prerequisites should have this, but a file can hold anything.
	root.Prereqs = append(root.Prereqs, sp.Clone(root))
	hash := gurps.Hash64(root)
	p, _ := showPrereqPanel(t, screen, &root, false)
	for _, path := range []string{"r.0", "r.1.0", "r.1.1", "r.2", "r.3"} {
		screen.Do(func() {
			p.open = path
			p.rebuild(path + keyFirst)
		})
		c.Equal(hash, gurps.Hash64(root), "opening %s changes nothing", path)
	}
	screen.Do(func() {
		popup, ok := p.FindRefKey("r.3:powercmp").Self.(*unison.PopupMenu[string])
		c.True(ok, "the power source chip is shown")
		c.Equal(0, popup.SelectedIndex(), "showing the same power source as this spell's even where it can't apply")
		c.Nil(p.FindRefKey("r.2"+keyFirst), "an unknown prerequisite doesn't open")
	})
	c.True(sp.SamePowerSource)

	var missing *gurps.PrereqList
	p, _ = showPrereqPanel(t, screen, &missing, false)
	c.Nil(missing, "a missing list stays missing until something is added")
	screen.Do(func() { menuAction(p.addEntries(p.tree(), treeRootPath), "Trait")() })
	c.Equal(1, len(missing.Prereqs), "then it is made")
}

// TestPrereqPanelUndo checks that typing is recorded as one snapshot of the tree, that undo and redo install copies of
// their snapshots, and that a field's own undo is never recorded.
func TestPrereqPanelUndo(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		field, ok := p.FindRefKey("r.0:name").Self.(*StringField)
		c.True(ok, "the open trait row has a name field")
		field.RequestFocus()
	})
	screen.Type("Magery")
	trait := func() *gurps.TraitPrereq {
		one, ok := root.Prereqs[0].(*gurps.TraitPrereq)
		c.True(ok)
		return one
	}
	c.Equal("Magery", trait().NameCriteria.Qualifier)
	c.Equal("Undo Name", host.mgr.UndoTitle())
	screen.Do(host.mgr.Undo)
	c.False(host.mgr.CanUndo(), "the typing was one edit")
	c.Equal("", trait().NameCriteria.Qualifier)
	screen.Do(host.mgr.Redo)
	c.Equal("Magery", trait().NameCriteria.Qualifier)
	screen.Do(func() {
		field, ok := p.FindRefKey("r.0:name").Self.(*StringField)
		c.True(ok, "the row stays open across undo and redo")
		c.Equal("Magery", field.Text())
		c.Equal(field.AsPanel(), field.Window().Focus(), "and the field keeps the focus")
	})

	screen.Do(func() { menuAction(p.moreEntries(root.Prereqs[0], "r.0"), "Delete")() })
	c.Equal(2, len(root.Prereqs))
	c.Equal("", p.open, "deleting the open row leaves none open")
	installed := root
	screen.Do(host.mgr.Undo)
	c.Equal(3, len(root.Prereqs))
	c.True(installed != root, "undo installs a copy")
	screen.Do(func() { host.mgr.Undo() })
	c.False(host.mgr.CanUndo())
}

// TestPrereqPanelUndoFocusesChipField checks that undoing typing in a chip's field gives the focus back to that field,
// not to the chip around it, and that typing after a row closes and opens again is a step of its own.
func TestPrereqPanelUndoFocusesChipField(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	trait, ok := root.Prereqs[0].(*gurps.TraitPrereq)
	c.True(ok)
	trait.NotesCriteria.Compare = criteria.IsText
	p, host := showPrereqPanel(t, screen, &root, false)
	typeInNotes := func(text string) {
		screen.Do(func() { p.toggle("r.0") })
		screen.Do(func() {
			field, isField := p.FindRefKey("r.0:notes").Self.(*StringField)
			c.True(isField, "the notes chip's key is its field's")
			field.RequestFocus()
			field.SetSelection(len(field.Text()), len(field.Text()))
		})
		screen.Type(text)
	}
	typeInNotes("Ma")
	screen.Do(func() { p.toggle("r.0") })
	typeInNotes("gic")
	c.Equal("Magic", trait.NotesCriteria.Qualifier)
	screen.Do(host.mgr.Undo)
	screen.Do(func() {
		trait, ok = root.Prereqs[0].(*gurps.TraitPrereq) // Undo installs a copy.
		c.True(ok)
		c.Equal("Ma", trait.NotesCriteria.Qualifier, "closing the row ended the first run of typing")
		c.Equal("r.0:notes", p.Window().Focus().RefKey, "the field gets the focus back")
	})
}

// TestPrereqPanelUndoFocus checks that undo gives the focus to the widget that made the change, in the row it restores,
// and that redo gives it to where the change put it.
func TestPrereqPanelUndoFocus(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	for _, one := range []struct {
		name, undo, redo string
		act              func(p *prereqPanel)
	}{
		{"adding a chip", "r.0:add notes", "r.0:notescmp", func(p *prereqPanel) {
			if b, ok := p.FindRefKey("r.0:add notes").Self.(*unison.Button); ok {
				b.ClickCallback()
			}
		}},
		{"a move", "r.0" + keyMore, "r.0.0" + keyMore, func(p *prereqPanel) {
			menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Move Down")()
		}},
	} {
		root := newTestPrereqTree()
		p, host := showPrereqPanel(t, screen, &root, false)
		screen.Do(func() { p.toggle("r.0") })
		screen.Do(func() { one.act(p) })
		screen.Do(host.mgr.Undo)
		screen.Do(func() {
			c.Equal(one.undo, p.Window().Focus().RefKey, "undoing %s", one.name)
			c.Equal(prereq.Trait, p.node("r.0").PrereqType())
		})
		screen.Do(host.mgr.Redo)
		screen.Do(func() { c.Equal(one.redo, p.Window().Focus().RefKey, "redoing %s", one.name) })
	}
}

// TestPrereqPanelOpenRowFollowsRestructures checks that the open row stays open wherever a change to the tree's shape
// takes it, even within a group that is copied, and that none is open once it is deleted.
func TestPrereqPanelOpenRowFollowsRestructures(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	for _, one := range []struct {
		open, path, label, want string
	}{
		{"r.1.1", "r.1", "Wrap in Group", "r.1.0.1"},
		{"r.1.1", "r.1", "Move Up", "r.0.1"},
		{"r.1.1", "r.1", "Ungroup", "r.2"},
		{"r.1.1", "r.1", "Delete", ""},
		{"r.0", "r.0", "Delete", ""},
		{"r.0", "r.0", "Move Down", "r.0.0"},
	} {
		root := newTestPrereqTree()
		if list, ok := root.Prereqs[1].(*gurps.PrereqList); ok {
			list.All = true
		}
		p, _ := showPrereqPanel(t, screen, &root, false)
		screen.Do(func() { p.toggle(one.open) })
		screen.Do(func() {
			opened := p.node(one.open).PrereqType()
			menuAction(p.moreEntries(p.node(one.path), one.path), one.label)()
			c.Equal(one.want, p.open, "%s %s with %s open", one.label, one.path, one.open)
			if one.want != "" {
				c.Equal(opened, p.node(one.want).PrereqType())
			}
		})
	}
}

// TestPrereqPanelShowsAnyNumberItCantOffer checks that a numeric comparison of "anything", which only a file edited by
// hand can hold where it isn't offered, is shown rather than left blank, and that showing it changes nothing.
func TestPrereqPanelShowsAnyNumberItCantOffer(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	sp := gurps.NewSpellPrereq()
	sp.QuantityCriteria.Compare = criteria.AnyNumber
	root.Prereqs = gurps.Prereqs{sp.Clone(root)}
	hash := gurps.Hash64(root)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		popup, isPopup := p.FindRefKey("r.0:quantitycmp").Self.(*unison.PopupMenu[criteria.NumericComparison])
		c.True(isPopup)
		c.Equal("anything", popup.Text())
	})
	c.Equal(hash, gurps.Hash64(root))
}

// TestPrereqPanelWeightAnything checks that a contained weight of "anything", which only a file edited by hand can
// hold, shows its comparison without a weight field.
func TestPrereqPanelWeightAnything(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	wp := gurps.NewContainedWeightPrereq(nil)
	wp.WeightCriteria.Compare = criteria.AnyNumber
	root.Prereqs = gurps.Prereqs{wp.Clone(root)}
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.0") })
	screen.Do(func() {
		popup, isPopup := p.FindRefKey("r.0:weightcmp").Self.(*unison.PopupMenu[criteria.NumericComparison])
		c.True(isPopup)
		if isPopup {
			c.Equal("which is anything", popup.Text())
		}
		c.Nil(p.FindRefKey("r.0:weight"), "and no weight field")
	})
}

// TestPrereqPanelLayoutFollowsRebuild checks that a rebuild outside a dock, as in a dialog, lays the window's content
// out again before it hands out the focus, so that the panel already has the size of its new rows when the focus
// lands in them, rather than only at the next draw.
func TestPrereqPanelLayoutFollowsRebuild(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	var before, atFocus, pref float32
	screen.Do(func() {
		before = p.FrameRect().Height
		host.FocusChangeInHierarchyCallback = func(_, _ *unison.Panel) {
			atFocus = p.FrameRect().Height
			_, size, _ := p.Sizes(geom.Size{Width: p.FrameRect().Width})
			pref = size.Height
		}
		p.toggle("r.0")
	})
	c.True(atFocus > before, "the panel grows: %v from %v", atFocus, before)
	c.Equal(pref, atFocus)
}

// TestPrereqPanelMoves checks that Move up and Move down step into an adjacent group and out of the ends of one, and
// that Ungroup is offered only where it keeps the meaning.
func TestPrereqPanelMoves(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	move := func(path, label string) {
		screen.Do(func() {
			act := menuAction(p.moreEntries(p.node(path), path), label)
			c.NotNil(act, "%s offers %s", path, label)
			if act != nil {
				act()
			}
		})
	}
	screen.Do(func() {
		c.Nil(menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Move Up"), "nothing above the top")
	})
	move("r.0", "Move Down")
	c.Equal([]prereq.Type{prereq.List, prereq.Trait, prereq.Skill, prereq.Script, prereq.Unknown}, prereqShape(root),
		"down into the group below, at its top")
	move("r.0.0", "Move Up")
	c.Equal([]prereq.Type{prereq.Trait, prereq.List, prereq.Skill, prereq.Script, prereq.Unknown}, prereqShape(root),
		"up out of the top of a group")
	move("r.2", "Move Up")
	c.Equal([]prereq.Type{prereq.Trait, prereq.List, prereq.Skill, prereq.Script, prereq.Unknown}, prereqShape(root),
		"up into the group above, at its bottom")
	list, ok := root.Prereqs[1].(*gurps.PrereqList)
	c.True(ok)
	c.Equal(prereq.Unknown, list.Prereqs[2].PrereqType())
	c.True(list.Prereqs[2].ParentList() == list, "a moved node belongs to its new list")

	screen.Do(func() {
		c.Nil(menuAction(p.moreEntries(list, "r.1"), "Ungroup"), "an any-of group in an all-of list stays")
		list.All = true
		c.NotNil(menuAction(p.moreEntries(list, "r.1"), "Ungroup"), "unless the modes match")
		list.WhenTL.Compare = criteria.AtLeastNumber
		c.Nil(menuAction(p.moreEntries(list, "r.1"), "Ungroup"), "and it has no tech level")
	})
}

// TestPrereqPanelEscapeClosesTheOpenRow checks that Escape closes an open row, returning the focus to its sentence,
// that Escape within the panel never reaches the editor, where it would discard the changes, and that Escape outside the
// panel still does.
func TestPrereqPanelEscapeClosesTheOpenRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.1.0") })
	screen.Do(func() {
		c.NotNil(p.Window().Focus(), "opening a row focuses its first control")
		grip := func(path string) float32 {
			row := p.FindRefKey(path + keyMore).Parent()
			return row.Children()[0].RectToRoot(row.Children()[0].ContentRect(true)).X
		}
		c.Equal(grip("r.1.1"), grip("r.1.0"), "an open row's grip lines up with a closed one's")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("", p.open)
	c.Equal(0, host.escapes)
	screen.Do(func() { c.Equal(p.FindRefKey("r.1.0"+keySentence), p.Window().Focus()) })
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(0, host.escapes, "with no row open, Escape in the panel does nothing")
	screen.Do(func() {
		outside := unison.NewField()
		host.AddChild(outside)
		host.MarkForLayoutAndRedraw()
		outside.RequestFocus()
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal(1, host.escapes, "Escape outside the panel reaches the editor")
}

// TestPrereqPanelDoneKeyActsOnce checks that Return on an open row's Done button closes the row, and that a held
// Return's repeats, even once they reach another Done button, don't close that row as well.
func TestPrereqPanelDoneKeyActsOnce(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	done := func(path string) {
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { p.FindRefKey(path + ":done").RequestFocus() })
	}
	done("r.0")
	screen.KeyDown(unison.KeyReturn, mod.None)
	screen.Do(func() { c.Equal("", p.open, "Return closes the row") })
	done("r.1.0")
	screen.KeyDown(unison.KeyReturn, mod.None)
	screen.Do(func() { c.Equal("r.1.0", p.open, "a repeat of the held Return leaves the row open") })
	screen.KeyUp(unison.KeyReturn, mod.None)
	screen.KeyPress(unison.KeyReturn, mod.None)
	screen.Do(func() { c.Equal("", p.open, "until it is pressed again") })
}

// TestPrereqPanelCollapse checks that the title bar collapses the panel to a paragraph describing the tree and expands
// it again, as does the paragraph, from a click or the keyboard, that a screen reader hears whether it is expanded, that
// collapsing hands the focus from the rows to the title bar, and that it is no edit and keeps the open row open.
func TestPrereqPanelCollapse(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	c.Equal(3, len(root.Prereqs), "precondition: three prerequisites at the top")
	p, host := showPrereqPanel(t, screen, &root, false)
	hash := gurps.Hash64(root)
	expanded := func() bool {
		node := &accessibility.Node{}
		p.collapse.Accessibility.Callback(node)
		c.True(node.Expandable)
		return node.Expanded
	}
	summary := func() *sentenceButton {
		b, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		return b
	}
	// inOpenRow reports whether the focus is within the editor of the open row.
	inOpenRow := func() bool {
		focus := p.Window().Focus()
		editor := p.FindRefKey("r.0" + keyFirst)
		return focus != nil && editor != nil && editor.HasInSelfOrDescendants(func(one *unison.Panel) bool {
			return one == focus
		})
	}
	screen.Do(func() {
		c.Equal(role.DisclosureTriangle, p.collapse.Accessibility.Role)
		c.Equal("Prerequisites", p.collapse.Accessibility.Name, "the title bar is named for the title")
		c.True(p.collapse.Focusable())
		c.True(expanded(), "precondition: the panel is expanded")
		c.Nil(p.FindRefKey(sectionSummaryKey))
		p.toggle("r.0")
	})
	screen.Do(func() {
		c.True(inOpenRow(), "precondition: the focus is in the open row")
		p.collapse.MouseUpCallback(geom.Point{X: 1, Y: 1}, 0, mod.None)
	})
	screen.Do(func() {
		c.False(expanded(), "clicking the title bar collapses the panel")
		for _, key := range []string{"r.0" + keyFirst, "r.1.0" + keySentence, treeRootPath + keyPill} {
			c.Nil(p.FindRefKey(key), "collapsing hides %s", key)
		}
		c.Equal(0, len(p.views), "and the statuses of the rows")
		c.Equal(`Has trait "" and (has skill "" at level at least 0 or a custom check) and `+
			`meets an unknown type of prerequisite ("future") that needs a newer version of GCS.`,
			summary().Accessibility.Name, "the paragraph describes the tree")
		c.True(strings.HasSuffix(summary().plainText(), "."), "the paragraph ends with a period")
		c.Equal(p.collapse.AsPanel(), p.Window().Focus(), "the focus moves from the rows to the title bar")
		// Nothing but the title strip and the paragraph takes room: the border's insets, as under a plain title, and
		// the 2 point inset all round.
		paragraph := summary()
		_, pref, _ := p.Sizes(geom.Size{Width: p.FrameRect().Width})
		_, text, _ := paragraph.Sizes(geom.Size{Width: paragraph.FrameRect().Width})
		plain := &TitledBorder{Title: p.collapse.border.Title, Font: p.collapse.border.Font}
		c.Equal(plain.Insets().Height()+4+text.Height, pref.Height, "the collapsed panel is only as tall as it shows")
	})
	screen.KeyPress(unison.KeyEscape, mod.None)
	c.Equal("r.0", p.open, "the open row stays open, even through Escape")
	c.Equal(0, host.escapes, "which doesn't reach the editor")
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() {
		c.True(expanded(), "Space expands the panel")
		c.Nil(p.FindRefKey(sectionSummaryKey), "and hides the paragraph")
		c.Nil(p.paragraph)
		c.NotNil(p.FindRefKey("r.0"+keyFirst), "with the open row still open")
		c.Equal(p.collapse.AsPanel(), p.Window().Focus(), "and the focus left on the title bar")
	})
	screen.KeyPress(unison.KeyReturn, mod.None)
	screen.Do(func() {
		c.False(expanded(), "Return collapses it")
		summary().RequestFocus()
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() {
		c.True(expanded(), "the paragraph expands the panel")
		c.Nil(p.FindRefKey(sectionSummaryKey))
		c.True(inOpenRow(), "and the open row takes the focus from it")
	})
	c.Equal(hash, gurps.Hash64(root), "collapsing changes nothing")
	c.False(host.mgr.CanUndo(), "and is not an edit")
	c.Equal(0, host.modified, "nor marks the editor modified")
}

// TestPrereqPanelCollapsedFollowsTree checks that a collapsed panel's paragraph follows the tree when the editor syncs,
// without building the rows or their statuses.
func TestPrereqPanelCollapsedFollowsTree(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	trait := gurps.NewTraitPrereq()
	trait.NameCriteria.Qualifier = "Magery"
	broken := gurps.NewScriptPrereq()
	broken.Script = "nope("
	root := gurps.NewPrereqList()
	root.Prereqs = gurps.Prereqs{trait, broken}
	root = root.CloneAsPrereqList(nil)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(p.collapse.toggle)
	text := func() (s string) {
		screen.Do(func() {
			b, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
			c.True(ok, "a collapsed panel shows a paragraph")
			if ok {
				s = b.plainText()
			}
		})
		return s
	}
	c.Contains(text(), "Magery")
	screen.Do(func() {
		p.edit("", "", "", func() {
			if one, ok := root.Prereqs[0].(*gurps.TraitPrereq); ok {
				one.NameCriteria.Qualifier = "Luck"
			}
		})
	})
	waitForEvaluation(screen)
	c.Contains(text(), "Luck", "the paragraph follows the change once the editor syncs")
	c.True(strings.HasSuffix(text(), "."), "and still ends with a period")
	screen.Do(func() {
		c.Equal(gurps.Hash64(root), p.hash, "the change was taken in")
		c.Equal(0, len(p.views), "with no rows built to show a status")
		p.refresh()
	})
	c.Contains(text(), "Luck", "refreshing again changes nothing")
}

// TestPrereqPanelStartingState checks that a panel with prerequisites starts out collapsed, leaving the focus where the
// window put it, and one without starts out open with its placeholder, and that adding the first prerequisite or
// deleting the last leaves it open.
func TestPrereqPanelStartingState(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	var missing *gurps.PrereqList
	c.Equal(3, len(root.Prereqs), "precondition: three prerequisites at the top")
	var full, open *prereqPanel
	var field *unison.Field
	host := &sentenceUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
	screen.Do(func() {
		host.Self = host
		host.SetLayout(&unison.FlexLayout{Columns: 1})
		field = unison.NewField()
		host.AddChild(field)
		full = newPrereqPanel(gurps.NewEntity(), &root, prereq.TypesForNonEquipment, false)
		open = newPrereqPanel(gurps.NewEntity(), &missing, prereq.TypesForNonEquipment, false)
		host.AddChild(full)
		host.AddChild(open)
	})
	wnd := showInTestWindow(t, screen, 700, host)
	screen.Do(func() {
		c.True(full.collapse.collapsed, "a panel with prerequisites starts out collapsed")
		c.NotNil(full.FindRefKey(sectionSummaryKey), "showing the paragraph")
		c.Nil(full.FindRefKey("r.0"+keySentence), "in place of the rows")
		c.False(open.collapse.collapsed, "a panel without prerequisites starts out open")
		c.NotNil(open.FindRefKey(treeRootPath+":empty"), "showing the placeholder")
		c.Equal(field.AsPanel(), wnd.Focus(), "neither takes the focus from the field ahead of them")
		menuAction(open.addEntries(open.tree(), treeRootPath), "Trait")()
	})
	c.Equal(1, len(missing.Prereqs), "precondition: a prerequisite was added")
	screen.Do(func() {
		c.False(open.collapse.collapsed, "adding the first prerequisite leaves the panel open")
		menuAction(open.moreEntries(open.node("r.0"), "r.0"), "Delete")()
	})
	c.Equal(0, len(missing.Prereqs), "precondition: the prerequisite was deleted")
	screen.Do(func() {
		c.False(open.collapse.collapsed, "deleting the last leaves it open")
		c.NotNil(open.FindRefKey(treeRootPath + ":empty"))
	})
}

// TestPrereqPanelCollapseEmpty checks that a collapsed panel with no prerequisites says so, even when its root has a
// tech level condition.
func TestPrereqPanelCollapseEmpty(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var missing *gurps.PrereqList
	p, _ := showPrereqPanel(t, screen, &missing, false)
	screen.Do(p.collapse.toggle)
	screen.Do(func() {
		b, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		if ok {
			c.Equal("No prerequisites.", b.plainText())
		}
		c.Nil(p.FindRefKey(treeRootPath+":empty"), "in place of the placeholder")
	})
	c.Nil(missing, "collapsing doesn't make the missing list")

	// A tech level condition is all an empty root has to describe, which is still no prerequisites.
	screen.Do(p.collapse.toggle)
	screen.Do(func() { menuAction(p.addEntries(p.tree(), treeRootPath), "Only When TL…")() })
	screen.Do(p.collapse.toggle)
	screen.Do(func() {
		c.True(p.headed, "precondition: the root shows its head")
		b, ok := p.FindRefKey(sectionSummaryKey).Self.(*sentenceButton)
		c.True(ok, "a collapsed panel shows a paragraph")
		if ok {
			c.Equal("No prerequisites.", b.plainText(), "an empty root with a tech level condition says so too")
		}
	})
}

// TestPrereqPanelTitleBar checks that the title bar covers the strip the title is drawn in, and that the rows start
// where they would under a title that can't be clicked.
func TestPrereqPanelTitleBar(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() {
		border := p.collapse.border
		c.Equal(border.TitleStrip(p.FrameRect().Size), p.collapse.FrameRect())
		children := p.Children()
		c.True(len(children) > 1 && children[0] == p.collapse.AsPanel(), "the title bar comes first, then the rows")
		if len(children) > 1 {
			plain := &TitledBorder{Title: border.Title, Font: border.Font}
			c.Equal(plain.Insets().Top+2, children[1].FrameRect().Y, "the first row starts below the title")
		}
	})
}

// TestPrereqPanelStatus checks the status of each row and group against the sheet, as an icon, a tooltip and in the
// accessible names, and that the sentences follow the tree when it changes.
func TestPrereqPanelStatus(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	trait := gurps.NewTraitPrereq()
	trait.NameCriteria.Qualifier = "Magery"
	met := gurps.NewScriptPrereq()
	met.Script = "true"
	broken := gurps.NewScriptPrereq()
	broken.Script = "nope("
	later := gurps.NewPrereqList()
	later.WhenTL.Compare = criteria.AtLeastNumber
	later.WhenTL.Qualifier = fxp.Ten
	later.Prereqs = gurps.Prereqs{gurps.NewTraitPrereq()}
	root := gurps.NewPrereqList()
	holder := gurps.NewPrereqList()
	holder.Prereqs = gurps.Prereqs{gurps.NewPrereqList()}
	outOfTL := gurps.NewPrereqList()
	outOfTL.Prereqs = gurps.Prereqs{later.Clone(nil)}
	limbo := gurps.NewPrereqList()
	limbo.Prereqs = gurps.Prereqs{met.Clone(nil), broken.Clone(nil)}
	root.Prereqs = gurps.Prereqs{trait, met, broken, later, gurps.NewPrereqList(), holder, outOfTL, limbo}
	root = root.CloneAsPrereqList(nil)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() {
		for path, want := range map[string]*unison.SVG{
			"r": svg.Not, "r.0": svg.Not, "r.1": unison.CheckmarkSVG, "r.2": unison.TriangleExclamationSVG, "r.3": svg.CircledMinus,
			"r.3.0": svg.CircledMinus, "r.4": svg.CircledMinus, "r.5": svg.CircledMinus, "r.5.0": svg.CircledMinus,
			"r.7": unison.TriangleExclamationSVG,
		} {
			for _, v := range p.views {
				if v.node == p.node(path) {
					drawable, ok := v.icon.Drawable.(*unison.DrawableSVG)
					c.True(ok && drawable.SVG == want, "the status icon of %s", path)
				}
			}
		}
		sentence, ok := p.FindRefKey("r.0" + keySentence).Self.(*sentenceButton)
		c.True(ok)
		c.Equal("Has trait Magery, not met", sentence.Accessibility.Name)
		c.Nil(p.paragraph, "an expanded panel shows no summary over the tree")
		checks := p.checks()
		for path, want := range map[string]string{
			"r.0": "Not met: Has trait Magery", "r.3": "Doesn't apply at this tech level",
			"r.3.0": "Doesn't apply at this tech level", "r.4": "Empty group, left out of the check",
			"r.5": "Holds nothing to check, left out of the check", "r.5.0": "Empty group, left out of the check",
			"r.6": "Nothing in it applies at this tech level, left out of the check",
		} {
			_, tip, _ := p.status(p.node(path), checks)
			c.Equal(want, tip, path)
		}
		_, tip, suffix := p.status(p.node("r"), checks)
		c.Contains(tip, "A custom check (couldn't run: SyntaxError: ")
		c.Equal("not met", suffix, "an unmet requirement decides an all of group, even beside a script that couldn't run")
		_, tip, suffix = p.status(p.node("r.2"), checks)
		c.True(strings.HasPrefix(tip, "Not met: A custom check (couldn't run: SyntaxError: "), tip)
		c.True(strings.HasPrefix(suffix, "couldn't run: SyntaxError: "), suffix)
		_, _, suffix = p.status(p.node("r.7"), checks)
		c.Equal("couldn't be checked", suffix, "a group nothing decides fails when a script in it couldn't run")
		p.toggle("r.0")
	})
	screen.Do(func() {
		for _, v := range p.views {
			c.True(v.node != p.node("r.0"), "an open row shows no status")
		}
		p.toggle("r.0")
	})
	screen.Do(func() {
		p.edit("", "", "", func() {
			if one, ok := root.Prereqs[0].(*gurps.TraitPrereq); ok {
				one.NameCriteria.Qualifier = "Luck"
			}
		})
	})
	screen.Do(p.refresh)
	screen.Do(func() {
		sentence, ok := p.FindRefKey("r.0" + keySentence).Self.(*sentenceButton)
		c.True(ok)
		c.Equal("Has trait Luck, not met", sentence.Accessibility.Name, "the sentence follows the change, in place")
		sentence, ok = p.FindRefKey("r.3.0" + keySentence).Self.(*sentenceButton)
		c.True(ok)
		c.Equal(`Has trait "", doesn't apply at this tech level`, sentence.Accessibility.Name)
	})
	group := func() (name string) {
		screen.Do(func() { name = p.FindRefKey("r.3" + keyPill).Parent().Parent().Accessibility.Name })
		return name
	}
	c.Equal("All of, only when TL at least 10, doesn't apply at this tech level", group(), "a group's name has its status")
	screen.Do(func() {
		p.edit("", "", "", func() {
			if list, ok := root.Prereqs[3].(*gurps.PrereqList); ok {
				list.WhenTL.Qualifier = fxp.Twelve
			}
		})
	})
	screen.Do(p.refresh)
	c.Equal("All of, only when TL at least 12, doesn't apply at this tech level", group(),
		"a group's name follows its tech level")
	screen.Do(func() {
		p.edit("", "", "", func() { root.Prereqs = gurps.Prereqs{later} })
		status, tip, suffix := p.status(p.node("r"), p.checks())
		c.Equal(gurps.CheckMet, status, "a top level with nothing left to check is met, as on the sheet")
		c.Equal("Met", tip)
		c.Equal("met", suffix)
	})
}

// TestPrereqPanelScriptResult checks that the script editor's result line says the outcome, and that changing the
// description evaluates it again.
func TestPrereqPanelScriptResult(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	root.Prereqs = gurps.Prereqs{gurps.NewScriptPrereq()}
	root = root.CloneAsPrereqList(nil)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.0") })
	var editor *scriptEditor
	screen.Do(func() {
		for panel := p.FindRefKey("r.0:script"); editor == nil; panel = panel.Parent() {
			if one, ok := panel.Self.(*scriptEditor); ok {
				editor = one
			}
		}
	})
	result := func(script string) (text string) {
		screen.Do(func() { editor.field.SetText(script) })
		waitForEvaluation(screen)
		screen.Do(func() { text = editor.result.plainText() })
		return text
	}
	c.Equal("Passed", result("true"))
	c.Equal("Failed", result("false"))
	c.Equal("Failed: Too weak", result(`"Too weak"`))
	c.Equal("Failed: a\nb", result(`"a\nb"`))
	c.True(strings.HasPrefix(result(`throw new Error("boom")`), "Couldn't run: "), "an error says the script couldn't run")

	var evaluations int
	screen.Do(func() {
		evaluate := editor.opts.Evaluate
		editor.opts.Evaluate = func(script string) (gurps.CheckResult, string) {
			evaluations++
			return evaluate(script)
		}
		field, ok := p.FindRefKey("r.0:name").Self.(*StringField)
		c.True(ok)
		field.SetText("Strong")
	})
	waitForEvaluation(screen)
	c.Equal(1, evaluations, "changing the description evaluates the script again")
}

// TestPrereqPanelDragAndDrop checks where a dragged prerequisite goes before, after or into what it is dropped on, that
// nothing goes into itself, and that a drop is undone as one step.
func TestPrereqPanelDragAndDrop(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	drag := func(from, onto string, fraction float32) (accepted bool) {
		screen.Do(func() {
			target := p.FindRefKey(onto + keyMore).Parent()
			r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
			where := geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
			data := &rowDrag{panel: p.AsPanel(), path: from}
			p.dragOver(where, data)
			accepted = p.dropTarget != nil
			p.drop(where, data)
		})
		return accepted
	}
	c.False(drag("r.1", "r.1.0", 0.5), "a group can't go into itself")
	c.False(drag("r.0", "r.0", 0.2), "nor a row onto itself")
	c.True(drag("r.0", "r.1", 0.8), "below the top of a group's head is into it")
	c.Equal([]prereq.Type{prereq.List, prereq.Skill, prereq.Script, prereq.Trait, prereq.Unknown}, prereqShape(root))
	c.True(drag("r.1", "r.0.2", 0.2), "the top of a row is before it")
	c.Equal([]prereq.Type{prereq.List, prereq.Skill, prereq.Script, prereq.Unknown, prereq.Trait}, prereqShape(root))
	c.True(drag("r.0.3", "r.0", 0.2), "the top of a group's head is before the group")
	c.Equal([]prereq.Type{prereq.Trait, prereq.List, prereq.Skill, prereq.Script, prereq.Unknown}, prereqShape(root))
	screen.Do(host.mgr.Undo)
	c.Equal([]prereq.Type{prereq.List, prereq.Skill, prereq.Script, prereq.Unknown, prereq.Trait}, prereqShape(root),
		"a drop is one step to undo")
	c.True(drag("r.0.0", treeRootPath, 0.5), "the root's head is into the root")
	c.Equal([]prereq.Type{prereq.List, prereq.Script, prereq.Unknown, prereq.Trait, prereq.Skill}, prereqShape(root),
		"at its end")
}

// TestPrereqPanelDropAfterGroup checks that a drop beside the last child of a group that ends its parent goes after
// the group, at the end of the parent.
func TestPrereqPanelDropAfterGroup(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	root.Prereqs = root.Prereqs[:2]
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() {
		group := p.FindRefKey("r.1:group")
		last := p.FindRefKey("r.1.1" + keyMore).Parent()
		where := geom.NewPoint(p.RectFromRoot(group.RectToRoot(group.ContentRect(true))).X+4,
			p.RectFromRoot(last.RectToRoot(last.ContentRect(true))).Bottom()-2)
		data := &rowDrag{panel: p.AsPanel(), path: "r.0"}
		p.dragOver(where, data)
		c.Equal(group, p.dropTarget)
		c.Equal(dropAfter, p.dropWhere)
		p.drop(where, data)
	})
	c.Equal([]prereq.Type{prereq.List, prereq.Skill, prereq.Script, prereq.Trait}, prereqShape(root))
}

// TestPrereqPanelDragFromRow checks that a row can be dragged by its sentence, and that a click on a sentence, even a
// slow one, still opens its row.
func TestPrereqPanelDragFromRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	var sentence, head *unison.Panel
	var below geom.Point
	screen.Do(func() {
		registerWindowDragTypes(p.Window())
		sentence = p.FindRefKey("r.0" + keySentence)
		head = p.FindRefKey("r.1" + keyMore).Parent()
		below = geom.NewPoint(40, head.FrameRect().Height*0.8)
	})
	screen.Drag(screen.PanelCenter(sentence), screen.PanelPoint(head, below), 10)
	c.Equal([]prereq.Type{prereq.List, prereq.Skill, prereq.Script, prereq.Trait, prereq.Unknown}, prereqShape(root))
	c.Equal("", p.open, "the drag didn't also click")
	screen.Do(func() { sentence = p.FindRefKey("r.0.0" + keySentence) })
	screen.Click(screen.PanelCenter(sentence))
	c.Equal("r.0.0", p.open)

	screen.Do(func() { sentence = p.FindRefKey("r.0.1" + keySentence) })
	at := screen.PanelCenter(sentence)
	screen.MouseDown(at, unison.ButtonLeft, mod.None)
	time.Sleep(300 * time.Millisecond)
	screen.MouseMove(geom.NewPoint(at.X+2, at.Y), mod.None)
	screen.MouseUp(geom.NewPoint(at.X+2, at.Y), unison.ButtonLeft, mod.None)
	c.Equal("r.0.1", p.open, "a slow click that barely moves is still a click")
}

// TestPrereqPanelLevelChip checks that a skill's level of at least 0 is shown, that removing it leaves no level, and
// that a skill or trait the editor adds starts with none.
func TestPrereqPanelLevelChip(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { p.toggle("r.1.0") })
	screen.Do(func() {
		buttons := panelsOfType[*unison.Button](p.FindRefKey("r.1.0:level" + keyChip))
		c.NotEqual(0, len(buttons), "a skill's level of at least 0 is a chip")
		buttons[len(buttons)-1].ClickCallback()
		for _, label := range []string{"Skill", "Trait"} {
			menuAction(p.addEntries(p.tree(), treeRootPath), label)()
		}
	})
	skill, ok := p.node("r.1.0").(*gurps.SkillPrereq)
	c.True(ok)
	c.Equal(criteria.AnyNumber, skill.LevelCriteria.Compare, "removing the chip leaves no level")
	newSkill, ok := p.node("r.3").(*gurps.SkillPrereq)
	c.True(ok)
	c.Equal(criteria.AnyNumber, newSkill.LevelCriteria.Compare)
	newTrait, ok := p.node("r.4").(*gurps.TraitPrereq)
	c.True(ok)
	c.Equal(criteria.AnyNumber, newTrait.LevelCriteria.Compare)
}

// TestPrereqPanelChipOrder checks that the buttons adding unused criteria follow the chips in use, and that an added
// criterion takes its place among those chips.
func TestPrereqPanelChipOrder(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	order := func() []string {
		children := p.FindRefKey("r.1.0:level" + keyChip).Parent().Children()
		keys := make([]string, 0, len(children))
		for _, child := range children {
			keys = append(keys, child.RefKey[len("r.1.0:"):])
		}
		return keys
	}
	screen.Do(func() { p.toggle("r.1.0") })
	screen.Do(func() {
		c.Equal([]string{"level" + keyChip, "add specialization", "add optspecialization"}, order())
		panelsOfType[*unison.Button](p.FindRefKey("r.1.0:add specialization"))[0].ClickCallback()
	})
	screen.Do(func() {
		c.Equal([]string{"specialization" + keyChip, "level" + keyChip, "add optspecialization"}, order())
	})
}

// TestPrereqPanelEmptyRoot checks that an empty root shows only its placeholder and add button, without a summary,
// pill, status or more button, and isn't a group to a screen reader.
func TestPrereqPanelEmptyRoot(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var missing *gurps.PrereqList
	p, _ := showPrereqPanel(t, screen, &missing, false)
	screen.Do(func() {
		c.Nil(p.FindRefKey(sectionSummaryKey))
		c.Nil(p.FindRefKey(treeRootPath + keyPill))
		c.Equal(0, len(p.views))
		c.NotNil(p.FindRefKey(treeRootPath + ":empty"))
		c.Equal(p.FindRefKey(treeRootPath+":empty").Parent(), p.FindRefKey(treeRootPath+keyAdd).Parent(),
			"its add button beside the placeholder")
		c.Nil(p.FindRefKey(treeRootPath+keyMore), "and no more button")
		box := p.FindRefKey(treeRootPath + ":empty").Parent().Parent()
		c.NotEqual(role.Group, box.Accessibility.Role, "nor is it a group to a screen reader")
		c.Equal("", box.Accessibility.Name)
	})
}

// TestPrereqPanelEmptyRootGroupType checks that choosing a group type or a tech level from the Add menu of an empty
// root that shows its placeholder alone gives the root itself that type or condition, showing its head over a group's
// placeholder, that undo takes it back to the single line, and that once the root shows its head, choosing a group
// type adds a group rather than changing the root's.
func TestPrereqPanelEmptyRootGroupType(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var root *gurps.PrereqList
	p, host := showPrereqPanel(t, screen, &root, false)
	headed := func() (pill bool, placeholder string) {
		screen.Do(func() {
			pill = p.FindRefKey(treeRootPath+keyPill) != nil
			if b, ok := p.FindRefKey(treeRootPath + ":empty").Self.(*unison.Button); ok {
				placeholder = b.Text.String()
			}
		})
		return pill, placeholder
	}
	choose := func(label string) {
		screen.Do(func() { menuAction(p.addEntries(p.tree(), treeRootPath), label)() })
	}
	single := "No prerequisites. Click here to add one."
	group := "Empty group. Add a requirement or drag one here."
	pill, text := headed()
	c.False(pill)
	c.Equal(single, text)
	choose("Any of Group")
	pill, text = headed()
	c.True(pill, "the root takes the group type")
	c.Equal(group, text)
	screen.Do(func() {
		c.Equal(0, len(root.Prereqs))
		c.False(root.All)
		c.Nil(p.FindRefKey(treeRootPath+keyMore), "its head has no add button")
		adds := 0
		var count func(panel *unison.Panel)
		count = func(panel *unison.Panel) {
			if panel.RefKey == treeRootPath+keyAdd {
				adds++
			}
			for _, child := range panel.Children() {
				count(child)
			}
		}
		count(p.AsPanel())
		c.Equal(1, adds, "only the one beside its placeholder")
	})
	screen.Do(host.mgr.Undo)
	pill, text = headed()
	c.False(pill, "undo goes back to the single line")
	c.Equal(single, text)
	screen.Do(func() {
		focus := p.Window().Focus()
		c.NotNil(focus)
		if focus != nil {
			c.Equal(treeRootPath+keyAdd, focus.RefKey, "giving the focus to what adds to the root")
		}
	})
	choose("All of Group")
	pill, _ = headed()
	c.True(pill, "choosing the type the root already has still shows its head")
	screen.Do(host.mgr.Undo)
	choose("Only When TL…")
	pill, _ = headed()
	c.True(pill, "a tech level condition shows the head")
	screen.Do(func() {
		c.NotNil(p.FindRefKey(treeRootPath + ":tl" + keyChip))
		c.Equal(0, len(root.Prereqs))
		c.True(root.All)
	})
	screen.Do(host.mgr.Undo)
	pill, _ = headed()
	c.False(pill, "undo takes the condition away again")

	choose("Any of Group")
	choose("All of Group")
	screen.Do(func() {
		c.False(root.All, "the root keeps the type it was given")
		c.Equal(1, len(root.Prereqs), "and holds a new group")
		if len(root.Prereqs) == 1 {
			list, ok := root.Prereqs[0].(*gurps.PrereqList)
			c.True(ok && list.All, "of the type chosen")
		}
	})
}

// TestPrereqPanelEmptiedRootKeepsItsHead checks that a root emptied of its last prerequisite still shows its head when
// it is "Any of" or has a tech level condition, so that neither is hidden.
func TestPrereqPanelEmptiedRootKeepsItsHead(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	root.All = false
	root.WhenTL = criteria.Number{Compare: criteria.AtMostNumber, Qualifier: fxp.FromInteger(defaultWhenTL)}
	root.Prereqs = gurps.Prereqs{gurps.NewTraitPrereq()}
	root = root.CloneAsPrereqList(nil)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { menuAction(p.moreEntries(root.Prereqs[0], "r.0"), "Delete")() })
	screen.Do(func() {
		c.Equal(0, len(root.Prereqs))
		c.NotNil(p.FindRefKey(treeRootPath+keyPill), "its group type shows")
		c.NotNil(p.FindRefKey(treeRootPath+":tl"+keyChip), "as does its tech level condition")
	})
}

// TestPrereqPanelUndoReopensRow checks that a row the Add menu adds opens with the focus in its name field, and that
// undo and redo open whichever row was open when the change was made, or none.
func TestPrereqPanelUndoReopensRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { menuAction(p.addEntries(p.tree(), treeRootPath), "Trait")() })
	screen.Do(func() { c.Equal("r.3:name", p.Window().Focus().RefKey, "a new row focuses its name field") })
	screen.Type("Luck")
	screen.Do(func() { p.toggle("r.3") })
	screen.Do(host.mgr.Undo)
	screen.Do(func() {
		c.Equal("r.3", p.open, "the row the undone change was made in opens")
		c.Equal("r.3:name", p.Window().Focus().RefKey)
	})
	screen.Do(func() { p.toggle("r.3") })
	screen.Do(func() { menuAction(p.moreEntries(p.node("r.0"), "r.0"), "Duplicate")() })
	screen.Do(func() { p.toggle("r.1") })
	screen.Do(host.mgr.Undo)
	c.Equal("", p.open, "a change made with no row open closes the open row when undone")
	screen.Do(host.mgr.Redo)
	c.Equal("", p.open, "and when redone")
}

// TestPrereqPanelMenus checks the wording of the Add menu, and that a count of colleges reads as its sentence does.
func TestPrereqPanelMenus(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	sp := gurps.NewSpellPrereq()
	sp.SubType = spellcmp.CollegeCount
	root.Prereqs = gurps.Prereqs{sp.Clone(root)}
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() {
		for _, label := range []string{"Equipped Equipment", "All of Group", "Any of Group", "Only When TL…"} {
			c.NotNil(menuAction(p.addEntries(p.tree(), treeRootPath), label), label)
		}
		p.toggle("r.0")
	})
	screen.Do(func() {
		var keys []string
		for _, child := range p.FindRefKey("r.0:match").Parent().Children() {
			if child.RefKey != "" {
				keys = append(keys, child.RefKey)
			}
		}
		c.Equal([]string{"r.0:has", "r.0:type", "r.0:match", "r.0:quantitycmp", "r.0:quantity"}, keys)
	})
}

// TestPrereqPanelControlNamesDiffer checks that no two controls of an open row share an accessible name.
func TestPrereqPanelControlNamesDiffer(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := gurps.NewPrereqList()
	for _, one := range prereq.TypesForNonEquipment {
		root.Prereqs = append(root.Prereqs, (&prereqPanel{}).createPrereqForType(one, root))
	}
	seedEveryPrereqControl(root)
	p, _ := showPrereqPanel(t, screen, &root, false)
	for i := range prereq.TypesForNonEquipment {
		path := childPath(treeRootPath, i)
		screen.Do(func() { p.toggle(path) })
		names := make(map[string]bool)
		var controls []*unison.Panel
		var wnd *unison.Window
		screen.Do(func() {
			wnd = p.Window()
			p.FindRefKey(path + keyFirst).HasInSelfOrDescendants(func(one *unison.Panel) bool {
				if one.Focusable() {
					controls = append(controls, one)
				}
				return false
			})
		})
		c.NotNil(screen.AccessibilityTree(wnd))
		c.True(len(controls) > 2, path)
		for _, one := range controls {
			if node := screen.AccessibilityNodeFor(one); node != nil && node.Name != "" {
				c.False(names[node.Name], "%s: %s is shared", path, node.Name)
				names[node.Name] = true
			}
		}
	}
}

// TestPrereqPanelGroupMenusAdd checks that a nested group has no add button in its head and that its more menu starts
// with what can be added to it, under headings that say so, that the root has an add button in place of a more button,
// that each adds into its own group with the focus as the add button gave it, and that an empty group keeps its add
// button beside its placeholder.
func TestPrereqPanelGroupMenusAdd(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	more := func(path string) []menuEntry {
		var entries []menuEntry
		screen.Do(func() {
			c.NotNil(p.FindRefKey(path+keyMore), "%s has a more button, or the root its add button", path)
			// The menu the button would show, gathered rather than popped up.
			if path == treeRootPath {
				entries = p.addEntries(p.tree(), path)
			} else {
				entries = p.moreEntries(p.node(path), path)
			}
		})
		return entries
	}
	screen.Do(func() {
		for _, path := range []string{treeRootPath, "r.1"} {
			c.Nil(p.FindRefKey(path+keyAdd), "%s has no placeholder's add button", path)
			head := p.FindRefKey(path + keyPill).Parent()
			c.Equal(head, p.FindRefKey(path+keyMore).Parent(), "%s: its button is in its head", path)
		}
		c.Equal("More actions", tooltipText(p.FindRefKey("r.1"+keyMore).Tooltip), "a nested group's is a more button")
		add, ok := p.FindRefKey(treeRootPath + keyMore).Self.(*unison.Button)
		c.True(ok)
		if ok {
			c.Equal("Add to this group", tooltipText(add.Tooltip), "the root's is an add button")
			drawable, isSVG := add.Drawable.(*unison.DrawableSVG)
			c.True(isSVG && drawable.SVG == unison.CircledAddSVG, "showing the add icon")
		}
		right := func(key string) float32 {
			panel := p.FindRefKey(key)
			return p.RectFromRoot(panel.RectToRoot(panel.ContentRect(true))).Right()
		}
		c.Equal(right("r.0"+keyMore), right(treeRootPath+keyMore), "the root's more button lines up with the others")
	})
	// groupLen returns the number of children of the group at the path, or -1 if it isn't a group.
	groupLen := func(path string) int {
		if list, ok := p.node(path).(*gurps.PrereqList); ok {
			return len(list.Prereqs)
		}
		return -1
	}
	adds := menuLabels(more(treeRootPath))
	c.Equal("Requirement", adds[0])
	c.True(slices.Contains(adds, "Structure"), "the root's add menu has its headings")
	c.Equal("Only When TL…", adds[len(adds)-1], "with the tech level condition under Structure")
	nested := menuLabels(more("r.1"))
	want := slices.Clone(adds)
	for i, label := range want {
		if label == "Requirement" || label == "Structure" {
			want[i] = "Add " + label
		}
	}
	c.Equal(want, nested[:len(adds)], "a nested group's more menu starts with what can be added to it, saying so")
	c.Equal([]string{"-", "Duplicate"}, nested[len(adds):len(adds)+2], "then the rest, after a separator")
	c.Equal("Duplicate", menuLabels(more("r.0"))[0], "a row's more menu is as it was")

	// choose picks the entry with the label from the menu, on the UI thread, as a click on it would.
	choose := func(entries []menuEntry, label string) {
		screen.Do(menuAction(entries, label))
	}
	choose(more("r.1"), "Trait")
	screen.Do(func() {
		c.Equal(3, groupLen("r.1"), "the nested group's menu adds to it")
		c.Equal("r.1.2", p.open, "opening the new row")
		c.True(p.FindRefKey("r.1.2"+keyFirst).HasInSelfOrDescendants(func(one *unison.Panel) bool {
			return one == p.Window().Focus()
		}), "with the focus in it")
	})
	screen.Do(host.mgr.Undo)
	screen.Do(func() { c.Equal("r.1"+keyMore, p.Window().Focus().RefKey, "undo gives the focus to the more button") })
	choose(more(treeRootPath), "All of Group")
	screen.Do(func() {
		c.Equal(4, len(p.tree().Prereqs), "the root's menu adds to the root")
		c.Equal("r.3"+keyMore, p.Window().Focus().RefKey, "a new group takes the focus on its more button")
	})

	// An empty nested group keeps its add button, beside its placeholder.
	screen.Do(func() {
		empty := p.FindRefKey("r.3:empty")
		add := p.FindRefKey("r.3" + keyAdd)
		c.NotNil(add, "an empty group has an add button")
		c.Equal(empty.Parent(), add.Parent(), "beside its placeholder")
		c.True(add.Parent() != p.FindRefKey("r.3"+keyPill).Parent(), "not in its head")
		c.Equal("Add to this group", tooltipText(add.Tooltip))
	})
	choose(more("r.3"), "Skill")
	screen.Do(func() {
		c.Equal(1, groupLen("r.3"))
		c.Nil(p.FindRefKey("r.3"+keyAdd), "the add button goes once the group holds something")
		menuAction(p.moreEntries(p.node("r.3.0"), "r.3.0"), "Delete")()
	})
	screen.Sync()
	screen.Do(func() {
		c.Equal("r.3"+keyAdd, p.Window().Focus().RefKey, "deleting the last child focuses the placeholder's add button")
		menuAction(p.moreEntries(p.node("r.2"), "r.2"), "Delete")()
	})
	screen.Sync()
	screen.Do(func() {
		c.Equal("r.2"+keyMore, p.Window().Focus().RefKey, "deleting another focuses what comes after it")
		menuAction(p.moreEntries(p.node("r.2"), "r.2"), "Delete")()
	})
	screen.Sync()
	screen.Do(func() {
		c.Equal(treeRootPath+keyMore, p.Window().Focus().RefKey, "deleting the last focuses the root's add button")
	})
}

// TestPrereqPanelPlaceholderTextStaysPutOnFocus checks that the empty placeholder draws its text in the same place with
// and without the focus.
func TestPrereqPanelPlaceholderTextStaysPutOnFocus(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var root *gurps.PrereqList
	p, _ := showPrereqPanel(t, screen, &root, false)
	var empty, add *unison.Panel
	screen.Do(func() {
		empty = p.FindRefKey(treeRootPath + ":empty")
		add = p.FindRefKey(treeRootPath + keyAdd)
	})
	capture := func(focus *unison.Panel) *image.NRGBA {
		screen.Do(focus.RequestFocus)
		screen.Sync()
		return screen.CaptureWindow(p.Window())
	}
	focused, unfocused := capture(empty), capture(add)
	var r geom.Rect
	var scale float32
	screen.Do(func() {
		// Clear of the outline and the focus ring around it.
		r = empty.RectToRoot(empty.ContentRect(false)).Inset(geom.NewUniformInsets(9))
		scale = float32(focused.Bounds().Dx()) / p.Window().ContentRect().Width
	})
	differ := 0
	for y := int(r.Y * scale); y < int(r.Bottom()*scale); y++ {
		for x := int(r.X * scale); x < int(r.Right()*scale); x++ {
			if focused.NRGBAAt(x, y) != unfocused.NRGBAAt(x, y) {
				differ++
			}
		}
	}
	c.Equal(0, differ, "the text is drawn in the same place with the focus as without")
}

// TestPrereqPanelEquippedEquipmentTags checks that the tags field of an equipped equipment prerequisite says how to
// match any of several tags, since it matches each of the comma-separated tags in turn.
func TestPrereqPanelEquippedEquipmentTags(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	equipped := gurps.NewEquippedEquipmentPrereq()
	equipped.TagsCriteria = criteria.Text{Compare: criteria.IsText, Qualifier: "Sword, Axe"}
	root := gurps.NewPrereqList()
	root.Prereqs = gurps.Prereqs{equipped}
	p, _ := showPrereqPanel(t, screen, &root, false)
	path := childPath(treeRootPath, 0)
	screen.Do(func() { p.toggle(path) })
	screen.Do(func() {
		field := p.FindRefKey(path + ":tag")
		c.NotNil(field, "the prerequisite has a tags field")
		if field != nil {
			c.True(strings.Contains(tooltipText(field.Tooltip), "Separate multiple tags with commas"),
				"whose tooltip says to separate tags with commas")
		}
	})
}

// TestPrereqPanelGroupMoreButtonNames checks that a screen reader hears a group's more button named for the group, and
// for what it holds when it holds anything, so that two empty groups, or a group of one and its row, aren't alike.
func TestPrereqPanelGroupMoreButtonNames(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	emptyAny := gurps.NewPrereqList()
	emptyAny.All = false
	emptyAll := gurps.NewPrereqList()
	luck := gurps.NewTraitPrereq()
	luck.NameCriteria.Qualifier = "Luck"
	ofOne := gurps.NewPrereqList()
	ofOne.Prereqs = gurps.Prereqs{luck.Clone(nil)}
	root := gurps.NewPrereqList()
	root.Prereqs = gurps.Prereqs{emptyAny, emptyAll, ofOne, luck}
	root = root.CloneAsPrereqList(nil)
	c.Equal(4, len(root.Prereqs))
	p, _ := showPrereqPanel(t, screen, &root, false)
	name := func(path string) string {
		var button *unison.Panel
		screen.Do(func() { button = p.FindRefKey(path + keyMore) })
		c.NotNil(button, "%s has a more button", path)
		c.NotNil(screen.AccessibilityTree(p.Window()))
		if node := screen.AccessibilityNodeFor(button); node != nil {
			return node.Name
		}
		return ""
	}
	c.Equal("More actions for Any of", name("r.0"))
	c.Equal("More actions for All of", name("r.1"))
	c.Equal("More actions for All of: Has trait Luck", name("r.2"))
	c.Equal("More actions for Has trait Luck", name("r.3"))
}
