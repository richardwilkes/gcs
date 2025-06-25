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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/toolbox/xmath/rand"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var (
	_ FileBackedDockable           = &LootSheet{}
	_ unison.UndoManagerProvider   = &LootSheet{}
	_ ModifiableRoot               = &LootSheet{}
	_ Rebuildable                  = &LootSheet{}
	_ unison.TabCloser             = &LootSheet{}
	_ gurps.SheetSettingsResponder = &LootSheet{}
	_ KeyedDockable                = &LootSheet{}
)

type disclosureTables interface {
	FirstDisclosureState() (open, exists bool)
	SetDisclosureState(open bool)
	FirstNoteState() int
	ApplyNoteState(closed bool)
}

// LootSheet holds the view for a loot sheet.
type LootSheet struct {
	unison.Panel
	path              string
	targetMgr         *TargetMgr
	undoMgr           *unison.UndoManager
	toolbar           *unison.Panel
	scroll            *unison.ScrollPanel
	content           *unison.Panel
	loot              *gurps.Loot
	hash              uint64
	Equipment         *PageList[*gurps.Equipment]
	Notes             *PageList[*gurps.Note]
	dragReroutePanel  *unison.Panel
	searchTracker     *SearchTracker
	scale             int
	awaitingUpdate    bool
	needsSaveAsPrompt bool
}

// OpenLootSheets returns the currently open loot sheets.
func OpenLootSheets(exclude *LootSheet) []*LootSheet {
	var result []*LootSheet
	for _, one := range AllDockables() {
		if loot, ok := one.(*LootSheet); ok && loot != exclude {
			result = append(result, loot)
		}
	}
	return result
}

// NewLootSheetFromFile loads a loot sheet file and creates a new unison.Dockable for it.
func NewLootSheetFromFile(filePath string) (unison.Dockable, error) {
	loot, err := gurps.NewLootFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	l := NewLootSheet(filePath, loot)
	l.needsSaveAsPrompt = false
	return l, nil
}

// NewLootSheet creates a new unison.Dockable for loot sheet files.
func NewLootSheet(filePath string, loot *gurps.Loot) *LootSheet {
	l := &LootSheet{
		path:              filePath,
		undoMgr:           unison.NewUndoManager(200, func(err error) { errs.Log(err) }),
		scroll:            unison.NewScrollPanel(),
		content:           unison.NewPanel(),
		loot:              loot,
		scale:             gurps.GlobalSettings().General.InitialSheetUIScale,
		hash:              gurps.Hash64(loot),
		needsSaveAsPrompt: true,
	}
	l.Self = l
	l.targetMgr = NewTargetMgr(l)
	l.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})

	l.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
		l.RequestFocus()
		return false
	}
	l.DataDragOverCallback = func(_ unison.Point, data map[string]any) bool {
		l.dragReroutePanel = nil
		for _, key := range dropKeys {
			if _, ok := data[key]; ok {
				if l.dragReroutePanel = l.keyToPanel(key); l.dragReroutePanel != nil {
					l.dragReroutePanel.DataDragOverCallback(unison.Point{Y: 100000000}, data)
					return true
				}
				break
			}
		}
		return false
	}
	l.DataDragExitCallback = func() {
		if l.dragReroutePanel != nil {
			l.dragReroutePanel.DataDragExitCallback()
			l.dragReroutePanel = nil
		}
	}
	l.DataDragDropCallback = func(_ unison.Point, data map[string]any) {
		if l.dragReroutePanel != nil {
			l.dragReroutePanel.DataDragDropCallback(unison.Point{Y: 10000000}, data)
			l.dragReroutePanel = nil
		}
	}
	l.DrawOverCallback = func(gc *unison.Canvas, _ unison.Rect) {
		if l.dragReroutePanel != nil {
			r := l.RectFromRoot(l.dragReroutePanel.RectToRoot(l.dragReroutePanel.ContentRect(true)))
			paint := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			gc.DrawRect(r, paint)
		}
	}

	l.content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	l.content.AddChild(createLootTopBlock(l.loot, l.targetMgr))
	l.createLists()

	l.scroll.SetContent(l.content, behavior.Unmodified, behavior.Unmodified)
	l.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	l.createToolbar()
	l.AddChild(l.scroll)

	l.InstallCmdHandlers(SaveItemID, func(_ any) bool { return l.Modified() }, func(_ any) { l.save(false) })
	l.InstallCmdHandlers(SaveAsItemID, unison.AlwaysEnabled, func(_ any) { l.save(true) })
	l.installNewItemCmdHandlers(NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID, l.Equipment)
	l.installNewItemCmdHandlers(NewNoteItemID, NewNoteContainerItemID, l.Notes)
	l.InstallCmdHandlers(ExportAsPDFItemID, unison.AlwaysEnabled, func(_ any) { l.exportToPDF() })
	l.InstallCmdHandlers(ExportAsWEBPItemID, unison.AlwaysEnabled, func(_ any) { l.exportToWEBP() })
	l.InstallCmdHandlers(ExportAsPNGItemID, unison.AlwaysEnabled, func(_ any) { l.exportToPNG() })
	l.InstallCmdHandlers(ExportAsJPEGItemID, unison.AlwaysEnabled, func(_ any) { l.exportToJPEG() })
	l.InstallCmdHandlers(PrintItemID, unison.AlwaysEnabled, func(_ any) { l.print() })

	l.loot.EnsureAttachments()
	l.loot.SourceMatcher().PrepareHashes(l.loot)
	return l
}

