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
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/printing"
)

// SkipDeepSync is set on components that should not trigger a deep sync.
const SkipDeepSync = "!deepsync"

var (
	_ FileBackedDockable           = &Sheet{}
	_ ExportDockable               = &Sheet{}
	_ unison.UndoManagerProvider   = &Sheet{}
	_ ModifiableRoot               = &Sheet{}
	_ Rebuildable                  = &Sheet{}
	_ unison.TabCloser             = &Sheet{}
	_ KeyedDockable                = &Sheet{}
	_ gurps.DataOwnerProvider      = &Sheet{}
	_ gurps.SheetSettingsResponder = &Sheet{}

	printMgr    printing.PrintManager
	lastPrinter printing.PrinterID
	dropKeys    = []*uti.DataType{
		equipmentDragKey,
		skillDragKey,
		spellDragKey,
		traitDragKey,
		noteDragKey,
	}
)

type itemCreator interface {
	CreateItem(Rebuildable, ItemVariant)
}

// Sheet holds the view for a GURPS character sheet.
type Sheet struct {
	fileBackedPanel
	targetMgr            *TargetMgr
	undoMgr              *unison.UndoManager
	toolbar              *unison.Panel
	scroll               *unison.ScrollPanel
	entity               *gurps.Entity
	content              *unison.Panel
	contentLayout        *overlayStackLayout
	page                 *Page
	blocks               map[string]unison.Paneler
	layoutEditor         *sheetLayoutEditor
	layoutButton         *unison.Button
	layoutButtonGroup    *unison.Group
	Reactions            *PageList[*gurps.ConditionalModifier]
	ConditionalModifiers *PageList[*gurps.ConditionalModifier]
	MeleeWeapons         *PageList[*gurps.Weapon]
	RangedWeapons        *PageList[*gurps.Weapon]
	Traits               *PageList[*gurps.Trait]
	Skills               *PageList[*gurps.Skill]
	Spells               *PageList[*gurps.Spell]
	CarriedEquipment     *PageList[*gurps.Equipment]
	OtherEquipment       *PageList[*gurps.Equipment]
	Notes                *PageList[*gurps.Note]
	searchTracker        *SearchTracker
	scale                int
	awaitingUpdate       bool
}

// ActiveSheet returns the currently active sheet.
func ActiveSheet() *Sheet {
	d := ActiveDockable()
	if d == nil {
		return nil
	}
	if s, ok := d.(*Sheet); ok {
		return s
	}
	return nil
}

// OpenSheets returns the currently open sheets.
func OpenSheets(exclude *Sheet) []*Sheet {
	var sheets []*Sheet
	for _, d := range AllDockables() {
		if sheet, ok := d.(*Sheet); ok && sheet != exclude {
			sheets = append(sheets, sheet)
		}
	}
	return sheets
}

// NewSheetFromFile loads a GURPS character sheet file and creates a new unison.Dockable for it.
func NewSheetFromFile(filePath string) (unison.Dockable, error) {
	return openDockableFromFile(filePath, gurps.NewEntityFromFile, NewSheet)
}

// NewSheet creates a new unison.Dockable for GURPS character sheet files.
func NewSheet(filePath string, entity *gurps.Entity) *Sheet {
	s := &Sheet{
		undoMgr: unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:  unison.NewScrollPanel(),
		entity:  entity,
		scale:   gurps.GlobalSettings().General.InitialSheetUIScale,
		content: unison.NewPanel(),
	}
	s.Self = s
	s.initFileEditor(s, filePath, gurps.SheetExt, entity.Save, entity)
	s.targetMgr = NewTargetMgr(s)
	s.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})

	installDropRerouting(s.AsPanel(), dropKeys, s.keyToPanel)

	s.page = NewPage(s.entity)
	// The page is the only thing the content has ever held. The stacking layout adds the ability to put the layout
	// editor's overlay on top of it, at exactly the page's size, and is otherwise indistinguishable from the single
	// column the content used to be laid out as.
	s.contentLayout = &overlayStackLayout{page: s.page}
	s.content.SetLayout(s.contentLayout)
	s.content.AddChild(s.page)
	// Every block that isn't a list is built once, here, and kept for as long as the sheet lives, whether the layout
	// shows it or not. Re-parenting the same panels on each rebuild keeps the field focus and the target manager's
	// references to them valid, and a block that is hidden today can be shown again without rebuilding it.
	s.blocks = make(map[string]unison.Paneler, len(gurps.AllBlockKeys))
	for _, key := range gurps.AllBlockKeys {
		if block := newSheetBlockPanel(key, s.entity, s.targetMgr); !xreflect.IsNil(block) {
			s.blocks[key] = block
		}
	}
	s.buildLayout()
	s.scroll.SetContent(s.content, behavior.Unmodified, behavior.Unmodified)
	s.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	s.createToolbar()
	s.AddChild(s.scroll)

	s.InstallCmdHandlers(SaveItemID, func(_ any) bool { return s.Modified() }, func(_ any) { s.save(false) })
	s.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { s.save(true) })
	s.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, func() itemCreator { return s.Traits })
	s.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, func() itemCreator { return s.Skills })
	s.installNewItemCmdHandlers(NewTechniqueItemID, -1, func() itemCreator { return s.Skills })
	s.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, func() itemCreator { return s.Spells })
	s.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, func() itemCreator { return s.Spells })
	s.installNewItemCmdHandlers(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID,
		func() itemCreator { return s.CarriedEquipment })
	s.installNewItemCmdHandlers(NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID,
		func() itemCreator { return s.OtherEquipment })
	s.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, func() itemCreator { return s.Notes })
	s.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems(s, s.Traits.Table, s.entity.TraitList, s.entity.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
				return s.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(s.entity, nil))
	})
	s.InstallCmdHandlers(OrganizeTraitsItemID, unison.AlwaysEnabled, func(_ any) { organizeTraits(s, s.Traits.Table) })
	s.InstallCmdHandlers(SwapDefaultsItemID, s.canSwapDefaults, s.swapDefaults)
	InstallExportCmdHandlers(s)
	s.InstallCmdHandlers(ClearPortraitItemID, s.canClearPortrait, s.clearPortrait)
	s.InstallCmdHandlers(ExportPortraitItemID, s.canExportPortrait, s.exportPortrait)
	s.InstallCmdHandlers(CloneSheetItemID, unison.AlwaysEnabled, func(_ any) { s.cloneSheet() })
	s.InstallCmdHandlers(EditSheetLayoutItemID, unison.AlwaysEnabled, func(_ any) { s.toggleLayoutEditing() })
	return s
}

