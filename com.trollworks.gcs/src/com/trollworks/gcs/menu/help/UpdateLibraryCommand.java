/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.app.UpdateChecker;
import com.trollworks.gcs.menu.Command;

import java.awt.event.ActionEvent;

/** Provides the "Update Library" command. */
public class UpdateLibraryCommand extends Command {
    /** The action command this command will issue. */
    public static final String               CMD_CHECK_FOR_UPDATE = "CheckForLibraryUpdate";
    /** The singleton {@link UpdateLibraryCommand}. */
    public static final UpdateLibraryCommand INSTANCE             = new UpdateLibraryCommand();

    private UpdateLibraryCommand() {
        super(UpdateChecker.getDataResult(), CMD_CHECK_FOR_UPDATE);
    }

    @Override
    public void adjust() {
        setTitle(UpdateChecker.getDataResult());
        setEnabled(false);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
    }
}
