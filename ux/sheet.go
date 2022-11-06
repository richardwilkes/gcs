/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	gsettings "github.com/richardwilkes/gcs/v5/model/gurps/settings"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/printing"
)

// SkipDeepSync is set on components that should not trigger a deep sync.
const SkipDeepSync = "!deepsync"

var (
	_        FileBackedDockable         = &Sheet{}
	_        unison.UndoManagerProvider = &Sheet{}
	_        ModifiableRoot             = &Sheet{}
	_        Rebuildable                = &Sheet{}
	_        DockableKind               = &Sheet{}
	_        unison.TabCloser           = &Sheet{}
	dropKeys                            = []string{
		gid.Equipment,
		gid.Skill,
		gid.Spell,
		gid.Trait,
		gid.Note,
	}
)

type itemCreator interface {
	CreateItem(Rebuildable, ItemVariant)
}

// Sheet holds the view for a GURPS character sheet.
type Sheet struct {
	unison.Panel
	path                 string
	targetMgr            *TargetMgr
	undoMgr              *unison.UndoManager
	toolbar              *unison.Panel
	scroll               *unison.ScrollPanel
	entity               *gurps.Entity
	crc                  uint64
	content              *unison.Panel
	modifiedFunc         func()
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
	dragReroutePanel     *unison.Panel
	scale                int
	awaitingUpdate       bool
	needsSaveAsPrompt    bool
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
	ws := WorkspaceFromWindowOrAny(unison.ActiveWindow())
	ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, one := range dc.Dockables() {
			if sheet, ok := one.(*Sheet); ok && sheet != exclude {
				sheets = append(sheets, sheet)
			}
		}
		return false
	})
	return sheets
}

// NewSheetFromFile loads a GURPS character sheet file and creates a new unison.Dockable for it.
func NewSheetFromFile(filePath string) (unison.Dockable, error) {
	entity, err := gurps.NewEntityFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	s := NewSheet(filePath, entity)
	s.needsSaveAsPrompt = false
	return s, nil
}

