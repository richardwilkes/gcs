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
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.Collection;
import javax.swing.JOptionPane;

/** Provides the "Save" command. */
public class SaveCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_SAVE = "Save";

    /** The singleton {@link SaveCommand}. */
    public static final SaveCommand INSTANCE = new SaveCommand();

    private SaveCommand() {
        super(I18n.Text("Save"), CMD_SAVE, KeyEvent.VK_S);
    }

    @Override
    public void adjust() {
        Saveable saveable = getTarget(Saveable.class);
        Commitable.sendCommitToFocusOwner();
        setEnabled(saveable != null && saveable.isModified());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        save(getTarget(Saveable.class));
    }

    /**
     * Makes an attempt to save the specified {@link Saveable}s if any have been modified.
     *
     * @param saveables The {@link Saveable}s to work on.
     * @return {@code false} if a save was cancelled or failed.
     */
    public static boolean attemptSave(Collection<Saveable> saveables) {
        Commitable.sendCommitToFocusOwner();
        for (Saveable saveable : saveables) {
            if (!attemptSaveInternal(saveable)) {
                return false;
            }
        }
        return true;
    }

    /**
     * Makes an attempt to save the specified {@link Saveable} if it has been modified.
     *
     * @param saveable The {@link Saveable} to work on.
     * @return {@code false} if the save was cancelled or failed.
     */
    public static boolean attemptSave(Saveable saveable) {
        if (saveable != null) {
            Commitable.sendCommitToFocusOwner();
            return attemptSaveInternal(saveable);
        }
        return true;
    }

    private static boolean attemptSaveInternal(Saveable saveable) {
        if (saveable.isModified()) {
            saveable.toFrontAndFocus();
            int answer = JOptionPane.showConfirmDialog(UIUtilities.getComponentForDialog(saveable), MessageFormat.format(I18n.Text("Save changes to \"{0}\"?"), saveable.getSaveTitle()), I18n.Text("Save"), JOptionPane.YES_NO_CANCEL_OPTION);
            if (answer == JOptionPane.CANCEL_OPTION || answer == JOptionPane.CLOSED_OPTION) {
                return false;
            }
            if (answer == JOptionPane.YES_OPTION) {
                save(saveable);
                return !saveable.isModified();
            }
        }
        return true;
    }

    /**
     * Allows the user to save the file.
     *
     * @param saveable The {@link Saveable} to work on.
     * @return The path(s) actually written to. May be empty.
     */
    public static Path[] save(Saveable saveable) {
        if (saveable == null) {
            return new Path[0];
        }
        Path path = saveable.getBackingFile();
        if (path != null) {
            Path[] paths = saveable.saveTo(path);
            for (Path one : paths) {
                Preferences.getInstance().addRecentFile(one);
            }
            return paths;
        }
        return SaveAsCommand.saveAs(saveable);
    }
}
