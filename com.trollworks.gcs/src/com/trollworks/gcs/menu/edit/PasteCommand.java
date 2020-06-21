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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.awt.Component;
import java.awt.datatransfer.DataFlavor;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import javax.swing.text.JTextComponent;

/** Provides the "Paste" command. */
public class PasteCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_PASTE = "Paste";

    /** The singleton {@link PasteCommand}. */
    public static final PasteCommand INSTANCE = new PasteCommand();

    private PasteCommand() {
        super(I18n.Text("Paste"), CMD_PASTE, KeyEvent.VK_V);
    }

    @Override
    public void adjust() {
        boolean   enable = false;
        Component comp   = getFocusOwner();
        if (comp instanceof JTextComponent && comp.isEnabled()) {
            JTextComponent textComp = (JTextComponent) comp;
            if (textComp.isEditable()) {
                try {
                    enable = comp.getToolkit().getSystemClipboard().isDataFlavorAvailable(DataFlavor.stringFlavor);
                } catch (Exception exception) {
                    Log.warn(exception);
                }
            }
        } else {
            Pastable pastable = getTarget(Pastable.class);
            if (pastable != null) {
                enable = pastable.canPasteSelection();
            }
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component comp = getFocusOwner();
        if (comp instanceof JTextComponent) {
            ((JTextComponent) comp).paste();
        } else {
            Pastable pastable = getTarget(Pastable.class);
            if (pastable != null && pastable.canPasteSelection()) {
                pastable.pasteSelection();
            }
        }
    }
}
