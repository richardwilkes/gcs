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

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.StdMenuBar;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import javax.swing.JOptionPane;

/** Update the system Library folder. */
public class UpdateSystemLibraryCommand extends Command {
    /** Creates a new {@link UpdateSystemLibraryCommand}. */
    public UpdateSystemLibraryCommand() {
        super(I18n.Text("Update Master Library"), "update_master_library");
    }

    @Override
    public void adjust() {
        setEnabled(!StdMenuBar.SUPRESS_MENUS);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        askUserToUpdate();
    }

    public static void askUserToUpdate() {
        String no = I18n.Text("No");
        if (WindowUtils.showConfirmDialog(null, I18n.Text("Update the Master Library to the latest content?\n\nNote that any existing content will be removed and replaced.\nContent in the User Library will not be modified."), I18n.Text("Master Library Update"), JOptionPane.OK_CANCEL_OPTION, new String[]{I18n.Text("Update"), no}, no) == JOptionPane.OK_OPTION) {
            Library.download();
        }
    }
}
