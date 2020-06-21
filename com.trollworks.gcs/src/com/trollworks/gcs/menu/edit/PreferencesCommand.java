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
import com.trollworks.gcs.preferences.PreferencesWindow;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.desktop.PreferencesEvent;
import java.awt.desktop.PreferencesHandler;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Preferences..." command. */
public class PreferencesCommand extends Command implements PreferencesHandler {
    /** The action command this command will issue. */
    public static final String CMD_PREFERENCES = "Preferences";

    /** The singleton {@link PreferencesCommand}. */
    public static final PreferencesCommand INSTANCE = new PreferencesCommand();

    private PreferencesCommand() {
        super(I18n.Text("Preferences…"), CMD_PREFERENCES, KeyEvent.VK_COMMA);
    }

    @Override
    public void adjust() {
        setEnabled(!UIUtilities.inModalState());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        PreferencesWindow.display();
    }

    @Override
    public void handlePreferences(PreferencesEvent event) {
        PreferencesWindow.display();
    }
}
