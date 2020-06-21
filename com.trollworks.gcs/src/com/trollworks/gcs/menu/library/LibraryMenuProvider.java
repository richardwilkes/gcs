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

public class LibraryMenuProvider {
    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Library"));
        for (Library lib : Library.LIBRARIES) {
            if (lib != Library.USER) {
                menu.add(new JMenuItem(new LibraryUpdateCommand(lib)));
            }
            menu.add(new JMenuItem(new ShowLibraryFolderCommand(lib)));
            menu.addSeparator();
        }
        menu.add(new JMenuItem(ChangeLibraryLocationsCommand.INSTANCE));
        return menu;
    }
}
