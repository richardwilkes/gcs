/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.nio.file.Path;

/** Provides the "Save As..." command. */
public final class SaveAsCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_SAVE_AS = "SaveAs";

    /** The singleton {@link SaveAsCommand}. */
    public static final SaveAsCommand INSTANCE = new SaveAsCommand();

    private SaveAsCommand() {
        super(I18n.text("Save As…"), CMD_SAVE_AS, KeyEvent.VK_S, SHIFTED_COMMAND_MODIFIER);
    }

    @Override
    public void adjust() {
        setEnabled(!UIUtilities.inModalState() && getTarget(Saveable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        saveAs(getTarget(Saveable.class));
    }

    /**
     * Allows the user to save the file under another name.
     *
     * @param saveable The {@link Saveable} to work on.
     */
    public static void saveAs(Saveable saveable) {
        if (saveable != null) {
            FileType fileType = saveable.getFileType();
            String   name     = PathUtils.cleanNameForFile(saveable.getSaveTitle());
            if (name.isBlank()) {
                name = fileType.getUntitledDefaultFileName();
            }
            Path path = Modal.presentSaveFileDialog(UIUtilities.getComponentForDialog(saveable),
                    I18n.text("Save As…"), Dirs.GENERAL, name, fileType.getFilter());
            if (path != null && saveable.saveTo(path)) {
                Settings.getInstance().addRecentFile(path);
                LibraryExplorerDockable explorer = LibraryExplorerDockable.get();
                if (explorer != null) {
                    explorer.refresh();
                }
            }
        }
    }
}
