/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.modifier.EquipmentModifiersDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Equipment Modifier" command. */
public class NewEquipmentModifierCommand extends Command {
    /** The action command this command will issue. */
    public static final String                      CMD_NEW_EQUIPMENT_MODIFIER = "NewEquipmentModifier";
    /** The "New Equipment Modifier" command. */
    public static final NewEquipmentModifierCommand INSTANCE                   = new NewEquipmentModifierCommand(I18n.Text("New Equipment Modifier"), CMD_NEW_EQUIPMENT_MODIFIER, KeyEvent.VK_M, COMMAND_MODIFIER);

    private NewEquipmentModifierCommand(String title, String cmd, int keyCode, int modifiers) {
        super(title, cmd, keyCode, modifiers);
    }

    @Override
    public void adjust() {
        boolean                    enable   = false;
        EquipmentModifiersDockable dockable = getTarget(EquipmentModifiersDockable.class);
        if (dockable != null) {
            enable = !dockable.getOutline().getModel().isLocked();
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline                outline;
        DataFile                   dataFile;
        EquipmentModifiersDockable dockable = getTarget(EquipmentModifiersDockable.class);
        if (dockable == null) {
            return;
        }
        dataFile = dockable.getDataFile();
        outline = dockable.getOutline();
        if (outline.getModel().isLocked()) {
            return;
        }
        EquipmentModifier modifier = new EquipmentModifier(dataFile);
        outline.addRow(modifier, getTitle(), false);
        outline.getModel().select(modifier, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