// PageInfoProvider returns the page info provider for this sheet.
func (s *Sheet) PageInfoProvider() gurps.PageInfoProvider {
	return s.entity
}

// CloneSheet loads the specified sheet file and creates a new character sheet from it.
func CloneSheet(filePath string) {
	d, err := NewSheetFromFile(filePath)
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to load character sheet"), err)
		return
	}
	if s, ok := d.(*Sheet); ok {
		s.cloneSheet()
	}
}

func (s *Sheet) cloneSheet() {
	unableToCloneMsg := i18n.Text("Unable to clone character sheet")
	data, err := jio.Marshal(s.entity)
	if err != nil {
		Workspace.ErrorHandler(unableToCloneMsg, err)
		return
	}
	entity := gurps.NewEntity()
	if err = jio.Unmarshal(data, entity); err != nil {
		Workspace.ErrorHandler(unableToCloneMsg, err)
		return
	}
	entity.ID = tid.MustNewTID(kinds.Entity)
	entity.CreatedOn = jio.Now()
	entity.Profile.ApplyRandomizers(s.entity)
	entity.ModifiedOn = entity.CreatedOn
	sheet := NewSheet(entity.Profile.Name+gurps.SheetExt, entity)
	DisplayNewDockable(sheet)
	sheet.undoMgr.Clear()
	sheet.hash = 0
}

func (s *Sheet) createToolbar() {
	s.toolbar = newToolbar()
	s.AddChild(s.toolbar)
	s.toolbar.AddChild(NewDefaultInfoPop())

	addHelpButton(s.toolbar, "md:User%20Guide/Character%20Sheet%20Overview")
	addUIScaleField(s.toolbar, func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
		func() int { return s.scale }, func(scale int) { s.scale = scale }, true, s.scroll)

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = s.toggleHierarchy
	s.toolbar.AddChild(hierarchyButton)

	noteToggleButton := unison.NewSVGButton(svg.NotesToggle)
	noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	noteToggleButton.ClickCallback = s.toggleNotes
	s.toolbar.AddChild(noteToggleButton)

	sheetSettingsButton := unison.NewSVGButton(svg.Settings)
	sheetSettingsButton.Tooltip = newWrappedTooltip(i18n.Text("Sheet Settings"))
	sheetSettingsButton.ClickCallback = func() { ShowSheetSettings(s) }
	s.toolbar.AddChild(sheetSettingsButton)

	attributesButton := unison.NewSVGButton(svg.Attributes)
	attributesButton.Tooltip = newWrappedTooltip(i18n.Text("Attributes"))
	attributesButton.ClickCallback = func() { ShowAttributeSettings(s) }
	s.toolbar.AddChild(attributesButton)

	bodyTypeButton := unison.NewSVGButton(svg.BodyType)
	bodyTypeButton.Tooltip = newWrappedTooltip(i18n.Text("Body Type"))
	bodyTypeButton.ClickCallback = func() { ShowBodySettings(s) }
	s.toolbar.AddChild(bodyTypeButton)

	// The layout button latches, which unison only draws for a button that belongs to a group, so it is given one of
	// its own. The base is hidden while the button is off, the way every other toolbar button's is, and shown while it
	// is on, so that editing the layout reads as a filled chip rather than a mere change of tint.
	s.layoutButton = unison.NewSVGButton(svg.Layout)
	s.layoutButton.Sticky = true
	s.layoutButtonGroup = unison.NewGroup(s.layoutButton)
	s.layoutButton.Tooltip = newWrappedTooltip(i18n.Text("Edit the sheet's block layout"))
	s.layoutButton.ClickCallback = s.toggleLayoutEditing
	s.toolbar.AddChild(s.layoutButton)

	layoutMenuButton := unison.NewSVGButton(svg.CircledVerticalEllipsis)
	layoutMenuButton.Tooltip = newWrappedTooltip(i18n.Text("Block layout menu"))
	layoutMenuButton.ClickCallback = func() { s.showLayoutMenu(layoutMenuButton) }
	s.toolbar.AddChild(layoutMenuButton)

	cloneSheetButton := unison.NewSVGButton(svg.Clone)
	cloneSheetButton.Tooltip = newWrappedTooltip(cloneSheetAction.Title)
	cloneSheetButton.ClickCallback = s.cloneSheet
	s.toolbar.AddChild(cloneSheetButton)

	syncSourceButton := unison.NewSVGButton(svg.DownToBracket)
	syncSourceButton.Tooltip = newWrappedTooltip(i18n.Text("Sync with all sources in this sheet"))
	syncSourceButton.ClickCallback = s.syncWithAllSources
	s.toolbar.AddChild(syncSourceButton)

	calcButton := unison.NewSVGButton(svg.Calculator)
	calcButton.Tooltip = newWrappedTooltip(i18n.Text("Calculators (jumping, throwing, hiking, etc.)"))
	calcButton.ClickCallback = func() { DisplayCalculator(s) }
	s.toolbar.AddChild(calcButton)

	s.searchTracker = installListSearchTracker(s.toolbar, s.lists)

	finishToolbarLayout(s.toolbar)
}

