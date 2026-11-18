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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
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
