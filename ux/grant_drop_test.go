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
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
)

// grantAt returns the panel's feature at the index, which must be a grant.
func grantAt(t *testing.T, p *featuresPanel, index int) *gurps.Grant {
	t.Helper()
	g, ok := (*p.features)[index].(*gurps.Grant)
	if !ok {
		t.Fatalf("feature %d is a %T, not a grant", index, (*p.features)[index])
	}
	return g
}

// newGrantTestTrait returns a trait outside any sheet with the given name and levels, leveled when levels isn't zero.
func newGrantTestTrait(name string, levels fxp.Int) *gurps.Trait {
	t := gurps.NewTrait(nil, nil, false)
	t.Name = name
	if levels != 0 {
		t.CanLevel = true
		t.PointsPerLevel = fxp.One
		t.Levels = levels
	}
	return t
}

// grantItemNames returns the names of the items, which must be traits or skills.
func grantItemNames(t *testing.T, items []*grantLibraryItem) []string {
	t.Helper()
	names := make([]string, 0, len(items))
	for _, item := range items {
		switch one := item.node.(type) {
		case *gurps.Trait:
			names = append(names, one.Name)
		case *gurps.Skill:
			names = append(names, one.Name)
		default:
			t.Fatalf("expected a trait or skill, found %T", item.node)
		}
	}
	return names
}

// TestGrantItemsFromDraggedRows checks that dragged trait, skill and spell rows become one grant item each, for a grant
// of their kind, techniques included, that a container stands for the items within it, and that an item that grants
// something itself is left out.
func TestGrantItemsFromDraggedRows(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	vision := gurps.NewTrait(nil, nil, true)
	vision.Name = "Vision"
	for _, name := range []string{"Night Vision", "Dark Vision"} {
		child := newGrantTestTrait(name, 0)
		child.SetParent(vision)
		vision.Children = append(vision.Children, child)
	}
	granting := newGrantTestTrait("Granting", 0)
	granting.Features = gurps.Features{gurps.NewGrant(feature.GrantTrait)}
	rows := []*Node[*gurps.Trait]{
		NewNode(table, nil, vision, false),
		NewNode(table, nil, newGrantTestTrait("Fit", 0), false),
		NewNode(table, nil, granting, false),
	}
	kind, items := grantItemsFrom(&unison.TableDragData[*Node[*gurps.Trait]]{Table: table, Rows: rows})
	c.Equal(feature.GrantTrait, kind)
	c.Equal([]string{"Night Vision", "Dark Vision", "Fit"}, grantItemNames(t, items))

	_, skillTable := NewNodeTable(NewSkillsProvider(gurps.NewEntity(), false), nil)
	kick := gurps.NewTechnique(nil, nil, "Karate")
	kick.Name = "Kick"
	skillRows := []*Node[*gurps.Skill]{
		NewNode(skillTable, nil, newTestSkill("Karate", fxp.One, nil), false),
		NewNode(skillTable, nil, kick, false),
	}
	kind, items = grantItemsFrom(&unison.TableDragData[*Node[*gurps.Skill]]{Table: skillTable, Rows: skillRows})
	c.Equal(feature.GrantSkill, kind)
	c.Equal([]string{"Karate", "Kick"}, grantItemNames(t, items))

	_, spellTable := NewNodeTable(NewSpellsProvider(gurps.NewEntity(), false), nil)
	spellRows := []*Node[*gurps.Spell]{NewNode(spellTable, nil, newTestSpell("Light", fxp.One, nil), false)}
	kind, items = grantItemsFrom(&unison.TableDragData[*Node[*gurps.Spell]]{Table: spellTable, Rows: spellRows})
	c.Equal(feature.GrantSpell, kind)
	c.Equal(1, len(items))

	kind, _ = grantItemsFrom(nil)
	c.Equal(feature.Unknown, kind, "anything else isn't a drag of items to grant")
}

