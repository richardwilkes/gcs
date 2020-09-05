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
import com.trollworks.gcs.utility.FileType;

import java.awt.EventQueue;
import java.awt.event.ActionEvent;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import javax.swing.SwingUtilities;

/** A command that will open a specific data file. */
public class OpenDataFileCommand extends Command implements Runnable {
    private static final String     CMD_PREFIX  = "OpenDataFile[";
    private static final String     CMD_POSTFIX = "]";
    private static       boolean    PASS_THROUGH;
    private static       List<Path> PENDING;
    private              Path       mPath;
    private              boolean    mVerify;

    /** @param path The file to open. */
    public static synchronized void open(Path path) {
        if (PASS_THROUGH) {
            OpenDataFileCommand opener = new OpenDataFileCommand(path);
            if (SwingUtilities.isEventDispatchThread()) {
                opener.run();
            } else {
                EventQueue.invokeLater(opener);
            }
        } else {
            if (PENDING == null) {
                PENDING = new ArrayList<>();
            }
            PENDING.add(path);
        }
    }

    /**
     * Enables the pass-through mode so that future calls to {@link #open(Path)} will no longer
     * queue files for later opening. All queued files will now be opened.
     */
    public static synchronized void enablePassThrough() {
        PASS_THROUGH = true;
        if (PENDING != null) {
            for (Path path : PENDING) {
                open(path);
            }
            PENDING = null;
        }
    }

    /**
     * Creates a new {@link OpenDataFileCommand}.
     *
     * @param title The title to use.
     * @param path  The file to open.
     */
    public OpenDataFileCommand(String title, Path path) {
        super(title, CMD_PREFIX + path + CMD_POSTFIX, FileType.getIconForFileName(path.getFileName().toString()));
        mPath = path;
    }

    /**
     * Creates a new {@link OpenDataFileCommand} that can only be invoked successfully if {@link
     * OpenCommand} is enabled.
     *
     * @param path The file to open.
     */
    public OpenDataFileCommand(Path path) {
        super(path.getFileName().toString(), CMD_PREFIX + path + CMD_POSTFIX);
        mPath = path;
        mVerify = true;
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        run();
    }

    @Override
    public void run() {
        if (mVerify) {
            OpenCommand.INSTANCE.adjust();
            if (!OpenCommand.INSTANCE.isEnabled()) {
                return;
            }
        }
        OpenCommand.open(mPath);
    }
}
