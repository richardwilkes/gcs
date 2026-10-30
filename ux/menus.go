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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/export"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
)

// Menu, Item & Action IDs
const (
	NewSheetItemID = unison.UserBaseID + iota
	NewTemplateItemID
	NewLootSheetItemID
	NewCampaignItemID
	NewTraitsLibraryItemID
	NewTraitModifiersLibraryItemID
	NewEquipmentLibraryItemID
	NewEquipmentModifiersLibraryItemID
	NewNotesLibraryItemID
	NewSkillsLibraryItemID
	NewSpellsLibraryItemID
	NewMarkdownFileItemID
	NewAncestryItemID
	NewNameGeneratorItemID
	OpenItemID
	CloseTabID
	FileNewMenuID
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
	DownloadRulesFileItemID
	ExportPortraitItemID
	ClearPortraitItemID
	ClearSourceItemID
	SyncWithSourceItemID
	JumpToSearchFilterItemID
	ConvertToContainerItemID
	ConvertToNonContainerItemID
	ConvertToChoiceContainerItemID
	ConvertToGroupContainerItemID
	ToggleStateItemID
	IncrementItemID
	DecrementItemID
	IncrementUsesItemID
	DecrementUsesItemID
	ResetUsesToMaxItemID
	IncrementSkillLevelItemID
	DecrementSkillLevelItemID
	IncrementTechLevelItemID
	DecrementTechLevelItemID
	IncrementEquipmentLevelItemID
	DecrementEquipmentLevelItemID
	SwapDefaultsItemID
	MoveToOtherEquipmentItemID
	MoveToCarriedEquipmentItemID
	MoveUpItemID
	MoveDownItemID
	MoveOutOfContainerItemID
	MoveIntoContainerItemID
	InsertMenuID
	InsertTraitMenuID
	InsertSkillMenuID
	InsertSpellMenuID
	InsertEquipmentMenuID
	InsertNoteMenuID
	ItemMenuID
	MoveMenuID
	AdjustMenuID
	ConvertMenuID
	SheetMenuID
	AddNaturalAttacksItemID
	OrganizeTraitsItemID
	OpenEditorItemID
	CloneSheetItemID
	CopyToSheetItemID
	CopyToTemplateItemID
	ApplyTemplateItemID
	ApplyModifierItemID
	AddAsGrantItemID
	NewSheetFromTemplateItemID
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
	ListFilterSettingsItemID
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
	ScaleToMenuID
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
	EditSheetLayoutItemID
	DockUnDockItemID
	CalculatorItemID
	SellEquipmentItemID
	BuyEquipmentItemID
	BuySelectedEquipmentItemID
	ShoppingRecordItemID

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

	FirstGroupContainerMarker // Keep this block grouped together
	NewCarriedEquipmentGroupItemID
	NewOtherEquipmentGroupItemID
	LastGroupContainerMarker

	FirstModifierChoiceMarker // Keep this block grouped together
	NewEquipmentModifierChoiceItemID
	NewTraitModifierChoiceItemID
	LastModifierChoiceMarker

	FirstAlternateNonContainerMarker // Keep this block grouped together
	NewRitualMagicSpellItemID
	NewTechniqueItemID
	LastAlternateNonContainerMarker

	// These are only offered by templates, the only place template choices may be made. They must stay outside the
	// marker blocks above, since NewTableDockable maps an ID inside one to that block's item variant.
	NewEquipmentChoiceContainerItemID
	NewSkillChoiceContainerItemID
	NewSpellChoiceContainerItemID
	NewTraitChoiceContainerItemID

	NewMeleeWeaponItemID
	NewRangedWeaponItemID

	RecentFieldBaseItemID  = NewRangedWeaponItemID + 500
	ExportToTextBaseItemID = RecentFieldBaseItemID + 500
)

var registerKeyBindingsOnce sync.Once

// ContextMenuItem holds the title and ID of a context menu item that is derived from a command ID.
type ContextMenuItem struct {
	Title string
	ID    int
}