// TestFeaturesPanelDropsGrants checks that items dropped on the row of an empty grant of their kind fill it, with any
// more added after it, and that items dropped elsewhere become new grants, a leveled trait at its level, opened at
// their level field, each drop a single undoable edit.
func TestFeaturesPanelDropsGrants(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	empty := gurps.NewGrant(feature.GrantTrait)
	features := gurps.Features{empty}
	p, host := showFeaturesPanel(t, screen, entity, owner, &features, false)
	items := func(names ...string) []*grantLibraryItem {
		list := make([]*grantLibraryItem, 0, len(names))
		for _, name := range names {
			list = append(list, &grantLibraryItem{node: newGrantTestTrait(name, fxp.Three)})
		}
		return list
	}
	screen.Do(func() {
		var row *unison.Panel
		for _, child := range p.Children() {
			if path, ok := child.ClientData()[featureDropKey].(string); ok && path == "0" {
				row = child
			}
		}
		c.NotNil(row)
		where := p.PointFromRoot(row.PointToRoot(geom.NewPoint(1, 1)))
		target, at := p.grantDropAt(where, feature.GrantTrait)
		c.True(target == row, "a drop on the row of an empty grant goes into it")
		c.Equal(dropInto, at)
		target, at = p.grantDropAt(where, feature.GrantSkill)
		c.True(target == row, "unless the grant is of another kind, when it goes beside it")
		c.Equal(dropBefore, at)
		p.dropGrants(row, dropInto, feature.GrantTrait, items("Night Vision", "Dark Vision"))
	})
	screen.Do(func() {
		c.Equal(2, len(features))
		c.True(features[0] == empty, "the empty grant is filled, not replaced")
		c.Equal("Night Vision", empty.ItemName())
		c.Equal(fxp.Three, empty.Amount, "the filled grant takes the trait's level too")
		second, ok := features[1].(*gurps.Grant)
		c.True(ok)
		c.Equal("Dark Vision", second.ItemName(), "a second item is added after it")
		c.Equal("0", p.open)
		p.dropGrants(nil, 0, feature.GrantTrait, items("Fit"))
	})
	screen.Do(func() {
		c.Equal(3, len(features), "dropped anywhere else, it goes at the end")
		last, ok := features[2].(*gurps.Grant)
		c.True(ok)
		c.Equal("Fit", last.ItemName())
		c.Equal(fxp.Three, last.Amount, "a leveled trait is granted at its level")
		c.Equal("2", p.open)
		c.Equal("2:amount", p.Window().Focus().RefKey, "the new grant's level field takes the focus")
		host.mgr.Undo()
	})
	screen.Do(func() {
		c.Equal(2, len(features), "undo takes the drop back")
		host.mgr.Undo()
	})
	screen.Do(func() {
		c.Equal(1, len(features))
		g, ok := features[0].(*gurps.Grant)
		c.True(ok)
		c.Equal("", string(g.Key()), "and empties the grant again")
	})
}

// TestGrantsOnSpellsStartSwitchable checks that a grant added to a spell, whether made new, turned into one from another
// kind of feature or dropped, starts out switchable, that one added to a trait doesn't, and that a grant switched to
// another kind keeps its switchability as it was.
func TestGrantsOnSpellsStartSwitchable(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	traitPanel, _ := showFeaturesPanel(t, screen, entity, gurps.NewTrait(entity, nil, false), &gurps.Features{}, false)
	bonus := gurps.NewReactionBonus()
	unswitched := gurps.NewGrant(feature.GrantSkill)
	features := gurps.Features{bonus, unswitched}
	c.False(bonus.IsSwitchable() || unswitched.IsSwitchable(), "precondition: neither feature is switchable")
	p, _ := showFeaturesPanel(t, screen, entity, gurps.NewSpell(entity, nil, false), &features, false)
	defer func(last feature.Type) { lastFeatureTypeUsed = last }(lastFeatureTypeUsed)
	c.True(p.createFeatureForType(feature.GrantTrait).IsSwitchable(), "a new grant on a spell is switchable")
	c.False(traitPanel.createFeatureForType(feature.GrantTrait).IsSwitchable(), "one on a trait isn't")
	choose := func(path string, t feature.Type) {
		screen.Do(func() { p.toggle(path) })
		screen.Do(func() { featuresPopup[featureTypeEntry](c, p, path+":type").Select(featureTypeEntry{featureType: t}) })
	}
	choose("0", feature.GrantTrait)
	c.Equal(feature.GrantTrait, features[0].FeatureType(), "precondition: the bonus became a grant")
	c.True(features[0].IsSwitchable(), "a bonus turned into a grant on a spell is switchable")
	choose("1", feature.GrantSpell)
	c.Equal(feature.GrantSpell, features[1].FeatureType(), "precondition: the grant changed kind")
	c.False(features[1].IsSwitchable(), "a grant changing kind keeps its switchability")
	screen.Do(func() {
		p.dropGrants(nil, 0, feature.GrantTrait, []*grantLibraryItem{{node: newGrantTestTrait("Fit", 0)}})
	})
	c.Equal(3, len(features), "precondition: the drop added a grant")
	c.True(features[2].IsSwitchable(), "a grant dropped on a spell is switchable")
}

