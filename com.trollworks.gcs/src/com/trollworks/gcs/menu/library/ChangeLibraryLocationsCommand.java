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

import com.trollworks.gcs.library.LibraryLocationsPanel;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

public class ChangeLibraryLocationsCommand extends Command {
    public static final ChangeLibraryLocationsCommand INSTANCE = new ChangeLibraryLocationsCommand();

    private ChangeLibraryLocationsCommand() {
        super(I18n.Text("Change Library Locations"), "change_library_locations");
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        LibraryLocationsPanel.showDialog();
    }
}