// DockKey implements KeyedDockable.
func (l *LootSheet) DockKey() string {
	return filePrefix + l.path
}

func (l *LootSheet) createToolbar() {
	l.toolbar = unison.NewPanel()
	l.AddChild(l.toolbar)
	l.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0,
		unison.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	l.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	l.toolbar.AddChild(NewDefaultInfoPop())

	l.toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialSheetUIScale },
			func() int { return l.scale },
			func(scale int) { l.scale = scale },
			nil,
			false,
			l.scroll,
		),
	)

	hierarchyButton := unison.NewSVGButton(svg.Hierarchy)
	hierarchyButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all hierarchical rows"))
	hierarchyButton.ClickCallback = l.toggleHierarchy
	l.toolbar.AddChild(hierarchyButton)

	noteToggleButton := unison.NewSVGButton(svg.NotesToggle)
	noteToggleButton.Tooltip = newWrappedTooltip(i18n.Text("Opens/closes all embedded notes"))
	noteToggleButton.ClickCallback = l.toggleNotes
	l.toolbar.AddChild(noteToggleButton)

	syncSourceButton := unison.NewSVGButton(svg.DownToBracket)
	syncSourceButton.Tooltip = newWrappedTooltip(i18n.Text("Sync with all sources in this sheet"))
	syncSourceButton.ClickCallback = func() { l.syncWithAllSources() }
	l.toolbar.AddChild(syncSourceButton)

	treasureButton := unison.NewSVGButton(svg.MagicWand)
	treasureButton.Tooltip = newWrappedTooltip(i18n.Text("Generate a treasure horde from this loot sheet"))
	treasureButton.ClickCallback = func() { l.generateTreasure() }
	l.toolbar.AddChild(treasureButton)

	l.searchTracker = InstallSearchTracker(l.toolbar, func() {
		l.Equipment.Table.ClearSelection()
		l.Notes.Table.ClearSelection()
	}, func(refList *[]*searchRef, text string, namesOnly bool) {
		searchSheetTable(refList, text, namesOnly, l.Equipment)
		searchSheetTable(refList, text, namesOnly, l.Notes)
	})

	l.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(l.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})
}

const (
	lootPanelFieldPrefix         = "loot:"
	lootPanelNameFieldRefKey     = lootPanelFieldPrefix + "name"
	lootPanelDescFieldRefKey     = lootPanelFieldPrefix + "desc"
	lootPanelLocationFieldRefKey = lootPanelFieldPrefix + "location"
	lootPanelSessionFieldRefKey  = lootPanelFieldPrefix + "session"
)

func createLootTopBlock(loot *gurps.Loot, targetMgr *TargetMgr) *Page {
	page := NewPage(loot)
	top := unison.NewPanel()
	top.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	top.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	top.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Loot")},
		unison.NewEmptyBorder(unison.Insets{
			Top:    1,
			Left:   2,
			Bottom: 1,
			Right:  2,
		})))
	addLootTextField(top, targetMgr, i18n.Text("Name"), lootPanelNameFieldRefKey, &loot.Name)
	addLootTextField(top, targetMgr, i18n.Text("Location"), lootPanelLocationFieldRefKey, &loot.Location)
	addLootTextField(top, targetMgr, i18n.Text("Session"), lootPanelSessionFieldRefKey, &loot.Session)
	page.AddChild(top)
	return page
}

func addLootTextField(parent *unison.Panel, targetMgr *TargetMgr, title, fieldRefKey string, field *string) {
	parent.AddChild(NewPageLabel(title))
	parent.AddChild(NewStringPageField(targetMgr, fieldRefKey, title,
		func() string { return *field },
		func(s string) { *field = s }))
}