// DataOwner implements gurps.DataOwnerProvider.
func (s *Sheet) DataOwner() gurps.DataOwner {
	return s.entity
}

func (s *Sheet) canExportPortrait(_ any) bool {
	return s.entity.Profile.CanExportPortrait()
}

func (s *Sheet) exportPortrait(_ any) {
	if s.entity.Profile.CanExportPortrait() {
		if ext := s.entity.Profile.PortraitExtension(); ext != "" {
			s.Window().ShowCursor()
			dialog := unison.NewSaveDialog()
			backingFilePath := s.BackingFilePath()
			dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
			dialog.SetAllowedExtensions(ext)
			dialog.SetInitialFileName(xfilepath.SanitizeName(xfilepath.BaseName(backingFilePath)))
			if dialog.RunModal() {
				if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), ext, false); ok {
					if err := s.entity.Profile.ExportPortrait(filePath); err != nil {
						Workspace.ErrorHandler(i18n.Text("Unable to export portrait"), err)
					}
				}
			}
		}
	}
}

func (s *Sheet) canClearPortrait(_ any) bool {
	return len(s.entity.Profile.PortraitData) != 0
}

func (s *Sheet) clearPortrait(_ any) {
	if s.canClearPortrait(nil) {
		s.undoMgr.Add(&unison.UndoEdit[[]byte]{
			ID:         unison.NextUndoID(),
			EditName:   clearPortraitAction.Title,
			UndoFunc:   func(edit *unison.UndoEdit[[]byte]) { s.updatePortrait(edit.BeforeData) },
			RedoFunc:   func(edit *unison.UndoEdit[[]byte]) { s.updatePortrait(edit.AfterData) },
			BeforeData: s.entity.Profile.PortraitData,
			AfterData:  nil,
		})
		s.updatePortrait(nil)
	}
}

func (s *Sheet) updatePortrait(data []byte) {
	s.entity.Profile.SetPortraitData(data)
	s.MarkForRedraw()
	s.MarkModified(s)
}

func (s *Sheet) keyToPanel(key *uti.DataType) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = s.CarriedEquipment.Table
	case skillDragKey:
		p = s.Skills.Table
	case spellDragKey:
		p = s.Spells.Table
	case traitDragKey:
		p = s.Traits.Table
	case noteDragKey:
		p = s.Notes.Table
	default:
		return nil
	}
	if !unison.AncestorIs(p, s.page) {
		// The layout doesn't place the list, so it is no more a drop target than it is something the sheet can draw
		// the drop feedback over.
		return nil
	}
	return p.AsPanel()
}

// installNewItemCmdHandlers installs the handlers for the "New ..." commands that add an item to one of the sheet's
// lists. The list is looked up through the getter each time a command is invoked rather than captured here, since a
// list whose set of columns has to change can only do so by being replaced outright (a table's columns are fixed at
// creation -- see PageList.needReconstruction), which leaves the list that was captured orphaned. Creating an item in
// an orphaned list still updates the model, but everything that goes with it is aimed at a table nobody is looking at:
// the undo edit can't even find the undo manager, so the insertion isn't undoable and the user's next undo silently
// takes back the edit before it, and the new row is neither selected nor scrolled into view in the list that is
// actually on screen.
func (s *Sheet) installNewItemCmdHandlers(itemID, containerID int, creator func() itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		s.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator().CreateItem(s, ContainerItemVariant) })
	}
	s.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator().CreateItem(s, variant) })
}

