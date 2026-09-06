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
	_ "embed"
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
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
	downloadRulesFileAction        *unison.Action
	duplicateAction                *unison.Action
	editSheetLayoutAction          *unison.Action
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
	moveDownAction                 *unison.Action
	moveIntoContainerAction        *unison.Action
	moveOutOfContainerAction       *unison.Action
	moveToCarriedEquipmentAction   *unison.Action
	moveToOtherEquipmentAction     *unison.Action
	moveUpAction                   *unison.Action
	newAncestryAction              *unison.Action
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
	newNameGeneratorAction              *unison.Action
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
	organizeTraitsAction                *unison.Action
	pageRefMappingsAction               *unison.Action
	perSheetAttributeSettingsAction     *unison.Action
	perSheetBodyTypeSettingsAction      *unison.Action
	perSheetSettingsAction              *unison.Action
	printAction                         *unison.Action
	redoAction                          *unison.Action
	resetUsesToMaxAction                *unison.Action
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
	addNaturalAttacksAction = registerFocusAction("add.natural.attacks", AddNaturalAttacksItemID,
		i18n.Text("Add Natural Attacks"), unison.KeyBinding{})
	applyTemplateAction = registerFocusAction("apply.template", ApplyTemplateItemID,
		i18n.Text("Apply Template to Character Sheet"),
		unison.KeyBinding{KeyCode: unison.KeyA, Modifiers: mod.Shift | mod.OSMenuCommand()})
	cloneSheetAction = registerFocusAction("clone.sheet", CloneSheetItemID,
		i18n.Text("Clone Character Sheet & Re-Randomize Fields"), unison.KeyBinding{})
	organizeTraitsAction = registerFocusAction("organize.traits", OrganizeTraitsItemID, i18n.Text("Organize Traits"),
		unison.KeyBinding{})
	newSheetFromTemplateAction = registerFocusAction("new.sheet.from.template", NewSheetFromTemplateItemID,
		i18n.Text("New Character Sheet from Template"),
		unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: mod.Option | mod.OSMenuCommand()})
	downloadRulesFileAction = registerKeyBindableAction("download.rules.file", &unison.Action{
		ID:              DownloadRulesFileItemID,
		Title:           i18n.Text("Download GURPS Rules Lookup File"),
		ExecuteCallback: func(_ *unison.Action, _ any) { downloadRulesLookupFile() },
	})
	exportPortraitAction = registerFocusAction("export.portrait", ExportPortraitItemID, i18n.Text("Export Portrait"),
		unison.KeyBinding{})
	clearPortraitAction = registerFocusAction("clear.portrait", ClearPortraitItemID, i18n.Text("Clear Portrait"),
		unison.KeyBinding{})
	clearSourceAction = registerFocusAction("clear.source", ClearSourceItemID, i18n.Text("Clear Source"),
		unison.KeyBinding{})
	closeTabAction = registerKeyBindableAction("close", &unison.Action{
		ID:         CloseTabID,
		Title:      i18n.Text("Close"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyW, Modifiers: mod.OSMenuCommand()},
		EnabledCallback: func(_ *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if Workspace.Window != wnd {
					return true // not the workspace, so allow regular window close
				}
				if d := wnd.Focus().Ancestor[unison.Dockable](); d != nil {
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
				} else if d := wnd.Focus().Ancestor[unison.Dockable](); d != nil {
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
	convertToContainerAction = registerFocusAction("convert.to_container", ConvertToContainerItemID,
		i18n.Text("Convert to Container"), unison.KeyBinding{})
	convertToNonContainerAction = registerFocusAction("convert.to_non_container", ConvertToNonContainerItemID,
		i18n.Text("Convert to Non-Container"), unison.KeyBinding{})
	copyToSheetAction = registerFocusAction("copy.to_sheet", CopyToSheetItemID, i18n.Text("Copy to Character Sheet"),
		unison.KeyBinding{KeyCode: unison.KeyC, Modifiers: mod.Shift | mod.OSMenuCommand()})
	copyToTemplateAction = registerFocusAction("copy.to_template", CopyToTemplateItemID, i18n.Text("Copy to Template"),
		unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: mod.Shift | mod.OSMenuCommand()})
	decreaseEquipmentLevelAction = registerFocusAction("dec.eqp.lvl", DecrementEquipmentLevelItemID,
		i18n.Text("Decrease Equipment Level"), unison.KeyBinding{})
	decreaseSkillLevelAction = registerFocusAction("dec.sl", DecrementSkillLevelItemID,
		i18n.Text("Decrease Skill Level"), unison.KeyBinding{KeyCode: unison.KeyPeriod, Modifiers: mod.OSMenuCommand()})
	decreaseTechLevelAction = registerFocusAction("dec.tl", DecrementTechLevelItemID, i18n.Text("Decrease Tech Level"),
		unison.KeyBinding{KeyCode: unison.KeyOpenBracket, Modifiers: mod.OSMenuCommand()})
	decreaseUsesAction = registerFocusAction("dec.uses", DecrementUsesItemID, i18n.Text("Decrease Uses"),
		unison.KeyBinding{KeyCode: unison.KeyDown, Modifiers: mod.OSMenuCommand()})
	decrementAction = registerFocusAction("dec", DecrementItemID, i18n.Text("Decrement"),
		unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: mod.OSMenuCommand()})
	defaultAttributeSettingsAction = registerKeyBindableAction("settings.attributes.default", &unison.Action{
		ID:              DefaultAttributeSettingsItemID,
		Title:           i18n.Text("Default Attributes…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowAttributeSettings(nil) },
	})
	defaultBodyTypeSettingsAction = registerKeyBindableAction("settings.body_type.default", &unison.Action{
		ID:              DefaultBodyTypeSettingsItemID,
		Title:           i18n.Text("Default Body Type…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowBodySettings(globalBodySettings) },
	})
	defaultSheetSettingsAction = registerKeyBindableAction("settings.sheet.default", &unison.Action{
		ID:              DefaultSheetSettingsItemID,
		Title:           i18n.Text("Default Sheet Settings…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: mod.OSMenuCommand()},
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowSheetSettings(nil) },
	})
	dockUnDockAction = registerFocusAction("dock_undock", DockUnDockItemID, i18n.Text("Undock From Workspace"),
		unison.KeyBinding{KeyCode: unison.KeySlash, Modifiers: mod.Option | mod.OSMenuCommand()})
	duplicateAction = registerFocusAction("duplicate", DuplicateItemID, i18n.Text("Duplicate"),
		unison.KeyBinding{KeyCode: unison.KeyU, Modifiers: mod.OSMenuCommand()})
	editSheetLayoutAction = registerFocusAction("edit.sheet.layout", EditSheetLayoutItemID,
		i18n.Text("Edit Sheet Layout"), unison.KeyBinding{})
	exportAsJPEGAction = registerFocusAction("export.jpeg", ExportAsJPEGItemID, i18n.Text("JPEG"), unison.KeyBinding{})
	exportAsPDFAction = registerFocusAction("export.pdf", ExportAsPDFItemID, i18n.Text("PDF"),
		unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: mod.Shift | mod.OSMenuCommand()})
	exportAsPNGAction = registerFocusAction("export.png", ExportAsPNGItemID, i18n.Text("PNG"), unison.KeyBinding{})
	exportAsWEBPAction = registerFocusAction("export.webp", ExportAsWEBPItemID, i18n.Text("WEBP"), unison.KeyBinding{})
	jumpToSearchFilterAction = registerFocusAction("jump-to-search", JumpToSearchFilterItemID,
		i18n.Text("Jump to Search/Filter Field"),
		unison.KeyBinding{KeyCode: unison.KeyJ, Modifiers: mod.OSMenuCommand()})
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
	increaseEquipmentLevelAction = registerFocusAction("inc.eqp.lvl", IncrementEquipmentLevelItemID,
		i18n.Text("Increase Equipment Level"), unison.KeyBinding{})
	increaseSkillLevelAction = registerFocusAction("inc.sl", IncrementSkillLevelItemID,
		i18n.Text("Increase Skill Level"), unison.KeyBinding{KeyCode: unison.KeySlash, Modifiers: mod.OSMenuCommand()})
	increaseTechLevelAction = registerFocusAction("inc.tl", IncrementTechLevelItemID, i18n.Text("Increase Tech Level"),
		unison.KeyBinding{KeyCode: unison.KeyCloseBracket, Modifiers: mod.OSMenuCommand()})
	increaseUsesAction = registerFocusAction("inc.uses", IncrementUsesItemID, i18n.Text("Increase Uses"),
		unison.KeyBinding{KeyCode: unison.KeyUp, Modifiers: mod.OSMenuCommand()})
	resetUsesToMaxAction = registerFocusAction("reset.uses.to.max", ResetUsesToMaxItemID,
		i18n.Text("Reset Uses to Maximum"), unison.KeyBinding{})
	incrementAction = registerFocusAction("inc", IncrementItemID, i18n.Text("Increment"),
		unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: mod.OSMenuCommand()})
	menuKeySettingsAction = registerKeyBindableAction("settings.keys", &unison.Action{
		ID:              MenuKeySettingsItemID,
		Title:           i18n.Text("Menu Keys…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowMenuKeySettings() },
	})
	// The plain command-key arrows already belong to Increase Uses and Decrease Uses, so the repositioning commands
	// take the shifted ones.
	moveUpAction = registerFocusAction("move.up", MoveUpItemID, MoveUp.Title(),
		unison.KeyBinding{KeyCode: unison.KeyUp, Modifiers: mod.Shift | mod.OSMenuCommand()})
	moveDownAction = registerFocusAction("move.down", MoveDownItemID, MoveDown.Title(),
		unison.KeyBinding{KeyCode: unison.KeyDown, Modifiers: mod.Shift | mod.OSMenuCommand()})
	moveOutOfContainerAction = registerFocusAction("move.out", MoveOutOfContainerItemID, MoveOutOfContainer.Title(),
		unison.KeyBinding{KeyCode: unison.KeyLeft, Modifiers: mod.Shift | mod.OSMenuCommand()})
	moveIntoContainerAction = registerFocusAction("move.in", MoveIntoContainerItemID, MoveIntoContainer.Title(),
		unison.KeyBinding{KeyCode: unison.KeyRight, Modifiers: mod.Shift | mod.OSMenuCommand()})
	moveToCarriedEquipmentAction = registerFocusAction("move.to.carried", MoveToCarriedEquipmentItemID,
		i18n.Text("Move to Carried Equipment"), unison.KeyBinding{})
	moveToOtherEquipmentAction = registerFocusAction("move.to.other", MoveToOtherEquipmentItemID,
		i18n.Text("Move to Other Equipment"), unison.KeyBinding{})
	newAncestryAction = registerKeyBindableAction("new.ancestry", &unison.Action{
		ID:              NewAncestryItemID,
		Title:           i18n.Text("New Ancestry"),
		ExecuteCallback: func(_ *unison.Action, _ any) { newAncestryDocument() },
	})
	newCarriedEquipmentAction = registerFocusAction("new.eqp", NewCarriedEquipmentItemID,
		i18n.Text("New Carried Equipment"), unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: mod.OSMenuCommand()})
	newCarriedEquipmentContainerAction = registerFocusAction("new.eqp.container", NewCarriedEquipmentContainerItemID,
		i18n.Text("New Carried Equipment Container"),
		unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newCharacterSheetAction = registerKeyBindableAction("new.char.sheet", &unison.Action{
		ID:         NewSheetItemID,
		Title:      i18n.Text("New Character Sheet"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: mod.OSMenuCommand()},
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
	newEquipmentContainerModifierAction = registerFocusAction("new.eqm.container", NewEquipmentContainerModifierItemID,
		i18n.Text("New Equipment Modifier Container"),
		unison.KeyBinding{KeyCode: unison.KeyF, Modifiers: mod.Shift | mod.Option | mod.OSMenuCommand()})
	newEquipmentLibraryAction = registerLibraryAction("new.eqp.lib", NewEquipmentLibraryItemID,
		i18n.Text("New Equipment Library"), "Equipment"+gurps.EquipmentExt, NewEquipmentTableDockable)
	newEquipmentModifierAction = registerFocusAction("new.eqm", NewEquipmentModifierItemID,
		i18n.Text("New Equipment Modifier"),
		unison.KeyBinding{KeyCode: unison.KeyF, Modifiers: mod.Option | mod.OSMenuCommand()})
	newEquipmentModifiersLibraryAction = registerLibraryAction("new.eqm.lib", NewEquipmentModifiersLibraryItemID,
		i18n.Text("New Equipment Modifiers Library"), "Equipment Modifiers"+gurps.EquipmentModifiersExt,
		NewEquipmentModifierTableDockable)
	newMarkdownFileAction = registerKeyBindableAction("new.markdown", &unison.Action{
		ID:    NewMarkdownFileItemID,
		Title: i18n.Text("New Markdown File"),
		ExecuteCallback: func(_ *unison.Action, _ any) {
			DisplayNewDockable(NewMarkdownDockableWithContent("untitled.md", "", true, true))
		},
	})
	newMeleeWeaponAction = registerFocusAction("new.melee", NewMeleeWeaponItemID, i18n.Text("New Melee Weapon"),
		unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newNameGeneratorAction = registerKeyBindableAction("new.names", &unison.Action{
		ID:              NewNameGeneratorItemID,
		Title:           i18n.Text("New Name Generator"),
		ExecuteCallback: func(_ *unison.Action, _ any) { newNameGeneratorDocument() },
	})
	newNoteAction = registerFocusAction("new.not", NewNoteItemID, i18n.Text("New Note"),
		unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newNoteContainerAction = registerFocusAction("new.not.container", NewNoteContainerItemID,
		i18n.Text("New Note Container"),
		unison.KeyBinding{KeyCode: unison.KeyN, Modifiers: mod.Shift | mod.Option | mod.OSMenuCommand()})
	newNotesLibraryAction = registerLibraryAction("new.not.lib", NewNotesLibraryItemID, i18n.Text("New Notes Library"),
		"Notes"+gurps.NotesExt, NewNoteTableDockable)
	newOtherEquipmentAction = registerFocusAction("new.eqp.other", NewOtherEquipmentItemID,
		i18n.Text("New Other Equipment"),
		unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: mod.Option | mod.OSMenuCommand()})
	newOtherEquipmentContainerAction = registerFocusAction("new.eqp.other.container", NewOtherEquipmentContainerItemID,
		i18n.Text("New Other Equipment Container"),
		unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: mod.Shift | mod.Option | mod.OSMenuCommand()})
	newRangedWeaponAction = registerFocusAction("new.ranged", NewRangedWeaponItemID, i18n.Text("New Ranged Weapon"),
		unison.KeyBinding{KeyCode: unison.KeyR, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newRitualMagicSpellAction = registerFocusAction("new.spl.ritual", NewRitualMagicSpellItemID,
		i18n.Text("New Ritual Magic Spell"),
		unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: mod.Shift | mod.Option | mod.OSMenuCommand()})
	newSkillAction = registerFocusAction("new.skl", NewSkillItemID, i18n.Text("New Skill"),
		unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: mod.OSMenuCommand()})
	newSkillContainerAction = registerFocusAction("new.skl.container", NewSkillContainerItemID,
		i18n.Text("New Skill Container"),
		unison.KeyBinding{KeyCode: unison.KeyK, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newSkillsLibraryAction = registerLibraryAction("new.skl.lib", NewSkillsLibraryItemID,
		i18n.Text("New Skills Library"), "Skills"+gurps.SkillsExt, NewSkillTableDockable)
	newSpellAction = registerFocusAction("new.spl", NewSpellItemID, i18n.Text("New Spell"),
		unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: mod.OSMenuCommand()})
	newSpellContainerAction = registerFocusAction("new.spl.container", NewSpellContainerItemID,
		i18n.Text("New Spell Container"),
		unison.KeyBinding{KeyCode: unison.KeyB, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newSpellsLibraryAction = registerLibraryAction("new.spl.lib", NewSpellsLibraryItemID,
		i18n.Text("New Spells Library"), "Spells"+gurps.SpellsExt, NewSpellTableDockable)
	newTechniqueAction = registerFocusAction("new.skl.technique", NewTechniqueItemID, i18n.Text("New Technique"),
		unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: mod.OSMenuCommand()})
	newTraitAction = registerFocusAction("new.adq", NewTraitItemID, i18n.Text("New Trait"),
		unison.KeyBinding{KeyCode: unison.KeyD, Modifiers: mod.OSMenuCommand()})
	newTraitContainerAction = registerFocusAction("new.adq.container", NewTraitContainerItemID,
		i18n.Text("New Trait Container"),
		unison.KeyBinding{KeyCode: unison.KeyD, Modifiers: mod.Shift | mod.OSMenuCommand()})
	newTraitContainerModifierAction = registerFocusAction("new.adm.container", NewTraitContainerModifierItemID,
		i18n.Text("New Trait Modifier Container"),
		unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: mod.Shift | mod.Option | mod.OSMenuCommand()})
	newTraitModifierAction = registerFocusAction("new.adm", NewTraitModifierItemID, i18n.Text("New Trait Modifier"),
		unison.KeyBinding{KeyCode: unison.KeyM, Modifiers: mod.Option | mod.OSMenuCommand()})
	newTraitModifiersLibraryAction = registerLibraryAction("new.adm.lib", NewTraitModifiersLibraryItemID,
		i18n.Text("New Trait Modifiers Library"), "Trait Modifiers"+gurps.TraitModifiersExt,
		NewTraitModifierTableDockable)
	newTraitsLibraryAction = registerLibraryAction("new.adq.lib", NewTraitsLibraryItemID,
		i18n.Text("New Traits Library"), "Traits"+gurps.TraitsExt, NewTraitTableDockable)
	openAction = registerKeyBindableAction("open", &unison.Action{
		ID:         OpenItemID,
		Title:      i18n.Text("Open…"),
		KeyBinding: unison.KeyBinding{KeyCode: unison.KeyO, Modifiers: mod.OSMenuCommand()},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if paths, ok := chooseFilesToOpen(gurps.DefaultLastDirKey, true, gurps.AcceptableExtensions()...); ok {
				OpenFiles(paths)
			}
		},
	})
	openEachPageReferenceAction = registerFocusAction("pageref.open.all", OpenEachPageReferenceItemID,
		i18n.Text("Open Each Page Reference"),
		unison.KeyBinding{KeyCode: unison.KeyG, Modifiers: mod.Shift | mod.OSMenuCommand()})
	openEditorAction = registerFocusAction("open.editor", OpenEditorItemID, i18n.Text("Open Detail Editor"),
		unison.KeyBinding{KeyCode: unison.KeyI, Modifiers: mod.OSMenuCommand()})
	openOnePageReferenceAction = registerFocusAction("pageref.open.first", OpenOnePageReferenceItemID,
		i18n.Text("Open Page Reference"), unison.KeyBinding{KeyCode: unison.KeyG, Modifiers: mod.OSMenuCommand()})
	pageRefMappingsAction = registerKeyBindableAction("settings.pagerefs", &unison.Action{
		ID:              PageRefMappingsItemID,
		Title:           i18n.Text("Page Reference Mappings…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { ShowPageRefMappings() },
	})
	perSheetAttributeSettingsAction = registerSheetAction("settings.attributes.per_sheet",
		PerSheetAttributeSettingsItemID, i18n.Text("Attributes…"), unison.KeyBinding{},
		func(s *Sheet) { ShowAttributeSettings(s) })
	perSheetBodyTypeSettingsAction = registerSheetAction("settings.body_type.per_sheet", PerSheetBodyTypeSettingsItemID,
		i18n.Text("Body Type…"), unison.KeyBinding{}, func(s *Sheet) { ShowBodySettings(s) })
	perSheetSettingsAction = registerSheetAction("settings.sheet.per_sheet", PerSheetSettingsItemID,
		i18n.Text("Sheet Settings…"),
		unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: mod.Shift | mod.OSMenuCommand()},
		func(s *Sheet) { ShowSheetSettings(s) })
	printAction = registerFocusAction("print", PrintItemID, i18n.Text("Print…"),
		unison.KeyBinding{KeyCode: unison.KeyP, Modifiers: mod.OSMenuCommand()})
	redoAction = registerKeyBindableAction("redo", undoRedoAction(RedoItemID, unison.KeyY, unison.CannotRedoTitle,
		(*unison.UndoManager).RedoTitle, (*unison.UndoManager).CanRedo, (*unison.UndoManager).Redo))
	saveAction = registerFocusAction("save", SaveItemID, i18n.Text("Save"),
		unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: mod.OSMenuCommand()})
	saveAsAction = registerFocusAction("save_as", SaveAsItemID, i18n.Text("Save As…"),
		unison.KeyBinding{KeyCode: unison.KeyS, Modifiers: mod.Shift | mod.OSMenuCommand()})
	scale25Action = registerFocusAction("scale.25", Scale25ItemID, i18n.Text("25% Scale"),
		unison.KeyBinding{KeyCode: unison.KeyQ, Modifiers: mod.OSMenuCommand() | mod.Option})
	scale50Action = registerFocusAction("scale.50", Scale50ItemID, i18n.Text("50% Scale"),
		unison.KeyBinding{KeyCode: unison.KeyH, Modifiers: mod.OSMenuCommand() | mod.Option})
	scale75Action = registerFocusAction("scale.75", Scale75ItemID, i18n.Text("75% Scale"),
		unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: mod.OSMenuCommand() | mod.Option})
	scale100Action = registerFocusAction("scale.100", Scale100ItemID, i18n.Text("100% Scale"),
		unison.KeyBinding{KeyCode: unison.Key1, Modifiers: mod.OSMenuCommand()})
	scale200Action = registerFocusAction("scale.200", Scale200ItemID, i18n.Text("200% Scale"),
		unison.KeyBinding{KeyCode: unison.Key2, Modifiers: mod.OSMenuCommand()})
	scale300Action = registerFocusAction("scale.300", Scale300ItemID, i18n.Text("300% Scale"),
		unison.KeyBinding{KeyCode: unison.Key3, Modifiers: mod.OSMenuCommand()})
	scale400Action = registerFocusAction("scale.400", Scale400ItemID, i18n.Text("400% Scale"),
		unison.KeyBinding{KeyCode: unison.Key4, Modifiers: mod.OSMenuCommand()})
	scale500Action = registerFocusAction("scale.500", Scale500ItemID, i18n.Text("500% Scale"),
		unison.KeyBinding{KeyCode: unison.Key5, Modifiers: mod.OSMenuCommand()})
	scale600Action = registerFocusAction("scale.600", Scale600ItemID, i18n.Text("600% Scale"),
		unison.KeyBinding{KeyCode: unison.Key6, Modifiers: mod.OSMenuCommand()})
	scaleDefaultAction = registerFocusAction("scale.default", ScaleDefaultItemID, i18n.Text("Default Scale"),
		unison.KeyBinding{KeyCode: unison.Key0, Modifiers: mod.OSMenuCommand()})
	scaleDownAction = registerFocusAction("scale.down", ScaleDownItemID, i18n.Text("Scale Down"),
		unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: mod.OSMenuCommand() | mod.Option})
	scaleUpAction = registerFocusAction("scale.up", ScaleUpItemID, i18n.Text("Scale Up"),
		unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: mod.OSMenuCommand() | mod.Option})
	syncWithSourceAction = registerFocusAction("clear.sync", SyncWithSourceItemID, i18n.Text("Sync with Source"),
		unison.KeyBinding{})
	swapDefaultsAction = registerFocusAction("swap.defaults", SwapDefaultsItemID, i18n.Text("Swap Defaults"),
		unison.KeyBinding{KeyCode: unison.KeyX, Modifiers: mod.Shift | mod.OSMenuCommand()})
	toggleStateAction = registerFocusAction("toggle", ToggleStateItemID, i18n.Text("Toggle State"),
		unison.KeyBinding{KeyCode: unison.KeyApostrophe, Modifiers: mod.OSMenuCommand()})
	undoAction = registerKeyBindableAction("undo", undoRedoAction(UndoItemID, unison.KeyZ, unison.CannotUndoTitle,
		(*unison.UndoManager).UndoTitle, (*unison.UndoManager).CanUndo, (*unison.UndoManager).Undo))

	// Actions that may not be assigned a key binding
	checkForAppUpdatesAction = &unison.Action{
		ID:    CheckForAppUpdatesItemID,
		Title: fmt.Sprintf(i18n.Text("Check for %s updates"), xos.AppName),
		// Usable whenever no check is already running, whatever the setting and whatever is already known: with an
		// update known, a fresh check reopens the update window, which is what the settings tooltip and the release
		// notes promise. A quiet check counts as running, so that the item doesn't offer to start a second request for
		// an answer that is already on its way.
		EnabledCallback: func(_ *unison.Action, _ any) bool {
			return !AppUpdateCheckInProgress()
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
		Title: fmt.Sprintf(i18n.Text("Make a One-time Donation for %s Development"), xos.AppName),
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
		Title: fmt.Sprintf(i18n.Text("Sponsor %s Development"), xos.AppName),
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
			HandleLink(nil, "md:User%20Guide/Home")
		},
	}
}