func (l *LootSheet) installNewItemCmdHandlers(itemID, containerID int, creator itemCreator) {
	variant := NoItemVariant
	if containerID == -1 {
		variant = AlternateItemVariant
	} else {
		l.InstallCmdHandlers(containerID, unison.AlwaysEnabled,
			func(_ any) { creator.CreateItem(l, ContainerItemVariant) })
	}
	l.InstallCmdHandlers(itemID, unison.AlwaysEnabled, func(_ any) { creator.CreateItem(l, variant) })
}

func (l *LootSheet) keyToPanel(key string) *unison.Panel {
	var p unison.Paneler
	switch key {
	case equipmentDragKey:
		p = l.Equipment.Table
	case noteDragKey:
		p = l.Notes.Table
	default:
		return nil
	}
	return p.AsPanel()
}

// Entity implements gurps.EntityProvider
func (l *LootSheet) Entity() *gurps.Entity {
	return nil
}

// DockableKind implements widget.DockableKind
func (l *LootSheet) DockableKind() string {
	return LootSheetDockableKind
}

// UndoManager implements undo.Provider
func (l *LootSheet) UndoManager() *unison.UndoManager {
	return l.undoMgr
}

// TitleIcon implements workspace.FileBackedDockable
func (l *LootSheet) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(l.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (l *LootSheet) Title() string {
	return fs.BaseName(l.path)
}

func (l *LootSheet) String() string {
	return l.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (l *LootSheet) Tooltip() string {
	return l.path
}

// BackingFilePath implements workspace.FileBackedDockable
func (l *LootSheet) BackingFilePath() string {
	if l.needsSaveAsPrompt {
		name := strings.TrimSpace(l.loot.Name)
		if name == "" {
			name = i18n.Text("Unnamed Loot")
		}
		return name + gurps.LootExt
	}
	return l.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (l *LootSheet) SetBackingFilePath(p string) {
	l.path = p
	UpdateTitleForDockable(l)
}

// Modified implements workspace.FileBackedDockable
func (l *LootSheet) Modified() bool {
	return l.hash != gurps.Hash64(l.loot)
}

// MarkModified implements widget.ModifiableRoot.
func (l *LootSheet) MarkModified(_ unison.Paneler) {
	if !l.awaitingUpdate {
		l.awaitingUpdate = true
		h, v := l.scroll.Position()
		focusRefKey := l.targetMgr.CurrentFocusRef()
		l.loot.ModifiedOn = jio.Now()
		DeepSync(l)
		UpdateTitleForDockable(l)
		l.awaitingUpdate = false
		l.searchTracker.Refresh()
		l.targetMgr.ReacquireFocus(focusRefKey, l.toolbar, l.scroll.Content())
		l.scroll.SetPosition(h, v)
	}
}

// MayAttemptClose implements unison.TabCloser.
func (l *LootSheet) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(l)
}

// AttemptClose implements unison.TabCloser.
func (l *LootSheet) AttemptClose() bool {
	if AttemptSaveForDockable(l) {
		return AttemptCloseForDockable(l)
	}
	return false
}

func (l *LootSheet) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || l.needsSaveAsPrompt {
		success = SaveDockableAs(l, gurps.LootExt, l.loot.Save, func(path string) {
			l.hash = gurps.Hash64(l.loot)
			l.path = path
		})
	} else {
		success = SaveDockable(l, l.loot.Save, func() { l.hash = gurps.Hash64(l.loot) })
	}
	if success {
		l.needsSaveAsPrompt = false
	}
	return success
}

type lootTablesUndoData struct {
	equipment *TableUndoEditData[*gurps.Equipment]
	notes     *TableUndoEditData[*gurps.Note]
}

func newLootTablesUndoData(l *LootSheet) *lootTablesUndoData {
	return &lootTablesUndoData{
		equipment: NewTableUndoEditData(l.Equipment.Table),
		notes:     NewTableUndoEditData(l.Notes.Table),
	}
}

func (l *lootTablesUndoData) Apply() {
	l.equipment.Apply()
	l.notes.Apply()
}

func (l *LootSheet) syncWithAllSources() {
	var undo *unison.UndoEdit[*lootTablesUndoData]
	mgr := unison.UndoManagerFor(l)
	if mgr != nil {
		undo = &unison.UndoEdit[*lootTablesUndoData]{
			ID:         unison.NextUndoID(),
			EditName:   syncWithSourceAction.Title,
			UndoFunc:   func(e *unison.UndoEdit[*lootTablesUndoData]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*lootTablesUndoData]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*lootTablesUndoData], _ unison.Undoable) bool { return false },
			BeforeData: newLootTablesUndoData(l),
		}
	}
	l.loot.SyncWithLibrarySources()
	l.Equipment.Table.SyncToModel()
	l.Notes.Table.SyncToModel()
	if mgr != nil && undo != nil {
		undo.AfterData = newLootTablesUndoData(l)
		mgr.Add(undo)
	}
	l.Rebuild(true)
}

