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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Open Item" command. */
public class OpenItemCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_OPEN_ITEM = "Open Item";

    /** The singleton {@link OpenItemCommand}. */
    public static final OpenItemCommand INSTANCE = new OpenItemCommand();

    private OpenItemCommand() {
        super(I18n.Text("Open Item"), CMD_OPEN_ITEM, KeyEvent.VK_ENTER);
    }

    @Override
    public void adjust() {
        boolean  enable   = false;
        Openable openable = getTarget(Openable.class);
        if (openable != null) {
            enable = openable.canOpenSelection();
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Openable openable = getTarget(Openable.class);
        if (openable != null) {
            openable.openSelection();
        }
    }
}
