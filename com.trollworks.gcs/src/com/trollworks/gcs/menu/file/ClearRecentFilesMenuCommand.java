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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.util.ArrayList;

/** Provides the "Clear" command in the {@link RecentFilesMenu}. */
public final class ClearRecentFilesMenuCommand extends Command {

    /** The action command this command will issue. */
    public static final String CMD_CLEAR_RECENT_FILES_MENU = "ClearRecentFilesMenu";

    /** The singleton {@link ClearRecentFilesMenuCommand}. */
    public static final ClearRecentFilesMenuCommand INSTANCE = new ClearRecentFilesMenuCommand();

    private ClearRecentFilesMenuCommand() {
        super(I18n.text("Clear"), CMD_CLEAR_RECENT_FILES_MENU);
    }

    @Override
    public void adjust() {
        setEnabled(!Settings.getInstance().getRecentFiles().isEmpty());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Settings.getInstance().setRecentFiles(new ArrayList<>());
    }
}
