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
	"fmt"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/namegen"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/role"
)

// TestEveryControlHasAnAccessibleName opens every kind of view the application has -- the sheet, each editor seeded
// with every kind of feature and prerequisite, the settings views, the calculators, the libraries and the file editors
// -- and asks for the description an assistive technology would be handed of each, failing for every control in them
// that would be announced with no name. A field, popup or color well is named by the label laid out before it, by the
// label it points at with Accessibility.LabeledBy, or by an Accessibility.Name of its own; an icon-only button by its
// tooltip. A control with none of those is announced as "edit text" or "pop up button" and nothing more, which leaves a
// screen reader user guessing at what they are about to change.
//
// The tree is walked rather than the panels, so that what a table describes without a panel apiece -- its column
// headers and the cells of its rows -- is checked along with everything else.
func TestEveryControlHasAnAccessibleName(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	audit := &axNameAudit{t: t, screen: screen, wnd: wnd}
	audit.check("workspace", wnd.Content())

	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var entity *gurps.Entity
	screen.Do(func() {
		entity = sheet.Entity()
		// Give the sheet's lists a row apiece so that their cells are described.
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Audit Trait"
		entity.Traits = append(entity.Traits, tr)
		sk := gurps.NewSkill(entity, nil, false)
		sk.Name = "Audit Skill"
		entity.Skills = append(entity.Skills, sk)
		eq := gurps.NewEquipment(entity, nil, false)
		eq.Name = "Audit Equipment"
		entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
		entity.Recalculate()
		sheet.Rebuild(true)
	})
	audit.check("character sheet", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })
	audit.check("character sheet layout editing", sheet)
	screen.Do(func() { sheet.toggleLayoutEditing() })

	// Every feature and prerequisite type has a row of its own in the editors, so each editor is seeded with one of
	// every kind it can hold.
	allFeatures := func(owner fmt.Stringer, forEquipmentModifier bool) gurps.Features {
		var list gurps.Features
		fp := newFeaturesPanel(entity, owner, &gurps.Features{}, forEquipmentModifier)
		for _, ft := range feature.Types {
			if ft == feature.Unknown {
				continue
			}
			if f := fp.createFeatureForType(ft); f != nil {
				list = append(list, f)
			}
		}
		return list
	}
	allPrereqs := func(types []prereq.Type, ownerIsSpell bool) *gurps.PrereqList {
		root := gurps.NewPrereqList()
		pp := newPrereqPanel(entity, &root, types, ownerIsSpell)
		for _, pt := range types {
			if pr := pp.createPrereqForType(pt, root); pr != nil {
				root.Prereqs = append(root.Prereqs, pr)
			}
		}
		return root
	}
	studies := func() []*gurps.Study {
		return []*gurps.Study{{Type: study.Self, Hours: fxp.Ten, Note: "audit"}}
	}

	audit.checkOpened("trait editor", func() {
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Audit Trait"
		tr.Features = allFeatures(tr, false)
		tr.Prereq = allPrereqs(prereq.TypesForNonEquipment, false)
		tr.Study = studies()
		tr.Weapons = []*gurps.Weapon{gurps.NewWeapon(tr, true), gurps.NewWeapon(tr, false)}
		EditTrait(sheet, tr)
	})
	audit.checkOpened("trait modifier editor", func() {
		m := gurps.NewTraitModifier(entity, nil, false)
		m.Features = allFeatures(m, false)
		EditTraitModifier(sheet, m)
	})
	audit.checkOpened("skill editor", func() {
		s := gurps.NewSkill(entity, nil, false)
		s.Prereq = allPrereqs(prereq.TypesForNonEquipment, false)
		s.Features = allFeatures(s, false)
		s.Study = studies()
		s.Defaults = []*gurps.SkillDefault{{DefaultType: gurps.DexterityID}, {DefaultType: gurps.SkillID}}
		EditSkill(sheet, s)
	})
	audit.checkOpened("technique editor", func() {
		EditSkill(sheet, gurps.NewTechnique(entity, nil, "Audit Skill"))
	})
	audit.checkOpened("spell editor", func() {
		s := gurps.NewSpell(entity, nil, false)
		s.Prereq = allPrereqs(prereq.TypesForNonEquipment, true)
		s.Study = studies()
		EditSpell(sheet, s)
	})
	audit.checkOpened("equipment editor", func() {
		e := gurps.NewEquipment(entity, nil, false)
		e.Features = allFeatures(e, false)
		e.Prereq = allPrereqs(prereq.TypesForEquipment, false)
		EditEquipment(sheet, e, true)
	})
	audit.checkOpened("equipment modifier editor", func() {
		m := gurps.NewEquipmentModifier(entity, nil, false)
		m.Features = allFeatures(m, true)
		EditEquipmentModifier(sheet, m)
	})
	audit.checkOpened("note editor", func() { EditNote(sheet, gurps.NewNote(entity, nil, false)) })
	audit.checkOpened("melee weapon editor", func() {
		EditWeapon(sheet, gurps.NewWeapon(gurps.NewTrait(entity, nil, false), true))
	})
	audit.checkOpened("ranged weapon editor", func() {
		EditWeapon(sheet, gurps.NewWeapon(gurps.NewTrait(entity, nil, false), false))
	})
	audit.checkOpened("points editor", func() {
		entity.PointsRecord = append(entity.PointsRecord, &gurps.PointsRecord{
			When:   jio.Now(),
			Points: fxp.Five,
			Reason: "audit",
		})
		displayPointsEditor(sheet, entity)
	})

	audit.checkOpened("sheet settings", func() { ShowSheetSettings(sheet) })
	audit.checkOpened("attribute settings", func() { ShowAttributeSettings(sheet) })
	audit.checkOpened("body settings", func() { ShowBodySettings(sheet) })
	audit.checkOpened("general settings", ShowGeneralSettings)
	audit.checkOpened("color settings", ShowColorSettings)
	audit.checkOpened("font settings", ShowFontSettings)
	audit.checkOpened("menu key settings", ShowMenuKeySettings)
	audit.checkOpened("page reference mappings", ShowPageRefMappings)
	audit.checkOpened("library settings", func() { ShowLibrarySettings(gurps.GlobalSettings().Libraries.User()) })

	if calc, isCalc := audit.open(func() { DisplayCalculator(sheet) }).(*Calculator); isCalc {
		for i := range calc.tabs {
			var title string
			screen.Do(func() {
				calc.tabBar.selectTab(i)
				title = calc.tabs[i].title()
			})
			audit.check("calculator: "+title, calc)
		}
	} else {
		t.Error("the calculator did not open")
	}

	audit.checkOpened("character template", func() { newCharacterTemplateAction.Execute(nil) })
	audit.checkOpened("loot sheet", func() { newLootSheetAction.Execute(nil) })
	audit.checkOpened("traits library", func() { newTraitsLibraryAction.Execute(nil) })
	audit.checkOpened("equipment library", func() { newEquipmentLibraryAction.Execute(nil) })
	audit.checkOpened("markdown file", func() { newMarkdownFileAction.Execute(nil) })

	if d, isOne := audit.open(func() { newAncestryAction.Execute(nil) }).(*ancestryEditorDockable); isOne {
		screen.Do(func() {
			d.model.CommonOptions.HairOptions = append(d.model.CommonOptions.HairOptions,
				&gurps.WeightedStringOption{Weight: 1, Value: "Brown"})
			d.model.GenderOptions = append(d.model.GenderOptions, &gurps.WeightedAncestryOptions{
				Weight: 1,
				Value:  &gurps.AncestryOptions{Name: "Audit"},
			})
			d.sync()
		})
		audit.check("ancestry editor", d)
	} else {
		t.Error("the ancestry editor did not open")
	}
	if d, isOne := audit.open(func() { newNameGeneratorAction.Execute(nil) }).(*nameGeneratorEditorDockable); isOne {
		screen.Do(func() {
			d.model.Type = namegen.Compound
			d.model.Compound = []*gurps.NameGenerator{
				{Type: namegen.Simple, Entries: []*gurps.WeightedStringOption{{Weight: 1, Value: "Audit"}}},
			}
			d.sync()
		})
		audit.check("name generator editor", d)
	} else {
		t.Error("the name generator editor did not open")
	}

	if len(audit.failures) != 0 {
		t.Errorf("%d controls would be announced with no name:\n%s", len(audit.failures),
			strings.Join(audit.failures, "\n"))
	}
}

