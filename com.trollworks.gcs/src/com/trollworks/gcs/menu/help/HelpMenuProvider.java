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

import com.trollworks.gcs.GCS;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** Provides the standard "Help" menu. */
public class HelpMenuProvider {
    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Help"));
        if (!Platform.isMacintosh()) {
            menu.add(new JMenuItem(AboutCommand.INSTANCE));
            menu.addSeparator();
        }
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Sponsor GCS Development"), "https://github.com/sponsors/richardwilkes")));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Make A Donation For GCS Development"), "https://paypal.me/GURPSCharacterSheet")));
        menu.addSeparator();
        menu.add(new JMenuItem(UpdateAppCommand.INSTANCE));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Release Notes"), "https://github.com/richardwilkes/gcs/releases")));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("License"), "https://github.com/richardwilkes/gcs/blob/master/LICENSE")));
        menu.addSeparator();
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Web Site"), GCS.WEB_SITE)));
        menu.add(new JMenuItem(new OpenURICommand(I18n.Text("Mailing Lists"), "https://groups.io/g/gcs")));
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
