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

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.desktop.OpenFilesEvent;
import java.awt.desktop.OpenFilesHandler;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;
import java.io.IOException;
import java.nio.file.Path;

/** Provides the "Open..." command. */
public class OpenCommand extends Command implements OpenFilesHandler {
    /** The action command this command will issue. */
    public static final String CMD_OPEN = "Open";

    /** The singleton {@link OpenCommand}. */
    public static final OpenCommand INSTANCE = new OpenCommand();

    private OpenCommand() {
        super(I18n.Text("Open…"), CMD_OPEN, KeyEvent.VK_O);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        open();
    }

    /** Ask the user to open a file. */
    public static void open() {
        Path path = StdFileDialog.showOpenDialog(getFocusOwner(), I18n.Text("Open…"), FileType.createFileFilters(I18n.Text("All Readable Files"), FileType.ALL_OPENABLE.toArray(new FileType[0])));
        if (path != null) {
            open(path);
        }
    }

    /** @param path The file to open. */
    public static void open(Path path) {
        if (path != null) {
            try {
                LibraryExplorerDockable library = LibraryExplorerDockable.get();
                FileProxy proxy = library == null ? null : library.open(path);
                if (proxy != null) {
                    proxy.toFrontAndFocus();
                    Preferences.getInstance().addRecentFile(path);
                } else {
                    throw new IOException(I18n.Text("unknown file extension"));
                }
            } catch (Exception exception) {
                Log.error(exception);
                StdFileDialog.showCannotOpenMsg(getFocusOwner(), path.toString(), exception);
            }
        }
    }

    @Override
    public void openFiles(OpenFilesEvent event) {
        for (File file : event.getFiles()) {
            // We call this rather than directly to open(Path) above to allow the file opening to be
            // deferred until startup has finished
            OpenDataFileCommand.open(file.toPath());
        }
    }
}
