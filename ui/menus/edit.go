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
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var (
	// Undo the last action.
	Undo *unison.Action
	// Redo the last action.
	Redo *unison.Action
	// Duplicate the currently selected content.
	Duplicate *unison.Action
	// OpenEditor opens an editor for the selected item(s).
	OpenEditor *unison.Action
	// CopyToSheet copies the selected items to the foremost character sheet.
	CopyToSheet *unison.Action
	// CopyToTemplate copies the selected items to the foremost template.
	CopyToTemplate *unison.Action
	// ApplyTemplate applies the foremost template to the foremost character sheet.
	ApplyTemplate *unison.Action
	// Increment the points of the selection.
	Increment *unison.Action
	// Decrement the points of the selection.
	Decrement *unison.Action
	// IncreaseUses increments the uses of the selection.
	IncreaseUses *unison.Action
	// DecreaseUses decrements the uses of the selection.
	DecreaseUses *unison.Action
	// IncreaseSkillLevel increments the uses of the skill level.
	IncreaseSkillLevel *unison.Action
	// DecreaseSkillLevel decrements the uses of the skill level.
	DecreaseSkillLevel *unison.Action
	// IncreaseTechLevel increments the uses of the tech level.
	IncreaseTechLevel *unison.Action
	// DecreaseTechLevel decrements the uses of the tech level.
	DecreaseTechLevel *unison.Action
	// ToggleState switches the state of the selected item(s).
	ToggleState *unison.Action
	// SwapDefaults swaps the defaults of the selected skill.
	SwapDefaults *unison.Action
	// ConvertToContainer converts the currently selected item into a container.
	ConvertToContainer *unison.Action
)

func registerEditMenuActions() {
	Undo = &unison.Action{
		ID:         constants.UndoItemID,
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
	}
	Redo = &unison.Action{
		ID:         constants.RedoItemID,
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
	}
	Duplicate = &unison.Action{
		ID:              constants.DuplicateItemID,
		Title:           i18n.Text("Duplicate"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyU, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	OpenEditor = &unison.Action{
		ID:              constants.OpenEditorItemID,
		Title:           i18n.Text("Open Detail Editor"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyI, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	CopyToSheet = &unison.Action{
		ID:              constants.CopyToSheetItemID,
		Title:           i18n.Text("Copy to Character Sheet"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyC, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	CopyToTemplate = &unison.Action{
		ID:              constants.CopyToTemplateItemID,
		Title:           i18n.Text("Copy to Template"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ApplyTemplate = &unison.Action{
		ID:              constants.ApplyTemplateItemID,
		Title:           i18n.Text("Apply Template to Character Sheet"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyA, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Increment = &unison.Action{
		ID:              constants.IncrementItemID,
		Title:           i18n.Text("Increment"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Decrement = &unison.Action{
		ID:              constants.DecrementItemID,
		Title:           i18n.Text("Decrement"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	IncreaseUses = &unison.Action{
		ID:              constants.IncrementUsesItemID,
		Title:           i18n.Text("Increase Uses"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyUp, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	DecreaseUses = &unison.Action{
		ID:              constants.DecrementUsesItemID,
		Title:           i18n.Text("Decrease Uses"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyDown, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	IncreaseSkillLevel = &unison.Action{
		ID:              constants.IncrementSkillLevelItemID,
		Title:           i18n.Text("Increase Skill Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeySlash, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	DecreaseSkillLevel = &unison.Action{
		ID:              constants.DecrementSkillLevelItemID,
		Title:           i18n.Text("Decrease Skill Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyPeriod, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	IncreaseTechLevel = &unison.Action{
		ID:              constants.IncrementTechLevelItemID,
		Title:           i18n.Text("Increase Tech Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyCloseBracket, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	DecreaseTechLevel = &unison.Action{
		ID:              constants.DecrementTechLevelItemID,
		Title:           i18n.Text("Decrease Tech Level"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyOpenBracket, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ToggleState = &unison.Action{
		ID:              constants.ToggleStateItemID,
		Title:           i18n.Text("Toggle State"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyApostrophe, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	SwapDefaults = &unison.Action{
		ID:              constants.SwapDefaultsItemID,
		Title:           i18n.Text("Swap Defaults"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyX, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ConvertToContainer = &unison.Action{
		ID:              constants.ConvertToContainerItemID,
		Title:           i18n.Text("Convert to Container"),
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}

	settings.RegisterKeyBinding("undo", Undo)
	settings.RegisterKeyBinding("redo", Redo)
	settings.RegisterKeyBinding("cut", unison.CutAction)
	settings.RegisterKeyBinding("copy", unison.CopyAction)
	settings.RegisterKeyBinding("paste", unison.PasteAction)
	settings.RegisterKeyBinding("duplicate", Duplicate)
	settings.RegisterKeyBinding("delete", unison.DeleteAction)
	settings.RegisterKeyBinding("select.all", unison.SelectAllAction)
	settings.RegisterKeyBinding("open.editor", OpenEditor)
	settings.RegisterKeyBinding("copy.to_sheet", CopyToSheet)
	settings.RegisterKeyBinding("copy.to_template", CopyToTemplate)
	settings.RegisterKeyBinding("apply.template", ApplyTemplate)
	settings.RegisterKeyBinding("inc", Increment)
	settings.RegisterKeyBinding("dec", Decrement)
	settings.RegisterKeyBinding("inc.uses", IncreaseUses)
	settings.RegisterKeyBinding("dec.uses", DecreaseUses)
	settings.RegisterKeyBinding("inc.sl", IncreaseSkillLevel)
	settings.RegisterKeyBinding("dec.sl", DecreaseSkillLevel)
	settings.RegisterKeyBinding("inc.tl", IncreaseTechLevel)
	settings.RegisterKeyBinding("dec.tl", DecreaseTechLevel)
	settings.RegisterKeyBinding("toggle", ToggleState)
	settings.RegisterKeyBinding("swap.defaults", SwapDefaults)
	settings.RegisterKeyBinding("convert.to_container", ConvertToContainer)
}

func setupEditMenu(bar unison.Menu) {
	f := bar.Factory()
	m := bar.Menu(unison.EditMenuID)

	i := insertItem(m, 0, Undo.NewMenuItem(f))
	i = insertItem(m, i, Redo.NewMenuItem(f))
	insertSeparator(m, i)

	m.InsertItem(m.Item(unison.DeleteItemID).Index(), Duplicate.NewMenuItem(f))

	i = insertSeparator(m, m.Item(unison.SelectAllItemID).Index()+1)
	i = insertItem(m, i, OpenEditor.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, CopyToSheet.NewMenuItem(f))
	i = insertItem(m, i, CopyToTemplate.NewMenuItem(f))
	i = insertItem(m, i, ApplyTemplate.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, Increment.NewMenuItem(f))
	i = insertItem(m, i, Decrement.NewMenuItem(f))
	i = insertItem(m, i, IncreaseUses.NewMenuItem(f))
	i = insertItem(m, i, DecreaseUses.NewMenuItem(f))
	i = insertItem(m, i, IncreaseSkillLevel.NewMenuItem(f))
	i = insertItem(m, i, DecreaseSkillLevel.NewMenuItem(f))
	i = insertItem(m, i, IncreaseTechLevel.NewMenuItem(f))
	i = insertItem(m, i, DecreaseTechLevel.NewMenuItem(f))

	i = insertSeparator(m, i)
	i = insertItem(m, i, ToggleState.NewMenuItem(f))
	i = insertItem(m, i, SwapDefaults.NewMenuItem(f))
	insertItem(m, i, ConvertToContainer.NewMenuItem(f))
}
