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
	"github.com/richardwilkes/unison/enums/mod"
)

// prereqUndoHost stands in for the editor that holds a prereqPanel: it provides the undo manager, syncs when marked
// modified, and records the Escape that would discard the editor's changes.
type prereqUndoHost struct {
	unison.Panel
	mgr     *unison.UndoManager
	escapes int
}

func (h *prereqUndoHost) UndoManager() *unison.UndoManager {
	return h.mgr
}

// MarkModified implements ModifiableRoot, syncing as the editor does.
func (h *prereqUndoHost) MarkModified(_ unison.Paneler) {
	DeepSync(h)
}

// showPrereqPanel shows a prereqPanel for the root in a window, within a host, so that its rebuilds run and its edits
// can be undone.
func showPrereqPanel(t *testing.T, screen *unison.HeadlessScreen, root **gurps.PrereqList, ownerIsSpell bool) (*prereqPanel, *prereqUndoHost) {
	var p *prereqPanel
	host := &prereqUndoHost{mgr: unison.NewUndoManager(100, func(error) {})}
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

// waitForEvaluation waits out scriptEvaluationDelay, then for the work it put off.
func waitForEvaluation(screen *unison.HeadlessScreen) {
	time.Sleep(2 * scriptEvaluationDelay)
	screen.Sync()
}

// prereqMenuAction returns the action of the entry with the label, or nil.
func prereqMenuAction(entries []menuEntry, label string) func() {
	for _, one := range entries {
		if one.Label == label {
			return one.Act
		}
	}
	return nil
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
	screen.Do(func() { prereqMenuAction(p.addEntries(p.tree(), prereqRootPath), "Trait")() })
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

	screen.Do(func() { prereqMenuAction(p.moreEntries(root.Prereqs[0], "r.0"), "Delete")() })
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
			prereqMenuAction(p.moreEntries(p.node("r.0"), "r.0"), "Move Down")()
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
			prereqMenuAction(p.moreEntries(p.node(one.path), one.path), one.label)()
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

// TestPrereqPanelMoves checks that Move up and Move down step into an adjacent group and out of the ends of one, and
// that Ungroup is offered only where it keeps the meaning.
func TestPrereqPanelMoves(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, _ := showPrereqPanel(t, screen, &root, false)
	move := func(path, label string) {
		screen.Do(func() {
			act := prereqMenuAction(p.moreEntries(p.node(path), path), label)
			c.NotNil(act, "%s offers %s", path, label)
			if act != nil {
				act()
			}
		})
	}
	screen.Do(func() {
		c.Nil(prereqMenuAction(p.moreEntries(p.node("r.0"), "r.0"), "Move Up"), "nothing above the top")
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
		c.Nil(prereqMenuAction(p.moreEntries(list, "r.1"), "Ungroup"), "an any-of group in an all-of list stays")
		list.All = true
		c.NotNil(prereqMenuAction(p.moreEntries(list, "r.1"), "Ungroup"), "unless the modes match")
		list.WhenTL.Compare = criteria.AtLeastNumber
		c.Nil(prereqMenuAction(p.moreEntries(list, "r.1"), "Ungroup"), "and it has no tech level")
	})
}

// TestPrereqPanelEscapeClosesTheOpenRow checks that Escape closes an open row, returning the focus to its sentence,
// without reaching the editor, where it would discard the changes, and that it reaches the editor once no row is open.
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
	c.Equal(1, host.escapes)
}

// TestPrereqPanelStatus checks the status of each row and group against the sheet, as an icon, a tooltip and in the
// accessible names, and that the summary and sentences follow the tree when it changes.
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
	root.Prereqs = gurps.Prereqs{trait, met, broken, later, gurps.NewPrereqList()}
	root = root.CloneAsPrereqList(nil)
	p, _ := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() {
		for path, want := range map[string]*unison.SVG{
			"r.0": svg.Not, "r.1": unison.CheckmarkSVG, "r.2": unison.TriangleExclamationSVG, "r.3": unison.DashSVG,
			"r.3.0": unison.DashSVG, "r.4": unison.DashSVG,
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
		c.Contains(p.summary.plainText(), "Magery")
		c.True(strings.HasSuffix(p.summary.plainText(), ")."), "the summary ends with a period")
		for path, want := range map[string]string{
			"r.0": "Not met: Has trait Magery", "r.4": "Empty group, always met",
		} {
			_, tip, _ := p.status(p.node(path))
			c.Equal(want, tip, path)
		}
	})
	screen.Do(func() {
		p.edit("", "", func() {
			if one, ok := root.Prereqs[0].(*gurps.TraitPrereq); ok {
				one.NameCriteria.Qualifier = "Luck"
			}
		})
	})
	waitForEvaluation(screen)
	screen.Do(func() {
		c.Contains(p.summary.plainText(), "Luck", "the summary follows the change")
		sentence, ok := p.FindRefKey("r.0" + keySentence).Self.(*sentenceButton)
		c.True(ok)
		c.Equal("Has trait Luck, not met", sentence.Accessibility.Name, "as does the sentence, in place")
		sentence, ok = p.FindRefKey("r.3.0" + keySentence).Self.(*sentenceButton)
		c.True(ok)
		c.Equal(`Has trait whose name is "", doesn't apply at this tech level`, sentence.Accessibility.Name)
	})
	group := func() (name string) {
		screen.Do(func() { name = p.FindRefKey("r.3" + keyPill).Parent().Parent().Accessibility.Name })
		return name
	}
	c.Equal("All of, only when TL at least 10, doesn't apply at this tech level", group(), "a group's name has its status")
	screen.Do(func() {
		p.edit("", "", func() {
			if list, ok := root.Prereqs[3].(*gurps.PrereqList); ok {
				list.WhenTL.Qualifier = fxp.Twelve
			}
		})
	})
	waitForEvaluation(screen)
	c.Equal("All of, only when TL at least 12, doesn't apply at this tech level", group(),
		"a group's name follows its tech level")
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
			key := onto + keyMore
			if onto == prereqRootPath {
				key = onto + keyAdd
			}
			target := p.FindRefKey(key).Parent()
			r := p.RectFromRoot(target.RectToRoot(target.ContentRect(true)))
			where := geom.NewPoint(r.X+r.Width/3, r.Y+r.Height*fraction)
			data := &prereqDrag{panel: p, path: from}
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
	c.True(drag("r.0.0", prereqRootPath, 0.5), "the root's head is into the root")
	c.Equal([]prereq.Type{prereq.List, prereq.Script, prereq.Unknown, prereq.Trait, prereq.Skill}, prereqShape(root),
		"at its end")
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
			prereqMenuAction(p.addEntries(p.tree(), prereqRootPath), label)()
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

// TestPrereqPanelEmptyRoot checks that an empty root shows only its placeholder, without a summary, pill or status.
func TestPrereqPanelEmptyRoot(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var missing *gurps.PrereqList
	p, _ := showPrereqPanel(t, screen, &missing, false)
	screen.Do(func() {
		c.Nil(p.summary.Parent())
		c.Nil(p.FindRefKey(prereqRootPath + keyPill))
		c.Equal(0, len(p.views))
		c.NotNil(p.FindRefKey(prereqRootPath + ":empty"))
	})
}

// TestPrereqPanelUndoReopensRow checks that undoing typing in a row that has since closed opens it again, and that a
// row the Add menu adds opens with the focus in its name field.
func TestPrereqPanelUndoReopensRow(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	root := newTestPrereqTree()
	p, host := showPrereqPanel(t, screen, &root, false)
	screen.Do(func() { prereqMenuAction(p.addEntries(p.tree(), prereqRootPath), "Trait")() })
	screen.Do(func() { c.Equal("r.3:name", p.Window().Focus().RefKey, "a new row focuses its name field") })
	screen.Type("Luck")
	screen.Do(func() { p.toggle("r.3") })
	screen.Do(host.mgr.Undo)
	screen.Do(func() {
		c.Equal("r.3", p.open, "the row the undone change was made in opens")
		c.Equal("r.3:name", p.Window().Focus().RefKey)
	})
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
			c.NotNil(prereqMenuAction(p.addEntries(p.tree(), prereqRootPath), label), label)
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
		path := childPath(prereqRootPath, i)
		screen.Do(func() { p.toggle(path) })
		names := make(map[string]bool)
		var controls []*unison.Panel
		screen.Do(func() {
			p.FindRefKey(path + keyFirst).HasInSelfOrDescendants(func(one *unison.Panel) bool {
				if one.Focusable() {
					controls = append(controls, one)
				}
				return false
			})
		})
		c.NotNil(screen.AccessibilityTree(p.Window()))
		c.True(len(controls) > 2, path)
		for _, one := range controls {
			if node := screen.AccessibilityNodeFor(one); node != nil && node.Name != "" {
				c.False(names[node.Name], "%s: %s is shared", path, node.Name)
				names[node.Name] = true
			}
		}
	}
}

