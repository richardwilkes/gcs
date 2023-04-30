/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	xfs "github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/slices"
)

// Menu, Item & Action IDs
const (
	NewSheetItemID = unison.UserBaseID + iota
	NewTemplateItemID
	NewTraitsLibraryItemID
	NewTraitModifiersLibraryItemID
	NewEquipmentLibraryItemID
	NewEquipmentModifiersLibraryItemID
	NewNotesLibraryItemID
	NewSkillsLibraryItemID
	NewSpellsLibraryItemID
	NewMarkdownFileItemID
	OpenItemID
	CloseTabID
	RecentFilesMenuID
	SaveItemID
	SaveAsItemID
	ExportToMenuID
	ExportAsPDFItemID
	ExportAsWEBPItemID
	ExportAsPNGItemID
	ExportAsJPEGItemID
	PrintItemID
	UndoItemID
	RedoItemID
	DuplicateItemID
	ClearPortraitItemID
	ConvertToContainerItemID
	ConvertToNonContainerItemID
	ToggleStateItemID
	IncrementItemID
	DecrementItemID
	IncrementUsesItemID
	DecrementUsesItemID
	IncrementSkillLevelItemID
	DecrementSkillLevelItemID
	IncrementTechLevelItemID
	DecrementTechLevelItemID
	SwapDefaultsItemID
	ItemMenuID
	AddNaturalAttacksItemID
	OpenEditorItemID
	CopyToSheetItemID
	CopyToTemplateItemID
	ApplyTemplateItemID
	OpenOnePageReferenceItemID
	OpenEachPageReferenceItemID
	SettingsMenuID
	PerSheetSettingsItemID
	PerSheetAttributeSettingsItemID
	PerSheetBodyTypeSettingsItemID
	DefaultSheetSettingsItemID
	DefaultAttributeSettingsItemID
	DefaultBodyTypeSettingsItemID
	GeneralSettingsItemID
	PageRefMappingsItemID
	ColorSettingsItemID
	FontSettingsItemID
	MenuKeySettingsItemID
	SponsorGCSDevelopmentItemID
	MakeDonationItemID
	UpdateAppStatusItemID
	CheckForAppUpdatesItemID
	ReleaseNotesItemID
	LicenseItemID
	WebSiteItemID
	MailingListItemID
	UserGuideItemID
	ViewMenuID
	ScaleDefaultItemID
	ScaleUpItemID
	ScaleDownItemID
	Scale25ItemID
	Scale50ItemID
	Scale75ItemID
	Scale100ItemID
	Scale200ItemID
	Scale300ItemID
	Scale400ItemID
	Scale500ItemID
	Scale600ItemID
	DockUnDockItemID

	FirstNonContainerMarker // Keep this block grouped together
	NewCarriedEquipmentItemID
	NewEquipmentModifierItemID
	NewNoteItemID
	NewOtherEquipmentItemID
	NewSkillItemID
	NewSpellItemID
	NewTraitItemID
	NewTraitModifierItemID
	LastNonContainerMarker

	FirstContainerMarker // Keep this block grouped together
	NewCarriedEquipmentContainerItemID
	NewEquipmentContainerModifierItemID
	NewNoteContainerItemID
	NewOtherEquipmentContainerItemID
	NewSkillContainerItemID
	NewSpellContainerItemID
	NewTraitContainerItemID
	NewTraitContainerModifierItemID
	LastContainerMarker

	FirstAlternateNonContainerMarker // Keep this block grouped together
	NewRitualMagicSpellItemID
	NewTechniqueItemID
	LastAlternateNonContainerMarker

	NewMeleeWeaponItemID
	NewRangedWeaponItemID

	OpenInWindowMenuID

	LibraryBaseItemID      = OpenInWindowMenuID + int(gurps.LastDockableGroup) + 2
	RecentFieldBaseItemID  = LibraryBaseItemID + 500
	ExportToTextBaseItemID = RecentFieldBaseItemID + 500
)

var registerKeyBindingsOnce sync.Once

// ContextMenuItem holds the title and ID of a context menu item that is derived from a command ID.
type ContextMenuItem struct {
	Title string
	ID    int
}

