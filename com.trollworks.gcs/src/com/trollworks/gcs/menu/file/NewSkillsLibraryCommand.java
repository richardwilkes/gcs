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
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.skill.SkillsDockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

/** Provides the "New Skills Library" command. */
public class NewSkillsLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String                  CMD_NEW_LIBRARY = "NewSkillsLibrary";
    /** The singleton {@link NewSkillsLibraryCommand}. */
    public static final NewSkillsLibraryCommand INSTANCE        = new NewSkillsLibraryCommand();

    private NewSkillsLibraryCommand() {
        super(I18n.Text("New Skills Library"), CMD_NEW_LIBRARY);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        newSkillsLibrary();
    }

    /** @return The newly created a new {@link SkillsDockable}. */
    public static SkillsDockable newSkillsLibrary() {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            SkillList list = new SkillList();
            list.getModel().setLocked(false);
            SkillsDockable dockable = new SkillsDockable(list);
            library.dockLibrary(dockable);
            return dockable;
        }
        return null;
    }
}
