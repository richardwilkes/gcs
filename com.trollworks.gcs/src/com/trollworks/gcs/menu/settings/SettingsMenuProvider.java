/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.settings;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

public final class SettingsMenuProvider {
    private SettingsMenuProvider() {
    }

    public static List<Command> getModifiableCommands() {
        List<Command> cmds = new ArrayList<>();
        cmds.add(SettingsCommand.INSTANCE);
        cmds.add(AttributeSettingsCommand.PER_SHEET);
        cmds.add(HitLocationSettingsCommand.PER_SHEET);
        cmds.add(PreferencesCommand.INSTANCE);
        cmds.add(AttributeSettingsCommand.DEFAULTS);
        cmds.add(HitLocationSettingsCommand.DEFAULTS);
        cmds.add(PageRefMappingsCommand.INSTANCE);
        cmds.add(ThemeSettingsCommand.INSTANCE);
        cmds.add(MenuKeySettingsCommand.INSTANCE);
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Settings"));
        menu.add(new DynamicMenuItem(SettingsCommand.INSTANCE));
        menu.add(new DynamicMenuItem(AttributeSettingsCommand.PER_SHEET));
        menu.add(new DynamicMenuItem(HitLocationSettingsCommand.PER_SHEET));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(PreferencesCommand.INSTANCE));
        menu.add(new DynamicMenuItem(AttributeSettingsCommand.DEFAULTS));
        menu.add(new DynamicMenuItem(HitLocationSettingsCommand.DEFAULTS));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(PageRefMappingsCommand.INSTANCE));
        menu.add(new DynamicMenuItem(ThemeSettingsCommand.INSTANCE));
        menu.add(new DynamicMenuItem(MenuKeySettingsCommand.INSTANCE));
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