// SetupMenuBar the menu bar for the window.
func SetupMenuBar(wnd *unison.Window) {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	gurps.GlobalSettings().KeyBindings.MakeCurrent()
	unison.DefaultMenuFactory().BarForWindow(wnd, func(bar unison.Menu) {
		unison.InsertStdMenus(bar, ShowAbout, nil, nil)
		std := bar.Item(unison.PreferencesItemID)
		if std != nil {
			std.Menu().RemoveItem(std.Index())
		}
		var s menuBarScope
		s.setupFileMenu(bar)
		s.setupEditMenu(bar)
		f := bar.Factory()
		i := s.insertMenu(bar, bar.Item(unison.EditMenuID).Index()+1, s.createItemMenu(f))
		i = s.insertMenu(bar, i, s.createSettingsMenu(f))
		s.insertMenu(bar, i, s.createViewMenu(f))
		s.setupWindowMenu(bar)
		s.setupHelpMenu(bar)
	})
}

type menuBarScope struct{} // Just here to provide some level of scoping

func (s menuBarScope) setupFileMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.FileMenuID)
	i := s.insertMenuItem(m, 0, newCharacterSheetAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newCharacterTemplateAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newMarkdownFileAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, newTraitsLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newTraitModifiersLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newSkillsLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newSpellsLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newEquipmentLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newEquipmentModifiersLibraryAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, newNotesLibraryAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, openAction.NewMenuItem(f))
	s.insertMenu(m, i, f.NewMenu(RecentFilesMenuID, i18n.Text("Recent Files"), s.recentFilesUpdater))

	i = m.Item(unison.CloseItemID).Index()
	m.RemoveItem(i)
	i = s.insertMenuItem(m, i, closeTabAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, saveAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, saveAsAction.NewMenuItem(f))
	i = s.insertMenu(m, i, f.NewMenu(ExportToMenuID, i18n.Text("Export To…"), s.exportToUpdater))

	i = s.insertMenuSeparator(m, i)
	s.insertMenuItem(m, i, printAction.NewMenuItem(f))
}

func (s menuBarScope) setupEditMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.EditMenuID)

	i := s.insertMenuItem(m, 0, undoAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, redoAction.NewMenuItem(f))
	s.insertMenuSeparator(m, i)

	deleteIndex := m.Item(unison.DeleteItemID).Index()
	m.InsertItem(deleteIndex+1, clearPortraitAction.NewMenuItem(f))
	m.InsertItem(deleteIndex, duplicateAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, m.Item(unison.SelectAllItemID).Index()+1)
	i = s.insertMenuItem(m, i, openEditorAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, copyToSheetAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, copyToTemplateAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, applyTemplateAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, incrementAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, decrementAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, increaseUsesAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, decreaseUsesAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, increaseSkillLevelAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, decreaseSkillLevelAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, increaseTechLevelAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, decreaseTechLevelAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, toggleStateAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, swapDefaultsAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, convertToContainerAction.NewMenuItem(f))
	s.insertMenuItem(m, i, convertToNonContainerAction.NewMenuItem(f))
}

func (s menuBarScope) createItemMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(ItemMenuID, i18n.Text("Item"), nil)

	m.InsertItem(-1, newTraitAction.NewMenuItem(f))
	m.InsertItem(-1, newTraitContainerAction.NewMenuItem(f))
	m.InsertItem(-1, newTraitModifierAction.NewMenuItem(f))
	m.InsertItem(-1, newTraitContainerModifierAction.NewMenuItem(f))
	m.InsertItem(-1, addNaturalAttacksAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, newSkillAction.NewMenuItem(f))
	m.InsertItem(-1, newSkillContainerAction.NewMenuItem(f))
	m.InsertItem(-1, newTechniqueAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, newSpellAction.NewMenuItem(f))
	m.InsertItem(-1, newSpellContainerAction.NewMenuItem(f))
	m.InsertItem(-1, newRitualMagicSpellAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, newCarriedEquipmentAction.NewMenuItem(f))
	m.InsertItem(-1, newCarriedEquipmentContainerAction.NewMenuItem(f))
	m.InsertItem(-1, newOtherEquipmentAction.NewMenuItem(f))
	m.InsertItem(-1, newOtherEquipmentContainerAction.NewMenuItem(f))
	m.InsertItem(-1, newEquipmentModifierAction.NewMenuItem(f))
	m.InsertItem(-1, newEquipmentContainerModifierAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, newNoteAction.NewMenuItem(f))
	m.InsertItem(-1, newNoteContainerAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, newMeleeWeaponAction.NewMenuItem(f))
	m.InsertItem(-1, newRangedWeaponAction.NewMenuItem(f))

	m.InsertSeparator(-1, false)
	m.InsertItem(-1, openOnePageReferenceAction.NewMenuItem(f))
	m.InsertItem(-1, openEachPageReferenceAction.NewMenuItem(f))
	return m
}

