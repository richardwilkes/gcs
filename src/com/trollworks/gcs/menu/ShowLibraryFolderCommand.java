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

import com.trollworks.gcs.app.GCSCmdLine;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Desktop;
import java.awt.Desktop.Action;
import java.awt.event.ActionEvent;
import java.io.File;
import java.util.Arrays;

/** Shows the user the Library folder. */
public class ShowLibraryFolderCommand extends Command {
    /** Creates a new {@link ShowLibraryFolderCommand}. */
    public ShowLibraryFolderCommand() {
        super(I18n.Text("Show GCS Library on Disk"), "show_gcs_library");
    }

    @Override
    public void adjust() {
        // Not used. Always enabled.
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        try {
            File    dir     = GCSCmdLine.getLibraryRootPath().toFile();
            Desktop desktop = Desktop.getDesktop();
            if (desktop.isSupported(Action.BROWSE_FILE_DIR)) {
                File[] contents = dir.listFiles();
                if (contents != null && contents.length > 0) {
                    Arrays.sort(contents);
                    dir = contents[0];
                }
                desktop.browseFileDirectory(dir.getCanonicalFile());
            } else {
                desktop.open(dir);
            }
        } catch (Exception exception) {
            WindowUtils.showError(null, exception.getMessage());
        }
    }
}
