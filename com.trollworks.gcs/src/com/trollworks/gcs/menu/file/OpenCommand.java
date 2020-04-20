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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.common.BaseWindow;
import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;

import java.awt.desktop.OpenFilesEvent;
import java.awt.desktop.OpenFilesHandler;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;
import java.io.IOException;

/** Provides the "Open..." command. */
public class OpenCommand extends Command implements OpenFilesHandler {
    /** The action command this command will issue. */
    public static final String CMD_OPEN = "Open";

    /** The singleton {@link OpenCommand}. */
    public static final OpenCommand INSTANCE = new OpenCommand();

    private OpenCommand() {
        super(I18n.Text("Open\u2026"), CMD_OPEN, KeyEvent.VK_O);
    }

    @Override
    public void adjust() {
        // Do nothing. Always enabled.
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        open();
    }

    /** Ask the user to open a file. */
    public static void open() {
        open(StdFileDialog.showOpenDialog(getFocusOwner(), I18n.Text("Open\u2026"), FileType.createFileFilters(I18n.Text("All Readable Files"), FileType.OPENABLE)));
    }

    /** @param file The file to open. */
    public static void open(File file) {
        if (file != null) {
            try {
                FileProxy proxy = BaseWindow.findFileProxy(file);
                if (proxy == null) {
                    LibraryExplorerDockable library = LibraryExplorerDockable.get();
                    if (library != null) {
                        proxy = library.open(file.toPath());
                    }
                }
                if (proxy != null) {
                    proxy.toFrontAndFocus();
                    RecentFilesMenu.addRecent(file);
                } else {
                    throw new IOException(I18n.Text("unknown file extension"));
                }
            } catch (Exception exception) {
                Log.error(exception);
                StdFileDialog.showCannotOpenMsg(getFocusOwner(), file.getName(), exception);
            }
        }
    }

    @Override
    public void openFiles(OpenFilesEvent event) {
        for (File file : event.getFiles()) {
            // We call this rather than directly to open(File) above to allow the file opening to be
            // deferred until startup has finished
            OpenDataFileCommand.open(file);
        }
    }
}