// SetupMenuBar sets up the menu bar for the window.
func SetupMenuBar(wnd *unison.Window) {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	gurps.GlobalSettings().KeyBindings.MakeCurrent()
	unison.DefaultMenuFactory().BarForWindow(wnd, menuBarScope{}.populate)
}

type menuBarScope struct{} // Just here to provide some level of scoping

// populate fills bar with the application's menus: the standard ones, less the Preferences item, and between Edit and
// Window the Insert, Item, Sheet, Settings and View menus.
func (s menuBarScope) populate(bar unison.Menu) {
	unison.InsertStdMenus(bar, ShowAbout, nil, nil)
	if std := bar.Item(unison.PreferencesItemID); std != nil {
		std.Menu().RemoveItem(std.Index())
	}
	s.setupFileMenu(bar)
	s.setupEditMenu(bar)
	f := bar.Factory()
	i := s.insertMenu(bar, bar.Item(unison.EditMenuID).Index()+1, s.createInsertMenu(f))
	i = s.insertMenu(bar, i, s.createItemMenu(f))
	i = s.insertMenu(bar, i, s.createSheetMenu(f))
	i = s.insertMenu(bar, i, s.createSettingsMenu(f))
	s.insertMenu(bar, i, s.createViewMenu(f))
	s.setupWindowMenu(bar)
	s.setupHelpMenu(bar)
}

func (s menuBarScope) setupFileMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.FileMenuID)
	i := s.insertMenu(m, 0, s.createFileNewMenu(f))
	i = s.insertMenuItem(m, i, openAction.NewMenuItem(f))
	s.insertMenu(m, i, f.NewMenu(RecentFilesMenuID, i18n.Text("Open Recent"), s.recentFilesUpdater))

	i = m.Item(unison.CloseItemID).Index()
	m.RemoveItem(i)
	i = s.insertMenuItem(m, i, closeTabAction.NewMenuItem(f))

	i = s.insertMenuSeparator(m, i)
	i = s.insertMenuItem(m, i, saveAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, saveAsAction.NewMenuItem(f))
	i = s.insertMenu(m, i, s.createExportMenu(f))
	s.insertMenuItem(m, i, printAction.NewMenuItem(f))
}

func (s menuBarScope) createFileNewMenu(f unison.MenuFactory) unison.Menu {
	return s.newActionMenu(f, FileNewMenuID, i18n.Text("New"),
		newCharacterSheetAction,
		newCharacterTemplateAction,
		newLootSheetAction,
		// TODO: Re-enable Campaign files
		// newCampaignAction,
		nil,
		newTraitsLibraryAction,
		newTraitModifiersLibraryAction,
		newSkillsLibraryAction,
		newSpellsLibraryAction,
		newEquipmentLibraryAction,
		newEquipmentModifiersLibraryAction,
		newNotesLibraryAction,
		nil,
		newAncestryAction,
		newNameGeneratorAction,
		newMarkdownFileAction,
	)
}

func (s menuBarScope) setupEditMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.EditMenuID)
	i := s.insertMenuItem(m, 0, undoAction.NewMenuItem(f))
	i = s.insertMenuItem(m, i, redoAction.NewMenuItem(f))
	s.insertMenuSeparator(m, i)
	m.InsertItem(m.Item(unison.DeleteItemID).Index(), duplicateAction.NewMenuItem(f))
	i = s.insertMenuSeparator(m, m.Item(unison.SelectAllItemID).Index()+1)
	s.insertMenuItem(m, i, jumpToSearchFilterAction.NewMenuItem(f))
}

