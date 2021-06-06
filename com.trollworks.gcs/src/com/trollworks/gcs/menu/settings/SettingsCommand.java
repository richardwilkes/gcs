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

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.SheetSettingsWindow;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

public final class SettingsCommand extends Command {
    public static final SettingsCommand PER_SHEET = new SettingsCommand(false);
    public static final SettingsCommand DEFAULTS  = new SettingsCommand(true);
    private             boolean         mForDefaults;

    private SettingsCommand(boolean defaults) {
        super(defaults ? I18n.text("Default Sheet Settings…") : I18n.text("Sheet Settings…"),
                defaults ? "default_sheet_settings" : "sheet_settings",
                KeyEvent.VK_COMMA, defaults ? COMMAND_MODIFIER : SHIFTED_COMMAND_MODIFIER);
        mForDefaults = defaults;
    }

    @Override
    public void adjust() {
        boolean shouldEnable = !UIUtilities.inModalState();
        if (shouldEnable && !mForDefaults) {
            shouldEnable = getTarget(SheetDockable.class) != null;
        }
        setEnabled(shouldEnable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!UIUtilities.inModalState()) {
            if (mForDefaults) {
                SheetSettingsWindow.display(null);
            } else {
                SheetDockable target = getTarget(SheetDockable.class);
                if (target != null) {
                    SheetSettingsWindow.display(target.getDataFile());
                }
            }
        }
    }
}