// TestGrantDropZoneReplacesTheChosenItem checks that the open row of a grant keeps its drop zone once an item is
// chosen, that a drop on the zone replaces the item rather than adding a grant, and that a drop elsewhere on the row
// still goes beside it, a new grant going before or after the row as the drop does.
func TestGrantDropZoneReplacesTheChosenItem(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
	features := gurps.Features{g}
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	screen.Do(func() { p.toggle("0") })
	var zone *unison.Panel
	screen.Do(func() { zone = p.FindRefKey("0:drop") })
	if zone == nil {
		t.Fatal("the open row of a chosen grant has no drop zone")
	}
	screen.Do(func() {
		where := p.PointFromRoot(zone.PointToRoot(zone.ContentRect(true).Center()))
		target, at := p.grantDropAt(where, feature.GrantTrait)
		c.True(target == zone, "a drop on the zone goes into the grant")
		c.Equal(dropInto, at)
		row, _, _ := p.rowAt(where)
		target, at = p.grantDropAt(p.PointFromRoot(row.PointToRoot(geom.NewPoint(1, 1))), feature.GrantTrait)
		c.True(target == row, "a drop elsewhere on the row goes beside it")
		c.Equal(dropBefore, at)
		p.dropGrants(zone, dropInto, feature.GrantTrait, []*grantLibraryItem{{node: newGrantTestTrait("Dark Vision", 0)}})
	})
	screen.Do(func() {
		c.Equal(1, len(features), "the drop replaces the item rather than adding a grant")
		c.True(features[0] == g, "in the same grant")
		c.Equal("Dark Vision", g.ItemName())
	})
	rowFor := func(path string) *unison.Panel {
		for _, child := range p.Children() {
			if rowPath, ok := child.ClientData()[featureDropKey].(string); ok && rowPath == path {
				return child
			}
		}
		return nil
	}
	names := func() []string {
		list := make([]string, 0, len(features))
		for i := range features {
			list = append(list, grantAt(t, p, i).ItemName())
		}
		return list
	}
	screen.Do(func() {
		p.dropGrants(rowFor("0"), dropAfter, feature.GrantTrait,
			[]*grantLibraryItem{{node: newGrantTestTrait("Fit", 0)}})
	})
	screen.Do(func() {
		c.Equal([]string{"Dark Vision", "Fit"}, names(), "a drop after a row goes after it")
		p.dropGrants(rowFor("1"), dropBefore, feature.GrantTrait,
			[]*grantLibraryItem{{node: newGrantTestTrait("Acute Hearing", 0)}})
	})
	screen.Do(func() {
		c.Equal([]string{"Dark Vision", "Acute Hearing", "Fit"}, names(), "and one before a row before it")
	})
}

// TestFeaturesPanelTakesDraggedRowsAsGrants checks that trait rows dragged from a table onto a features panel are taken
// as a copy and dropped as grants, expanding the panel when it is collapsed, while a drag of nothing but the panel's
// own item is turned down, and a read-only panel takes no such drag at all.
func TestFeaturesPanelTakesDraggedRowsAsGrants(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	owner := gurps.NewTrait(entity, nil, false)
	var features gurps.Features
	p, _ := showFeaturesPanel(t, screen, entity, owner, &features, false)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	uxtest.SwapForTest(t, &draggedTableData, nil)
	di := &fakeDragInfo{types: []string{traitDragKey.UTI}}
	var accepted, dropped bool
	var op drag.Op
	dragOnto := func(traits ...*gurps.Trait) {
		rows := make([]*Node[*gurps.Trait], 0, len(traits))
		for _, one := range traits {
			rows = append(rows, NewNode(table, nil, one, false))
		}
		draggedTableData = &unison.TableDragData[*Node[*gurps.Trait]]{Table: table, Rows: rows}
		screen.Do(func() {
			where := geom.NewPoint(10, 10)
			accepted = p.CanAcceptDropCallback(di)
			op = p.DragEnteredCallback(di, where, 0)
			dropped = p.DropCallback(di, where, 0)
		})
	}

	dragOnto(owner)
	c.Equal(drag.None, op, "the panel's own item can't be granted to it")
	c.False(dropped)
	c.Equal(0, len(features))

	screen.Do(p.collapse.toggle)
	dragOnto(owner, newGrantTestTrait("Fit", 0))
	c.True(accepted, "rows a grant can give are accepted")
	c.Equal(drag.Copy, op, "as a copy")
	c.True(dropped)
	c.Equal(1, len(features), "a grant is added for each row but the panel's own item")
	if len(features) == 1 {
		c.Equal("Fit", grantAt(t, p, 0).ItemName())
	}
	c.False(p.collapse.collapsed, "the drop expands the panel")

	screen.Do(p.makeReadOnly)
	dragOnto(newGrantTestTrait("Night Vision", 0))
	c.False(accepted, "a read-only panel takes no grants")
}

// newTestSheetWithGrant returns a sheet holding the skill Stealth and, inside a closed container, the trait Training,
// which grants the skill Brawling.
func newTestSheetWithGrant(t *testing.T) (sheet *Sheet, training *gurps.Trait, granted, stealth *gurps.Skill) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	group := gurps.NewTrait(entity, nil, true)
	group.SetOpen(false)
	training = gurps.NewTrait(entity, group, false)
	training.Name = "Training"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	group.Children = []*gurps.Trait{training}
	entity.Traits = []*gurps.Trait{group}
	stealth = newTestSkill("Stealth", fxp.One, nil)
	stealth.SetDataOwner(entity)
	entity.Skills = []*gurps.Skill{stealth}
	sheet.Rebuild(true)
	for _, one := range entity.Skills {
		if gurps.IsGranted(one) {
			granted = one
		}
	}
	return sheet, training, granted, stealth
}

