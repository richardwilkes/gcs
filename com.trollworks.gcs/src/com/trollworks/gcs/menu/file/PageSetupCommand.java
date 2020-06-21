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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Page Setup..." command. */
public class PageSetupCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_PAGE_SETUP = "PageSetup";

    /** The singleton {@link PageSetupCommand}. */
    public static final PageSetupCommand INSTANCE = new PageSetupCommand();

    private PageSetupCommand() {
        super(I18n.Text("Page Setup…"), CMD_PAGE_SETUP, KeyEvent.VK_P, SHIFTED_COMMAND_MODIFIER);
    }

    @Override
    public void adjust() {
        setEnabled(getTarget(PrintProxy.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        PrintProxy proxy = getTarget(PrintProxy.class);
        if (proxy != null) {
            PrintManager mgr = proxy.getPrintManager();
            if (mgr != null) {
                mgr.pageSetup(proxy);
            } else {
                WindowUtils.showError(UIUtilities.getComponentForDialog(proxy), I18n.Text("There is no system printer available."));
            }
        }
    }
}