// DockableKind implements widget.DockableKind
func (s *Sheet) DockableKind() string {
	return SheetDockableKind
}

// Entity returns the entity this is displaying information for.
func (s *Sheet) Entity() *gurps.Entity {
	return s.entity
}

// UndoManager implements undo.Provider
func (s *Sheet) UndoManager() *unison.UndoManager {
	return s.undoMgr
}

// BackingFilePath implements FileBackedDockable. A sheet that has never been saved goes by its character's name.
func (s *Sheet) BackingFilePath() string {
	if s.needsSaveAsPrompt {
		return unsavedFileName(s.entity.Profile.Name, i18n.Text("Unnamed Character"), gurps.SheetExt)
	}
	return s.path
}

// MarkModified implements widget.ModifiableRoot.
func (s *Sheet) MarkModified(src unison.Paneler) {
	if !s.awaitingUpdate {
		s.awaitingUpdate = true
		// Everything below reads the derived state -- the panels, the tables, and the calculator all display skill
		// levels, points and the like -- so the entity is brought up to date first. This used to happen by accident,
		// as a side effect of the tab asking whether the sheet had unsaved changes, which recalculated the entity on
		// its way to hashing it.
		s.entity.Recalculate()
		s.bumpModificationTimestamp()
		UpdateTitleForDockable(s)
		skipDeepSync := false
		if !xreflect.IsNil(src) {
			_, skipDeepSync = src.AsPanel().ClientData()[SkipDeepSync]
		}
		if skipDeepSync {
			// The deep sync is what rebuilds the tables, and it is also the only thing here that can disturb the
			// focus and scroll position or change which rows match the active search. When it is skipped (e.g. while
			// typing into a simple field such as the name or title), saving and restoring the focus and scroll
			// position and refreshing the search results is just wasted work, and that overhead is enough to make
			// interactive typing stutter on slower platforms. So none of it is done in that case.
			s.awaitingUpdate = false
		} else {
			h, v := s.scroll.Position()
			focusRefKey := s.targetMgr.CurrentFocusRef()
			// TODO: This can be too slow when the lists have many rows of content, impinging upon interactive typing.
			//       Looks like most of the time is spent in updating the tables. Unfortunately, there isn't a fast way
			//       to determine that the content of a table doesn't need to be refreshed.
			DeepSync(s)
			s.awaitingUpdate = false
			s.searchTracker.Refresh()
			s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
			s.scroll.SetPosition(h, v)
		}
		UpdateCalculator(s)
	}
}

// bumpModificationTimestamp implements modificationTimestampBumper.
func (s *Sheet) bumpModificationTimestamp() {
	if miscPanel, ok := s.blocks[gurps.BlockMiscellaneousKey].(*MiscPanel); ok {
		miscPanel.UpdateModified()
	}
}

// buildLayout (re)builds the page from the sheet's block layout tree. The panels themselves are not rebuilt: the
// blocks that aren't lists are built once and kept (see NewSheet), and a list is only built anew when the columns it
// has to show no longer match the ones it has, since a table's columns are fixed at creation. Anything that captured a
// list has to allow for it being replaced -- see installNewItemCmdHandlers.
func (s *Sheet) buildLayout() {
	// Everything the sheet keeps is detached first, so that a block the layout doesn't place is left without a parent.
	// That is how the rest of the sheet tells a block that is on the page from one that isn't: removing the page's
	// children only detaches the bands, leaving anything nested inside them still pointing at a band nobody can see.
	for _, block := range s.blocks {
		block.AsPanel().RemoveFromParent()
	}
	for _, list := range s.lists() {
		if !xreflect.IsNil(list) {
			list.AsPanel().RemoveFromParent()
		}
	}
	s.page.RemoveAllChildren()
	layout := s.entity.SheetSettings.Layout
	for _, band := range buildLayoutBands(layout.Root, s.layoutLeaf) {
		s.page.AddChild(band)
	}
	// A list the layout doesn't show is still brought into being, since the undo data, the disclosure handling, the
	// searching and the rebuild's selection tracking all reach for all ten of them without asking whether they are on
	// the page.
	for _, key := range gurps.AllBlockKeys {
		if gurps.IsListBlockKey(key) && !layout.Contains(key) {
			s.layoutLeaf(key)
		}
	}
	s.page.ApplyPreferredSize()
	if s.layoutEditing() {
		// The page just changed size and every block moved, so the overlay has to be brought back over it and the
		// regions it works from thrown away.
		s.layoutEditor.syncFrame()
	}
}