// TestGrantedRowsOnSheet checks that the commands that change rows are off for a selection of granted rows alone, and
// that the mark of a granted row selects the item granting it, opening its container.
func TestGrantedRowsOnSheet(t *testing.T) {
	c := check.New(t)
	sheet, training, granted, stealth := newTestSheetWithGrant(t)
	c.NotNil(granted)
	table := sheet.Skills.Table
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true})
	c.False(hasUngrantedSelection(table), "a granted row changes only with its grant")
	table.SetSelectionMap(map[tid.TID]bool{granted.ID(): true, stealth.ID(): true})
	c.True(hasUngrantedSelection(table))

	selectGrantingItem(table, gurps.GrantingItem(granted))
	c.True(training.Parent().IsOpen(), "the container holding the grantor is opened")
	selected := sheet.Traits.Table.SelectedRows(false)
	c.Equal(1, len(selected))
	c.True(selected[0].Data() == training, "the grantor is selected")
}

// TestGrantedItemEditorIsReadOnly checks that a granted item's editor only shows it.
func TestGrantedItemEditorIsReadOnly(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	screen.Do(func() {
		sheet, _, granted, stealth := newTestSheetWithGrant(t)
		DisplayNewDockable(sheet)
		c.True(EditSkill(sheet, granted).readOnly)
		c.False(EditSkill(sheet, stealth).readOnly)
	})
}

// TestGrantItemEditorOffersNoGrants checks that the editor of a grant's item, and an editor opened from it for one of
// the item's modifiers, offer no grant feature types and take no dropped grants, whether rows from a table or a grant
// from another features panel, since a granted item grants nothing.
func TestGrantItemEditorOffersNoGrants(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	nightVision := newGrantTestTrait("Night Vision", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Mod"
	nightVision.Modifiers = []*gurps.TraitModifier{modifier}
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(nightVision, gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{training}
	sheet.Rebuild(true)
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var outer *editor[*gurps.Trait, *gurps.TraitEditData]
	var outerFeatures *featuresPanel
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer = EditTrait(sheet, training)
		panels := uxtest.PanelsOfType[*featuresPanel](outer.content)
		c.Equal(1, len(panels))
		outerFeatures = panels[0]
		outerFeatures.editGrantedItem(grantAt(t, outerFeatures, 0))
	})
	nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
	var modEditor *editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]
	screen.Do(func() {
		_, isNested := nested.owner.(*grantItemOwner)
		c.True(isNested, "precondition: the other trait editor is the grant item's")
		EditTraitModifier(nested, nested.editorData.Modifiers[0])
	})
	modEditor = uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	screen.Do(func() {
		for _, e := range []*unison.Panel{nested.content, modEditor.content} {
			panels := uxtest.PanelsOfType[*featuresPanel](e)
			c.Equal(1, len(panels))
			c.False(slices.ContainsFunc(panels[0].featureTypesList(), gurps.IsGrant), "no grant types are offered")
			c.False(panels[0].takesGrants(), "and no grants are taken")
			c.False(panels[0].readOnly, "precondition: the panel takes other copies")
			target, _ := panels[0].dropAt(geom.NewPoint(4, 4), &rowDrag{panel: outerFeatures.AsPanel(), path: "0"})
			c.Nil(target, "nor a grant from another features panel")
		}
	})
}

// TestGrantItemEditorLeavesLevelAndEnabledToTheGrant checks that the editor of a grant's trait offers neither a Level
// field nor an Enabled check box, since the grant's row sets the level of the trait it puts on the sheet and that trait
// is enabled whenever the grant is in effect, so nothing would read either.
func TestGrantItemEditorLeavesLevelAndEnabledToTheGrant(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	training := gurps.NewTrait(entity, nil, false)
	training.Name = "Training"
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(newGrantTestTrait("Acute Vision", fxp.Two), gurps.LibraryFile{})
	training.Features = gurps.Features{g}
	entity.Traits = []*gurps.Trait{training}
	sheet.Rebuild(true)
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var outer *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		DisplayNewDockable(sheet)
		outer = EditTrait(sheet, training)
		panels := uxtest.PanelsOfType[*featuresPanel](outer.content)
		c.Equal(1, len(panels))
		panels[0].editGrantedItem(grantAt(t, panels[0], 0))
	})
	nested := uxtest.OtherEditor(t, screen, outer, isTraitEditor)
	hasLevelField := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*DecimalField](content), func(f *DecimalField) bool {
			return f.undoTitle == "Level"
		})
	}
	hasEnabledCheckBox := func(content *unison.Panel) bool {
		return slices.ContainsFunc(uxtest.PanelsOfType[*CheckBox](content), func(cb *CheckBox) bool {
			return cb.Text.String() == "Enabled"
		})
	}
	screen.Do(func() {
		_, isNested := nested.owner.(*grantItemOwner)
		c.True(isNested, "precondition: the other trait editor is the grant item's")
		c.True(hasLevelField(outer.content), "precondition: a trait's own editor has a Level field")
		c.True(hasEnabledCheckBox(outer.content), "and an Enabled check box")
		c.False(hasLevelField(nested.content), "the grant's row sets the level")
		c.False(hasEnabledCheckBox(nested.content), "and the granted trait is enabled while the grant is in effect")
	})
}

