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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.equipment.EquipmentDockable;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Equipment Library" command. */
public class NewEquipmentLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                     CMD_NEW_LIBRARY = "NewEquipmentLibrary";
    /** The singleton {@link NewEquipmentLibraryCommand}. */
    public static final NewEquipmentLibraryCommand INSTANCE        = new NewEquipmentLibraryCommand();

    private NewEquipmentLibraryCommand() {
        super(I18n.Text("New Equipment Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newEquipmentLibrary();
    }

    /** @return The newly created a new {@link EquipmentDockable}. */
    public static EquipmentDockable newEquipmentLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            EquipmentList list = new EquipmentList();
            list.getModel().setLocked(false);
            EquipmentDockable dockable = new EquipmentDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
