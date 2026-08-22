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
	"net/url"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
)

// fakeDragInfo is a minimal drag.Info carrying only a set of data types, letting tests drive the drag callbacks
// without a native drag session.
type fakeDragInfo struct {
	types []string
}

func (f *fakeDragInfo) SourceDragOpMask() drag.Op        { return drag.Copy | drag.Move }
func (f *fakeDragInfo) DataTypes() []string              { return f.types }
func (f *fakeDragInfo) HasString() bool                  { return false }
func (f *fakeDragInfo) HasFilePaths() bool               { return false }
func (f *fakeDragInfo) HasURLs() bool                    { return false }
func (f *fakeDragInfo) HasDataType(dataType string) bool { return slices.Contains(f.types, dataType) }
func (f *fakeDragInfo) Text() string                     { return "" }
func (f *fakeDragInfo) FilePaths() []string              { return nil }
func (f *fakeDragInfo) URLs() []*url.URL                 { return nil }
func (f *fakeDragInfo) Data(_ string) []byte             { return nil }

// fakeAltDropProvider is the minimal TableProvider needed to exercise InstallTableDropSupport's alternate drop path
// (dropping a modifier onto a row) in a headless test. Only the methods that path touches do anything.
type fakeAltDropProvider struct {
	unison.SimpleTableModel[*Node[*gurps.Trait]]
	altDrops []int
}

func (p *fakeAltDropProvider) DataOwner() gurps.DataOwner                    { return nil }
func (p *fakeAltDropProvider) SetTable(_ *unison.Table[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) RootData() []*gurps.Trait                      { return nil }
func (p *fakeAltDropProvider) SetRootData(_ []*gurps.Trait)                  {}
func (p *fakeAltDropProvider) DragKey() *uti.DataType                        { return traitDragKey }
func (p *fakeAltDropProvider) DragSVG() *unison.SVG                          { return nil }
func (p *fakeAltDropProvider) ItemNames() (singular, plural string)          { return "Trait", "Traits" }
func (p *fakeAltDropProvider) ColumnIDs() []int                              { return nil }
func (p *fakeAltDropProvider) HierarchyColumnID() int                        { return -1 }
func (p *fakeAltDropProvider) ExcessWidthColumnID() int                      { return -1 }
func (p *fakeAltDropProvider) ContextMenuItems() []ContextMenuItem           { return nil }
func (p *fakeAltDropProvider) Serialize() ([]byte, error)                    { return nil, nil }
func (p *fakeAltDropProvider) Deserialize(_ []byte) error                    { return nil }
func (p *fakeAltDropProvider) RefKey() string                                { return "" }
func (p *fakeAltDropProvider) AllTags() []string                             { return nil }
func (p *fakeAltDropProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Trait]] {
	return nil
}
func (p *fakeAltDropProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) DropShouldMoveData(_, _ *unison.Table[*Node[*gurps.Trait]]) bool {
	return false
}
func (p *fakeAltDropProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.Trait]])        {}
func (p *fakeAltDropProvider) OpenEditor(_ Rebuildable, _ *unison.Table[*Node[*gurps.Trait]]) {}
func (p *fakeAltDropProvider) CreateItem(_ Rebuildable, _ *unison.Table[*Node[*gurps.Trait]], _ ItemVariant) {
}

func (p *fakeAltDropProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: traitModifierDragKey,
		Drop:    func(rowIndex int, _ any) { p.altDrops = append(p.altDrops, rowIndex) },
	}
}

// newAltDropTestTable returns a table with drop support installed and a single root row, whose height is the table's
// MinimumRowHeight since no columns are configured, so points with a small Y are over the row and large Y values miss.
func newAltDropTestTable(provider *fakeAltDropProvider) *unison.Table[*Node[*gurps.Trait]] {
	table := unison.NewTable(provider)
	InstallTableDropSupport(table, provider)
	table.SetRootRows([]*Node[*gurps.Trait]{NewNode(table, nil, gurps.NewTrait(nil, nil, false), false)})
	return table
}