// NewSheet creates a new unison.Dockable for GURPS character sheet files.
func NewSheet(filePath string, entity *gurps.Entity) *Sheet {
	s := &Sheet{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { jot.Error(err) }),
		scroll:            unison.NewScrollPanel(),
		entity:            entity,
		crc:               entity.CRC64(),
		scale:             settings.Global().General.InitialSheetUIScale,
		content:           unison.NewPanel(),
		needsSaveAsPrompt: true,
	}
	s.Self = s
	s.targetMgr = NewTargetMgr(s)
	s.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  unison.FillAlignment,
		VAlign:  unison.FillAlignment,
	})

	s.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
		s.RequestFocus()
		return false
	}
	s.DataDragOverCallback = func(_ unison.Point, data map[string]any) bool {
		s.dragReroutePanel = nil
		for _, key := range dropKeys {
			if _, ok := data[key]; ok {
				if s.dragReroutePanel = s.keyToPanel(key); s.dragReroutePanel != nil {
					s.dragReroutePanel.DataDragOverCallback(unison.Point{Y: 100000000}, data)
					return true
				}
				break
			}
		}
		return false
	}
	s.DataDragExitCallback = func() {
		if s.dragReroutePanel != nil {
			s.dragReroutePanel.DataDragExitCallback()
			s.dragReroutePanel = nil
		}
	}
	s.DataDragDropCallback = func(_ unison.Point, data map[string]any) {
		if s.dragReroutePanel != nil {
			s.dragReroutePanel.DataDragDropCallback(unison.Point{Y: 10000000}, data)
			s.dragReroutePanel = nil
		}
	}
	s.DrawOverCallback = func(gc *unison.Canvas, rect unison.Rect) {
		if s.dragReroutePanel != nil {
			r := s.RectFromRoot(s.dragReroutePanel.RectToRoot(s.dragReroutePanel.ContentRect(true)))
			paint := unison.DropAreaColor.Paint(gc, r, unison.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	s.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	var top *Page
	top, s.modifiedFunc = createPageTopBlock(s.entity, s.targetMgr)
	s.content.AddChild(top)
	s.createLists()
	s.scroll.SetContent(s.content, unison.UnmodifiedBehavior, unison.UnmodifiedBehavior)
	s.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	s.scroll.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, theme.PageVoidColor.Paint(gc, rect, unison.Fill))
	}

	sheetSettingsButton := unison.NewSVGButton(svg.Settings)
	sheetSettingsButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Sheet Settings"))
	sheetSettingsButton.ClickCallback = func() { ShowSheetSettings(s) }

	attributesButton := unison.NewSVGButton(svg.Attributes)
	attributesButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Attributes"))
	attributesButton.ClickCallback = func() { ShowAttributeSettings(s) }

	bodyTypeButton := unison.NewSVGButton(svg.BodyType)
	bodyTypeButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Body Type"))
	bodyTypeButton.ClickCallback = func() { ShowBodySettings(s) }

	s.toolbar = unison.NewPanel()
	s.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	s.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	s.toolbar.AddChild(NewDefaultInfoPop())
	s.toolbar.AddChild(NewScaleField(gsettings.InitialUIScaleMin, gsettings.InitialUIScaleMax,
		func() int { return settings.Global().General.InitialSheetUIScale }, func() int { return s.scale },
		func(scale int) { s.scale = scale }, s.scroll, nil, false))
	s.toolbar.AddChild(sheetSettingsButton)
	s.toolbar.AddChild(attributesButton)
	s.toolbar.AddChild(bodyTypeButton)
	s.toolbar.AddChild(NewToolbarSeparator())
	installSearchTracker(s.toolbar, func() {
		s.Reactions.Table.ClearSelection()
		s.ConditionalModifiers.Table.ClearSelection()
		s.MeleeWeapons.Table.ClearSelection()
		s.RangedWeapons.Table.ClearSelection()
		s.Traits.Table.ClearSelection()
		s.Skills.Table.ClearSelection()
		s.Spells.Table.ClearSelection()
		s.CarriedEquipment.Table.ClearSelection()
		s.OtherEquipment.Table.ClearSelection()
		s.Notes.Table.ClearSelection()
	}, func(refList *[]*searchRef, text string) {
		searchSheetTable(refList, text, s.Traits)
		searchSheetTable(refList, text, s.Skills)
		searchSheetTable(refList, text, s.Spells)
		searchSheetTable(refList, text, s.CarriedEquipment)
		searchSheetTable(refList, text, s.OtherEquipment)
		searchSheetTable(refList, text, s.Notes)
	})
	s.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(s.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	s.AddChild(s.toolbar)
	s.AddChild(s.scroll)

	s.InstallCmdHandlers(SaveItemID, func(_ any) bool { return s.Modified() }, func(_ any) { s.save(false) })
	s.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { s.save(true) })
	s.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, s.Traits)
	s.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, s.Skills)
	s.installNewItemCmdHandlers(NewTechniqueItemID, -1, s.Skills)
	s.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, s.Spells)
	s.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, s.Spells)
	s.installNewItemCmdHandlers(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID,
		s.CarriedEquipment)
	s.installNewItemCmdHandlers(NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID,
		s.OtherEquipment)
	s.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, s.Notes)
	s.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems[*gurps.Trait](s, s.Traits.Table, s.entity.TraitList, s.entity.SetTraitList,
			func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
				return s.Traits.provider.RootRows()
			}, gurps.NewNaturalAttacks(s.entity, nil))
	})
	s.InstallCmdHandlers(SwapDefaultsItemID, s.canSwapDefaults, s.swapDefaults)
	s.InstallCmdHandlers(ExportAsPDFItemID, unison.AlwaysEnabled, func(_ any) { s.exportToPDF() })
	s.InstallCmdHandlers(ExportAsWEBPItemID, unison.AlwaysEnabled, func(_ any) { s.exportToWEBP() })
	s.InstallCmdHandlers(ExportAsPNGItemID, unison.AlwaysEnabled, func(_ any) { s.exportToPNG() })
	s.InstallCmdHandlers(ExportAsJPEGItemID, unison.AlwaysEnabled, func(_ any) { s.exportToJPEG() })
	s.InstallCmdHandlers(PrintItemID, unison.AlwaysEnabled, func(_ any) { s.print() })

	return s
}

func (s *Sheet) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case gid.Equipment:
		p = s.CarriedEquipment.Table
	case gid.Skill:
		p = s.Skills.Table
	case gid.Spell:
		p = s.Spells.Table
	case gid.Trait:
		p = s.Traits.Table
	case gid.Note:
		p = s.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

func (s *Sheet) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		s.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(s, ContainerItemVariant) })
	}
	s.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(s, variant) })
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

