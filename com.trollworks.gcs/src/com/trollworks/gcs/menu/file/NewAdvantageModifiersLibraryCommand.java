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

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.modifier.AdvantageModifierList;
import com.trollworks.gcs.modifier.AdvantageModifiersDockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Advantage Modifiers Library" command. */
public class NewAdvantageModifiersLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                              CMD_NEW_LIBRARY = "NewAdvantageModifiersLibrary";
    /** The singleton {@link NewAdvantageModifiersLibraryCommand}. */
    public static final NewAdvantageModifiersLibraryCommand INSTANCE        = new NewAdvantageModifiersLibraryCommand();

    private NewAdvantageModifiersLibraryCommand() {
        super(I18n.Text("New Advantage Modifiers Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newAdvantageModifiersLibrary();
    }

    /** @return The newly created a new {@link AdvantageModifiersDockable}. */
    public static AdvantageModifiersDockable newAdvantageModifiersLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            AdvantageModifierList list = new AdvantageModifierList();
            list.getModel().setLocked(false);
            AdvantageModifiersDockable dockable = new AdvantageModifiersDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
