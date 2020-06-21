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

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Advantages Library" command. */
public class NewAdvantagesLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                      CMD_NEW_LIBRARY = "NewAdvantagesLibrary";
    /** The singleton {@link NewAdvantagesLibraryCommand}. */
    public static final NewAdvantagesLibraryCommand INSTANCE        = new NewAdvantagesLibraryCommand();

    private NewAdvantagesLibraryCommand() {
        super(I18n.Text("New Advantages Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newAdvantagesLibrary();
    }

    /** @return The newly created a new {@link AdvantagesDockable}. */
    public static AdvantagesDockable newAdvantagesLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            AdvantageList list = new AdvantageList();
            list.getModel().setLocked(false);
            AdvantagesDockable dockable = new AdvantagesDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