// TestGrantItemEditorShowsTheCostAtTheGrantsLevel checks that the editor of a grant's trait, which leaves the level to
// the grant, shows the point cost at the grant's level rather than at the level the grant's copy of the trait holds,
// which the grant's row never changes, and a cost of 0 for a level below 1, at which the grant puts no trait on the
// sheet.
func TestGrantItemEditorShowsTheCostAtTheGrantsLevel(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kind   grantadjust.Kind
		amount fxp.Int
		want   string
	}{
		{name: "level set", kind: grantadjust.SetLevel, amount: fxp.Three, want: "8"},
		{name: "level set to 1", kind: grantadjust.SetLevel, amount: fxp.One, want: "6"},
		{name: "levels taken away", kind: grantadjust.AddLevels, amount: -fxp.Two, want: "0"},
		{name: "level set below 1", kind: grantadjust.SetLevel, amount: fxp.Half, want: "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			entity := gurps.NewEntity()
			carrier := gurps.NewTrait(entity, nil, false)
			carrier.Name = "Training"
			trait := newGrantTestTrait("Acute Vision", fxp.One)
			// A base cost, so that the cost at level 0 isn't 0.
			trait.BasePoints = fxp.Five
			g := gurps.NewGrant(feature.GrantTrait)
			g.SetItem(trait, gurps.LibraryFile{})
			carrier.Features = gurps.Features{g}
			g.Adjust = tc.kind
			g.Amount = tc.amount
			root := &stubRebuildable{}
			root.Self = root
			item := g.Trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
			owner := &grantItemOwner{
				Rebuildable: root,
				panel:       newFeaturesPanel(entity, carrier, &carrier.Features, false),
				key:         g.Key(),
				hash:        grantItemHash(g),
				item:        item,
			}
			_, content := buildEditorContent(owner, item, initTraitEditor)
			fields := uxtest.PanelsOfType[*NonEditableField](content)
			if len(fields) == 0 {
				t.Fatal("the editor must show the point cost")
			}
			c.Equal(tc.want, fields[0].String())
		})
	}
}

// TestGrantDestinationsAreToldApart checks that the editors offered as destinations for Add as Grant are named so that
// two modifier editors with the same title, opened from the editors of different traits, can be told apart.
func TestGrantDestinationsAreToldApart(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traits := make([]*gurps.Trait, 0, 2)
	for _, name := range []string{"Training", "Schooling"} {
		trait := gurps.NewTrait(entity, nil, false)
		trait.Name = name
		modifier := gurps.NewTraitModifier(entity, nil, false)
		modifier.Name = "Mod"
		trait.Modifiers = []*gurps.TraitModifier{modifier}
		gurps.AttachModifiers(trait, trait.Modifiers)
		traits = append(traits, trait)
	}
	entity.Traits = traits
	sheet.Rebuild(true)
	var names []string
	screen.Do(func() {
		DisplayNewDockable(sheet)
		for _, trait := range traits {
			e := EditTrait(sheet, trait)
			EditTraitModifier(e, e.editorData.Modifiers[0])
		}
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{newGrantTestTrait("Night Vision", 0)})
		for _, d := range grantDestinations(items) {
			names = append(names, d.String())
		}
	})
	c.Equal(4, len(names))
	c.Equal(len(names), len(slices.Compact(slices.Clone(names))), "every destination has a name of its own: %v", names)
	c.True(slices.ContainsFunc(names, func(s string) bool {
		return strings.HasPrefix(s, "Mod (") && strings.Contains(s, "Schooling")
	}), "a modifier's editor names the trait it was opened from: %v", names)
}