// seedEveryPrereqControl turns on every optional criterion of the prerequisites in the root, and adds a group with a
// tech level and nothing in it, so that an audit sees every control the panel can show.
func seedEveryPrereqControl(root *gurps.PrereqList) {
	on := criteria.Text{Compare: criteria.IsText}
	for _, one := range root.Prereqs {
		switch pr := one.(type) {
		case *gurps.TraitPrereq:
			pr.LevelCriteria.Qualifier = fxp.One
			pr.NotesCriteria = on
		case *gurps.SkillPrereq:
			pr.SpecializationCriteria = on
			pr.OptionalSpecializationCriteria = on
		case *gurps.SpellPrereq:
			if !pr.SamePowerSource {
				pr.PowerSourceCriteria = on
			}
		case *gurps.AttributePrereq:
			pr.CombinedWith = gurps.DexterityID
		case *gurps.EquippedEquipmentPrereq:
			pr.TagsCriteria = on
		default:
		}
	}
	group := gurps.NewPrereqList()
	group.All = false
	group.WhenTL.Compare = criteria.AtMostNumber
	group.Parent = root
	root.Prereqs = append(root.Prereqs, group)
}

// checkPrereqs opens a dockable with fn and checks the controls in it, then again with each row of its prerequisites
// panel open in turn, since a closed row shows only its sentence.
func (a *axNameAudit) checkPrereqs(view string, fn func()) {
	a.t.Helper()
	d := a.open(fn)
	if d == nil {
		return
	}
	a.check(view, d)
	var p *prereqPanel
	var paths []string
	a.screen.Do(func() {
		if found := panelsOfType[*prereqPanel](d.AsPanel()); len(found) == 1 {
			p = found[0]
			for i, one := range p.tree().Prereqs {
				if one.PrereqType() != prereq.List && one.PrereqType() != prereq.Unknown {
					paths = append(paths, childPath(prereqRootPath, i))
				}
			}
		}
	})
	if p == nil {
		a.t.Errorf("%s: no prerequisites panel", view)
		return
	}
	for _, path := range paths {
		a.screen.Do(func() { p.toggle(path) })
		a.check(view+", "+path+" open", d)
	}
}
