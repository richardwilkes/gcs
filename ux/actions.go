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
	_ "embed"
	"fmt"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/cmdline"
	"github.com/richardwilkes/toolbox/desktop"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// WebSiteDomain holds the web site domain for GCS.
const WebSiteDomain = "gurpscharactersheet.com"

//go:embed license.md
var licenseMarkdownContent string

// These actions are registered for key bindings.
var (
	addNaturalAttacksAction        *unison.Action
	applyTemplateAction            *unison.Action
	clearPortraitAction            *unison.Action
	clearSourceAction              *unison.Action
	cloneSheetAction               *unison.Action
	closeTabAction                 *unison.Action
	colorSettingsAction            *unison.Action
	convertToContainerAction       *unison.Action
	convertToNonContainerAction    *unison.Action
	copyToSheetAction              *unison.Action
	copyToTemplateAction           *unison.Action
	decreaseEquipmentLevelAction   *unison.Action
	decreaseSkillLevelAction       *unison.Action
	decreaseTechLevelAction        *unison.Action
	decreaseUsesAction             *unison.Action
	decrementAction                *unison.Action
	defaultAttributeSettingsAction *unison.Action
	defaultBodyTypeSettingsAction  *unison.Action
	defaultSheetSettingsAction     *unison.Action
	dockUnDockAction               *unison.Action
	duplicateAction                *unison.Action
	exportAsJPEGAction             *unison.Action
	exportAsPDFAction              *unison.Action
	exportAsPNGAction              *unison.Action
	exportAsWEBPAction             *unison.Action
	exportPortraitAction           *unison.Action
	fontSettingsAction             *unison.Action
	generalSettingsAction          *unison.Action
	increaseEquipmentLevelAction   *unison.Action
	increaseSkillLevelAction       *unison.Action
	increaseTechLevelAction        *unison.Action
	increaseUsesAction             *unison.Action
	incrementAction                *unison.Action
	jumpToSearchFilterAction       *unison.Action
	menuKeySettingsAction          *unison.Action
	moveToCarriedEquipmentAction   *unison.Action
	moveToOtherEquipmentAction     *unison.Action
	// TODO: Re-enable Campaign files
	// newCampaignAction                   *unison.Action
	newCarriedEquipmentAction           *unison.Action
	newCarriedEquipmentContainerAction  *unison.Action
	newCharacterSheetAction             *unison.Action
	newCharacterTemplateAction          *unison.Action
	newLootSheetAction                  *unison.Action
	newEquipmentContainerModifierAction *unison.Action
	newEquipmentLibraryAction           *unison.Action
	newEquipmentModifierAction          *unison.Action
	newEquipmentModifiersLibraryAction  *unison.Action
	newMarkdownFileAction               *unison.Action
	newMeleeWeaponAction                *unison.Action
	newNoteAction                       *unison.Action
	newNoteContainerAction              *unison.Action
	newNotesLibraryAction               *unison.Action
	newOtherEquipmentAction             *unison.Action
	newOtherEquipmentContainerAction    *unison.Action
	newRangedWeaponAction               *unison.Action
	newRitualMagicSpellAction           *unison.Action
	newSheetFromTemplateAction          *unison.Action
	newSkillAction                      *unison.Action
	newSkillContainerAction             *unison.Action
	newSkillsLibraryAction              *unison.Action
	newSpellAction                      *unison.Action
	newSpellContainerAction             *unison.Action
	newSpellsLibraryAction              *unison.Action
	newTechniqueAction                  *unison.Action
	newTraitAction                      *unison.Action
	newTraitContainerAction             *unison.Action
	newTraitContainerModifierAction     *unison.Action
	newTraitModifierAction              *unison.Action
	newTraitModifiersLibraryAction      *unison.Action
	newTraitsLibraryAction              *unison.Action
	openAction                          *unison.Action
	openEachPageReferenceAction         *unison.Action
	openEditorAction                    *unison.Action
	openOnePageReferenceAction          *unison.Action
	pageRefMappingsAction               *unison.Action
	perSheetAttributeSettingsAction     *unison.Action
	perSheetBodyTypeSettingsAction      *unison.Action
	perSheetSettingsAction              *unison.Action
	printAction                         *unison.Action
	redoAction                          *unison.Action
	saveAction                          *unison.Action
	saveAsAction                        *unison.Action
	scale100Action                      *unison.Action
	scale200Action                      *unison.Action
	scale25Action                       *unison.Action
	scale300Action                      *unison.Action
	scale400Action                      *unison.Action
	scale500Action                      *unison.Action
	scale50Action                       *unison.Action
	scale600Action                      *unison.Action
	scale75Action                       *unison.Action
	scaleDefaultAction                  *unison.Action
	scaleDownAction                     *unison.Action
	scaleUpAction                       *unison.Action
	syncWithSourceAction                *unison.Action
	swapDefaultsAction                  *unison.Action
	toggleStateAction                   *unison.Action
	undoAction                          *unison.Action
)