func (s menuBarScope) createSettingsMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(SettingsMenuID, i18n.Text("Settings"), nil)
	m.InsertItem(-1, perSheetSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, perSheetAttributeSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, perSheetBodyTypeSettingsAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, defaultSheetSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, defaultAttributeSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, defaultBodyTypeSettingsAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, generalSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, pageRefMappingsAction.NewMenuItem(f))
	m.InsertItem(-1, colorSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, fontSettingsAction.NewMenuItem(f))
	m.InsertItem(-1, menuKeySettingsAction.NewMenuItem(f))
	s.insertMenu(m, -1, s.createOpenInWindowMenu(f))
	return m
}

func (s menuBarScope) createOpenInWindowMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(OpenInWindowMenuID, i18n.Text("Use Separate Windows For…"), nil)
	settings := gurps.GlobalSettings()
	for _, group := range gurps.AllDockableGroup {
		mi := s.createOpenInWindowAction(group).NewMenuItem(f)
		mi.SetCheckState(unison.CheckStateFromBool(slices.Contains(settings.OpenInWindow, group)))
		m.InsertItem(-1, mi)
	}
	return m
}

func (s menuBarScope) createOpenInWindowAction(group gurps.DockableGroup) *unison.Action {
	return &unison.Action{
		ID:    OpenInWindowMenuID + 1 + int(group),
		Title: group.String(),
		ExecuteCallback: func(_ *unison.Action, item any) {
			if mi, ok := item.(unison.MenuItem); ok {
				settings := gurps.GlobalSettings()
				if i := slices.Index(settings.OpenInWindow, group); i != -1 {
					settings.OpenInWindow = slices.Delete(settings.OpenInWindow, i, i+1)
					mi.SetCheckState(unison.OffCheckState)
				} else {
					settings.OpenInWindow = gurps.SanitizeDockableGroups(append(settings.OpenInWindow, group))
					mi.SetCheckState(unison.OnCheckState)
				}
			}
		},
	}
}

func (s menuBarScope) createViewMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(ViewMenuID, i18n.Text("View"), nil)
	m.InsertItem(-1, scaleDefaultAction.NewMenuItem(f))
	m.InsertItem(-1, scaleUpAction.NewMenuItem(f))
	m.InsertItem(-1, scaleDownAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, scale25Action.NewMenuItem(f))
	m.InsertItem(-1, scale50Action.NewMenuItem(f))
	m.InsertItem(-1, scale75Action.NewMenuItem(f))
	m.InsertItem(-1, scale100Action.NewMenuItem(f))
	m.InsertItem(-1, scale200Action.NewMenuItem(f))
	m.InsertItem(-1, scale300Action.NewMenuItem(f))
	m.InsertItem(-1, scale400Action.NewMenuItem(f))
	m.InsertItem(-1, scale500Action.NewMenuItem(f))
	m.InsertItem(-1, scale600Action.NewMenuItem(f))
	platformViewMenuAddition(m)
	return m
}

func (s menuBarScope) setupWindowMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.WindowMenuID)
	s.insertMenuItem(m, -1, dockUnDockAction.NewMenuItem(f))
}

func (s menuBarScope) setupHelpMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.HelpMenuID)
	m.InsertItem(-1, userGuideAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, sponsorDevelopmentAction.NewMenuItem(f))
	m.InsertItem(-1, makeDonationAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, updateAppStatusAction.NewMenuItem(f))
	m.InsertItem(-1, checkForAppUpdatesAction.NewMenuItem(f))
	m.InsertItem(-1, releaseNotesAction.NewMenuItem(f))
	m.InsertItem(-1, licenseAction.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, webSiteAction.NewMenuItem(f))
	m.InsertItem(-1, mailingListAction.NewMenuItem(f))
}

func (s menuBarScope) recentFilesUpdater(menu unison.Menu) {
	menu.RemoveAll()
	list := gurps.GlobalSettings().ListRecentFiles()
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
		menu.InsertItem(-1, s.createOpenRecentFileAction(i, f, title).NewMenuItem(menu.Factory()))
	}
	if menu.Count() == 0 {
		s.appendDisabledMenuItem(menu, i18n.Text("No recent files available"))
	}
}

func (s menuBarScope) createOpenRecentFileAction(index int, path, title string) *unison.Action {
	return &unison.Action{
		ID:              RecentFieldBaseItemID + index,
		Title:           title,
		ExecuteCallback: func(_ *unison.Action, _ any) { OpenFile(path) },
	}
}

