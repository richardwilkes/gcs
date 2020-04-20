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

package com.trollworks.gcs.menu.help;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.Desktop;
import java.awt.Desktop.Action;
import java.awt.event.ActionEvent;
import java.io.File;
import java.util.Arrays;

/** Shows the user the Library folder. */
public class ShowLibraryFolderCommand extends Command {
    private boolean mSystem;

    /** Creates a new {@link ShowLibraryFolderCommand}. */
    public ShowLibraryFolderCommand(boolean system) {
        super(system ? I18n.Text("Show GCS Library on Disk") : I18n.Text("Show GCS User Library on Disk"), system ? "show_gcs_library" : "show_gcs_user_library");
        mSystem = system;
    }

    @Override
    public void adjust() {
        // Not used. Always enabled.
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        try {
            File    dir     = (mSystem ? Library.getSystemRootPath() : Library.getUserRootPath()).toFile();
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