func (s menuBarScope) createInsertMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(InsertMenuID, i18n.Text("Insert"), nil)
	m.InsertMenu(-1, s.newActionMenu(f, InsertTraitMenuID, i18n.Text("Trait"),
		newTraitAction,
		newTraitContainerAction,
		newTraitChoiceContainerAction,
		nil,
		newTraitModifierAction,
		newTraitContainerModifierAction,
		newTraitModifierChoiceAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, InsertSkillMenuID, i18n.Text("Skill"),
		newSkillAction,
		newSkillContainerAction,
		newSkillChoiceContainerAction,
		newTechniqueAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, InsertSpellMenuID, i18n.Text("Spell"),
		newSpellAction,
		newSpellContainerAction,
		newSpellChoiceContainerAction,
		newRitualMagicSpellAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, InsertEquipmentMenuID, i18n.Text("Equipment"),
		newCarriedEquipmentAction,
		newCarriedEquipmentContainerAction,
		newCarriedEquipmentGroupAction,
		nil,
		newOtherEquipmentAction,
		newOtherEquipmentContainerAction,
		newOtherEquipmentGroupAction,
		nil,
		newEquipmentChoiceContainerAction,
		nil,
		newEquipmentModifierAction,
		newEquipmentContainerModifierAction,
		newEquipmentModifierChoiceAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, InsertNoteMenuID, i18n.Text("Note"),
		newNoteAction,
		newNoteContainerAction,
	))
	s.appendActions(m,
		nil,
		newMeleeWeaponAction,
		newRangedWeaponAction,
	)
	return m
}

func (s menuBarScope) createItemMenu(f unison.MenuFactory) unison.Menu {
	m := s.newActionMenu(f, ItemMenuID, i18n.Text("Item"),
		openEditorAction,
		openOnePageReferenceAction,
		openEachPageReferenceAction,
		nil,
	)
	m.InsertMenu(-1, s.newActionMenu(f, MoveMenuID, i18n.Text("Move"),
		moveUpAction,
		moveDownAction,
		moveOutOfContainerAction,
		moveIntoContainerAction,
		nil,
		moveToCarriedEquipmentAction,
		moveToOtherEquipmentAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, AdjustMenuID, i18n.Text("Adjust"),
		incrementAction,
		decrementAction,
		nil,
		increaseUsesAction,
		decreaseUsesAction,
		resetUsesToMaxAction,
		nil,
		increaseSkillLevelAction,
		decreaseSkillLevelAction,
		nil,
		increaseTechLevelAction,
		decreaseTechLevelAction,
		nil,
		increaseEquipmentLevelAction,
		decreaseEquipmentLevelAction,
	))
	m.InsertMenu(-1, s.newActionMenu(f, ConvertMenuID, i18n.Text("Convert"),
		convertToContainerAction,
		convertToNonContainerAction,
		convertToChoiceContainerAction,
		convertToGroupContainerAction,
	))
	s.appendActions(m,
		toggleStateAction,
		swapDefaultsAction,
		applyModifierAction,
		addAsGrantAction,
		nil,
		copyToSheetAction,
		copyToTemplateAction,
		nil,
		syncWithSourceAction,
		clearSourceAction,
	)
	return m
}

func (s menuBarScope) createSheetMenu(f unison.MenuFactory) unison.Menu {
	return s.newActionMenu(f, SheetMenuID, i18n.Text("Sheet"),
		applyTemplateAction,
		newSheetFromTemplateAction,
		cloneSheetAction,
		nil,
		addNaturalAttacksAction,
		organizeTraitsAction,
		nil,
		buyEquipmentAction,
		buySelectedEquipmentAction,
		sellEquipmentAction,
		shoppingRecordAction,
		nil,
		clearPortraitAction,
		nil,
		perSheetSettingsAction,
		perSheetAttributeSettingsAction,
		perSheetBodyTypeSettingsAction,
		editSheetLayoutAction,
	)
}

func (s menuBarScope) createSettingsMenu(f unison.MenuFactory) unison.Menu {
	return s.newActionMenu(f, SettingsMenuID, i18n.Text("Settings"),
		generalSettingsAction,
		colorSettingsAction,
		fontSettingsAction,
		menuKeySettingsAction,
		listFilterSettingsAction,
		nil,
		defaultSheetSettingsAction,
		defaultAttributeSettingsAction,
		defaultBodyTypeSettingsAction,
		nil,
		pageRefMappingsAction,
		downloadRulesFileAction,
	)
}

