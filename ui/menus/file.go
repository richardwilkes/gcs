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

package menus

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/datafile"
	"github.com/richardwilkes/gcs/v5/model/gurps/export"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/workspace"
	"github.com/richardwilkes/gcs/v5/ui/workspace/lists"
	"github.com/richardwilkes/gcs/v5/ui/workspace/sheet"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

const outputTemplatesDirName = "Output Templates"

var (
	// NewCharacterSheet creates a new character sheet.
	NewCharacterSheet *unison.Action
	// NewCharacterTemplate creates a new character template.
	NewCharacterTemplate *unison.Action
	// NewTraitsLibrary creates a new traits library.
	NewTraitsLibrary *unison.Action
	// NewTraitModifiersLibrary creates a new trait modifiers library.
	NewTraitModifiersLibrary *unison.Action
	// NewEquipmentLibrary creates a new equipment library.
	NewEquipmentLibrary *unison.Action
	// NewEquipmentModifiersLibrary creates a new equipment modifiers library.
	NewEquipmentModifiersLibrary *unison.Action
	// NewNotesLibrary creates a new notes library.
	NewNotesLibrary *unison.Action
	// NewSkillsLibrary creates a new skills library.
	NewSkillsLibrary *unison.Action
	// NewSpellsLibrary creates a new spells library.
	NewSpellsLibrary *unison.Action
	// Open a file.
	Open *unison.Action
	// CloseTab closes a workspace tab if the workspace is foremost, or the current window if not.
	CloseTab *unison.Action
	// Save a file.
	Save *unison.Action
	// SaveAs saves to a new file.
	SaveAs *unison.Action
	// ExportAsPDF exports the content as a PDF.
	ExportAsPDF *unison.Action
	// Print the content.
	Print *unison.Action
)