func (s menuBarScope) exportToUpdater(menu unison.Menu) {
	const outputTemplatesDirName = "Output Templates"
	menu.RemoveAll()
	factory := menu.Factory()
	menu.InsertItem(-1, exportAsPDFAction.NewMenuItem(factory))
	menu.InsertItem(-1, exportAsWEBPAction.NewMenuItem(factory))
	menu.InsertItem(-1, exportAsPNGAction.NewMenuItem(factory))
	menu.InsertItem(-1, exportAsJPEGAction.NewMenuItem(factory))
	menu.InsertSeparator(-1, false)
	index := 0
	for _, lib := range gurps.GlobalSettings().Libraries().List() {
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
			s.appendDisabledMenuItem(menu, lib.Title)
			txt.SortStringsNaturalAscending(list)
			for _, one := range list {
				menu.InsertItem(-1, s.createExportToTextAction(index, one).NewMenuItem(factory))
				index++
			}
		}
	}
	if menu.Count() == 2 {
		s.appendDisabledMenuItem(menu, i18n.Text("No export templates available"))
	}
}

func (s menuBarScope) createExportToTextAction(index int, path string) *unison.Action {
	return &unison.Action{
		ID:              ExportToTextBaseItemID + index,
		Title:           "    " + xfs.TrimExtension(filepath.Base(path)),
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if sheet := ActiveSheet(); sheet != nil {
				dialog := unison.NewSaveDialog()
				ext := filepath.Ext(path)
				settings := gurps.GlobalSettings()
				dialog.SetInitialDirectory(settings.LastDir(gurps.DefaultLastDirKey))
				dialog.SetAllowedExtensions(ext)
				if dialog.RunModal() {
					if filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), ext, false); ok {
						settings.SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
						if err := gurps.LegacyExport(sheet.Entity(), path, filePath); err != nil {
							unison.ErrorDialogWithError(i18n.Text("Export failed"), err)
						}
					}
				}
			}
		},
	}
}

func (s menuBarScope) insertMenuSeparator(parent unison.Menu, atIndex int) int {
	parent.InsertSeparator(atIndex, false)
	return atIndex + 1
}

func (s menuBarScope) insertMenuItem(parent unison.Menu, atIndex int, item unison.MenuItem) int {
	parent.InsertItem(atIndex, item)
	return atIndex + 1
}

func (s menuBarScope) insertMenu(parent unison.Menu, atIndex int, menu unison.Menu) int {
	parent.InsertMenu(atIndex, menu)
	return atIndex + 1
}

func (s menuBarScope) appendDisabledMenuItem(menu unison.Menu, title string) {
	item := menu.Factory().NewItem(0, title, unison.KeyBinding{}, func(_ unison.MenuItem) bool { return false }, nil)
	menu.InsertItem(-1, item)
}

// AppendDefaultContextMenuItems appends the default set of context menu items for lists.
func AppendDefaultContextMenuItems(list []ContextMenuItem) []ContextMenuItem {
	return append(list,
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Open Detail Editor"), OpenEditorItemID},
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Duplicate"), DuplicateItemID},
		ContextMenuItem{i18n.Text("Delete"), unison.DeleteItemID},
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Apply Template to Character Sheet"), ApplyTemplateItemID},
		ContextMenuItem{i18n.Text("Copy to Character Sheet"), CopyToSheetItemID},
		ContextMenuItem{i18n.Text("Copy to Template"), CopyToTemplateItemID},
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Increment"), IncrementItemID},
		ContextMenuItem{i18n.Text("Decrement"), DecrementItemID},
		ContextMenuItem{i18n.Text("Increase Uses"), IncrementUsesItemID},
		ContextMenuItem{i18n.Text("Decrease Uses"), DecrementUsesItemID},
		ContextMenuItem{i18n.Text("Increase Skill Level"), IncrementSkillLevelItemID},
		ContextMenuItem{i18n.Text("Decrease Skill Level"), DecrementSkillLevelItemID},
		ContextMenuItem{i18n.Text("Increase Tech Level"), IncrementTechLevelItemID},
		ContextMenuItem{i18n.Text("Decrease Tech Level"), DecrementTechLevelItemID},
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Toggle State"), ToggleStateItemID},
		ContextMenuItem{i18n.Text("Swap Defaults"), SwapDefaultsItemID},
		ContextMenuItem{i18n.Text("Convert to Container"), ConvertToContainerItemID},
		ContextMenuItem{i18n.Text("Convert to Non-Container"), ConvertToNonContainerItemID},
		ContextMenuItem{"", -1},
		ContextMenuItem{i18n.Text("Open Page Reference"), OpenOnePageReferenceItemID},
		ContextMenuItem{i18n.Text("Open Each Page Reference"), OpenEachPageReferenceItemID},
	)
}