// Rebuild implements widget.Rebuildable.
func (l *LootSheet) Rebuild(full bool) {
	gurps.DiscardGlobalResolveCache()
	l.loot.EnsureAttachments()
	l.loot.SourceMatcher().PrepareHashes(l.loot)
	h, v := l.scroll.Position()
	focusRefKey := l.targetMgr.CurrentFocusRef()
	if full {
		equipmentSelMap := l.Equipment.RecordSelection()
		notesSelMap := l.Notes.RecordSelection()
		defer func() {
			l.Equipment.ApplySelection(equipmentSelMap)
			l.Notes.ApplySelection(notesSelMap)
		}()
		l.createLists()
	}
	DeepSync(l)
	UpdateTitleForDockable(l)
	l.searchTracker.Refresh()
	l.targetMgr.ReacquireFocus(focusRefKey, l.toolbar, l.scroll.Content())
	l.scroll.SetPosition(h, v)
}

func (l *LootSheet) createLists() {
	children := l.content.Children()
	if len(children) == 0 {
		return
	}
	page, ok := children[0].Self.(*Page)
	if !ok {
		return
	}
	if children = page.Children(); len(children) == 0 {
		return
	}
	for i := len(children) - 1; i > 0; i-- {
		page.RemoveChildAtIndex(i)
	}
	if l.Equipment.needReconstruction() {
		l.Equipment = NewOtherEquipmentPageList(l, l.loot)
	} else {
		l.Equipment.Sync()
	}
	l.Equipment.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	page.AddChild(l.Equipment)
	if l.Notes.needReconstruction() {
		l.Notes = NewNotesPageList(l, l.loot)
	} else {
		l.Notes.Sync()
	}
	l.Notes.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	page.AddChild(l.Notes)
	page.ApplyPreferredSize()
}

func (l *LootSheet) print() {
	data, err := newLootPageExporter(l.loot).exportAsPDFBytes()
	if err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to create PDF!"), err)
		return
	}
	dialog := printMgr.NewJobDialog(lastPrinter, "application/pdf", nil)
	if dialog.RunModal() {
		go backgroundPrint(l.loot.Name, dialog.Printer(), dialog.JobAttributes(), data)
	}
	if p := dialog.Printer(); p != nil {
		lastPrinter = p.PrinterID
	}
}

func (l *LootSheet) exportToPDF() {
	l.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := l.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("pdf")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "pdf", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newLootPageExporter(l.loot).exportAsPDFFile(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as PDF!"), err)
			}
		}
	}
}

func (l *LootSheet) exportToWEBP() {
	l.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := l.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("webp")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "webp", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newLootPageExporter(l.loot).exportAsWEBPs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as WEBP!"), err)
			}
		}
	}
}

func (l *LootSheet) exportToPNG() {
	l.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := l.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("png")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "png", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newLootPageExporter(l.loot).exportAsPNGs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as PNG!"), err)
			}
		}
	}
}

func (l *LootSheet) exportToJPEG() {
	l.Window().ShowCursor()
	dialog := unison.NewSaveDialog()
	backingFilePath := l.BackingFilePath()
	dialog.SetInitialDirectory(filepath.Dir(backingFilePath))
	dialog.SetAllowedExtensions("jpeg")
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(backingFilePath)))
	if dialog.RunModal() {
		if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), "jpeg", false); ok {
			gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
			if err := newLootPageExporter(l.loot).exportAsJPEGs(filePath); err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to export as JPEG!"), err)
			}
		}
	}
}

// SheetSettingsUpdated implements gurps.SheetSettingsResponder.
func (l *LootSheet) SheetSettingsUpdated(_ *gurps.Entity, blockLayout bool) {
	l.MarkModified(nil)
	l.Rebuild(blockLayout)
}

func (l *LootSheet) disclosureTables() []disclosureTables {
	return []disclosureTables{
		l.Equipment,
		l.Notes,
	}
}

func (l *LootSheet) toggleHierarchy() {
	tables := l.disclosureTables()
	var open, exists bool
	for _, table := range tables {
		if open, exists = table.FirstDisclosureState(); exists {
			break
		}
	}
	open = !open
	for _, table := range tables {
		table.SetDisclosureState(open)
	}
	l.Rebuild(true)
}

