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

import com.trollworks.gcs.character.SettingsEditor;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

public class SettingsCommand extends Command {
    public static final SettingsCommand INSTANCE = new SettingsCommand();

    private SettingsCommand() {
        super(I18n.Text("Sheet…"), "SheetSettings", KeyEvent.VK_COMMA, SHIFTED_COMMAND_MODIFIER);
    }

    @Override
    public void adjust() {
        setEnabled(!UIUtilities.inModalState() && getTarget(SheetDockable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!UIUtilities.inModalState()) {
            SheetDockable target = getTarget(SheetDockable.class);
            if (target != null) {
                SettingsEditor.display(target.getDataFile());
            }
        }
    }
}
