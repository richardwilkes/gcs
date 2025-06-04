// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/tid"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/printing"
)

// SkipDeepSync is set on components that should not trigger a deep sync.
const SkipDeepSync = "!deepsync"

var (
	_ FileBackedDockable           = &Sheet{}
	_ unison.UndoManagerProvider   = &Sheet{}
	_ ModifiableRoot               = &Sheet{}
	_ Rebuildable                  = &Sheet{}
	_ unison.TabCloser             = &Sheet{}
	_ KeyedDockable                = &Sheet{}
	_ gurps.DataOwnerProvider      = &Sheet{}
	_ gurps.SheetSettingsResponder = &Sheet{}

	printMgr    printing.PrintManager
	lastPrinter printing.PrinterID
	dropKeys    = []string{
		equipmentDragKey,
		gurps.SkillID,
		gurps.SpellID,
		traitDragKey,
		noteDragKey,
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
	hash                 uint64
	content              *unison.Panel
	modifiedFunc         func()
	syncDisclosureFunc   func()
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
	searchTracker        *SearchTracker
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
	for _, d := range AllDockables() {
		if sheet, ok := d.(*Sheet); ok && sheet != exclude {
			sheets = append(sheets, sheet)
		}
	}
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
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:            unison.NewScrollPanel(),
		entity:            entity,
		hash:              gurps.Hash64(entity),
		scale:             gurps.GlobalSettings().General.InitialSheetUIScale,
		content:           unison.NewPanel(),
		needsSaveAsPrompt: true,
	}
	s.Self = s
	s.targetMgr = NewTargetMgr(s)
	s.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
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
	s.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
		if s.dragReroutePanel != nil {
			r := s.RectFromRoot(s.dragReroutePanel.RectToRoot(s.dragReroutePanel.ContentRect(true)))
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	s.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	var top *Page
	top, s.modifiedFunc, s.syncDisclosureFunc = createPageTopBlock(s.entity, s.targetMgr)
	s.content.AddChild(top)
	s.createLists()
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
	s.installNewItemCmdHandlers(NewTraitItemID, NewTraitContainerItemID, s.Traits)
	s.installNewItemCmdHandlers(NewSkillItemID, NewSkillContainerItemID, s.Skills)
	s.installNewItemCmdHandlers(NewTechniqueItemID, -1, s.Skills)
	s.installNewItemCmdHandlers(NewSpellItemID, NewSpellContainerItemID, s.Spells)
	s.installNewItemCmdHandlers(NewRitualMagicSpellItemID, -1, s.Spells)
	s.installNewItemCmdHandlers(NewCarriedEquipmentItemID, NewCarriedEquipmentContainerItemID, s.CarriedEquipment)
	s.installNewItemCmdHandlers(NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID, s.OtherEquipment)
	s.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, s.Notes)
	s.InstallCmdHandlers(AddNaturalAttacksItemID, unison.AlwaysEnabled, func(_ any) {
		InsertItems(s, s.Traits.Table, s.entity.TraitList, s.entity.SetTraitList,
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
	s.InstallCmdHandlers(ClearPortraitItemID, s.canClearPortrait, s.clearPortrait)
	s.InstallCmdHandlers(ExportPortraitItemID, s.canExportPortrait, s.exportPortrait)
	s.InstallCmdHandlers(CloneSheetItemID, unison.AlwaysEnabled, func(_ any) { s.cloneSheet() })
	return s
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
	data, err := s.entity.MarshalJSON()
	if err != nil {
		Workspace.ErrorHandler(unableToCloneMsg, err)
		return
	}
	entity := gurps.NewEntity()
	if err = entity.UnmarshalJSON(data); err != nil {
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

// DockKey implements KeyedDockable.
func (s *Sheet) DockKey() string {
	return filePrefix + s.path
}

func (s *Sheet) createToolbar() {
	s.toolbar = unison.NewPanel()
	s.AddChild(s.toolbar)
	s.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0,
		unison.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	s.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})

	s.toolbar.AddChild(NewDefaultInfoPop())

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Character Sheet") }
	s.toolbar.AddChild(helpButton)
	s.toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
			func() int { return s.scale },
			func(scale int) { s.scale = scale },
			nil,
			false,
			s.scroll,
		),
	)

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

	s.searchTracker = InstallSearchTracker(s.toolbar, func() {
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
	}, func(refList *[]*searchRef, text string, namesOnly bool) {
		searchSheetTable(refList, text, namesOnly, s.Reactions)
		searchSheetTable(refList, text, namesOnly, s.ConditionalModifiers)
		searchSheetTable(refList, text, namesOnly, s.MeleeWeapons)
		searchSheetTable(refList, text, namesOnly, s.RangedWeapons)
		searchSheetTable(refList, text, namesOnly, s.Traits)
		searchSheetTable(refList, text, namesOnly, s.Skills)
		searchSheetTable(refList, text, namesOnly, s.Spells)
		searchSheetTable(refList, text, namesOnly, s.CarriedEquipment)
		searchSheetTable(refList, text, namesOnly, s.OtherEquipment)
		searchSheetTable(refList, text, namesOnly, s.Notes)
	})

	s.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(s.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
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
			dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
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
	s.entity.Profile.PortraitData = data
	s.entity.Profile.PortraitImage = nil
	s.MarkForRedraw()
	s.MarkModified(s)
}

func (s *Sheet) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = s.CarriedEquipment.Table
	case gurps.SkillID:
		p = s.Skills.Table
	case gurps.SpellID:
		p = s.Spells.Table
	case traitDragKey:
		p = s.Traits.Table
	case noteDragKey:
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
		SVG:  gurps.FileInfoFor(s.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (s *Sheet) Title() string {
	return fs.BaseName(s.BackingFilePath())
}

func (s *Sheet) String() string {
	return s.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (s *Sheet) Tooltip() string {
	return s.BackingFilePath()
}

// BackingFilePath implements workspace.FileBackedDockable
func (s *Sheet) BackingFilePath() string {
	if s.needsSaveAsPrompt {
		name := strings.TrimSpace(s.entity.Profile.Name)
		if name == "" {
			name = i18n.Text("Unnamed Character")
		}
		return name + gurps.SheetExt
	}
	return s.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (s *Sheet) SetBackingFilePath(p string) {
	s.path = p
	UpdateTitleForDockable(s)
}

// Modified implements workspace.FileBackedDockable
func (s *Sheet) Modified() bool {
	return s.hash != gurps.Hash64(s.entity)
}

// MarkModified implements widget.ModifiableRoot.
func (s *Sheet) MarkModified(src unison.Paneler) {
	if !s.awaitingUpdate {
		s.awaitingUpdate = true
		h, v := s.scroll.Position()
		focusRefKey := s.targetMgr.CurrentFocusRef()
		s.entity.DiscardCaches()
		s.modifiedFunc()
		UpdateTitleForDockable(s)
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
		s.awaitingUpdate = false
		s.searchTracker.Refresh()
		s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
		s.scroll.SetPosition(h, v)
		UpdateCalculator(s)
	}
}

// MayAttemptClose implements unison.TabCloser
func (s *Sheet) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(s)
}

// AttemptClose implements unison.TabCloser
func (s *Sheet) AttemptClose() bool {
	if AttemptSaveForDockable(s) {
		return AttemptCloseForDockable(s)
	}
	return false
}

func (s *Sheet) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || s.needsSaveAsPrompt {
		success = SaveDockableAs(s, gurps.SheetExt, s.entity.Save, func(path string) {
			s.hash = gurps.Hash64(s.entity)
			s.path = path
		})
	} else {
		success = SaveDockable(s, s.entity.Save, func() { s.hash = gurps.Hash64(s.entity) })
	}
	if success {
		s.needsSaveAsPrompt = false
	}
	return success
}

func (s *Sheet) print() {
	data, err := newPageExporter(s.entity).exportAsPDFBytes()
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to create PDF!"), err)
		return
	}
	dialog := printMgr.NewJobDialog(lastPrinter, "application/pdf", nil)
	if dialog.RunModal() {
		go backgroundPrint(s.entity.Profile.Name, dialog.Printer(), dialog.JobAttributes(), data)
	}
	if p := dialog.Printer(); p != nil {
		lastPrinter = p.PrinterID
	}
}

func backgroundPrint(title string, printer *printing.Printer, jobAttributes *printing.JobAttributes, data []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := printer.Print(ctx, title, "application/pdf", bytes.NewBuffer(data), len(data), jobAttributes); err != nil {
		unison.InvokeTask(func() { Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Printing '%s' failed"), title), err) })
	}
}

func (s *Sheet) exportToPDF() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := s.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("pdf")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "pdf", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newPageExporter(s.entity).exportAsPDFFile(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as PDF!"), err)
			}
		}
	}
}

func (s *Sheet) exportToWEBP() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := s.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("webp")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "webp", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newPageExporter(s.entity).exportAsWEBPs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as WEBP!"), err)
			}
		}
	}
}