// TestAddAsGrantReachesThresholdFeatures checks that Add as Grant offers the features of a pool threshold in the
// attribute settings, which take grants dropped on them, naming its pool, and adds the grant there.
func TestAddAsGrantReachesThresholdFeatures(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var d *attributeSettingsDockable
	var names []string
	screen.Do(func() {
		defs := testAttrDefs("st", "hp")
		def := defs.Set["hp"]
		def.Type = attribute.Pool
		def.Thresholds = []*gurps.PoolThreshold{{State: "Reeling"}}
		d = newAttributeSettingsDockable(nil, defs)
		d.TabTitle = "Attributes"
		d.TabIcon = svg.Attributes
		d.show()
		expandAllAttributes(d)
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{newGrantTestTrait("Night Vision", 0)})
		for _, one := range grantDestinations(items) {
			names = append(names, one.String())
		}
		addAsGrant(feature.GrantTrait, items)
	})
	c.Equal([]string{"Reeling (HP, Attributes)"}, names, "the threshold is offered, named with its pool")
	var granted []string
	screen.Do(func() {
		for _, f := range d.model.Set["hp"].Thresholds[0].Features {
			if g, ok := f.(*gurps.Grant); ok {
				granted = append(granted, g.ItemName())
			}
		}
	})
	c.Equal([]string{"Night Vision"}, granted, "the grant is added to the threshold")
}

// TestGrantNotAddedToItsOwnEditor checks that an item is never added as a grant to the editor of that same item, since
// the item would then grant a copy of itself.
func TestGrantNotAddedToItsOwnEditor(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	fit := gurps.NewTrait(entity, nil, false)
	fit.Name = "Fit"
	entity.Traits = []*gurps.Trait{fit}
	sheet.Rebuild(true)
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	screen.Do(func() {
		DisplayNewDockable(sheet)
		e = EditTrait(sheet, fit)
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{fit, newGrantTestTrait("Night Vision", 0)})
		c.Equal(2, len(items))
		addAsGrant(feature.GrantTrait, items)
	})
	var names []string
	screen.Do(func() {
		for _, f := range e.editorData.Features {
			if g, ok := f.(*gurps.Grant); ok {
				names = append(names, g.ItemName())
			}
		}
	})
	c.Equal([]string{"Night Vision"}, names, "the item being edited is left out")

	// So is the trait whose modifier is being edited, since its modifiers' grants are its own.
	screen.Do(func() {
		modifier := gurps.NewTraitModifier(entity, nil, false)
		modifier.Name = "Long"
		e.editorData.Modifiers = []*gurps.TraitModifier{modifier}
		EditTraitModifier(e, modifier)
	})
	modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	names = nil
	screen.Do(func() {
		p := uxtest.PanelsOfType[*featuresPanel](modifierEditor.content)[0]
		items := collectGrantItems(gurps.LibraryFile{}, []*gurps.Trait{fit, newGrantTestTrait("Night Vision", 0)})
		names = grantItemNames(t, p.grantable(items))
	})
	c.Equal([]string{"Night Vision"}, names, "the trait whose modifier is being edited is left out")
}

// TestDeletingContainerKeepsGrantedRows checks that deleting a container the user moved a granted row into leaves that
// row on the sheet, since a granted row goes only with its grant.
func TestDeletingContainerKeepsGrantedRows(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var sheet *Sheet
	var granted *gurps.Skill
	var group *gurps.Skill
	isSkillEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Skill, *gurps.SkillEditData])
		return ok
	}
	screen.Do(func() {
		sheet, _, granted, _ = newTestSheetWithGrant(t)
		entity := sheet.Entity()
		group = gurps.NewSkill(entity, nil, true)
		group.Name = "Group"
		entity.Skills = slices.DeleteFunc(entity.Skills, func(s *gurps.Skill) bool { return s == granted })
		granted.SetParent(group)
		group.Children = []*gurps.Skill{granted}
		entity.Skills = append(entity.Skills, group)
		sheet.Rebuild(true)
		DisplayNewDockable(sheet)
		EditSkill(sheet, granted)
		sheet.Skills.Table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		DeleteSelection(sheet.Skills.Table, true)
	})
	var kept bool
	var groups, editors int
	screen.Do(func() {
		for _, s := range sheet.Entity().Skills {
			if s == granted {
				kept = true
			}
			if s.Container() {
				groups++
			}
		}
		editors = len(AllMatchingDockables(isSkillEditor))
	})
	c.Equal(0, groups, "the container goes")
	c.True(kept, "the granted row within it stays, as the row it was")
	c.Equal(1, editors, "and so does its editor")
}

// newTestSheetWithEquipmentGrant returns a sheet holding a piece of carried equipment, a physical container when
// physical is set, that grants the skill Brawling, along with the carrier and the granted row.
func newTestSheetWithEquipmentGrant(t *testing.T, physical bool) (sheet *Sheet, carrier *gurps.Equipment, granted *gurps.Skill) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	carrier = gurps.NewEquipment(entity, nil, physical)
	carrier.Name = "Goggles"
	g := gurps.NewGrant(feature.GrantSkill)
	g.SetItem(newTestSkill("Brawling", fxp.One, nil), gurps.LibraryFile{})
	carrier.Features = gurps.Features{g}
	entity.CarriedEquipment = []*gurps.Equipment{carrier}
	sheet.Rebuild(true)
	for _, one := range entity.Skills {
		if gurps.IsGranted(one) {
			granted = one
		}
	}
	return sheet, carrier, granted
}