// layoutLeaf returns the panel to show for the block with the given key, or nil if that block has nothing to show. The
// four lists that are derived from the character rather than edited directly are omitted when they are empty.
func (s *Sheet) layoutLeaf(key string) unison.Paneler {
	switch key {
	case gurps.BlockReactionsKey:
		return derivedListLeaf(s, &s.Reactions, func() *PageList[*gurps.ConditionalModifier] {
			return NewReactionsPageList(s.entity)
		})
	case gurps.BlockConditionalModifiersKey:
		return derivedListLeaf(s, &s.ConditionalModifiers, func() *PageList[*gurps.ConditionalModifier] {
			return NewConditionalModifiersPageList(s.entity)
		})
	case gurps.BlockMeleeKey:
		return derivedListLeaf(s, &s.MeleeWeapons, func() *PageList[*gurps.Weapon] {
			return NewMeleeWeaponsPageList(s.entity)
		})
	case gurps.BlockRangedKey:
		return derivedListLeaf(s, &s.RangedWeapons, func() *PageList[*gurps.Weapon] {
			return NewRangedWeaponsPageList(s.entity)
		})
	case gurps.BlockTraitsKey:
		return syncOrRebuildList(&s.Traits, func() *PageList[*gurps.Trait] { return NewTraitsPageList(s, s.entity) })
	case gurps.BlockSkillsKey:
		return syncOrRebuildList(&s.Skills, func() *PageList[*gurps.Skill] { return NewSkillsPageList(s, s.entity) })
	case gurps.BlockSpellsKey:
		return syncOrRebuildList(&s.Spells, func() *PageList[*gurps.Spell] { return NewSpellsPageList(s, s.entity) })
	case gurps.BlockEquipmentKey:
		return syncOrRebuildList(&s.CarriedEquipment, func() *PageList[*gurps.Equipment] {
			return NewCarriedEquipmentPageList(s, s.entity)
		})
	case gurps.BlockOtherEquipmentKey:
		return syncOrRebuildList(&s.OtherEquipment, func() *PageList[*gurps.Equipment] {
			return NewOtherEquipmentPageList(s, s.entity)
		})
	case gurps.BlockNotesKey:
		return syncOrRebuildList(&s.Notes, func() *PageList[*gurps.Note] { return NewNotesPageList(s, s.entity) })
	default:
		block, exists := s.blocks[key]
		if !exists {
			return nil
		}
		// The same panel is used for as long as the sheet lives, so it has to be taken out of the band it was in the
		// last time the page was built before it can go into a new one.
		block.AsPanel().RemoveFromParent()
		return block
	}
}

// derivedListLeaf syncs or rebuilds one of the sheet's four lists that are derived from the character rather than
// edited directly (see syncOrRebuildList) and returns it as the panel to show for its block, or nil when it has no
// rows, since an empty derived list is left off the page. These lists are built without an owner, so the sheet has to
// be attached to their tables as the data owner provider by hand.
func derivedListLeaf[T gurps.Node[T]](s *Sheet, list **PageList[T], build func() *PageList[T]) unison.Paneler {
	l := syncOrRebuildList(list, build)
	SetDataOwnerProvider(l.Table, s)
	if l.Table.RootRowCount() == 0 {
		return nil
	}
	return l
}

// blockPanel returns the panel the sheet uses for the block with the given key, or nil if it has none. Unlike
// layoutLeaf, this neither creates nor synchronizes anything, so it is what the layout editor maps a panel it found on
// the page back to a block key with.
func (s *Sheet) blockPanel(key string) unison.Paneler {
	if gurps.IsListBlockKey(key) {
		return s.list(key)
	}
	return s.blocks[key]
}

// list returns the sheet's list for the given block key, or nil if the key isn't one of the ten list blocks. A list the
// sheet hasn't built yet -- which is only the case while the sheet is first being put together -- comes back as a nil
// *PageList inside the interface, which xreflect.IsNil sees through.
func (s *Sheet) list(key string) sheetList {
	switch key {
	case gurps.BlockReactionsKey:
		return s.Reactions
	case gurps.BlockConditionalModifiersKey:
		return s.ConditionalModifiers
	case gurps.BlockMeleeKey:
		return s.MeleeWeapons
	case gurps.BlockRangedKey:
		return s.RangedWeapons
	case gurps.BlockTraitsKey:
		return s.Traits
	case gurps.BlockSkillsKey:
		return s.Skills
	case gurps.BlockSpellsKey:
		return s.Spells
	case gurps.BlockEquipmentKey:
		return s.CarriedEquipment
	case gurps.BlockOtherEquipmentKey:
		return s.OtherEquipment
	case gurps.BlockNotesKey:
		return s.Notes
	default:
		return nil
	}
}

// lists returns the sheet's ten lists, in the canonical block order (see gurps.AllBlockKeys). See list for what comes
// back for a list the sheet hasn't built yet.
func (s *Sheet) lists() []sheetList {
	return listsForKeys(s.list, gurps.IsListBlockKey)
}