func (s *Sheet) exportToPNG() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := s.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("png")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "png", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newPageExporter(s.entity).exportAsPNGs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as PNG!"), err)
			}
		}
	}
}

func (s *Sheet) exportToJPEG() {
	s.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := s.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("jpeg")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "jpeg", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newPageExporter(s.entity).exportAsJPEGs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as JPEG!"), err)
			}
		}
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
		for _, c := range col {
			switch c {
			case gurps.BlockLayoutReactionsKey:
				if s.Reactions == nil {
					s.Reactions = NewReactionsPageList(s.entity)
				} else {
					s.Reactions.Sync()
				}
				SetDataOwnerProvider(s.Reactions.Table, s)
				if s.Reactions.Table.RootRowCount() > 0 {
					rowPanel.AddChild(s.Reactions)
				}
			case gurps.BlockLayoutConditionalModifiersKey:
				if s.ConditionalModifiers == nil {
					s.ConditionalModifiers = NewConditionalModifiersPageList(s.entity)
				} else {
					s.ConditionalModifiers.Sync()
				}
				SetDataOwnerProvider(s.ConditionalModifiers.Table, s)
				if s.ConditionalModifiers.Table.RootRowCount() > 0 {
					rowPanel.AddChild(s.ConditionalModifiers)
				}
			case gurps.BlockLayoutMeleeKey:
				if s.MeleeWeapons == nil {
					s.MeleeWeapons = NewMeleeWeaponsPageList(s.entity)
				} else {
					s.MeleeWeapons.Sync()
				}
				SetDataOwnerProvider(s.MeleeWeapons.Table, s)
				if s.MeleeWeapons.Table.RootRowCount() > 0 {
					rowPanel.AddChild(s.MeleeWeapons)
				}
			case gurps.BlockLayoutRangedKey:
				if s.RangedWeapons == nil {
					s.RangedWeapons = NewRangedWeaponsPageList(s.entity)
				} else {
					s.RangedWeapons.Sync()
				}
				SetDataOwnerProvider(s.RangedWeapons.Table, s)
				if s.RangedWeapons.Table.RootRowCount() > 0 {
					rowPanel.AddChild(s.RangedWeapons)
				}
			case gurps.BlockLayoutTraitsKey:
				if s.Traits.needReconstruction() {
					s.Traits = NewTraitsPageList(s, s.entity)
				} else {
					s.Traits.Sync()
				}
				rowPanel.AddChild(s.Traits)
			case gurps.BlockLayoutSkillsKey:
				if s.Skills.needReconstruction() {
					s.Skills = NewSkillsPageList(s, s.entity)
				} else {
					s.Skills.Sync()
				}
				rowPanel.AddChild(s.Skills)
			case gurps.BlockLayoutSpellsKey:
				if s.Spells.needReconstruction() {
					s.Spells = NewSpellsPageList(s, s.entity)
				} else {
					s.Spells.Sync()
				}
				rowPanel.AddChild(s.Spells)
			case gurps.BlockLayoutEquipmentKey:
				if s.CarriedEquipment.needReconstruction() {
					s.CarriedEquipment = NewCarriedEquipmentPageList(s, s.entity)
				} else {
					s.CarriedEquipment.Sync()
				}
				rowPanel.AddChild(s.CarriedEquipment)
			case gurps.BlockLayoutOtherEquipmentKey:
				if s.OtherEquipment.needReconstruction() {
					s.OtherEquipment = NewOtherEquipmentPageList(s, s.entity)
				} else {
					s.OtherEquipment.Sync()
				}
				rowPanel.AddChild(s.OtherEquipment)
			case gurps.BlockLayoutNotesKey:
				if s.Notes.needReconstruction() {
					s.Notes = NewNotesPageList(s, s.entity)
				} else {
					s.Notes.Sync()
				}
				rowPanel.AddChild(s.Notes)
			}
		}
		if len(rowPanel.Children()) != 0 {
			rowPanel.SetLayout(&unison.FlexLayout{
				Columns:      len(rowPanel.Children()),
				HSpacing:     1,
				HAlign:       align.Fill,
				VAlign:       align.Fill,
				EqualColumns: true,
			})
			rowPanel.SetLayoutData(&unison.FlexLayoutData{
				HAlign: align.Fill,
				HGrab:  true,
			})
			page.AddChild(rowPanel)
		}
	}
	page.ApplyPreferredSize()
}

