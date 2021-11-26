/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.Collection;

/** Provides the "Save" command. */
public final class SaveCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_SAVE = "Save";

    /** The singleton {@link SaveCommand}. */
    public static final SaveCommand INSTANCE = new SaveCommand();

    private SaveCommand() {
        super(I18n.text("Save"), CMD_SAVE, KeyEvent.VK_S);
    }

    @Override
    public void adjust() {
        if (UIUtilities.inModalState()) {
            setEnabled(false);
            return;
        }
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
     * @return The result of the save attempt.
     */
    public static SaveResult attemptSave(Collection<Saveable> saveables) {
        Commitable.sendCommitToFocusOwner();
        for (Saveable saveable : saveables) {
            if (attemptSaveInternal(saveable) == SaveResult.CANCEL) {
                return SaveResult.CANCEL;
            }
        }
        return SaveResult.SUCCESS;
    }

    /**
     * Makes an attempt to save the specified {@link Saveable} if it has been modified.
     *
     * @param saveable The {@link Saveable} to work on.
     * @return The result of the save attempt.
     */
    public static SaveResult attemptSave(Saveable saveable) {
        if (saveable != null) {
            Commitable.sendCommitToFocusOwner();
            return attemptSaveInternal(saveable);
        }
        return SaveResult.SUCCESS;
    }

    private static SaveResult attemptSaveInternal(Saveable saveable) {
        if (saveable.isModified()) {
            saveable.toFrontAndFocus();
            Modal dialog = Modal.prepareToShowMessage(UIUtilities.getComponentForDialog(saveable),
                    I18n.text("Save"), MessageType.QUESTION,
                    MessageFormat.format(I18n.text("Save changes to \"{0}\"?"), saveable.getSaveTitle()));
            dialog.addButton(I18n.text("Cancel"), Modal.CLOSED);
            dialog.addButton(I18n.text("Discard"), Modal.CANCEL);
            dialog.addButton(I18n.text("Save"), Modal.OK);
            dialog.presentToUser();
            switch (dialog.getResult()) {
                case Modal.OK:
                    save(saveable);
                    if (saveable.isModified()) {
                        return SaveResult.CANCEL;
                    }
                    return SaveResult.SUCCESS;
                case Modal.CANCEL: // No
                    return SaveResult.NO_SAVE;
                default:
                    return SaveResult.CANCEL;
            }
        }
        return SaveResult.SUCCESS;
    }

    /**
     * Allows the user to save the file.
     *
     * @param saveable The {@link Saveable} to work on.
     */
    public static void save(Saveable saveable) {
        if (saveable != null) {
            Path path = saveable.getBackingFile();
            if (path != null) {
                if (saveable.saveTo(path)) {
                    Settings.getInstance().addRecentFile(path);
                }
                return;
            }
            SaveAsCommand.saveAs(saveable);
        }
    }
}