// syncDisclosure brings the blocks that show state which can be disclosed back into line with the model. The tables
// do this for themselves; these panels have to be told.
func (s *Sheet) syncDisclosure() {
	for _, key := range []string{
		gurps.BlockPrimaryAttributesKey,
		gurps.BlockSecondaryAttributesKey,
		gurps.BlockPointPoolsKey,
	} {
		if attrPanel, ok := s.blocks[key].(*AttrPanel); ok {
			attrPanel.forceSync()
		}
	}
	if damagePanel, ok := s.blocks[gurps.BlockDamageKey].(*DamagePanel); ok {
		damagePanel.forceSync()
	}
	if bodyPanel, ok := s.blocks[gurps.BlockBodyKey].(*BodyPanel); ok {
		bodyPanel.sync(true)
	}
}

func (s *Sheet) canSwapDefaults(_ any) bool {
	canSwap := false
	for _, skillNode := range s.Skills.SelectedNodes(true) {
		skill := skillNode.Data()
		if skill.IsTechnique() {
			return false
		}
		if !skill.CanSwapDefaultsWith(skill.DefaultSkill()) &&
			skill.BestSwappableSkill() == nil &&
			!skill.AlternateDefaultsAvailable() {
			return false
		}
		canSwap = true
	}
	return canSwap
}

func (s *Sheet) swapDefaults(_ any) {
	undo := &unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]{
		ID:       unison.NextUndoID(),
		EditName: swapDefaultsAction.Title,
		UndoFunc: func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]) { e.BeforeData.Apply() },
		RedoFunc: func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]) { e.AfterData.Apply() },
		AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]], _ unison.Undoable) bool {
			return false
		},
		BeforeData: NewTableUndoEditData(s.Skills.Table),
	}
	for _, skillNode := range s.Skills.SelectedNodes(true) {
		skill := skillNode.Data()
		if !skill.CanSwapDefaults() {
			continue
		}
		if skill.AlternateDefaultsAvailable() {
			skill.SwapToNextDefault()
		} else if swap := skill.DefaultSkill(); skill.CanSwapDefaultsWith(swap) {
			skill.DefaultedFrom = nil
			swap.SwapDefaults()
		} else if other := skill.BestSwappableSkill(); other != nil {
			skill.DefaultedFrom = nil
			other.SwapDefaults()
		}
	}
	// Marking the sheet as modified recalculates the entity and re-syncs everything that shows a skill level, which
	// swapping a default can change well beyond the skills list (weapons, for one), and also bumps the modification
	// timestamp and updates the title, none of which recalculating and syncing the skills table by hand did.
	s.MarkModified(nil)
	undo.AfterData = NewTableUndoEditData(s.Skills.Table)
	s.UndoManager().Add(undo)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (s *Sheet) SheetSettingsUpdated(entity *gurps.Entity, fullRebuild bool) {
	if s.entity == entity {
		// A single rebuild both reports the change and refreshes everything the settings affect; marking the sheet as
		// modified first would only perform the same update a second time (see rebuildAsModified).
		rebuildAsModified(s, fullRebuild)
	}
}

// layoutEditing returns true if the sheet is currently in block layout editing mode.
func (s *Sheet) layoutEditing() bool {
	return s.layoutEditor != nil
}

// toggleLayoutEditing turns block layout editing mode on or off.
func (s *Sheet) toggleLayoutEditing() {
	if s.layoutEditing() {
		editor := s.layoutEditor
		s.layoutEditor = nil
		editor.stop()
	} else {
		s.layoutEditor = newSheetLayoutEditor(s)
		s.layoutEditor.start()
	}
	editing := s.layoutEditing()
	if s.layoutButton != nil {
		s.layoutButton.HideBase = !editing
		if editing {
			s.layoutButtonGroup.Select(s.layoutButton)
		} else {
			s.layoutButtonGroup.Select(nil)
		}
		s.layoutButton.MarkForLayoutAndRedraw()
	}
}

// applySheetLayout replaces the sheet's block layout with a copy of the one given and shows the result. The sheet owns
// this rather than the editor, since an undo can be asked for from the Edit menu long after editing mode was left.
func (s *Sheet) applySheetLayout(l *gurps.SheetLayout) {
	s.entity.SheetSettings.Layout = l.Clone()
	rebuildAsModified(s, true)
}

// recordLayoutUndo adds an undo edit that moves the sheet's block layout between the two states given. Both are taken
// over by the edit, which hands out copies of them, so neither may be modified afterwards.
func (s *Sheet) recordLayoutUndo(name string, before, after *gurps.SheetLayout) {
	mgr := s.UndoManager()
	if mgr == nil {
		return
	}
	mgr.Add(&unison.UndoEdit[*gurps.SheetLayout]{
		ID:         unison.NextUndoID(),
		EditName:   name,
		UndoFunc:   func(e *unison.UndoEdit[*gurps.SheetLayout]) { s.applySheetLayout(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*gurps.SheetLayout]) { s.applySheetLayout(e.AfterData) },
		AbsorbFunc: func(_ *unison.UndoEdit[*gurps.SheetLayout], _ unison.Undoable) bool { return false },
		BeforeData: before,
		AfterData:  after,
	})
}

