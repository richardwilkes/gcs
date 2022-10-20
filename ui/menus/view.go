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
	"runtime"

	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// Scale menu items
var (
	ScaleDefault *unison.Action
	ScaleUp      *unison.Action
	ScaleDown    *unison.Action
	Scale25      *unison.Action
	Scale50      *unison.Action
	Scale75      *unison.Action
	Scale100     *unison.Action
	Scale200     *unison.Action
	Scale300     *unison.Action
	Scale400     *unison.Action
	Scale500     *unison.Action
	Scale600     *unison.Action
)

func registerViewMenuActions() {
	ScaleDefault = &unison.Action{
		ID:              constants.ScaleDefaultItemID,
		Title:           i18n.Text("Default Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key0, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ScaleUp = &unison.Action{
		ID:              constants.ScaleUpItemID,
		Title:           i18n.Text("Scale Up"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyEqual, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	ScaleDown = &unison.Action{
		ID:              constants.ScaleDownItemID,
		Title:           i18n.Text("Scale Down"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyMinus, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale25 = &unison.Action{
		ID:              constants.Scale25ItemID,
		Title:           i18n.Text("25% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyQ, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale50 = &unison.Action{
		ID:              constants.Scale50ItemID,
		Title:           i18n.Text("50% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyH, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale75 = &unison.Action{
		ID:              constants.Scale75ItemID,
		Title:           i18n.Text("75% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.KeyT, Modifiers: unison.OSMenuCmdModifier() | unison.OptionModifier},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale100 = &unison.Action{
		ID:              constants.Scale100ItemID,
		Title:           i18n.Text("100% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key1, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale200 = &unison.Action{
		ID:              constants.Scale200ItemID,
		Title:           i18n.Text("200% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key2, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale300 = &unison.Action{
		ID:              constants.Scale300ItemID,
		Title:           i18n.Text("300% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key3, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale400 = &unison.Action{
		ID:              constants.Scale400ItemID,
		Title:           i18n.Text("400% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key4, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale500 = &unison.Action{
		ID:              constants.Scale500ItemID,
		Title:           i18n.Text("500% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key5, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}
	Scale600 = &unison.Action{
		ID:              constants.Scale600ItemID,
		Title:           i18n.Text("600% Scale"),
		KeyBinding:      unison.KeyBinding{KeyCode: unison.Key6, Modifiers: unison.OSMenuCmdModifier()},
		EnabledCallback: unison.RouteActionToFocusEnabledFunc,
		ExecuteCallback: unison.RouteActionToFocusExecuteFunc,
	}

	settings.RegisterKeyBinding("scale.default", ScaleDefault)
	settings.RegisterKeyBinding("scale.up", ScaleUp)
	settings.RegisterKeyBinding("scale.down", ScaleDown)
	settings.RegisterKeyBinding("scale.25", Scale25)
	settings.RegisterKeyBinding("scale.50", Scale50)
	settings.RegisterKeyBinding("scale.75", Scale75)
	settings.RegisterKeyBinding("scale.100", Scale100)
	settings.RegisterKeyBinding("scale.200", Scale200)
	settings.RegisterKeyBinding("scale.300", Scale300)
	settings.RegisterKeyBinding("scale.400", Scale400)
	settings.RegisterKeyBinding("scale.500", Scale500)
	settings.RegisterKeyBinding("scale.600", Scale600)
}

func createViewMenu(f unison.MenuFactory) unison.Menu {
	m := f.NewMenu(constants.ViewMenuID, i18n.Text("View"), nil)

	m.InsertItem(-1, ScaleDefault.NewMenuItem(f))
	m.InsertItem(-1, ScaleUp.NewMenuItem(f))
	m.InsertItem(-1, ScaleDown.NewMenuItem(f))
	m.InsertSeparator(-1, false)
	m.InsertItem(-1, Scale25.NewMenuItem(f))
	m.InsertItem(-1, Scale50.NewMenuItem(f))
	m.InsertItem(-1, Scale75.NewMenuItem(f))
	m.InsertItem(-1, Scale100.NewMenuItem(f))
	m.InsertItem(-1, Scale200.NewMenuItem(f))
	m.InsertItem(-1, Scale300.NewMenuItem(f))
	m.InsertItem(-1, Scale400.NewMenuItem(f))
	m.InsertItem(-1, Scale500.NewMenuItem(f))
	m.InsertItem(-1, Scale600.NewMenuItem(f))

	if runtime.GOOS == toolbox.MacOS {
		m.InsertSeparator(-1, false)
	}

	return m
}
