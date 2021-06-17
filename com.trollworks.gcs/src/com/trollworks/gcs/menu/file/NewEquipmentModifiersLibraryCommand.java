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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.modifier.EquipmentModifierList;
import com.trollworks.gcs.modifier.EquipmentModifiersDockable;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

public final class NewEquipmentModifiersLibraryCommand extends Command {
    public static final NewEquipmentModifiersLibraryCommand INSTANCE = new NewEquipmentModifiersLibraryCommand();

    private NewEquipmentModifiersLibraryCommand() {
        super(I18n.text("New Equipment Modifiers Library"), "NewEquipmentModifiersLibrary");
    }

    @Override
    public void adjust() {
        setEnabled(!UIUtilities.inModalState());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            EquipmentModifierList list = new EquipmentModifierList();
            list.getModel().setLocked(false);
            EquipmentModifiersDockable dockable = new EquipmentModifiersDockable(list);
            library.dockLibrary(dockable);
        }
    }
}