// TestAltDropDragFeedbackFlushed verifies that dragging a modifier over a table row immediately flushes the drawing,
// since a native drag has no continuous redraw loop and the row highlight never appears without an explicit flush.
// This covers every table built through InstallTableDropSupport: character sheets, loot sheets, templates and editors.
func TestAltDropDragFeedbackFlushed(t *testing.T) {
	c := check.New(t)
	var flushes []*unison.Panel
	original := flushDragFeedback
	flushDragFeedback = func(panel *unison.Panel) { flushes = append(flushes, panel) }
	defer func() { flushDragFeedback = original }()

	provider := &fakeAltDropProvider{}
	table := newAltDropTestTable(provider)
	overRow := geom.Point{X: 1, Y: table.MinimumRowHeight / 2}
	offRows := geom.Point{X: 1, Y: 10000}
	di := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}

	// Entering over a row must offer a copy and flush the highlight to the screen.
	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "enter over row")
	c.Equal(1, len(flushes), "enter over row must flush")
	c.Equal(table.AsPanel(), flushes[0], "the table itself must be flushed")

	// Each update over a row must flush as well, since the highlight may have moved to a different row.
	c.Equal(drag.Copy, table.DragUpdatedCallback(di, overRow, mod.None), "update over row")
	c.Equal(2, len(flushes), "update over row must flush")

	// Moving off of the rows offers no drop and must flush again to erase the previous highlight.
	c.Equal(drag.None, table.DragUpdatedCallback(di, offRows, mod.None), "update off rows")
	c.Equal(3, len(flushes), "update off rows must flush to erase the highlight")

	// Exiting with no highlight showing has nothing to erase.
	table.DragExitedCallback()
	c.Equal(3, len(flushes), "exit without a highlight showing need not flush")

	// Exiting while a highlight is showing must flush to erase it.
	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "re-enter over row")
	c.Equal(4, len(flushes), "re-enter over row must flush")
	table.DragExitedCallback()
	c.Equal(5, len(flushes), "exit with a highlight showing must flush to erase it")

	// A drag of some other type must not trigger the alternate drop feedback.
	flushes = nil
	c.Equal(drag.None, table.DragEnteredCallback(&fakeDragInfo{types: []string{noteDragKey.UTI}}, overRow, mod.None),
		"unrelated drag type")
	c.Equal(0, len(flushes), "unrelated drag type must not flush")
}

// TestAltDropDeliversRowIndex verifies that releasing an alternate drop over a row hands the row index to the
// provider's drop handler and erases the highlight, while a release that isn't over a row does nothing.
func TestAltDropDeliversRowIndex(t *testing.T) {
	c := check.New(t)
	var flushes int
	original := flushDragFeedback
	flushDragFeedback = func(_ *unison.Panel) { flushes++ }
	defer func() { flushDragFeedback = original }()

	provider := &fakeAltDropProvider{}
	table := newAltDropTestTable(provider)
	overRow := geom.Point{X: 1, Y: table.MinimumRowHeight / 2}
	di := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}

	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "enter over row")
	flushes = 0
	c.True(table.DropCallback(di, overRow, mod.None), "drop over row must be handled")
	c.Equal([]int{0}, provider.altDrops, "drop must deliver the row index")
	c.Equal(1, flushes, "drop must flush to erase the highlight")

	// With no row targeted, the drop must be declined and the handler left uncalled.
	c.False(table.DropCallback(di, overRow, mod.None), "drop without a targeted row must be declined")
	c.Equal([]int{0}, provider.altDrops, "declined drop must not invoke the handler")
}

// TestAltDropNotifiesTheTable verifies that an alternate drop reports the change through the table's
// DropOccurredCallback, just as unison does for a normal drop. The providers' drop handlers only rebuild when the data
// owner has an owning entity, so without this notification a template or a traits/equipment list dockable keeps showing
// itself as unmodified after the drop, with no modified marker on its tab and no prompt to save on the way out.
func TestAltDropNotifiesTheTable(t *testing.T) {
	c := check.New(t)
	original := flushDragFeedback
	flushDragFeedback = func(_ *unison.Panel) {}
	defer func() { flushDragFeedback = original }()

	provider := &fakeAltDropProvider{}
	table := newAltDropTestTable(provider)
	c.NotNil(table.DropOccurredCallback, "drop support must install a drop notification")
	notified := 0
	table.DropOccurredCallback = func() { notified++ }
	overRow := geom.Point{X: 1, Y: table.MinimumRowHeight / 2}
	di := &fakeDragInfo{types: []string{traitModifierDragKey.UTI}}

	c.Equal(drag.Copy, table.DragEnteredCallback(di, overRow, mod.None), "enter over row")
	c.True(table.DropCallback(di, overRow, mod.None), "drop over row must be handled")
	c.Equal(1, notified, "a completed drop must notify the table")

	// With no row targeted, nothing was dropped, so there is nothing to report.
	c.False(table.DropCallback(di, overRow, mod.None), "drop without a targeted row must be declined")
	c.Equal(1, notified, "a declined drop must not notify the table")
}