// TitleIcon implements workspace.FileBackedDockable
func (s *Sheet) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  library.FileInfoFor(s.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (s *Sheet) Title() string {
	return fs.BaseName(s.path)
}

func (s *Sheet) String() string {
	return s.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (s *Sheet) Tooltip() string {
	return s.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (s *Sheet) BackingFilePath() string {
	return s.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (s *Sheet) SetBackingFilePath(p string) {
	s.path = p
	if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
		dc.UpdateTitle(s)
	}
}

// Modified implements workspace.FileBackedDockable
func (s *Sheet) Modified() bool {
	return s.crc != s.entity.CRC64()
}

// MarkModified implements widget.ModifiableRoot.
func (s *Sheet) MarkModified(src unison.Paneler) {
	if !s.awaitingUpdate {
		s.awaitingUpdate = true
		h, v := s.scroll.Position()
		focusRefKey := s.targetMgr.CurrentFocusRef()
		s.entity.DiscardCaches()
		s.modifiedFunc()
		// TODO: This can be too slow when the lists have many rows of content, impinging upon interactive typing.
		//       Looks like most of the time is spent in updating the tables. Unfortunately, there isn't a fast way to
		//       determine that the content of a table doesn't need to be refreshed.
		skipDeepSync := false
		if !toolbox.IsNil(src) {
			_, skipDeepSync = src.AsPanel().ClientData()[SkipDeepSync]
		}
		if !skipDeepSync {
			DeepSync(s)
		}
		if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
			dc.UpdateTitle(s)
		}
		s.awaitingUpdate = false
		s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
		s.scroll.SetPosition(h, v)
	}
}

// MayAttemptClose implements unison.TabCloser
func (s *Sheet) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(s)
}

// AttemptClose implements unison.TabCloser
func (s *Sheet) AttemptClose() bool {
	if !CloseGroup(s) {
		return false
	}
	if s.Modified() {
		switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), s.Title()), "") {
		case unison.ModalResponseDiscard:
		case unison.ModalResponseOK:
			if !s.save(false) {
				return false
			}
		case unison.ModalResponseCancel:
			return false
		}
	}
	if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
		dc.Close(s)
	}
	return true
}

func (s *Sheet) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || s.needsSaveAsPrompt {
		success = SaveDockableAs(s, library.SheetExt, s.entity.Save, func(path string) {
			s.crc = s.entity.CRC64()
			s.path = path
		})
	} else {
		success = SaveDockable(s, s.entity.Save, func() { s.crc = s.entity.CRC64() })
	}
	if success {
		s.needsSaveAsPrompt = false
	}
	return success
}

func (s *Sheet) print() {
	data, err := newPageExporter(s.entity).exportAsPDFBytes()
	if err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to create PDF!"), err)
		return
	}
	dialog := PrintMgr.NewJobDialog(printing.PrinterID{}, "application/pdf", nil)
	if dialog.RunModal() {
		dialog.JobAttributes()
		go backgroundPrint(s.entity.Profile.Name, dialog.Printer(), dialog.JobAttributes(), data)
	}
}

func backgroundPrint(title string, printer *printing.Printer, jobAttributes *printing.JobAttributes, data []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := printer.Print(ctx, title, "application/pdf", bytes.NewBuffer(data), len(data), jobAttributes); err != nil {
		unison.InvokeTask(func() { unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Printing '%s' failed"), title), err) })
	}
}

func (s *Sheet) exportToPDF() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	dialog.SetInitialDirectory(filepath.Dir(s.BackingFilePath()))
	dialog.SetAllowedExtensions("pdf")
	if dialog.RunModal() {
		unison.InvokeTaskAfter(func() {
			if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "pdf", false); ok {
				if err := newPageExporter(s.entity).exportAsPDFFile(filePath); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to export as PDF!"), err)
				}
			}
		}, time.Millisecond)
	}
}

func (s *Sheet) exportToWEBP() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	dialog.SetInitialDirectory(filepath.Dir(s.BackingFilePath()))
	dialog.SetAllowedExtensions("webp")
	if dialog.RunModal() {
		unison.InvokeTaskAfter(func() {
			if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "webp", false); ok {
				if err := newPageExporter(s.entity).exportAsWEBPs(filePath); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to export as WEBP!"), err)
				}
			}
		}, time.Millisecond)
	}
}

