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
import com.trollworks.gcs.utility.UpdateChecker;

import java.awt.event.ActionEvent;

/** Provides the "Update App" command. */
public class UpdateAppCommand extends Command {
    /** The action command this command will issue. */
    public static final String           CMD_CHECK_FOR_UPDATE = "CheckForAppUpdate";
    /** The singleton {@link UpdateAppCommand}. */
    public static final UpdateAppCommand INSTANCE             = new UpdateAppCommand();

    private UpdateAppCommand() {
        super(UpdateChecker.getAppResult(), CMD_CHECK_FOR_UPDATE);
    }

    @Override
    public void adjust() {
        setTitle(UpdateChecker.getAppResult());
        setEnabled(UpdateChecker.isNewAppVersionAvailable());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        UpdateChecker.goToUpdate();
    }
}