func (s menuBarScope) createViewMenu(f unison.MenuFactory) unison.Menu {
	m := s.newActionMenu(f, ViewMenuID, i18n.Text("View"),
		scaleDefaultAction,
		scaleUpAction,
		scaleDownAction,
	)
	m.InsertMenu(-1, s.newActionMenu(f, ScaleToMenuID, i18n.Text("Scale To"),
		scale25Action,
		scale50Action,
		scale75Action,
		scale100Action,
		scale200Action,
		scale300Action,
		scale400Action,
		scale500Action,
		scale600Action,
	))
	s.appendActions(m, nil, calculatorAction)
	platformViewMenuAddition(m)
	return m
}

func (s menuBarScope) setupWindowMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.WindowMenuID)
	s.insertMenuItem(m, -1, dockUnDockAction.NewMenuItem(f))
}

func (s menuBarScope) setupHelpMenu(bar unison.Menu) {
	m := bar.Menu(unison.HelpMenuID)
	s.appendActions(m,
		userGuideAction,
		nil,
		sponsorDevelopmentAction,
		makeDonationAction,
		nil,
		updateAppStatusAction,
		checkForAppUpdatesAction,
		releaseNotesAction,
		licenseAction,
		nil,
		webSiteAction,
		mailingListAction,
	)
}

// newActionMenu returns a new menu holding an item for each of actions, in order, with a separator for each nil.
func (s menuBarScope) newActionMenu(f unison.MenuFactory, id int, title string, actions ...*unison.Action) unison.Menu {
	m := f.NewMenu(id, title, nil)
	s.appendActions(m, actions...)
	return m
}

// appendActions appends an item for each of actions to m, in order, with a separator for each nil.
func (s menuBarScope) appendActions(m unison.Menu, actions ...*unison.Action) {
	f := m.Factory()
	for _, action := range actions {
		if action == nil {
			m.InsertSeparator(-1, false)
		} else {
			m.InsertItem(-1, action.NewMenuItem(f))
		}
	}
}

func (s menuBarScope) recentFilesUpdater(menu unison.Menu) {
	menu.RemoveAll()
	list := gurps.GlobalSettings().ListRecentFiles()
	m := make(map[string]int, len(list))
	for _, f := range list {
		title := filepath.Base(f)
		m[title]++
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
		ExecuteCallback: func(_ *unison.Action, _ any) { OpenFile(path, gurps.PageInfo{}) },
	}
}

// createExportMenu returns the File menu's Export submenu. The image and portrait exports head it from the start, so
// that they are in the menu whenever it is looked at and not only once it has been shown, and the templates the
// libraries offer are filled in under them each time it is shown.
func (s menuBarScope) createExportMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(ExportToMenuID, i18n.Text("Export"), s.exportTemplatesUpdater)
	s.appendActions(m, exportMenuPrologue()...)
	return m
}

// exportMenuPrologue returns the actions that head the Export menu whatever the libraries hold, with a separator after
// each group.
func exportMenuPrologue() []*unison.Action {
	return []*unison.Action{
		exportAsPDFAction,
		exportAsWEBPAction,
		exportAsPNGAction,
		exportAsJPEGAction,
		nil,
		exportPortraitAction,
		nil,
	}
}

// exportTemplatesUpdater replaces whatever follows the Export menu's prologue with the templates each library offers,
// under the library's title, or a note that there are none.
func (s menuBarScope) exportTemplatesUpdater(menu unison.Menu) {
	for prologue := len(exportMenuPrologue()); menu.Count() > prologue; {
		menu.RemoveItem(menu.Count() - 1)
	}
	factory := menu.Factory()
	index := 0
	for _, lib := range gurps.GlobalSettings().Libraries.List() {
		dir := lib.Path(false)
		entries, err := fs.ReadDir(os.DirFS(dir), library.OutputTemplatesDirName)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				errs.Log(err, "dir", dir)
			}
			continue
		}
		list := make([]string, 0, len(entries))
		for _, entry := range entries {
			name := entry.Name()
			fullPath := filepath.Join(dir, library.OutputTemplatesDirName, name)
			if !strings.HasPrefix(name, ".") && xos.FileExists(fullPath) {
				list = append(list, fullPath)
			}
		}
		if len(list) > 0 || lib.IsMaster() {
			s.appendDisabledMenuItem(menu, lib.Data().Title)
			xstrings.SortStringsNaturalAscending(list)
			for _, one := range list {
				menu.InsertItem(-1, s.createExportToTextAction(index, one).NewMenuItem(factory))
				index++
			}
		}
	}
	// Count the templates that were added rather than the menu items, since the menu also holds the fixed prologue
	// above and a title for each library that contributed, none of which are export templates.
	if index == 0 {
		s.appendDisabledMenuItem(menu, i18n.Text("No export templates available"))
	}
}