// TestUndoOfTakingGrantsAwayRestoresGrantedRowInOtherList checks that undoing an edit that takes a carrier's grants
// away -- deleting the carrier, applying an editor or a sync with its source that removes the grant, disabling the
// carrier, dragging, moving or organizing it into a disabled container, moving or dragging it to the other equipment
// list, selling it, converting it to a group or disabling the ancestry holding it for one copied onto the sheet --
// brings back the row the grant put in the skills list as the row it was, where the user had moved it, rather than
// having the restored grant make a fresh one at the end of the list.
func TestUndoOfTakingGrantsAwayRestoresGrantedRowInOtherList(t *testing.T) {
	for _, tc := range []struct {
		name string
		// equipment is set when the carrier is a piece of carried equipment rather than a trait, and container when
		// that is a physical container.
		equipment bool
		container bool
		edit      func(t *testing.T, sheet *Sheet, trait *gurps.Trait, eqp *gurps.Equipment)
	}{
		{name: "delete selection", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, trait)
			DeleteSelection(sheet.Traits.Table, true)
		}},
		{name: "editor removes the grant", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			e, _ := buildEditorContent(sheet, trait, initTraitEditor)
			e.editorData.Features = nil
			e.applyEdits()
		}},
		{name: "toggle enablement", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, trait)
			toggleDisabled(sheet, sheet.Traits.Table)
		}},
		{name: "sync with a source that no longer grants it", edit: func(t *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			_, user := uxtest.UseTestLibraries(t, check.New(t))
			RegisterKnownFileTypes()
			lib := trait.Clone(gurps.LibraryFile{}, nil, nil, gurps.Copy)
			lib.Features = nil
			libFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
			if err := gurps.SaveTraits([]*gurps.Trait{lib},
				filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))); err != nil {
				t.Fatal(err)
			}
			trait.Source = gurps.Source{LibraryFile: libFile, TID: lib.ID()}
			selectGrantingItem(sheet.Skills.Table, trait)
			SyncWithSourceForSelection(sheet.Traits.Table)
		}},
		{name: "drag into a disabled container", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			entity := sheet.Entity()
			disabled := gurps.NewTrait(entity, nil, true)
			disabled.Disabled = true
			entity.Traits = append(entity.Traits, disabled)
			sheet.Rebuild(true)
			// Driven the way unison drives a drop within one table: the undo data is collected, the rows are moved
			// and the drop completion is called.
			selectGrantingItem(sheet.Skills.Table, trait)
			table := sheet.Traits.Table
			draggedTableData = &unison.TableDragData[*Node[*gurps.Trait]]{Table: table, Rows: table.SelectedRows(true)}
			defer func() { draggedTableData = nil }()
			undo := willDropCallback(table, table, true)
			trait.Parent().Children = nil
			trait.SetParent(disabled)
			disabled.Children = []*gurps.Trait{trait}
			table.SyncToModel()
			table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
			didDropCallback(undo, table, table, true)
		}},
		{name: "move into a disabled container", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			group := trait.Parent()
			disabled := gurps.NewTrait(sheet.Entity(), group, true)
			disabled.Disabled = true
			group.Children = append(group.Children, disabled)
			sheet.Rebuild(true)
			selectGrantingItem(sheet.Skills.Table, trait)
			MoveSelection(sheet.Traits.Table, MoveIntoContainer)
		}},
		{name: "organize into a disabled container", edit: func(_ *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			// Organizing files a top-level trait into the top-level group named for its category, even a disabled one.
			entity := sheet.Entity()
			trait.Parent().Children = nil
			trait.SetParent(nil)
			trait.BasePoints = fxp.Five
			disabled := gurps.NewTrait(entity, nil, true)
			disabled.Name = "Advantages"
			disabled.Disabled = true
			entity.Traits = append(entity.Traits, disabled, trait)
			sheet.Rebuild(true)
			sheet.AsPanel().PerformCmd(nil, OrganizeTraitsItemID)
		}},
		{name: "move to other equipment", equipment: true, edit: func(_ *testing.T, sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			selectGrantingItem(sheet.Skills.Table, eqp)
			moveSelectedEquipment(sheet, sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table)
		}},
		{name: "sell all of it", equipment: true, edit: func(_ *testing.T, sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			entity := sheet.Entity()
			recordTrade(sheet, "Sell Equipment", func() {
				entity.CommitSale([]gurps.SaleItem{{Equipment: eqp, Quantity: fxp.One}}, nil, false, "Sold", jio.Now())
			})
		}},
		{name: "drag to other equipment", equipment: true, edit: func(_ *testing.T, sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			// Driven the way unison drives a drop from one table into another: the undo data is collected, the rows
			// are moved and the drop completion is called.
			from, to := sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table
			selectGrantingItem(sheet.Skills.Table, eqp)
			draggedTableData = &unison.TableDragData[*Node[*gurps.Equipment]]{Table: from, Rows: from.SelectedRows(true)}
			defer func() { draggedTableData = nil }()
			undo := willDropCallback(from, to, true)
			entity := sheet.Entity()
			entity.CarriedEquipment = nil
			entity.OtherEquipment = []*gurps.Equipment{eqp}
			from.SyncToModel()
			to.SyncToModel()
			to.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
			didDropCallback(undo, from, to, true)
		}},
		{name: "convert to a group", equipment: true, container: true, edit: func(t *testing.T, sheet *Sheet, _ *gurps.Trait, eqp *gurps.Equipment) {
			// A group can't hold features, so converting a container to one drops its grant.
			uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool { return true })
			selectGrantingItem(sheet.Skills.Table, eqp)
			table := sheet.CarriedEquipment.Table
			table.PerformCmd(table, ConvertToGroupContainerItemID)
		}},
		{name: "copy an ancestry, disabling the one holding it", edit: func(t *testing.T, sheet *Sheet, trait *gurps.Trait, _ *gurps.Equipment) {
			// The carrier's container becomes the sheet's ancestry, which answering the question disables.
			ancestry := trait.Parent()
			ancestry.ContainerType = container.Ancestry
			ancestry.Ancestry = "Human"
			sheet.Rebuild(true)
			uxtest.SwapForTest(t, &askToDisableExistingAncestry, func(_ promptOperation, _, _ []string) bool { return true })
			uxtest.SwapForTest(t, &askToRandomizeAgain, func(_ promptOperation) bool { return false })
			copySelectionTo(newTestTemplateWithTraits(newAncestryTrait("Human")).Traits.Table, []*Sheet{sheet})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			var sheet *Sheet
			var trait *gurps.Trait
			var eqp *gurps.Equipment
			var granted *gurps.Skill
			if tc.equipment {
				sheet, eqp, granted = newTestSheetWithEquipmentGrant(t, tc.container)
			} else {
				sheet, trait, granted, _ = newTestSheetWithGrant(t)
			}
			if granted == nil {
				t.Fatal("the grant must put its row on the sheet")
			}
			entity := sheet.Entity()
			group := gurps.NewSkill(entity, nil, true)
			group.Name = "Group"
			entity.Skills = slices.DeleteFunc(entity.Skills, func(s *gurps.Skill) bool { return s == granted })
			granted.SetParent(group)
			group.Children = []*gurps.Skill{granted}
			entity.Skills = append(entity.Skills, group)
			sheet.Rebuild(true)
			tc.edit(t, sheet, trait, eqp)
			grantedSkill := func() *gurps.Skill {
				var found *gurps.Skill
				gurps.Traverse(func(s *gurps.Skill) bool {
					if gurps.IsGranted(s) {
						found = s
					}
					return false
				}, false, false, entity.Skills...)
				return found
			}
			c.Nil(grantedSkill(), "the granted row goes with its grant")
			if !sheet.undoMgr.CanUndo() {
				t.Fatal("the edit must be undoable")
			}
			sheet.undoMgr.Undo()
			restored := grantedSkill()
			if restored == nil {
				t.Fatal("the granted row must be back")
			}
			c.Equal(granted.ID(), restored.ID(), "the granted row comes back as the row it was")
			c.True(restored.Parent() == group, "and where the user had put it")
		})
	}
}

