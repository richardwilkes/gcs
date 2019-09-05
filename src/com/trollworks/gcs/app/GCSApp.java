/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.app;

import com.trollworks.gcs.common.ListCollectionThread;
import com.trollworks.gcs.common.Workspace;
import com.trollworks.gcs.menu.HelpMenuProvider;
import com.trollworks.gcs.menu.edit.EditMenuProvider;
import com.trollworks.gcs.menu.file.FileMenuProvider;
import com.trollworks.gcs.menu.item.ItemMenuProvider;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.preferences.ReferenceLookupPreferences;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.preferences.SystemPreferences;
import com.trollworks.toolkit.ui.App;
import com.trollworks.toolkit.ui.UpdateChecker;
import com.trollworks.toolkit.ui.menu.StdMenuBar;
import com.trollworks.toolkit.ui.preferences.FontPreferences;
import com.trollworks.toolkit.ui.preferences.MenuKeyPreferences;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.widget.AppWindow;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.Platform;
import com.trollworks.toolkit.utility.cmdline.CmdLine;
import com.trollworks.toolkit.utility.text.Text;

/** The main application user interface. */
public class GCSApp extends App {
    /** The one and only instance of this class. */
    public static final GCSApp INSTANCE  = new GCSApp();
    public static final String WEB_SITE  = "http://gurpscharactersheet.com";
    private static String      XATTR_CMD = "xattr -d com.apple.quarantine \"GURPS Character Sheet.app\"";

    private GCSApp() {
        super();
        AppWindow.setDefaultWindowIcons(GCSImages.getAppIcons());
    }

    @Override
    public void configureApplication(CmdLine cmdLine) {
        UpdateChecker.check("gcs", WEB_SITE + "/versions.txt", WEB_SITE);

        ListCollectionThread.get();

        StdMenuBar.configure(new FileMenuProvider(), new EditMenuProvider(), new ItemMenuProvider(), new HelpMenuProvider());
        OutputPreferences.initialize(); // Must come before SheetPreferences.initialize()
        SheetPreferences.initialize();
        PreferencesWindow.addCategory(SheetPreferences::new);
        PreferencesWindow.addCategory(OutputPreferences::new);
        PreferencesWindow.addCategory(FontPreferences::new);
        PreferencesWindow.addCategory(MenuKeyPreferences::new);
        PreferencesWindow.addCategory(ReferenceLookupPreferences::new);
        PreferencesWindow.addCategory(SystemPreferences::new);
    }

    @Override
    public void noWindowsAreOpenAtStartup(boolean finalChance) {
        Workspace.get();
    }

    @Override
    public void finalStartup() {
        super.finalStartup();
        setDefaultMenuBar(new StdMenuBar());
        if (Platform.isMacintosh() && App.getHomePath().toString().toLowerCase().contains("/apptranslocation/")) {
            WindowUtils.showError(null, Text.wrapToCharacterCount(I18n.Text("macOS has translocated GCS, restricting access to the file system and preventing access to the data library. To fix this, you must quit GCS, then run the following command in the terminal after cd'ing into the GURPS Character Sheet folder:\n\n"), 60) + XATTR_CMD);
        }
    }
}