// These actions aren't registered for key bindings.
var (
	checkForAppUpdatesAction *unison.Action
	licenseAction            *unison.Action
	mailingListAction        *unison.Action
	makeDonationAction       *unison.Action
	releaseNotesAction       *unison.Action
	sponsorDevelopmentAction *unison.Action
	updateAppStatusAction    *unison.Action
	webSiteAction            *unison.Action
	userGuideAction          *unison.Action
)

func registerActions() {
	// Standard actions that may be assigned a key binding
	gurps.RegisterKeyBinding("cut", unison.CutAction())
	gurps.RegisterKeyBinding("copy", unison.CopyAction())
	gurps.RegisterKeyBinding("paste", unison.PasteAction())
	gurps.RegisterKeyBinding("delete", unison.DeleteAction())
	gurps.RegisterKeyBinding("select.all", unison.SelectAllAction())

	// Actions that may be assigned a key binding
	addNaturalAttacksAction = registerKeyBindableAction("add.natural.attacks", &unison.Action{
		ID:              AddNaturalAttacksItemID,
		Title:           i18n.Text("Add Natural Attacks"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	applyTemplateAction = registerKeyBindableAction("apply.template", &unison.Action{
		ID:              ApplyTemplateItemID,
		Title:           i18n.Text("Apply Template to Character Sheet"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyA, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	cloneSheetAction = registerKeyBindableAction("clone.sheet", &unison.Action{
		ID:              CloneSheetItemID,
		Title:           i18n.Text("Clone Character Sheet & Re-Randomize Fields"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSheetFromTemplateAction = registerKeyBindableAction("new.sheet.from.template", &unison.Action{
		ID:              NewSheetFromTemplateItemID,
		Title:           i18n.Text("New Character Sheet from Template"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	exportPortraitAction = registerKeyBindableAction("export.portrait", &unison.Action{
		ID:              ExportPortraitItemID,
		Title:           i18n.Text("Export Portrait"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	clearPortraitAction = registerKeyBindableAction("clear.portrait", &unison.Action{
		ID:              ClearPortraitItemID,
		Title:           i18n.Text("Clear Portrait"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	clearSourceAction = registerKeyBindableAction("clear.source", &unison.Action{
		ID:              ClearSourceItemID,
		Title:           i18n.Text("Clear Source"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	closeTabAction = registerKeyBindableAction("close", &unison.Action{
		ID:         CloseTabID,
		Title:      i18n.Text("Close"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyW, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: func(_ *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if Workspace.Window != wnd {
					return true // not the workspace, so allow regular window close
				}
				if d := unison.Ancestor[unison.Dockable](wnd.Focus()); d != nil {
					if _, ok := d.AsPanel().Self.(unison.TabCloser); ok {
						return true
					}
				}
			}
			return false
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if Workspace.Window != wnd {
					// not the workspace, so allow regular window close
					wnd.AttemptClose()
				} else if d := unison.Ancestor[unison.Dockable](wnd.Focus()); d != nil {
					if closer, ok := d.AsPanel().Self.(unison.TabCloser); ok {
						closer.AttemptClose()
					}
				}
			}
		},
	})
	colorSettingsAction = registerKeyBindableAction("settings.colors", &unison.Action{
		ID:              ColorSettingsItemID,
		Title:           i18n.Text("Colors…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowColorSettings() },
	})
	convertToContainerAction = registerKeyBindableAction("convert.to_container", &unison.Action{
		ID:              ConvertToContainerItemID,
		Title:           i18n.Text("Convert to Container"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	convertToNonContainerAction = registerKeyBindableAction("convert.to_non_container", &unison.Action{
		ID:              ConvertToNonContainerItemID,
		Title:           i18n.Text("Convert to Non-Container"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	copyToSheetAction = registerKeyBindableAction("copy.to_sheet", &unison.Action{
		ID:              CopyToSheetItemID,
		Title:           i18n.Text("Copy to Character Sheet"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyC, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	copyToTemplateAction = registerKeyBindableAction("copy.to_template", &unison.Action{
		ID:              CopyToTemplateItemID,
		Title:           i18n.Text("Copy to Template"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	decreaseEquipmentLevelAction = registerKeyBindableAction("dec.eqp.lvl", &unison.Action{
		ID:              DecrementEquipmentLevelItemID,
		Title:           i18n.Text("Decrease Equipment Level"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	decreaseSkillLevelAction = registerKeyBindableAction("dec.sl", &unison.Action{
		ID:              DecrementSkillLevelItemID,
		Title:           i18n.Text("Decrease Skill Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyPeriod, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	decreaseTechLevelAction = registerKeyBindableAction("dec.tl", &unison.Action{
		ID:              DecrementTechLevelItemID,
		Title:           i18n.Text("Decrease Tech Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyOpenBracket, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	decreaseUsesAction = registerKeyBindableAction("dec.uses", &unison.Action{
		ID:              DecrementUsesItemID,
		Title:           i18n.Text("Decrease Uses"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyDown, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	decrementAction = registerKeyBindableAction("dec", &unison.Action{
		ID:              DecrementItemID,
		Title:           i18n.Text("Decrement"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	defaultAttributeSettingsAction = registerKeyBindableAction("settings.attributes.default", &unison.Action{
		ID:              DefaultAttributeSettingsItemID,
		Title:           i18n.Text("Default Attributes…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowAttributeSettings(nil) },
	})
	defaultBodyTypeSettingsAction = registerKeyBindableAction("settings.body_type.default", &unison.Action{
		ID:              DefaultBodyTypeSettingsItemID,
		Title:           i18n.Text("Default Body Type…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowBodySettings(&globalBodySettingsOwner{}) },
	})
	defaultSheetSettingsAction = registerKeyBindableAction("settings.sheet.default", &unison.Action{
		ID:              DefaultSheetSettingsItemID,
		Title:           i18n.Text("Default Sheet Settings…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowSheetSettings(nil) },
	})
	dockUnDockAction = registerKeyBindableAction("dock_undock", &unison.Action{
		ID:              DockUnDockItemID,
		Title:           i18n.Text("Undock From Workspace"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeySlash, Modifiers: unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	duplicateAction = registerKeyBindableAction("duplicate", &unison.Action{
		ID:              DuplicateItemID,
		Title:           i18n.Text("Duplicate"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyU, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	exportAsJPEGAction = registerKeyBindableAction("export.jpeg", &unison.Action{
		ID:              ExportAsJPEGItemID,
		Title:           i18n.Text("JPEG"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	exportAsPDFAction = registerKeyBindableAction("export.pdf", &unison.Action{
		ID:              ExportAsPDFItemID,
		Title:           i18n.Text("PDF"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	exportAsPNGAction = registerKeyBindableAction("export.png", &unison.Action{
		ID:              ExportAsPNGItemID,
		Title:           i18n.Text("PNG"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	exportAsWEBPAction = registerKeyBindableAction("export.webp", &unison.Action{
		ID:              ExportAsWEBPItemID,
		Title:           i18n.Text("WEBP"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	jumpToSearchFilterAction = registerKeyBindableAction("jump-to-search", &unison.Action{
		ID:              JumpToSearchFilterItemID,
		Title:           i18n.Text("Jump to Search/Filter Field"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyJ, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	fontSettingsAction = registerKeyBindableAction("settings.fonts", &unison.Action{
		ID:              FontSettingsItemID,
		Title:           i18n.Text("Fonts…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowFontSettings() },
	})
	generalSettingsAction = registerKeyBindableAction("settings.general", &unison.Action{
		ID:              GeneralSettingsItemID,
		Title:           i18n.Text("General Settings…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowGeneralSettings() },
	})
	increaseEquipmentLevelAction = registerKeyBindableAction("inc.eqp.lvl", &unison.Action{
		ID:              IncrementEquipmentLevelItemID,
		Title:           i18n.Text("Increase Equipment Level"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	increaseSkillLevelAction = registerKeyBindableAction("inc.sl", &unison.Action{
		ID:              IncrementSkillLevelItemID,
		Title:           i18n.Text("Increase Skill Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeySlash, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	increaseTechLevelAction = registerKeyBindableAction("inc.tl", &unison.Action{
		ID:              IncrementTechLevelItemID,
		Title:           i18n.Text("Increase Tech Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyCloseBracket, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	increaseUsesAction = registerKeyBindableAction("inc.uses", &unison.Action{
		ID:              IncrementUsesItemID,
		Title:           i18n.Text("Increase Uses"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyUp, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	incrementAction = registerKeyBindableAction("inc", &unison.Action{
		ID:              IncrementItemID,
		Title:           i18n.Text("Increment"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	menuKeySettingsAction = registerKeyBindableAction("settings.keys", &unison.Action{
		ID:              MenuKeySettingsItemID,
		Title:           i18n.Text("Menu Keys…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowMenuKeySettings() },
	})
	moveToCarriedEquipmentAction = registerKeyBindableAction("move.to.carried", &unison.Action{
		ID:              MoveToCarriedEquipmentItemID,
		Title:           i18n.Text("Move to Carried Equipment"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	moveToOtherEquipmentAction = registerKeyBindableAction("move.to.other", &unison.Action{
		ID:              MoveToOtherEquipmentItemID,
		Title:           i18n.Text("Move to Other Equipment"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newCarriedEquipmentAction = registerKeyBindableAction("new.eqp", &unison.Action{
		ID:              NewCarriedEquipmentItemID,
		Title:           i18n.Text("New Carried Equipment"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newCarriedEquipmentContainerAction = registerKeyBindableAction("new.eqp.container", &unison.Action{
		ID:              NewCarriedEquipmentContainerItemID,
		Title:           i18n.Text("New Carried Equipment Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newCharacterSheetAction = registerKeyBindableAction("new.char.sheet", &unison.Action{
		ID:         NewSheetItemID,
		Title:      i18n.Text("New Character Sheet"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			e := gurps.NewEntity()
			DisplayNewDockable(NewSheet(e.Profile.Name+gurps.SheetExt, e))
		},
	})
	newCharacterTemplateAction = registerKeyBindableAction("new.char.template", &unison.Action{
		ID:    NewTemplateItemID,
		Title: i18n.Text("New Character Template"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewTemplate("untitled"+gurps.TemplatesExt, gurps.NewTemplate()))
		},
	})
	newLootSheetAction = registerKeyBindableAction("new.loot", &unison.Action{
		ID:    NewLootSheetItemID,
		Title: i18n.Text("New Loot Sheet"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewLootSheet("untitled"+gurps.LootExt, gurps.NewLoot()))
		},
	})
	// TODO: Re-enable Campaign files
	// newCampaignAction = registerKeyBindableAction("new.campaign", &unison.Action{
	// 	ID:    NewCampaignItemID,
	// 	Title: i18n.Text("New Campaign"),
	// 	ExecuteCallback: func(_ *unison.Action, _ any) {
	// 		DisplayNewDockable(NewCampaign("untitled"+gurps.CampaignExt, gurps.NewCampaign()))
	// 	},
	// })
	newEquipmentContainerModifierAction = registerKeyBindableAction("new.eqm.container", &unison.Action{
		ID:              NewEquipmentContainerModifierItemID,
		Title:           i18n.Text("New Equipment Modifier Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyF, Modifiers: unison.ShiftModifier | unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newEquipmentLibraryAction = registerKeyBindableAction("new.eqp.lib", &unison.Action{
		ID:    NewEquipmentLibraryItemID,
		Title: i18n.Text("New Equipment Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewEquipmentTableDockable("Equipment"+gurps.EquipmentExt, nil))
		},
	})
	newEquipmentModifierAction = registerKeyBindableAction("new.eqm", &unison.Action{
		ID:              NewEquipmentModifierItemID,
		Title:           i18n.Text("New Equipment Modifier"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyF, Modifiers: unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newEquipmentModifiersLibraryAction = registerKeyBindableAction("new.eqp.lib", &unison.Action{
		ID:    NewEquipmentModifiersLibraryItemID,
		Title: i18n.Text("New Equipment Modifiers Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewEquipmentModifierTableDockable("Equipment Modifiers"+gurps.EquipmentModifiersExt, nil))
		},
	})
	newMarkdownFileAction = registerKeyBindableAction("new.markdown", &unison.Action{
		ID:    NewMarkdownFileItemID,
		Title: i18n.Text("New Markdown File"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			d, err := NewMarkdownDockableWithContent("untitled.md", "", true, true)
			if err != nil {
				Workspace.ErrorHandler(i18n.Text("Unable to create new markdown file"), err)
			} else {
				DisplayNewDockable(d)
			}
		},
	})
	newMeleeWeaponAction = registerKeyBindableAction("new.melee", &unison.Action{
		ID:              NewMeleeWeaponItemID,
		Title:           i18n.Text("New Melee Weapon"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newNoteAction = registerKeyBindableAction("new.not", &unison.Action{
		ID:              NewNoteItemID,
		Title:           i18n.Text("New Note"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newNoteContainerAction = registerKeyBindableAction("new.not.container", &unison.Action{
		ID:              NewNoteContainerItemID,
		Title:           i18n.Text("New Note Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: unison.ShiftModifier | unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newNotesLibraryAction = registerKeyBindableAction("new.not.lib", &unison.Action{
		ID:    NewNotesLibraryItemID,
		Title: i18n.Text("New Notes Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewNoteTableDockable("Notes"+gurps.NotesExt, nil))
		},
	})
	newOtherEquipmentAction = registerKeyBindableAction("new.eqp.other", &unison.Action{
		ID:              NewOtherEquipmentItemID,
		Title:           i18n.Text("New Other Equipment"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newOtherEquipmentContainerAction = registerKeyBindableAction("new.eqp.other.container", &unison.Action{
		ID:              NewOtherEquipmentContainerItemID,
		Title:           i18n.Text("New Other Equipment Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: unison.ShiftModifier | unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newRangedWeaponAction = registerKeyBindableAction("new.ranged", &unison.Action{
		ID:              NewRangedWeaponItemID,
		Title:           i18n.Text("New Ranged Weapon"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyR, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newRitualMagicSpellAction = registerKeyBindableAction("new.spl.ritual", &unison.Action{
		ID:              NewRitualMagicSpellItemID,
		Title:           i18n.Text("New Ritual Magic Spell"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: unison.ShiftModifier | unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSkillAction = registerKeyBindableAction("new.skl", &unison.Action{
		ID:              NewSkillItemID,
		Title:           i18n.Text("New Skill"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSkillContainerAction = registerKeyBindableAction("new.skl.container", &unison.Action{
		ID:              NewSkillContainerItemID,
		Title:           i18n.Text("New Skill Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSkillsLibraryAction = registerKeyBindableAction("new.skl.lib", &unison.Action{
		ID:    NewSkillsLibraryItemID,
		Title: i18n.Text("New Skills Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewSkillTableDockable("Skills"+gurps.SkillsExt, nil))
		},
	})
	newSpellAction = registerKeyBindableAction("new.spl", &unison.Action{
		ID:              NewSpellItemID,
		Title:           i18n.Text("New Spell"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSpellContainerAction = registerKeyBindableAction("new.spl.container", &unison.Action{
		ID:              NewSpellContainerItemID,
		Title:           i18n.Text("New Spell Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newSpellsLibraryAction = registerKeyBindableAction("new.spl.lib", &unison.Action{
		ID:    NewSpellsLibraryItemID,
		Title: i18n.Text("New Spells Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewSpellTableDockable("Spells"+gurps.SpellsExt, nil))
		},
	})
	newTechniqueAction = registerKeyBindableAction("new.skl.technique", &unison.Action{
		ID:              NewTechniqueItemID,
		Title:           i18n.Text("New Technique"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newTraitAction = registerKeyBindableAction("new.adq", &unison.Action{
		ID:              NewTraitItemID,
		Title:           i18n.Text("New Trait"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyD, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newTraitContainerAction = registerKeyBindableAction("new.adq.container", &unison.Action{
		ID:              NewTraitContainerItemID,
		Title:           i18n.Text("New Trait Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyD, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newTraitContainerModifierAction = registerKeyBindableAction("new.adm.container", &unison.Action{
		ID:              NewTraitContainerModifierItemID,
		Title:           i18n.Text("New Trait Modifier Container"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: unison.ShiftModifier | unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newTraitModifierAction = registerKeyBindableAction("new.adm", &unison.Action{
		ID:              NewTraitModifierItemID,
		Title:           i18n.Text("New Trait Modifier"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: unison.OptionModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	newTraitModifiersLibraryAction = registerKeyBindableAction("new.adm.lib", &unison.Action{
		ID:    NewTraitModifiersLibraryItemID,
		Title: i18n.Text("New Trait Modifiers Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewTraitModifierTableDockable("Trait Modifiers"+gurps.TraitModifiersExt, nil))
		},
	})
	newTraitsLibraryAction = registerKeyBindableAction("new.adq.lib", &unison.Action{
		ID:    NewTraitsLibraryItemID,
		Title: i18n.Text("New Traits Library"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewTraitTableDockable("Traits"+gurps.TraitsExt, nil))
		},
	})
	openAction = registerKeyBindableAction("open", &unison.Action{
		ID:         OpenItemID,
		Title:      i18n.Text("Open…"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyO, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			dialog := unison.NewOpenDialog()
			dialog.SetAllowsMultipleSelection(true)
			dialog.SetResolvesAliases(true)
			dialog.SetAllowedExtensions(gurps.AcceptableExtensions()...)
			dialog.SetCanChooseDirectories(false)
			dialog.SetCanChooseFiles(true)
			global := gurps.GlobalSettings()
			dialog.SetInitialDirectory(global.LastDir(gurps.DefaultLastDirKey))
			if dialog.RunModal() {
				paths := dialog.Paths()
				global.SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(paths[0]))
				OpenFiles(paths)
			}
		},
	})
	openEachPageReferenceAction = registerKeyBindableAction("pageref.open.all", &unison.Action{
		ID:              OpenEachPageReferenceItemID,
		Title:           i18n.Text("Open Each Page Reference"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyG, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	openEditorAction = registerKeyBindableAction("open.editor", &unison.Action{
		ID:              OpenEditorItemID,
		Title:           i18n.Text("Open Detail Editor"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyI, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	openOnePageReferenceAction = registerKeyBindableAction("pageref.open.first", &unison.Action{
		ID:              OpenOnePageReferenceItemID,
		Title:           i18n.Text("Open Page Reference"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyG, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	pageRefMappingsAction = registerKeyBindableAction("settings.pagerefs", &unison.Action{
		ID:              PageRefMappingsItemID,
		Title:           i18n.Text("Page Reference Mappings…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowPageRefMappings() },
	})
	perSheetAttributeSettingsAction = registerKeyBindableAction("settings.attributes.per_sheet", &unison.Action{
		ID:              PerSheetAttributeSettingsItemID,
		Title:           i18n.Text("Attributes…"),
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := ActiveSheet(); s != nil {
				ShowAttributeSettings(s)
			}
		},
	})
	perSheetBodyTypeSettingsAction = registerKeyBindableAction("settings.body_type.per_sheet", &unison.Action{
		ID:              PerSheetBodyTypeSettingsItemID,
		Title:           i18n.Text("Body Type…"),
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := ActiveSheet(); s != nil {
				ShowBodySettings(s)
			}
		},
	})
	perSheetSettingsAction = registerKeyBindableAction("settings.sheet.per_sheet", &unison.Action{
		ID:              PerSheetSettingsItemID,
		Title:           i18n.Text("Sheet Settings…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := ActiveSheet(); s != nil {
				ShowSheetSettings(s)
			}
		},
	})
	printAction = registerKeyBindableAction("print", &unison.Action{
		ID:              PrintItemID,
		Title:           i18n.Text("Print…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	redoAction = registerKeyBindableAction("redo", &unison.Action{
		ID:         RedoItemID,
		Title:      unison.CannotRedoTitle(),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyY, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: func(action *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					action.Title = mgr.RedoTitle()
					return mgr.CanRedo()
				}
			}
			action.Title = unison.CannotRedoTitle()
			return false
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					mgr.Redo()
				}
			}
		},
	})
	saveAction = registerKeyBindableAction("save", &unison.Action{
		ID:              SaveItemID,
		Title:           i18n.Text("Save"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	saveAsAction = registerKeyBindableAction("save_as", &unison.Action{
		ID:              SaveAsItemID,
		Title:           i18n.Text("Save As…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale25Action = registerKeyBindableAction("scale.25", &unison.Action{
		ID:              Scale25ItemID,
		Title:           i18n.Text("25% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyQ, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale50Action = registerKeyBindableAction("scale.50", &unison.Action{
		ID:              Scale50ItemID,
		Title:           i18n.Text("50% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyH, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale75Action = registerKeyBindableAction("scale.75", &unison.Action{
		ID:              Scale75ItemID,
		Title:           i18n.Text("75% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale100Action = registerKeyBindableAction("scale.100", &unison.Action{
		ID:              Scale100ItemID,
		Title:           i18n.Text("100% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key1, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale200Action = registerKeyBindableAction("scale.200", &unison.Action{
		ID:              Scale200ItemID,
		Title:           i18n.Text("200% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key2, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale300Action = registerKeyBindableAction("scale.300", &unison.Action{
		ID:              Scale300ItemID,
		Title:           i18n.Text("300% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key3, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale400Action = registerKeyBindableAction("scale.400", &unison.Action{
		ID:              Scale400ItemID,
		Title:           i18n.Text("400% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key4, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale500Action = registerKeyBindableAction("scale.500", &unison.Action{
		ID:              Scale500ItemID,
		Title:           i18n.Text("500% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key5, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scale600Action = registerKeyBindableAction("scale.600", &unison.Action{
		ID:              Scale600ItemID,
		Title:           i18n.Text("600% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key6, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scaleDefaultAction = registerKeyBindableAction("scale.default", &unison.Action{
		ID:              ScaleDefaultItemID,
		Title:           i18n.Text("Default Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key0, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scaleDownAction = registerKeyBindableAction("scale.down", &unison.Action{
		ID:              ScaleDownItemID,
		Title:           i18n.Text("Scale Down"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	scaleUpAction = registerKeyBindableAction("scale.up", &unison.Action{
		ID:              ScaleUpItemID,
		Title:           i18n.Text("Scale Up"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	syncWithSourceAction = registerKeyBindableAction("clear.sync", &unison.Action{
		ID:              SyncWithSourceItemID,
		Title:           i18n.Text("Sync with Source"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	swapDefaultsAction = registerKeyBindableAction("swap.defaults", &unison.Action{
		ID:              SwapDefaultsItemID,
		Title:           i18n.Text("Swap Defaults"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyX, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	toggleStateAction = registerKeyBindableAction("toggle", &unison.Action{
		ID:              ToggleStateItemID,
		Title:           i18n.Text("Toggle State"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyApostrophe, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
	undoAction = registerKeyBindableAction("undo", &unison.Action{
		ID:         UndoItemID,
		Title:      unison.CannotUndoTitle(),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyZ, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: func(action *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					action.Title = mgr.UndoTitle()
					return mgr.CanUndo()
				}
			}
			action.Title = unison.CannotUndoTitle()
			return false
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					mgr.Undo()
				}
			}
		},
	})

	// Actions that may not be assigned a key binding
	checkForAppUpdatesAction = &unison.Action{
		ID:    CheckForAppUpdatesItemID,
		Title: fmt.Sprintf(i18n.Text("Check for %s updates"), cmdline.AppName),
		EnabledCallback: func(_ *unison.Action, _ any) bool {
			_, releases, updating := AppUpdateResult()
			return !updating && releases == nil
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			gurps.GlobalSettings().LastSeenGCSVersion = ""
			CheckForAppUpdates()
		},
	}
	licenseAction = &unison.Action{
		ID:    LicenseItemID,
		Title: i18n.Text("License"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			ShowReadOnlyMarkdown(i18n.Text("License"), licenseMarkdownContent)
		},
	}
	mailingListAction = &unison.Action{
		ID:    MailingListItemID,
		Title: i18n.Text("Mailing Lists"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://groups.io/g/gcs")
		},
	}
	makeDonationAction = &unison.Action{
		ID:    MakeDonationItemID,
		Title: fmt.Sprintf(i18n.Text("Make a One-time Donation for %s Development"), cmdline.AppName),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://paypal.me/GURPSCharacterSheet")
		},
	}
	releaseNotesAction = &unison.Action{
		ID:    ReleaseNotesItemID,
		Title: i18n.Text("Release Notes"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://github.com/richardwilkes/gcs/releases")
		},
	}
	sponsorDevelopmentAction = &unison.Action{
		ID:    SponsorGCSDevelopmentItemID,
		Title: fmt.Sprintf(i18n.Text("Sponsor %s Development"), cmdline.AppName),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://github.com/sponsors/richardwilkes")
		},
	}
	updateAppStatusAction = &unison.Action{
		ID: UpdateAppStatusItemID,
		EnabledCallback: func(action *unison.Action, mi any) bool {
			title, releases, updating := AppUpdateResult()
			action.Title = title
			if menuItem, ok := mi.(unison.MenuItem); ok {
				menuItem.SetTitle(title)
			}
			return !updating && releases != nil
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if _, releases, updating := AppUpdateResult(); !updating && releases != nil {
				NotifyOfAppUpdate()
			}
		},
	}
	webSiteAction = &unison.Action{
		ID:    WebSiteItemID,
		Title: i18n.Text("Web Site"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			showWebPage("https://" + WebSiteDomain)
		},
	}
	userGuideAction = &unison.Action{
		ID:    UserGuideItemID,
		Title: i18n.Text("User Guide"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			HandleLink(nil, "md:Help/Interface/Overview")
		},
	}
}

func registerKeyBindableAction(key string, action *unison.Action) *unison.Action {
	gurps.RegisterKeyBinding(key, action)
	return action
}

func actionEnabledForSheet(_ *unison.Action, _ any) bool {
	return ActiveSheet() != nil
}

func showWebPage(uri string) {
	if err := desktop.Open(uri); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to open link"), err)
	}
}