func registerFileMenuActions() {
	NewCharacterSheet = &unison.Action{
		ID:         constants.NewSheetItemID,
		Title:      i18n.Text("New Character Sheet"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			entity := gurps.NewEntity(datafile.PC)
			workspace.DisplayNewDockable(nil, sheet.NewSheet(entity.Profile.Name+library.SheetExt, entity))
		},
	}
	NewCharacterTemplate = &unison.Action{
		ID:    constants.NewTemplateItemID,
		Title: i18n.Text("New Character Template"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, sheet.NewTemplate("untitled"+library.TemplatesExt, gurps.NewTemplate()))
		},
	}
	NewTraitsLibrary = &unison.Action{
		ID:    constants.NewTraitsLibraryItemID,
		Title: i18n.Text("New Traits Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, lists.NewTraitTableDockable("Traits"+library.TraitsExt, nil))
		},
	}
	NewTraitModifiersLibrary = &unison.Action{
		ID:    constants.NewTraitModifiersLibraryItemID,
		Title: i18n.Text("New Trait Modifiers Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil,
				lists.NewTraitModifierTableDockable("Trait Modifiers"+library.TraitModifiersExt, nil))
		},
	}
	NewEquipmentLibrary = &unison.Action{
		ID:    constants.NewEquipmentLibraryItemID,
		Title: i18n.Text("New Equipment Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, lists.NewEquipmentTableDockable("Equipment"+library.EquipmentExt, nil))
		},
	}
	NewEquipmentModifiersLibrary = &unison.Action{
		ID:    constants.NewEquipmentModifiersLibraryItemID,
		Title: i18n.Text("New Equipment Modifiers Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil,
				lists.NewEquipmentModifierTableDockable("Equipment Modifiers"+library.EquipmentModifiersExt, nil))
		},
	}
	NewNotesLibrary = &unison.Action{
		ID:    constants.NewNotesLibraryItemID,
		Title: i18n.Text("New Notes Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, lists.NewNoteTableDockable("Notes"+library.NotesExt, nil))
		},
	}
	NewSkillsLibrary = &unison.Action{
		ID:    constants.NewSkillsLibraryItemID,
		Title: i18n.Text("New Skills Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, lists.NewSkillTableDockable("Skills"+library.SkillsExt, nil))
		},
	}
	NewSpellsLibrary = &unison.Action{
		ID:    constants.NewSpellsLibraryItemID,
		Title: i18n.Text("New Spells Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.DisplayNewDockable(nil, lists.NewSpellTableDockable("Spells"+library.SpellsExt, nil))
		},
	}
	Open = &unison.Action{
		ID:         constants.OpenItemID,
		Title:      i18n.Text("Open…"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyO, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			dialog := unison.NewOpenDialog()
			dialog.SetAllowsMultipleSelection(true)
			dialog.SetResolvesAliases(true)
			dialog.SetAllowedExtensions(library.AcceptableExtensions()...)
			dialog.SetCanChooseDirectories(false)
			dialog.SetCanChooseFiles(true)
			global := settings.Global()
			dialog.SetInitialDirectory(global.LastDir(settings.DefaultLastDirKey))
			if dialog.RunModal() {
				paths := dialog.Paths()
				global.SetLastDir(settings.DefaultLastDirKey, filepath.Dir(paths[0]))
				workspace.OpenFiles(paths)
			}
		},
	}
	CloseTab = &unison.Action{
		ID:         constants.CloseTabID,
		Title:      i18n.Text("Close"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyW, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: func(_ *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if workspace.FromWindow(wnd) == nil {
					return true // not the workspace, so allow regular window close
				}
				if dc := unison.Ancestor[*unison.DockContainer](wnd.Focus()); dc != nil {
					if current := dc.CurrentDockable(); current != nil {
						if _, ok := current.(unison.TabCloser); ok {
							return true
						}
					}
				}
			}
			return false
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if workspace.FromWindow(wnd) == nil {
					// not the workspace, so allow regular window close
					wnd.AttemptClose()
				} else if dc := unison.Ancestor[*unison.DockContainer](wnd.Focus()); dc != nil {
					if current := dc.CurrentDockable(); current != nil {
						if closer, ok := current.(unison.TabCloser); ok {
							closer.AttemptClose()
						}
					}
				}
			}
		},
	}
	Save = &unison.Action{
		ID:              constants.SaveItemID,
		Title:           i18n.Text("Save"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	SaveAs = &unison.Action{
		ID:              constants.SaveAsItemID,
		Title:           i18n.Text("Save As…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ExportAsPDF = &unison.Action{
		ID:              constants.ExportAsPDFItemID,
		Title:           i18n.Text("PDF"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Print = &unison.Action{
		ID:              constants.PrintItemID,
		Title:           i18n.Text("Print…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}

	settings.RegisterKeyBinding("new.char.sheet", NewCharacterSheet)
	settings.RegisterKeyBinding("new.char.template", NewCharacterTemplate)
	settings.RegisterKeyBinding("new.adq.lib", NewTraitsLibrary)
	settings.RegisterKeyBinding("new.adm.lib", NewTraitModifiersLibrary)
	settings.RegisterKeyBinding("new.eqp.lib", NewEquipmentLibrary)
	settings.RegisterKeyBinding("new.eqp.lib", NewEquipmentModifiersLibrary)
	settings.RegisterKeyBinding("new.not.lib", NewNotesLibrary)
	settings.RegisterKeyBinding("new.skl.lib", NewSkillsLibrary)
	settings.RegisterKeyBinding("new.spl.lib", NewSpellsLibrary)
	settings.RegisterKeyBinding("open", Open)
	settings.RegisterKeyBinding("close", CloseTab)
	settings.RegisterKeyBinding("save", Save)
	settings.RegisterKeyBinding("save_as", SaveAs)
	settings.RegisterKeyBinding("export.pdf", ExportAsPDF)
	settings.RegisterKeyBinding("print", Print)
}

func setupFileMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.FileMenuID)
	i := insertItem(m, 0, NewCharacterSheet.NewMenuItem(f))
	i = insertItem(m, i, NewCharacterTemplate.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, NewTraitsLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewTraitModifiersLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewSkillsLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewSpellsLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewEquipmentLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewEquipmentModifiersLibrary.NewMenuItem(f))
	i = insertItem(m, i, NewNotesLibrary.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, Open.NewMenuItem(f))
	insertMenu(m, i, f.NewMenu(constants.RecentFilesMenuID, i18n.Text("Recent Files"), recentFilesUpdater))

	i = m.Item(unison.CloseItemID).Index()
	m.RemoveItem(i)
	i = insertItem(m, i, CloseTab.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, Save.NewMenuItem(f))
	i = insertItem(m, i, SaveAs.NewMenuItem(f))
	i = insertMenu(m, i, f.NewMenu(constants.ExportToMenuID, i18n.Text("Export To…"), exportToUpdater))

	i = insertSeparator(m, i)
	insertItem(m, i, Print.NewMenuItem(f))
}

func recentFilesUpdater(menu unison.Menu) {
	menu.RemoveAll()
	list := settings.Global().ListRecentFiles()
	m := make(map[string]int, len(list))
	for _, f := range list {
		title := filepath.Base(f)
		m[title] = m[title] + 1
	}
	for i, f := range list {
		title := filepath.Base(f)
		if m[title] > 1 {
			title = f
		}
		menu.InsertItem(-1, createOpenRecentFileAction(i, f, title).NewMenuItem(menu.Factory()))
	}
	if menu.Count() == 0 {
		appendDisabledMenuItem(menu, i18n.Text("No recent files available"))
	}
}

func createOpenRecentFileAction(index int, path, title string) *unison.Action {
	return &unison.Action{
		ID:    constants.RecentFieldBaseItemID + index,
		Title: title,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			workspace.OpenFile(nil, path)
		},
	}
}

func exportToUpdater(menu unison.Menu) {
	menu.RemoveAll()
	factory := menu.Factory()
	menu.InsertItem(-1, ExportAsPDF.NewMenuItem(factory))
	menu.InsertSeparator(-1, false)
	index := 0
	for _, lib := range settings.Global().Libraries().List() {
		dir := lib.Path()
		entries, err := fs.ReadDir(os.DirFS(dir), outputTemplatesDirName)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				jot.Error(errs.Wrap(err))
			}
			continue
		}
		list := make([]string, 0, len(entries))
		for _, entry := range entries {
			name := entry.Name()
			fullPath := filepath.Join(dir, outputTemplatesDirName, name)
			if !strings.HasPrefix(name, ".") && xfs.FileExists(fullPath) {
				list = append(list, fullPath)
			}
		}
		if len(list) > 0 || lib.IsMaster() {
			appendDisabledMenuItem(menu, lib.Title)
			txt.SortStringsNaturalAscending(list)
			for _, one := range list {
				menu.InsertItem(-1, createExportToTextAction(index, one).NewMenuItem(factory))
				index++
			}
		}
	}
	if menu.Count() == 2 {
		appendDisabledMenuItem(menu, i18n.Text("No export templates available"))
	}
}

func createExportToTextAction(index int, path string) *unison.Action {
	return &unison.Action{
		ID:              constants.ExportToTextBaseItemID + index,
		Title:           "    " + xfs.TrimExtension(filepath.Base(path)),
		EnabledCallback: enabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := sheet.ActiveSheet(); s != nil {
				dialog := unison.NewSaveDialog()
				ext := filepath.Ext(path)
				dialog.SetInitialDirectory(filepath.Dir(path))
				dialog.SetAllowedExtensions(ext)
				if dialog.RunModal() {
					if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), ext, false); ok {
						if err := export.LegacyExport(s.Entity(), path, filePath); err != nil {
							unison.ErrorDialogWithError(i18n.Text("Export failed"), err)
						}
					}
				}
			}
		},
	}
}

func appendDisabledMenuItem(menu unison.Menu, title string) {
	item := menu.Factory().NewItem(0, title, unison.KeyBinding{}, func(_ unison.MenuItem) bool { return false }, nil)
	menu.InsertItem(-1, item)
}