// TestDropWithinASheetSurvivesTheSourceTableBeingReplaced verifies that a drag from one list on a sheet to another is
// still recognized as coming from that sheet after the rebuild the drop triggers. Moving the only switchable item out
// of the carried equipment list takes the switch column away from it, and a list can only change its columns by being
// built anew, so the table the drag started from is left orphaned. Looked up again it resolves to the list that took
// its place; left stale it has no sheet above it at all, and the drop is mistaken for one arriving from a library,
// prompting the user to pick modifiers for an item that was already on the sheet.
func TestDropWithinASheetSurvivesTheSourceTableBeingReplaced(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = "Powered Armor"
	eqp.Features = gurps.Features{switchableSTBonus(nil)}
	modifier := gurps.NewEquipmentModifier(entity, nil, false)
	modifier.Name = "Cheap"
	eqp.Modifiers = []*gurps.EquipmentModifier{modifier}
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)

	from := sheet.CarriedEquipment.Table
	to := sheet.OtherEquipment.Table
	mgr := unison.UndoManagerFor(from)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.Equal(gurps.EquipmentSwitchColumn, from.Columns[1].ID, "the carried list must start out with the switch column")
	c.NotEqual(gurps.EquipmentSwitchColumn, to.Columns[0].ID,
		"the other equipment list must start out without the switch column")

	// Drive the move the way unison's drop handling does: collect the undo data, move the row in the model, bring both
	// tables up to date and select the moved row in the destination, then hand off to the drop completion.
	undo := willDropCallback(from, to, true)
	c.NotNil(undo, "the drop must be undoable")
	entity.CarriedEquipment = nil
	entity.OtherEquipment = []*gurps.Equipment{eqp}
	from.SyncToModel()
	to.SyncToModel()
	to.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
	didDropCallback(undo, from, to, true)

	c.NotEqual(from, sheet.CarriedEquipment.Table, "losing the switch column must replace the carried equipment table")
	c.NotEqual(to, sheet.OtherEquipment.Table, "gaining the switch column must replace the other equipment table")
	c.Equal(0, len(*prompts), "an item already on the sheet must not be prompted for its modifiers again")
	c.True(mgr.CanUndo(), "the move must have registered an undo edit")

	mgr.Undo()
	c.Equal(1, len(entity.CarriedEquipment), "undo must put the item back into the carried list")
	c.Equal(0, len(entity.OtherEquipment), "undo must take the item back out of the other equipment list")
	c.Equal(gurps.EquipmentSwitchColumn, sheet.CarriedEquipment.Table.Columns[1].ID,
		"undo must bring the switch column back to the carried list")
	c.Equal(0, len(*prompts), "undo must not prompt for modifiers either")
}

// TestReorderWithinATableIsNotTreatedAsAnAddition verifies that a drag that only reorders rows within a single list is
// still seen as such after the rebuild the drop triggers replaces that list's table. The two tables handed to the drop
// completion are the same one, and both have to be looked up again for them to stay that way; if only the destination
// is, the reorder reads as an arrival from somewhere else and the points of the moved rows are merged into the rows
// they match, deleting them.
func TestReorderWithinATableIsNotTreatedAsAnAddition(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	first := gurps.NewSkill(entity, nil, false)
	first.Name = "Brawling"
	first.Points = fxp.One
	second := gurps.NewSkill(entity, nil, false)
	second.Name = "Brawling"
	second.Points = fxp.One
	second.Features = gurps.Features{switchableSTBonus(nil)}
	entity.Skills = []*gurps.Skill{first, second}
	sheet.Rebuild(true)

	table := sheet.Skills.Table
	c.Equal(gurps.SkillSwitchColumn, table.Columns[0].ID, "the skills list must start out with the switch column")

	// Reorder the two rows and drop the switchable one, which takes the switch column away with it and so replaces the
	// table, then complete the drop as a move within that one table.
	undo := willDropCallback(table, table, true)
	entity.Skills = []*gurps.Skill{second, first}
	second.Features = nil
	table.SyncToModel()
	table.SetSelectionMap(map[tid.TID]bool{second.ID(): true})
	didDropCallback(undo, table, table, true)

	c.NotEqual(table, sheet.Skills.Table, "losing the switch column must replace the skills table")
	c.Equal(2, len(entity.Skills), "a reorder must not merge one row into the other")
	c.Equal(fxp.One, first.Points, "a reorder must leave the points alone")
	c.Equal(fxp.One, second.Points, "a reorder must leave the points alone")
}

