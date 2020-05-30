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

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.utility.UpdateChecker;

import java.awt.event.ActionEvent;

public class MasterLibraryStatusCommand extends Command {
    /** The singleton {@link MasterLibraryStatusCommand}. */
    public static final MasterLibraryStatusCommand INSTANCE = new MasterLibraryStatusCommand();

    private MasterLibraryStatusCommand() {
        super(UpdateChecker.getDataResult(), "MasterLibraryStatus");
    }

    @Override
    public void adjust() {
        if (StdMenuBar.SUPRESS_MENUS) {
            setEnabled(false);
            return;
        }
        setTitle(UpdateChecker.getDataResult());
        setEnabled(false);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
    }
}
