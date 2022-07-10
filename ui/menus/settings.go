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
	uisettings "github.com/richardwilkes/gcs/v5/ui/workspace/settings"
	"github.com/richardwilkes/gcs/v5/ui/workspace/sheet"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

var (
	// PerSheetSettings opens the settings for the front character sheet.
	PerSheetSettings *unison.Action
	// DefaultSheetSettings opens the default settings for the character sheet.
	DefaultSheetSettings *unison.Action
	// PerSheetAttributeSettings opens the attributes settings for the foremost character sheet.
	PerSheetAttributeSettings *unison.Action
	// DefaultAttributeSettings opens the default attributes settings.
	DefaultAttributeSettings *unison.Action
	// PerSheetBodyTypeSettings opens the body type settings for the foremost character sheet.
	PerSheetBodyTypeSettings *unison.Action
	// DefaultBodyTypeSettings opens the default body type settings.
	DefaultBodyTypeSettings *unison.Action
	// GeneralSettings opens the general settings.
	GeneralSettings *unison.Action
	// PageRefMappings opens the page reference mappings.
	PageRefMappings *unison.Action
	// ColorSettings opens the color settings.
	ColorSettings *unison.Action
	// FontSettings opens the font settings.
	FontSettings *unison.Action
	// MenuKeySettings opens the menu key settings.
	MenuKeySettings *unison.Action
)

func registerSettingsMenuActions() {
	PerSheetSettings = &unison.Action{
		ID:              constants.PerSheetSettingsItemID,
		Title:           i18n.Text("Sheet Settings…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: unison.ShiftModifier | unison.OSMenuCmdModifier()},
		EnabledCallback: enabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := sheet.ActiveSheet(); s != nil {
				uisettings.ShowSheetSettings(s)
			}
		},
	}
	DefaultSheetSettings = &unison.Action{
		ID:              constants.DefaultSheetSettingsItemID,
		Title:           i18n.Text("Default Sheet Settings…"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyComma, Modifiers: unison.OSMenuCmdModifier()},
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowSheetSettings(nil) },
	}
	PerSheetAttributeSettings = &unison.Action{
		ID:              constants.PerSheetAttributeSettingsItemID,
		Title:           i18n.Text("Attributes…"),
		EnabledCallback: enabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := sheet.ActiveSheet(); s != nil {
				uisettings.ShowAttributeSettings(s)
			}
		},
	}
	DefaultAttributeSettings = &unison.Action{
		ID:              constants.DefaultAttributeSettingsItemID,
		Title:           i18n.Text("Default Attributes…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowAttributeSettings(nil) },
	}
	PerSheetBodyTypeSettings = &unison.Action{
		ID:              constants.PerSheetBodyTypeSettingsItemID,
		Title:           i18n.Text("Body Type…"),
		EnabledCallback: enabledForSheet,
		ExecuteCallback: func(_ *unison.Action, _ any) {
			if s := sheet.ActiveSheet(); s != nil {
				uisettings.ShowBodyTypeSettings(s)
			}
		},
	}
	DefaultBodyTypeSettings = &unison.Action{
		ID:              constants.DefaultBodyTypeSettingsItemID,
		Title:           i18n.Text("Default Body Type…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowBodyTypeSettings(nil) },
	}
	GeneralSettings = &unison.Action{
		ID:              constants.GeneralSettingsItemID,
		Title:           i18n.Text("General Settings…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowGeneralSettings() },
	}
	PageRefMappings = &unison.Action{
		ID:              constants.PageRefMappingsItemID,
		Title:           i18n.Text("Page Reference Mappings…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowPageRefMappings() },
	}
	ColorSettings = &unison.Action{
		ID:              constants.ColorSettingsItemID,
		Title:           i18n.Text("Colors…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowColorSettings() },
	}
	FontSettings = &unison.Action{
		ID:              constants.FontSettingsItemID,
		Title:           i18n.Text("Fonts…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowFontSettings() },
	}
	MenuKeySettings = &unison.Action{
		ID:              constants.MenuKeySettingsItemID,
		Title:           i18n.Text("Menu Keys…"),
		ExecuteCallback: func(_ *unison.Action, _ any) { uisettings.ShowMenuKeySettings() },
	}

	settings.RegisterKeyBinding("settings.sheet.per_sheet", PerSheetSettings)
	settings.RegisterKeyBinding("settings.attributes.per_sheet", PerSheetAttributeSettings)
	settings.RegisterKeyBinding("settings.body_type.per_sheet", PerSheetBodyTypeSettings)
	settings.RegisterKeyBinding("settings.sheet.default", DefaultSheetSettings)
	settings.RegisterKeyBinding("settings.attributes.default", DefaultAttributeSettings)
	settings.RegisterKeyBinding("settings.body_type.default", DefaultBodyTypeSettings)
	settings.RegisterKeyBinding("settings.general", GeneralSettings)
	settings.RegisterKeyBinding("settings.pagerefs", PageRefMappings)
	settings.RegisterKeyBinding("settings.colors", ColorSettings)
	settings.RegisterKeyBinding("settings.fonts", FontSettings)
	settings.RegisterKeyBinding("settings.keys", MenuKeySettings)
}

func createSettingsMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(constants.SettingsMenuID, i18n.Text("Settings"), nil)
	m.InsertItem(-1, PerSheetSettings.NewMenuItem(f))
	m.InsertItem(-1, PerSheetAttributeSettings.NewMenuItem(f))
	m.InsertItem(-1, PerSheetBodyTypeSettings.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, DefaultSheetSettings.NewMenuItem(f))
	m.InsertItem(-1, DefaultAttributeSettings.NewMenuItem(f))
	m.InsertItem(-1, DefaultBodyTypeSettings.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, GeneralSettings.NewMenuItem(f))
	m.InsertItem(-1, PageRefMappings.NewMenuItem(f))
	m.InsertItem(-1, ColorSettings.NewMenuItem(f))
	m.InsertItem(-1, FontSettings.NewMenuItem(f))
	m.InsertItem(-1, MenuKeySettings.NewMenuItem(f))
	return m
}

func enabledForSheet(_ *unison.Action, _ any) bool {
	return sheet.ActiveSheet() != nil
}