// newLibraryStyleTraitsTable returns a traits table with nothing above it in the panel hierarchy, standing in for a
// library list: rows arriving from one count as coming from outside a sheet, so their modifiers and nameables are
// prompted for and their points are merged into any identical rows already present.
func newLibraryStyleTraitsTable(traits ...*gurps.Trait) *unison.Table[*Node[*gurps.Trait]] {
	data := gurps.NewTemplate()
	data.SetTraitList(traits)
	provider := NewTraitsProvider(data, false)
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.ClientData()[TableProviderClientKey] = provider
	table.SetRootRows(provider.RootRows())
	return table
}

// newSwitchableTraitModifier returns a disabled trait modifier carrying a switchable +1 ST bonus, so that enabling it
// is what gives its owner switchable features.
func newSwitchableTraitModifier(name string) *gurps.TraitModifier {
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = name
	modifier.Disabled = true
	modifier.Features = gurps.Features{switchableSTBonus(nil)}
	return modifier
}

// stubTraitModifierPrompt substitutes a non-interactive trait modifier prompt that hands the modifiers it was asked to
// show to the given responder and reports back whatever the responder returns, letting a test drive the rebuild that
// answering the prompt triggers. The count of prompts actually shown is returned, and the real prompt is restored when
// the test finishes.
func stubTraitModifierPrompt(t *testing.T, respond func(modifiers []*gurps.TraitModifier) bool) *int {
	t.Helper()
	original := promptForTraitModifiers
	t.Cleanup(func() { promptForTraitModifiers = original })
	shown := 0
	promptForTraitModifiers = func(_ string, modifiers []*gurps.TraitModifier) bool {
		if len(modifiers) == 0 {
			return false // The real prompt has nothing to show in this case, so it can't change anything either.
		}
		shown++
		return respond(modifiers)
	}
	return &shown
}

// enableAllModifiers is a stubTraitModifierPrompt responder standing in for the user turning on every modifier the
// prompt offers, and reporting the change so that the owner is rebuilt.
func enableAllModifiers(modifiers []*gurps.TraitModifier) bool {
	for _, one := range modifiers {
		one.Disabled = false
	}
	return true
}