// changeLayout applies the given alteration to the sheet's block layout, records it as a single undoable edit and
// shows the result. An alteration that leaves the layout as it was is dropped, so that a gesture that changed nothing
// doesn't put an edit that does nothing onto the undo stack.
func (s *Sheet) changeLayout(name string, alter func(l *gurps.SheetLayout)) {
	layout := s.entity.SheetSettings.Layout
	before := layout.Clone()
	alter(layout)
	if gurps.Hash64(before) == gurps.Hash64(layout) {
		return
	}
	s.recordLayoutUndo(name, before, layout.Clone())
	rebuildAsModified(s, true)
}

// hideLayoutBlock removes the block with the given key from the sheet.
func (s *Sheet) hideLayoutBlock(key string) {
	s.changeLayout(i18n.Text("Hide Block"), func(l *gurps.SheetLayout) { l.Hide(key) })
}

// showLayoutBlock puts a hidden block back onto the sheet, as a new band at the bottom.
func (s *Sheet) showLayoutBlock(key string) {
	s.changeLayout(i18n.Text("Show Block"), func(l *gurps.SheetLayout) { l.Show(key) })
}

// resetLayout returns the sheet to the block layout new sheets are given, which is what "Reset" in the per-sheet
// settings does for every other setting. Restoring the factory layout is the job of the global defaults' own reset.
func (s *Sheet) resetLayout() {
	s.changeLayout(i18n.Text("Reset Layout"), func(l *gurps.SheetLayout) {
		*l = *gurps.GlobalSettings().Sheet.Layout.Clone()
		l.EnsureValidity()
	})
}

// useLayoutAsDefault makes this sheet's block layout the one new sheets are created with.
func (s *Sheet) useLayoutAsDefault() {
	setDefaultSheetLayout(s.entity.SheetSettings.Layout.Clone())
}

// resetDefaultLayout returns the block layout new sheets are created with to the factory one. This sheet's own layout
// is left alone; "Reset Layout to Default" is what brings a sheet back to the default.
func (s *Sheet) resetDefaultLayout() {
	setDefaultSheetLayout(gurps.FactorySheetLayout())
}

// setDefaultSheetLayout installs the given layout as the one new sheets are created with. The global settings are
// written out at shutdown, as they are for every other in-place change to them, so nothing is saved here.
func setDefaultSheetLayout(layout *gurps.SheetLayout) {
	gurps.GlobalSettings().Sheet.Layout = layout
	// A sheet file without embedded settings clones the published snapshot of the global sheet settings rather than the
	// live ones, so the snapshot has to be republished for such a sheet opened from now on to get this layout.
	gurps.SyncGlobalSheetSettings()
	// Templates lay themselves out from the default layout, so they have to be told. A nil entity is what says the
	// change was to the defaults rather than to some sheet's own settings, which is why the open sheets ignore it.
	for _, one := range AllDockables() {
		if responder, ok := one.(gurps.SheetSettingsResponder); ok {
			responder.SheetSettingsUpdated(nil, true)
		}
	}
}

// showLayoutMenu pops up the block layout menu beneath the given toolbar button.
func (s *Sheet) showLayoutMenu(b *unison.Button) {
	f := unison.DefaultMenuFactory()
	m := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
	id := 1
	s.appendLayoutMenuItems(f, m, &id, "")
	m.Popup(b.RectToRoot(b.ContentRect(true)), 0)
	m.Dispose()
}

// appendLayoutMenuItems adds the block layout commands to the given menu, numbering the items it makes from the given
// counter. A non-empty hideKey adds the command that hides that particular block, which the overlay's context menu
// supplies and the toolbar's menu does not, since only the overlay knows which block was clicked on.
func (s *Sheet) appendLayoutMenuItems(f unison.MenuFactory, m unison.Menu, id *int, hideKey string) {
	editItem := f.NewItem(nextLayoutMenuItemID(id), i18n.Text("Edit Layout"), unison.KeyBinding{}, nil,
		func(_ unison.MenuItem) { s.toggleLayoutEditing() })
	if s.layoutEditing() {
		editItem.SetCheckState(check.On)
	} else {
		editItem.SetCheckState(check.Off)
	}
	m.InsertItem(-1, editItem)
	if hideKey != "" {
		m.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id),
			fmt.Sprintf(i18n.Text("Hide %s"), gurps.BlockTitle(hideKey)), unison.KeyBinding{}, nil,
			func(_ unison.MenuItem) { s.hideLayoutBlock(hideKey) }))
	}
	if hideKey == gurps.BlockPortraitKey && s.layoutEditing() {
		// Only the overlay supplies a hide key, and it is only in place while the layout is being edited, so this is
		// never reached from the toolbar's menu, which has no block to work on.
		m.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id), i18n.Text("Make Portrait Square"), unison.KeyBinding{},
			nil, func(_ unison.MenuItem) {
				if s.layoutEditing() {
					s.layoutEditor.squarePortrait()
				}
			}))
	}
	if hidden := s.entity.SheetSettings.Layout.HiddenKeys(); len(hidden) != 0 {
		sub := f.NewMenu(nextLayoutMenuItemID(id), i18n.Text("Show"), nil)
		for _, key := range hidden {
			sub.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id), gurps.BlockTitle(key), unison.KeyBinding{}, nil,
				func(_ unison.MenuItem) { s.showLayoutBlock(key) }))
		}
		m.InsertMenu(-1, sub)
	}
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id),
		i18n.Text("Use This Layout as the Default for New Sheets"), unison.KeyBinding{}, nil,
		func(_ unison.MenuItem) { s.useLayoutAsDefault() }))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id), i18n.Text("Reset Layout to Default"), unison.KeyBinding{},
		nil, func(_ unison.MenuItem) { s.resetLayout() }))
	m.InsertItem(-1, f.NewItem(nextLayoutMenuItemID(id),
		i18n.Text("Reset the Default Layout to Factory Settings"), unison.KeyBinding{}, nil,
		func(_ unison.MenuItem) { s.resetDefaultLayout() }))
}