// TestNestedEditorTitle checks that an editor opened from a document is named for the document, while one opened from
// another editor, for a grant's item or for one of the item's modifiers, is named for its item.
func TestNestedEditorTitle(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	outer := newGrantTestTrait("Training", 0)
	inner := newGrantTestTrait("Night Vision", 0)
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Mod"
	inner.Modifiers = []*gurps.TraitModifier{modifier}
	g := gurps.NewGrant(feature.GrantTrait)
	g.SetItem(inner, gurps.LibraryFile{})
	outer.Features = gurps.Features{g}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{outer}
	isTraitEditor := func(d unison.Dockable) bool {
		_, ok := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok
	}
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var outerTitle string
	screen.Do(func() {
		template := newTestTemplateDockable("Template", data)
		DisplayNewDockable(template)
		e = EditTrait(template, outer)
		outerTitle = e.Title()
		p := uxtest.PanelsOfType[*featuresPanel](e.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
	})
	nested := uxtest.OtherEditor(t, screen, e, isTraitEditor)
	var innerTitle, modTitle string
	screen.Do(func() {
		innerTitle = nested.Title()
		EditTraitModifier(nested, nested.editorData.Modifiers[0])
	})
	modEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		func(d unison.Dockable) bool {
			_, ok := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
			return ok
		})
	screen.Do(func() { modTitle = modEditor.Title() })
	c.Equal("Trait Editor for Template", outerTitle)
	c.Equal("Trait Editor for Night Vision", innerTitle)
	c.Equal("Trait Modifier Editor for Mod", modTitle)
}
