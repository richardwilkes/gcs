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

package com.trollworks.gcs.menu;

import com.trollworks.gcs.menu.edit.EditMenuProvider;
import com.trollworks.gcs.menu.file.FileMenuProvider;
import com.trollworks.gcs.menu.help.HelpMenuProvider;
import com.trollworks.gcs.menu.item.ItemMenuProvider;
import com.trollworks.gcs.menu.library.LibraryMenu;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenuBar;

/** The standard menu bar. */
public class StdMenuBar extends JMenuBar {
    /** @return The {@link Command}s that can have their accelerators modified. */
    public static final List<Command> getCommands() {
        List<Command> cmds = new ArrayList<>();
        cmds.addAll(FileMenuProvider.getModifiableCommands());
        cmds.addAll(EditMenuProvider.getModifiableCommands());
        cmds.addAll(ItemMenuProvider.getModifiableCommands());
        cmds.sort((Command c1, Command c2) -> NumericComparator.caselessCompareStrings(c1.getTitle(), c2.getTitle()));
        return cmds;
    }

    /** Creates a new {@link StdMenuBar}. */
    public StdMenuBar() {
        add(FileMenuProvider.createMenu());
        add(EditMenuProvider.createMenu());
        add(ItemMenuProvider.createMenu());
        add(new LibraryMenu());
        add(HelpMenuProvider.createMenu());
    }
}
