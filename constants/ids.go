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

package constants

import "github.com/richardwilkes/unison"

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
	OpenItemID
	CloseTabID
	RecentFilesMenuID
	SaveItemID
	SaveAsItemID
	ExportToMenuID
	PrintItemID
	UndoItemID
	RedoItemID
	DuplicateItemID
	ConvertToContainerItemID
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
	LibraryMenuID
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
	WebSiteItemID
	MailingListItemID
	ChangeLibraryLocationsItemID

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

	LibraryBaseItemID
	RecentFieldBaseItemID  = LibraryBaseItemID + 1000
	ExportToTextBaseItemID = RecentFieldBaseItemID + 1000
)