// nextLayoutMenuItemID hands out the next ID for an item of a temporary popup menu.
func nextLayoutMenuItemID(id *int) int {
	next := unison.PopupMenuTemporaryBaseID + *id
	*id++
	return next
}

func (s *Sheet) syncWithAllSources() {
	syncWithAllSources(s, s.entity, s.Traits, s.Skills, s.Spells, s.CarriedEquipment, s.OtherEquipment, s.Notes)
}

// Rebuild implements widget.Rebuildable.
func (s *Sheet) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	h, v := s.scroll.Position()
	focusRefKey := s.targetMgr.CurrentFocusRef()
	s.entity.Recalculate()
	if full {
		defer preserveSelections(s.lists)()
		s.buildLayout()
	}
	DeepSync(s)
	UpdateTitleForDockable(s)
	s.searchTracker.Refresh()
	s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
	s.scroll.SetPosition(h, v)
	if s.layoutEditing() {
		// Reacquiring the focus just took it back to whichever field held it before the overlay went up, so the overlay
		// has to ask for it again. Everything it drew is stale, too, since the page has been rebuilt underneath it.
		s.layoutEditor.invalidateRegions()
		s.layoutEditor.syncFrame()
		s.layoutEditor.overlay.RequestFocus()
	}
	UpdateCalculator(s)
}

func drawBandedBackground(p unison.Paneler, gc *unison.Canvas, rect geom.Rect, start, step int, overrideFunc func(rowIndex int, ink unison.Ink) unison.Ink) {
	gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
	children := p.AsPanel().Children()
	row := 0
	for i := start; i < len(children); i += step {
		var ink unison.Ink
		if ((i-start)/step)&1 == 1 {
			ink = unison.ThemeBanding
		} else {
			ink = unison.ThemeBelowSurface
		}
		if overrideFunc != nil {
			ink = overrideFunc(row, ink)
			row++
		}
		r := children[i].FrameRect()
		for j := i + 1; j < i+step; j++ {
			r = r.Union(children[j].FrameRect())
		}
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, ink.Paint(gc, r, paintstyle.Fill))
	}
}

// BodySettingsTitle implements BodySettingsOwner.
func (s *Sheet) BodySettingsTitle() string {
	return fmt.Sprintf(i18n.Text("Body Type: %s"), s.entity.Profile.Name)
}

// BodySettings implements BodySettingsOwner.
func (s *Sheet) BodySettings(forReset bool) *gurps.Body {
	if forReset {
		return gurps.GlobalSettings().Sheet.BodyType
	}
	return s.entity.SheetSettings.BodyType
}

// SetBodySettings implements BodySettingsOwner.
func (s *Sheet) SetBodySettings(body *gurps.Body) {
	s.entity.SheetSettings.BodyType = body
	for _, one := range AllDockables() {
		if responder, ok := one.(gurps.SheetSettingsResponder); ok {
			responder.SheetSettingsUpdated(s.entity, true)
		}
	}
}

func (s *Sheet) toggleHierarchy() {
	tables := s.lists()
	open, exists := s.entity.Attributes.FirstDisclosureState(s.entity)
	if !exists {
		if open, exists = s.entity.SheetSettings.BodyType.FirstDisclosureState(); !exists {
			for _, table := range tables {
				if open, exists = table.FirstDisclosureState(); exists {
					break
				}
			}
		}
	}
	open = !open
	s.entity.Attributes.SetDisclosureState(s.entity, open)
	s.entity.SheetSettings.BodyType.SetDisclosureState(open)
	for _, table := range tables {
		table.SetDisclosureState(open)
	}
	s.syncDisclosure()
	s.Rebuild(true)
}

func (s *Sheet) toggleNotes() {
	tables := s.lists()
	state := 0
	for _, table := range tables {
		if state = table.FirstNoteState(); state != 0 {
			break
		}
	}
	if state == 0 {
		return
	}
	var closed bool
	if state == 1 {
		closed = true
	}
	for _, table := range tables {
		table.ApplyNoteState(closed)
	}
	s.syncDisclosure()
	s.Rebuild(true)
}