// TestDropPromptingForModifiersSurvivesTheDestinationTableBeingReplaced verifies that a drop arriving from a library
// list completes against the list that is on screen even though answering the modifier prompt replaces it partway
// through. Only the modifiers that are enabled count toward a row having switchable features, so turning one on brings
// the switch column into view, and a list can only gain a column by being built anew -- leaving the table the drop was
// handed orphaned, with no sheet above it from which to reach the undo manager or ask for a rebuild.
func TestDropPromptingForModifiersSurvivesTheDestinationTableBeingReplaced(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	to := sheet.Traits.Table
	mgr := unison.UndoManagerFor(to)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.Equal(-1, switchColumnIndex(to.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	from := newLibraryStyleTraitsTable(trait)
	shown := stubTraitModifierPrompt(t, enableAllModifiers)

	// Drive the drop the way unison's drop handling does: collect the undo data, add the row to the model, bring the
	// destination up to date and select the new row, then hand off to the drop completion.
	undo := willDropCallback(from, to, false)
	c.NotNil(undo, "the drop must be undoable")
	entity.Traits = append(entity.Traits, trait)
	to.SyncToModel()
	to.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	didDropCallback(undo, from, to, false)

	c.Equal(1, *shown, "a drop from a library must prompt for the dropped row's modifiers")
	live := sheet.Traits.Table
	c.NotEqual(to, live, "gaining the switch column must replace the traits table")
	c.False(columnsOutOfSync(sheet.Traits.provider.ColumnIDs(), live.Columns),
		"the live traits list must show the columns its content calls for")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"enabling the switchable modifier must bring the switch column into view")
	c.True(live.CopySelectionMap()[trait.ID()], "the dropped row must be selected in the list the user is looking at")
	c.True(mgr.CanUndo(), "the drop must have registered an undo edit")

	mgr.Undo()
	c.Equal(originalTraits, len(entity.Traits), "undo must take the dropped trait back off the sheet")
	c.Equal(-1, switchColumnIndex(sheet.Traits.Table.Columns, gurps.TraitSwitchColumn),
		"undo must take the switch column away again")
}

// TestCopyRowsToRebuildsTheListThatReplacedTheOneItWasGiven verifies that the work CopyRowsTo does once its
// post-processing has finished -- scrolling to the new rows, recording the undo edit and rebuilding the owner -- is
// aimed at the table that is on screen. The post-processing prompts for modifiers, and answering that prompt rebuilds
// the sheet; since only enabled modifiers count toward a row having switchable features, the switch column can come or
// go, and a list can only change its columns by being built anew. Whatever the post-processing does after that point
// -- the remaining rows' prompts, the nameable substitutions, the point merge -- only reaches the screen if the
// closing rebuild does, and an orphaned table has no sheet above it to rebuild.
func TestCopyRowsToRebuildsTheListThatReplacedTheOneItWasGiven(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	target := sheet.Traits.Table
	mgr := unison.UndoManagerFor(target)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source := newLibraryStyleTraitsTable(trait)

	var copied *gurps.Trait
	CopyRowsTo(target, source.RootRows(), func(rows []*Node[*gurps.Trait]) {
		copied = rows[0].Data()
		// Stand in for the modifier prompt: turning the modifier on gives the trait a switchable feature, so the
		// traits list needs the switch column and can only get it by being built anew. The table CopyRowsTo is holding
		// is an orphan from here on.
		copied.Modifiers[0].Disabled = false
		sheet.Rebuild(true)
		// Stand in for everything the post-processing does after that prompt. It changes the model again -- here by
		// turning the modifier back off, which takes the switch column away again -- and counts on the rebuild at the
		// end of the copy to put the list back in step with its content.
		copied.Modifiers[0].Disabled = true
	}, true)

	live := sheet.Traits.Table
	c.NotEqual(target, live, "gaining the switch column must replace the traits table")
	c.False(columnsOutOfSync(sheet.Traits.provider.ColumnIDs(), live.Columns),
		"the copy must rebuild the list that replaced the one it was handed")
	c.Equal(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"the switch column must go away again once nothing is switchable")
	c.True(live.CopySelectionMap()[copied.ID()], "the copied row must be selected in the list the user is looking at")
	c.True(mgr.CanUndo(), "the copy must have registered an undo edit")

	mgr.Undo()
	c.Equal(originalTraits, len(entity.Traits), "undo must take the copied trait back off the sheet")
}

// TestCopyToSheetPostProcessingSurvivesTheTargetTableBeingReplaced verifies that the post-processing performed for a
// copy onto a sheet keeps working with the list that is on screen. It starts by prompting for the modifiers of the
// incoming rows, and answering that prompt rebuilds the sheet, which replaces the table when the toggled modifier
// carries a switchable feature; the nameable substitutions and the point merge that follow have to be applied to the
// list that took its place rather than to the orphan.
func TestCopyToSheetPostProcessingSurvivesTheTargetTableBeingReplaced(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	target := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(target.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")
	originalTraits := len(entity.Traits) // A new entity may come with traits of its own, such as the natural attacks.

	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Retractable")}
	source := newLibraryStyleTraitsTable(trait)
	shown := stubTraitModifierPrompt(t, enableAllModifiers)

	CopyRowsTo(target, source.RootRows(), func(_ []*Node[*gurps.Trait]) {
		processCopiedRows(source, target)
	}, true)

	c.Equal(1, *shown, "a copy from a library must prompt for the copied row's modifiers")
	c.Equal(originalTraits+1, len(entity.Traits), "the trait must have been copied onto the sheet")
	live := sheet.Traits.Table
	c.NotEqual(target, live, "gaining the switch column must replace the traits table")
	c.False(columnsOutOfSync(sheet.Traits.provider.ColumnIDs(), live.Columns),
		"the live traits list must show the columns its content calls for")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"enabling the switchable modifier must bring the switch column into view")
	c.True(live.CopySelectionMap()[entity.Traits[len(entity.Traits)-1].ID()],
		"the copied row must be selected in the list the user is looking at")
}
