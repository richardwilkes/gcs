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

package com.trollworks.gcs.menu;

import com.trollworks.gcs.app.GCSApp;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.MenuProvider;
import com.trollworks.toolkit.ui.menu.help.AboutCommand;
import com.trollworks.toolkit.ui.menu.help.OpenURICommand;
import com.trollworks.toolkit.ui.menu.help.UpdateCommand;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.Platform;

import java.util.Collections;
import java.util.Set;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** Provides the standard "Help" menu. */
public class HelpMenuProvider implements MenuProvider {
    public static final String NAME = "Help";

    @Override
    public Set<Command> getModifiableCommands() {
        return Collections.emptySet();
    }

    @Override
    public JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Help"));
        menu.setName(NAME);
        if (!Platform.isMacintosh()) {
            menu.add(new JMenuItem(AboutCommand.INSTANCE));
            menu.addSeparator();
        }
        menu.add(new JMenuItem(UpdateCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new JMenuItem(new ShowLibraryFolderCommand()));
        menu.addSeparator();
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Release Notes"), "https://github.com/richardwilkes/gcs/releases")));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("License"), "https://github.com/richardwilkes/gcs/blob/master/LICENSE")));
        menu.addSeparator();
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Web Site"), GCSApp.WEB_SITE)));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Mailing Lists"), "https://groups.io/g/gcs")));
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
