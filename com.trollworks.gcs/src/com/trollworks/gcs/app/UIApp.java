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

package com.trollworks.gcs.app;

import com.trollworks.gcs.common.Workspace;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.menu.edit.PreferencesCommand;
import com.trollworks.gcs.menu.file.OpenCommand;
import com.trollworks.gcs.menu.file.OpenDataFileCommand;
import com.trollworks.gcs.menu.file.PrintCommand;
import com.trollworks.gcs.menu.file.QuitCommand;
import com.trollworks.gcs.menu.help.AboutCommand;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Desktop;
import java.awt.Desktop.Action;
import java.awt.EventQueue;
import java.awt.desktop.QuitStrategy;
import java.nio.file.Path;
import java.util.List;

class UIApp {
    private static boolean NOTIFICATION_ALLOWED;

    static void startup(LaunchProxy launchProxy, List<Path> files) {
        if (Desktop.isDesktopSupported()) {
            Desktop desktop = Desktop.getDesktop();
            if (desktop.isSupported(Action.APP_ABOUT)) {
                desktop.setAboutHandler(AboutCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_PREFERENCES)) {
                desktop.setPreferencesHandler(PreferencesCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_OPEN_FILE)) {
                desktop.setOpenFileHandler(OpenCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_PRINT_FILE)) {
                desktop.setPrintFileHandler(PrintCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_QUIT_HANDLER)) {
                desktop.setQuitHandler(QuitCommand.INSTANCE);
            }
            if (desktop.isSupported(Action.APP_QUIT_STRATEGY)) {
                desktop.setQuitStrategy(QuitStrategy.NORMAL_EXIT);
            }
            if (desktop.isSupported(Action.APP_SUDDEN_TERMINATION)) {
                desktop.disableSuddenTermination();
            }
        }
        UpdateChecker.check();
        OutputPreferences.initialize(); // Must come before SheetPreferences.initialize()
        SheetPreferences.initialize();
        launchProxy.setReady(true);
        EventQueue.invokeLater(() -> {
            Workspace.get();
            OpenDataFileCommand.enablePassThrough();
            for (Path file : files) {
                OpenDataFileCommand.open(file.toFile());
            }
            if (Desktop.isDesktopSupported()) {
                Desktop desktop = Desktop.getDesktop();
                if (desktop.isSupported(Action.APP_MENU_BAR)) {
                    desktop.setDefaultMenuBar(new StdMenuBar());
                }
            }
            if (Platform.isMacintosh() && System.getProperty("java.home").toLowerCase().contains("/apptranslocation/")) {
                WindowUtils.showError(null, Text.wrapToCharacterCount(I18n.Text("macOS has translocated GCS, restricting access to the file system and preventing access to the data library. To fix this, you must quit GCS, then run the following command in the terminal after cd'ing into the GURPS Character Sheet folder:\n\n"), 60) + "xattr -d com.apple.quarantine \"GURPS Character Sheet.app\"");
            }
            setNotificationAllowed(true);
        });
    }

    /** @return Whether it is OK to put up a notification dialog yet. */
    static synchronized boolean isNotificationAllowed() {
        return NOTIFICATION_ALLOWED;
    }

    /** @param allowed Whether it is OK to put up a notification dialog yet. */
    static synchronized void setNotificationAllowed(boolean allowed) {
        NOTIFICATION_ALLOWED = allowed;
    }
}