func registerKeyBindableAction(key string, action *unison.Action) *unison.Action {
	gurps.RegisterKeyBinding(key, action)
	return action
}

// registerFocusAction registers a key-bindable action whose enabled state and execution are routed to the focused
// panel. Pass the zero KeyBinding for an action that has no default key binding.
func registerFocusAction(key string, id int, title string, binding unison.KeyBinding) *unison.Action {
	return registerKeyBindableAction(key, &unison.Action{
		ID:              id,
		Title:           title,
		KeyBinding:      binding,
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	})
}

// registerLibraryAction registers a key-bindable action that opens a new, empty library of the given type in a table
// dockable titled fileName.
func registerLibraryAction[T gurps.Node[T]](key string, id int, title, fileName string,
	newDockable func(filePath string, rows []T) *TableDockable[T],
) *unison.Action {
	return registerKeyBindableAction(key, &unison.Action{
		ID:              id,
		Title:           title,
		ExecuteCallback: func(_ *unison.Action, _ any) { DisplayNewDockable(newDockable(fileName, nil)) },
	})
}

// registerSheetAction registers a key-bindable action that is enabled only while a character sheet is active and
// which passes that sheet to show when executed. Pass the zero KeyBinding for an action that has no default key
// binding.
func registerSheetAction(key string, id int, title string, binding unison.KeyBinding, show func(*Sheet)) *unison.Action {
	return registerKeyBindableAction(key, &unison.Action{
		ID:              id,
		Title:           title,
		KeyBinding:      binding,
		EnabledCallback: actionEnabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := ActiveSheet(); s != nil {
				show(s)
			}
		},
	})
}

