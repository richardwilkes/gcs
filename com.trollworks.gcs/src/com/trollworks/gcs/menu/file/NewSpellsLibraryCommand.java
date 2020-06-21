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
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.spell.SpellsDockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Spells Library" command. */
public class NewSpellsLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                  CMD_NEW_LIBRARY = "NewSpellsLibrary";
    /** The singleton {@link NewSpellsLibraryCommand}. */
    public static final NewSpellsLibraryCommand INSTANCE        = new NewSpellsLibraryCommand();

    private NewSpellsLibraryCommand() {
        super(I18n.Text("New Spells Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newSpellsLibrary();
    }

    /** @return The newly created a new {@link SpellsDockable}. */
    public static SpellsDockable newSpellsLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            SpellList list = new SpellList();
            list.getModel().setLocked(false);
            SpellsDockable dockable = new SpellsDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
