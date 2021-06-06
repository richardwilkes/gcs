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
import com.trollworks.gcs.modifier.AdvantageModifierList;
import com.trollworks.gcs.modifier.AdvantageModifiersDockable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

public final class NewAdvantageModifiersLibraryCommand extends Command {
    public static final NewAdvantageModifiersLibraryCommand INSTANCE = new NewAdvantageModifiersLibraryCommand();

    private NewAdvantageModifiersLibraryCommand() {
        super(I18n.text("New Advantage Modifiers Library"), "NewAdvantageModifiersLibrary");
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        LibraryExplorerDockable library = LibraryExplorerDockable.get();
        if (library != null) {
            AdvantageModifierList list = new AdvantageModifierList();
            list.getModel().setLocked(false);
            AdvantageModifiersDockable dockable = new AdvantageModifiersDockable(list);
            library.dockLibrary(dockable);
        }
    }
}