// undoRedoAction returns an action that drives the active window's undo manager, or is disabled and titled with
// cannotTitle when there is no active window or it has no undo manager. title, can and do are the undo manager's
// UndoTitle/CanUndo/Undo or RedoTitle/CanRedo/Redo methods.
func undoRedoAction(id int, key unison.KeyCode, cannotTitle func() string, title func(*unison.UndoManager) string,
	can func(*unison.UndoManager) bool, do func(*unison.UndoManager),
) *unison.Action {
	return &unison.Action{
		ID:         id,
		Title:      cannotTitle(),
		KeyBinding: unison.KeyBinding{KeyCode: key, Modifiers: mod.OSMenuCommand()},
		EnabledCallback: func(action *unison.Action, _ any) bool {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					action.Title = title(mgr)
					return can(mgr)
				}
			}
			action.Title = cannotTitle()
			return false
		},
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if wnd := unison.ActiveWindow(); wnd != nil {
				if mgr := wnd.UndoManager(); mgr != nil {
					do(mgr)
				}
			}
		},
	}
}

func actionEnabledForSheet(_ *unison.Action, _ any) bool {
	return ActiveSheet() != nil
}

func showWebPage(uri string) {
	if err := xos.OpenBrowser(uri); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to open link"), err)
	}
}