func (l *LootSheet) toggleNotes() {
	tables := l.disclosureTables()
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
	l.Rebuild(true)
}

func (l *LootSheet) generateTreasure() {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	markdown := unison.NewMarkdown(false)
	markdown.SetContent(i18n.Text(`# Treasure Generation

This will generate a new Loot Sheet with items from the contents of this one.
Each top-level item in this sheet will be treated as a potential item to select from.
The quantity of that top-level item will be used to determine the likelihood of it
being selected, with larger numbers increasing the chance it is chosen.`), 400)
	content.AddChild(markdown)
	input := unison.NewPanel()
	input.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	input.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
	input.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdVSpacing * 4}))
	minValue := fxp.Thousand
	maxValue := fxp.TwoThousand
	var dialog *unison.Dialog
	var minField, maxField *DecimalField
	validateOK := func() {
		dialog.Button(unison.ModalResponseOK).SetEnabled(minValue <= maxValue && maxValue >= minValue &&
			!minField.Invalid() && !maxField.Invalid())
	}
	label := i18n.Text("Minimum Value")
	input.AddChild(NewFieldLeadingLabel(label, false))
	minField = NewDecimalField(nil, "", label,
		func() fxp.Int { return minValue },
		func(value fxp.Int) {
			minValue = value
			validateOK()
		},
		fxp.One, fxp.MaxSafeMultiply, false, false)
	input.AddChild(minField)
	label = i18n.Text("Maximum Value")
	input.AddChild(NewFieldLeadingLabel(label, false))
	maxField = NewDecimalField(nil, "", label,
		func() fxp.Int { return maxValue },
		func(value fxp.Int) {
			maxValue = value
			validateOK()
		},
		fxp.One, fxp.MaxSafeMultiply, false, false)
	input.AddChild(maxField)
	content.AddChild(input)
	icon := &unison.DrawableSVG{
		SVG:  svg.MagicWand,
		Size: unison.Size{Width: 48, Height: 48},
	}
	var err error
	if dialog, err = unison.NewDialog(icon, unison.DefaultDialogTheme.QuestionIconInk, content,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()},
		unison.FloatingWindowOption(), unison.NotResizableWindowOption()); err != nil {
		errs.Log(err)
		return
	}
	if dialog.RunModal() == unison.ModalResponseOK {
		var current fxp.Int
		r := rand.NewCryptoRand()
		m := make(map[*gurps.Equipment]int)
		for range 10 {
			choices, total, highest := pruneEquipmentList(maxValue-current, l.loot.Equipment)
			for len(choices) > 0 && current < minValue {
				found := false
				choice := fxp.Int(r.Intn(int(total)))
				for _, item := range choices {
					if item.Quantity >= choice {
						m[item]++
						current += item.ExtendedValueOfJustOne()
						found = true
						break
					}
					choice -= item.Quantity
				}
				if !found || current >= minValue {
					break
				}
				remaining := maxValue - current
				if highest > remaining {
					choices, total, highest = pruneEquipmentList(remaining, choices)
				}
			}
			if current >= minValue {
				break
			}
			current = 0
			clear(m)
		}
		if current < minValue {
			unison.ErrorDialogWithMessage(i18n.Text("Unable to generate treasure!"),
				fmt.Sprintf(i18n.Text(`The minimum value of $%s could not be reached while staying at
or under the maximum value of $%s with the available items.`),
					minValue.Comma(), maxValue.Comma()))
			return
		}
		loot := gurps.NewLoot()
		for item, quantity := range m {
			clone := item.Clone(gurps.LibraryFile{}, gurps.EntityFromNode(item), nil, false)
			clone.Quantity = fxp.From(quantity)
			loot.Equipment = append(loot.Equipment, clone)
		}
		loot.EnsureAttachments()
		sheet := NewLootSheet("untitled"+gurps.LootExt, loot)
		sheet.hash = 0 // Force it to be recognized as unsaved
		DisplayNewDockable(sheet)
	}
}

func pruneEquipmentList(remaining fxp.Int, items []*gurps.Equipment) (revisedItems []*gurps.Equipment, total, highest fxp.Int) {
	for _, item := range items {
		if item.Quantity > 0 {
			one := item.ExtendedValueOfJustOne()
			if one <= remaining {
				revisedItems = append(revisedItems, item)
				total += item.Quantity
				if highest < one {
					highest = one
				}
			}
		}
	}
	return
}