func (s *Sheet) canSwapDefaults(_ any) bool {
	canSwap := false
	for _, skillNode := range s.Skills.SelectedNodes(true) {
		skill := skillNode.Data()
		if skill.IsTechnique() {
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

type sheetTablesUndoData struct {
	traits           *TableUndoEditData[*gurps.Trait]
	skills           *TableUndoEditData[*gurps.Skill]
	spells           *TableUndoEditData[*gurps.Spell]
	carriedEquipment *TableUndoEditData[*gurps.Equipment]
	otherEquipment   *TableUndoEditData[*gurps.Equipment]
	notes            *TableUndoEditData[*gurps.Note]
}

func newSheetTablesUndoData(sheet *Sheet) *sheetTablesUndoData {
	return &sheetTablesUndoData{
		traits:           NewTableUndoEditData(sheet.Traits.Table),
		skills:           NewTableUndoEditData(sheet.Skills.Table),
		spells:           NewTableUndoEditData(sheet.Spells.Table),
		carriedEquipment: NewTableUndoEditData(sheet.CarriedEquipment.Table),
		otherEquipment:   NewTableUndoEditData(sheet.OtherEquipment.Table),
		notes:            NewTableUndoEditData(sheet.Notes.Table),
	}
}

func (s *sheetTablesUndoData) Apply() {
	s.traits.Apply()
	s.skills.Apply()
	s.spells.Apply()
	s.carriedEquipment.Apply()
	s.otherEquipment.Apply()
	s.notes.Apply()
}

func (s *Sheet) syncWithAllSources() {
	var undo *unison.UndoEdit[*sheetTablesUndoData]
	mgr := unison.UndoManagerFor(s)
	if mgr != nil {
		undo = &unison.UndoEdit[*sheetTablesUndoData]{
			ID:         unison.NextUndoID(),
			EditName:   syncWithSourceAction.Title,
			UndoFunc:   func(e *unison.UndoEdit[*sheetTablesUndoData]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*sheetTablesUndoData]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*sheetTablesUndoData], _ unison.Undoable) bool { return false },
			BeforeData: newSheetTablesUndoData(s),
		}
	}
	s.entity.SyncWithLibrarySources()
	s.Traits.Table.SyncToModel()
	s.Skills.Table.SyncToModel()
	s.Spells.Table.SyncToModel()
	s.CarriedEquipment.Table.SyncToModel()
	s.OtherEquipment.Table.SyncToModel()
	s.Notes.Table.SyncToModel()
	if mgr != nil && undo != nil {
		undo.AfterData = newSheetTablesUndoData(s)
		mgr.Add(undo)
	}
	s.Rebuild(true)
}

// Rebuild implements widget.Rebuildable.
func (s *Sheet) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
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
	UpdateTitleForDockable(s)
	s.searchTracker.Refresh()
	s.targetMgr.ReacquireFocus(focusRefKey, s.toolbar, s.scroll.Content())
	s.scroll.SetPosition(h, v)
	UpdateCalculator(s)
}

func drawBandedBackground(p unison.Paneler, gc *unison.Canvas, rect unison.Rect, start, step int, overrideFunc func(rowIndex int, ink unison.Ink) unison.Ink) {
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

func (s *Sheet) disclosureTables() []disclosureTables {
	return []disclosureTables{
		s.Reactions,
		s.ConditionalModifiers,
		s.MeleeWeapons,
		s.RangedWeapons,
		s.Traits,
		s.Skills,
		s.Spells,
		s.CarriedEquipment,
		s.OtherEquipment,
		s.Notes,
	}
}

func (s *Sheet) toggleHierarchy() {
	tables := s.disclosureTables()
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
	s.syncDisclosureFunc()
	s.Rebuild(true)
}

func (s *Sheet) toggleNotes() {
	tables := s.disclosureTables()
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
	s.syncDisclosureFunc()
	s.Rebuild(true)
}
