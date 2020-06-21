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

package com.trollworks.gcs.menu.library;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.utility.I18n;

import javax.swing.JMenu;
import javax.swing.JMenuItem;
import javax.swing.event.MenuEvent;
import javax.swing.event.MenuListener;

public class LibraryMenu extends JMenu implements MenuListener {
    public LibraryMenu() {
        super(I18n.Text("Library"));
        addMenuListener(this);
    }

    @Override
    public void menuCanceled(MenuEvent event) {
        // Nothing to do.
    }

    @Override
    public void menuDeselected(MenuEvent event) {
        // Nothing to do.
    }

    @Override
    public void menuSelected(MenuEvent event) {
        removeAll();
        for (Library lib : Library.LIBRARIES) {
            if (lib != Library.USER) {
                LibraryUpdateCommand item = new LibraryUpdateCommand(lib);
                item.adjust();
                add(new JMenuItem(item));
            }
            add(new JMenuItem(new ShowLibraryFolderCommand(lib)));
            addSeparator();
        }
        add(new JMenuItem(ChangeLibraryLocationsCommand.INSTANCE));
    }
}