func (s *Sheet) exportToPNG() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	dialog.SetInitialDirectory(filepath.Dir(s.BackingFilePath()))
	dialog.SetAllowedExtensions("png")
	if dialog.RunModal() {
		unison.InvokeTaskAfter(func() {
			if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "png", false); ok {
				if err := newPageExporter(s.entity).exportAsPNGs(filePath); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to export as PNG!"), err)
				}
			}
		}, time.Millisecond)
	}
}

func (s *Sheet) exportToJPEG() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	dialog.SetInitialDirectory(filepath.Dir(s.BackingFilePath()))
	dialog.SetAllowedExtensions("jpeg")
	if dialog.RunModal() {
		unison.InvokeTaskAfter(func() {
			if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "jpeg", false); ok {
				if err := newPageExporter(s.entity).exportAsJPEGs(filePath); err != nil {
					unison.ErrorDialogWithError(i18n.Text("Unable to export as JPEG!"), err)
				}
			}
		}, time.Millisecond)
	}
}

func (s *Sheet) createLists() {
	children := s.content.Children()
	if len(children) == 0 {
		return
	}
	page, ok := children[0].Self.(*Page)
	if !ok {
		return
	}
	children = page.Children()
	if len(children) < 2 {
		return
	}
	for i := len(children) - 1; i > 1; i-- {
		page.RemoveChildAtIndex(i)
	}
	// Add the various blocks, based on the layout preference.
	for _, col := range s.entity.SheetSettings.BlockLayout.ByRow() {
		rowPanel := unison.NewPanel()
		rowPanel.SetLayout(&unison.FlexLayout{
			Columns:      len(col),
			HSpacing:     1,
			HAlign:       unison.FillAlignment,
			VAlign:       unison.FillAlignment,
			EqualColumns: true,
		})
		rowPanel.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.FillAlignment,
			HGrab:  true,
		})
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutReactionsKey:
				if s.Reactions == nil {
					s.Reactions = NewReactionsPageList(s.entity)
				} else {
					s.Reactions.Sync()
				}
				rowPanel.AddChild(s.Reactions)
			case gurps.BlockLayoutConditionalModifiersKey:
				if s.ConditionalModifiers == nil {
					s.ConditionalModifiers = NewConditionalModifiersPageList(s.entity)
				} else {
					s.ConditionalModifiers.Sync()
				}
				rowPanel.AddChild(s.ConditionalModifiers)
			case gurps.BlockLayoutMeleeKey:
				if s.MeleeWeapons == nil {
					s.MeleeWeapons = NewMeleeWeaponsPageList(s.entity)
				} else {
					s.MeleeWeapons.Sync()
				}
				rowPanel.AddChild(s.MeleeWeapons)
			case gurps.BlockLayoutRangedKey:
				if s.RangedWeapons == nil {
					s.RangedWeapons = NewRangedWeaponsPageList(s.entity)
				} else {
					s.RangedWeapons.Sync()
				}
				rowPanel.AddChild(s.RangedWeapons)
			case gurps.BlockLayoutTraitsKey:
				if s.Traits == nil {
					s.Traits = NewTraitsPageList(s, s.entity)
				} else {
					s.Traits.Sync()
				}
				rowPanel.AddChild(s.Traits)
			case gurps.BlockLayoutSkillsKey:
				if s.Skills == nil {
					s.Skills = NewSkillsPageList(s, s.entity)
				} else {
					s.Skills.Sync()
				}
				rowPanel.AddChild(s.Skills)
			case gurps.BlockLayoutSpellsKey:
				if s.Spells == nil {
					s.Spells = NewSpellsPageList(s, s.entity)
				} else {
					s.Spells.Sync()
				}
				rowPanel.AddChild(s.Spells)
			case gurps.BlockLayoutEquipmentKey:
				if s.CarriedEquipment == nil {
					s.CarriedEquipment = NewCarriedEquipmentPageList(s, s.entity)
				} else {
					s.CarriedEquipment.Sync()
				}
				rowPanel.AddChild(s.CarriedEquipment)
			case gurps.BlockLayoutOtherEquipmentKey:
				if s.OtherEquipment == nil {
					s.OtherEquipment = NewOtherEquipmentPageList(s, s.entity)
				} else {
					s.OtherEquipment.Sync()
				}
				rowPanel.AddChild(s.OtherEquipment)
			case gurps.BlockLayoutNotesKey:
				if s.Notes == nil {
					s.Notes = NewNotesPageList(s, s.entity)
				} else {
					s.Notes.Sync()
				}
				rowPanel.AddChild(s.Notes)
			}
		}
		page.AddChild(rowPanel)
	}
	page.ApplyPreferredSize()
}

