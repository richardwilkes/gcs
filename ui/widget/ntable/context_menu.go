package ntable

import (
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// ContextMenuItem holds the title and ID of a context menu item that is derived from a command ID.
type ContextMenuItem struct {
	Title string
	ID    int
}

// DefaultContextMenuItems holds the default set of context menu items for lists.
var DefaultContextMenuItems = []ContextMenuItem{
	{"", -1},
	{i18n.Text("Open Detail Editor"), constants.OpenEditorItemID},
	{"", -1},
	{i18n.Text("Duplicate"), constants.DuplicateItemID},
	{i18n.Text("Delete"), unison.DeleteItemID},
	{"", -1},
	{i18n.Text("Copy to Character Sheet"), constants.CopyToSheetItemID},
	{i18n.Text("Copy to Template"), constants.CopyToTemplateItemID},
	{i18n.Text("Apply Template to Character Sheet"), constants.ApplyTemplateItemID},
	{"", -1},
	{i18n.Text("Increment"), constants.IncrementItemID},
	{i18n.Text("Decrement"), constants.DecrementItemID},
	{i18n.Text("Increase Uses"), constants.IncrementUsesItemID},
	{i18n.Text("Decrease Uses"), constants.DecrementUsesItemID},
	{i18n.Text("Increase Skill Level"), constants.IncrementSkillLevelItemID},
	{i18n.Text("Decrease Skill Level"), constants.DecrementSkillLevelItemID},
	{i18n.Text("Increase Tech Level"), constants.IncrementTechLevelItemID},
	{i18n.Text("Decrease Tech Level"), constants.DecrementTechLevelItemID},
	{"", -1},
	{i18n.Text("Toggle State"), constants.ToggleStateItemID},
	{i18n.Text("Swap Defaults"), constants.SwapDefaultsItemID},
	{i18n.Text("Convert to Container"), constants.ConvertToContainerItemID},
	{"", -1},
	{i18n.Text("Open Page Reference"), constants.OpenOnePageReferenceItemID},
	{i18n.Text("Open Each Page Reference"), constants.OpenEachPageReferenceItemID},
}

// CarriedEquipmentExtraContextMenuItems holds context menu items specific to the carried equipment list.
var CarriedEquipmentExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Carried Equipment"), constants.NewCarriedEquipmentItemID},
	{i18n.Text("New Carried Equipment Container"), constants.NewCarriedEquipmentContainerItemID},
}

// EquipmentModifierExtraContextMenuItems holds context menu items specific to the equipment modifier list.
var EquipmentModifierExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Equipment Modifier"), constants.NewEquipmentModifierItemID},
	{i18n.Text("New Equipment Modifier Container"), constants.NewEquipmentContainerModifierItemID},
}

// MeleeWeaponExtraContextMenuItems holds context menu items specific to the melee weapon list.
var MeleeWeaponExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Melee Weapon"), constants.NewMeleeWeaponItemID},
}

// NoteExtraContextMenuItems holds context menu items specific to the note list.
var NoteExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Note"), constants.NewNoteItemID},
	{i18n.Text("New Note Container"), constants.NewNoteContainerItemID},
}

// OtherEquipmentExtraContextMenuItems holds context menu items specific to the other equipment list.
var OtherEquipmentExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Other Equipment"), constants.NewOtherEquipmentItemID},
	{i18n.Text("New Other Equipment Container"), constants.NewOtherEquipmentContainerItemID},
}

// RangedWeaponExtraContextMenuItems holds context menu items specific to the ranged weapon list.
var RangedWeaponExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Ranged Weapon"), constants.NewRangedWeaponItemID},
}

// SkillExtraContextMenuItems holds context menu items specific to the skill list.
var SkillExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Skill"), constants.NewSkillItemID},
	{i18n.Text("New Skill Container"), constants.NewSkillContainerItemID},
	{i18n.Text("New Technique"), constants.NewTechniqueItemID},
}

// SpellExtraContextMenuItems holds context menu items specific to the spell list.
var SpellExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Spell"), constants.NewSpellItemID},
	{i18n.Text("New Spell Container"), constants.NewSpellContainerItemID},
	{i18n.Text("New Ritual Magic Spell"), constants.NewRitualMagicSpellItemID},
}

// TraitExtraContextMenuItems holds context menu items specific to the trait list.
var TraitExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Trait"), constants.NewTraitItemID},
	{i18n.Text("New Trait Container"), constants.NewTraitContainerItemID},
	{i18n.Text("Add Natural Attacks"), constants.AddNaturalAttacksItemID},
}

// TraitModifierExtraContextMenuItems holds context menu items specific to the trait modifier list.
var TraitModifierExtraContextMenuItems = []ContextMenuItem{
	{i18n.Text("New Trait Modifier"), constants.NewTraitModifierItemID},
	{i18n.Text("New Trait Modifier Container"), constants.NewTraitContainerModifierItemID},
}