// axNameAudit collects the controls in a headless workspace that an assistive technology would be handed with no name.
type axNameAudit struct {
	t        *testing.T
	screen   *unison.HeadlessScreen
	wnd      *unison.Window
	failures []string
}

// open runs fn on the UI thread and returns the dockable it opened, or nil -- and a test error -- if it opened some
// other number of them.
func (a *axNameAudit) open(fn func()) unison.Dockable {
	a.t.Helper()
	var opened []unison.Dockable
	a.screen.Do(func() {
		before := make(map[unison.Dockable]bool)
		for _, d := range AllDockables() {
			before[d] = true
		}
		fn()
		for _, d := range AllDockables() {
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

// checkOpened opens a dockable with fn and checks the controls in it.
func (a *axNameAudit) checkOpened(view string, fn func()) {
	a.t.Helper()
	if d := a.open(fn); d != nil {
		a.check(view, d)
	}
}

// check describes the window afresh and records every control under root that has no name.
func (a *axNameAudit) check(view string, root unison.Paneler) {
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

// axNodeNeedsAName reports whether a node is something a person is expected to read, change or act on, and so must be
// announced by name. A table's rows and cells are where its keyboard focus is reported, so they are focusable without
// being controls: a cell with nothing in it is rightly announced as blank, and one with a control in it is checked
// through that control.
func axNodeNeedsAName(node *accessibility.Node) bool {
	switch node.Role {
	case role.TextField, role.TextArea, role.SpinButton, role.ComboBox, role.PopupButton, role.Slider,
		role.ProgressBar, role.ColorWell, role.List, role.Table, role.Tree, role.Button, role.CheckBox,
		role.RadioButton, role.ToggleButton, role.Link, role.ColumnHeader:
		return true
	case role.Row, role.Cell:
		return false
	default:
		return node.Focusable
	}
}

// describe says where an unnamed control is, in terms that point at the code that built it.
func (a *axNameAudit) describe(view string, tree *accessibility.Tree, node *accessibility.Node, panels map[accessibility.NodeID]*unison.Panel) string {
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
	name = strings.TrimPrefix(name, "*github.com/richardwilkes/gcs/v5/ux.")
	name = strings.TrimPrefix(name, "*github.com/richardwilkes/unison.")
	name = strings.TrimPrefix(name, "*")
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