func (s *Sheet) canSwapDefaults(_ any) bool {
	canSwap := false
	for _, skillNode := range s.Skills.SelectedNodes(true) {
		skill := skillNode.Data()
		if skill.Type == gid.Technique {
			return false
		}
		if !skill.CanSwapDefaultsWith(skill.DefaultSkill()) && skill.BestSwappableSkill() == nil {
			return false
		}
		canSwap = true
	}
	return canSwap
}

func (s *Sheet) swapDefaults(_ any) {
	undo := &unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]{
		ID:       unison.NextUndoID(),
		EditName: i18n.Text("Swap Defaults"),
		UndoFunc: func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]) { e.BeforeData.Apply() },
		RedoFunc: func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]]) { e.AfterData.Apply() },
		AbsorbFunc: func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Skill]], other unison.Undoable) bool {
			return false
		},
		BeforeData: NewTableUndoEditData(s.Skills.Table),
	}
	for _, skillNode := range s.Skills.SelectedNodes(true) {
		skill := skillNode.Data()
		if !skill.CanSwapDefaults() {
			continue
		}
		swap := skill.DefaultSkill()
		if !skill.CanSwapDefaultsWith(swap) {
			swap = skill.BestSwappableSkill()
		}
		skill.DefaultedFrom = nil
		swap.SwapDefaults()
	}
	s.Skills.Sync()
	undo.AfterData = NewTableUndoEditData(s.Skills.Table)
	s.UndoManager().Add(undo)
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (s *Sheet) SheetSettingsUpdated(entity *gurps.Entity, blockLayout bool) {
	if s.entity == entity {
		s.MarkModified(nil)
		s.Rebuild(blockLayout)
	}
}

// Rebuild implements widget.Rebuildable.
func (s *Sheet) Rebuild(full bool) {
	h, v := s.scroll.Position()
	focusRefKey := s.targetMgr.CurrentFocusRef()
	s.entity.Recalculate()
	if full {
		reactionsSelMap := s.Reactions.RecordSelection()
		conditionalModifiersSelMap := s.ConditionalModifiers.RecordSelection()
		meleeWeaponsSelMap := s.MeleeWeapons.RecordSelection()
		rangedWeaponsSelMap := s.RangedWeapons.RecordSelection()
		traitsSelMap := s.Traits.RecordSelection()
		skillsSelMap := s.Skills.RecordSelection()
		spellsSelMap := s.Spells.RecordSelection()
		carriedEquipmentSelMap := s.CarriedEquipment.RecordSelection()
		otherEquipmentSelMap := s.OtherEquipment.RecordSelection()
		notesSelMap := s.Notes.RecordSelection()
		defer func() {
			s.Reactions.ApplySelection(reactionsSelMap)
			s.ConditionalModifiers.ApplySelection(conditionalModifiersSelMap)
			s.MeleeWeapons.ApplySelection(meleeWeaponsSelMap)
			s.RangedWeapons.ApplySelection(rangedWeaponsSelMap)
			s.Traits.ApplySelection(traitsSelMap)
			s.Skills.ApplySelection(skillsSelMap)
			s.Spells.ApplySelection(spellsSelMap)
			s.CarriedEquipment.ApplySelection(carriedEquipmentSelMap)
			s.OtherEquipment.ApplySelection(otherEquipmentSelMap)
			s.Notes.ApplySelection(notesSelMap)
		}()
		s.createLists()
	}
	DeepSync(s)
	if dc := unison.Ancestor[*unison.DockContainer](s); dc != nil {
		dc.UpdateTitle(s)
	}
	s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
	s.scroll.SetPosition(h, v)
}

func drawBandedBackground(p unison.Paneler, gc *unison.Canvas, rect unison.Rect, start, step int) {
	gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	children := p.AsPanel().Children()
	for i := start; i < len(children); i += step {
		var ink unison.Ink
		if ((i-start)/step)&1 == 1 {
			ink = unison.BandingColor
		} else {
			ink = unison.ContentColor
		}
		r := children[i].FrameRect()
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, ink.Paint(gc, r, unison.Fill))
	}
}