func (s menuBarScope) createExportToTextAction(index int, path string) *unison.Action {
	return &unison.Action{
		ID:              ExportToTextBaseItemID + index,
		Title:           "    " + xfilepath.TrimExtension(filepath.Base(path)),
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if sheet := ActiveSheet(); sheet != nil {
				if filePath, ok := chooseFileToSave(gurps.GlobalSettings().LastDir(gurps.DefaultLastDirKey),
					xfilepath.BaseName(sheet.BackingFilePath()), filepath.Ext(path), gurps.DefaultLastDirKey); ok {
					if err := export.Export(sheet.Entity(), path, filePath); err != nil {
						Workspace.ErrorHandler(i18n.Text("Export failed"), err)
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

// contextMenuItemFor returns the context menu item that invokes action, titled with the action's own title or the
// shorter one registered for it in contextMenuShortTitles, so the two can never drift apart. The action must already be
// registered (see registerActions), which SetupMenuBar ensures before any window, and so any context menu, exists.
func contextMenuItemFor(action *unison.Action) ContextMenuItem {
	if title, ok := contextMenuShortTitles[action.ID]; ok {
		return ContextMenuItem{Title: title, ID: action.ID}
	}
	return ContextMenuItem{Title: action.Title, ID: action.ID}
}

// AppendDefaultContextMenuItems appends the default set of context menu items for lists.
func AppendDefaultContextMenuItems(list []ContextMenuItem) []ContextMenuItem {
	return append(
		list,
		ContextMenuItem{"", -1},
		contextMenuItemFor(openEditorAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(duplicateAction),
		ContextMenuItem{unison.DeleteAction().Title, unison.DeleteItemID},
		ContextMenuItem{"", -1},
		contextMenuItemFor(moveUpAction),
		contextMenuItemFor(moveDownAction),
		contextMenuItemFor(moveOutOfContainerAction),
		contextMenuItemFor(moveIntoContainerAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(moveToCarriedEquipmentAction),
		contextMenuItemFor(moveToOtherEquipmentAction),
		contextMenuItemFor(copyToSheetAction),
		contextMenuItemFor(copyToTemplateAction),
		contextMenuItemFor(applyTemplateAction),
		contextMenuItemFor(applyModifierAction),
		contextMenuItemFor(addAsGrantAction),
		contextMenuItemFor(newSheetFromTemplateAction),
		contextMenuItemFor(cloneSheetAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(incrementAction),
		contextMenuItemFor(decrementAction),
		contextMenuItemFor(increaseUsesAction),
		contextMenuItemFor(decreaseUsesAction),
		contextMenuItemFor(resetUsesToMaxAction),
		contextMenuItemFor(increaseSkillLevelAction),
		contextMenuItemFor(decreaseSkillLevelAction),
		contextMenuItemFor(increaseTechLevelAction),
		contextMenuItemFor(decreaseTechLevelAction),
		contextMenuItemFor(increaseEquipmentLevelAction),
		contextMenuItemFor(decreaseEquipmentLevelAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(toggleStateAction),
		contextMenuItemFor(swapDefaultsAction),
		contextMenuItemFor(convertToContainerAction),
		contextMenuItemFor(convertToNonContainerAction),
		contextMenuItemFor(convertToChoiceContainerAction),
		contextMenuItemFor(convertToGroupContainerAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(openOnePageReferenceAction),
		contextMenuItemFor(openEachPageReferenceAction),
		ContextMenuItem{"", -1},
		contextMenuItemFor(syncWithSourceAction),
		contextMenuItemFor(clearSourceAction),
	)
}
