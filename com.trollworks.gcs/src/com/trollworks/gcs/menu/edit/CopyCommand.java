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

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import javax.swing.text.JTextComponent;

/** Provides the "Copy" command. */
public class CopyCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_COPY = "Copy";

    /** The singleton {@link CopyCommand}. */
    public static final CopyCommand INSTANCE = new CopyCommand();

    private CopyCommand() {
        super(I18n.Text("Copy"), CMD_COPY, KeyEvent.VK_C);
    }

    @Override
    public void adjust() {
        boolean   enable = false;
        Component comp   = getFocusOwner();
        if (comp instanceof JTextComponent) {
            JTextComponent textComp = (JTextComponent) comp;
            enable = textComp.getSelectionStart() != textComp.getSelectionEnd();
        } else {
            Copyable copyable = getTarget(Copyable.class);
            if (copyable != null) {
                enable = copyable.canCopySelection();
            }
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component comp = getFocusOwner();
        if (comp instanceof JTextComponent) {
            ((JTextComponent) comp).copy();
        } else {
            Copyable copyable = getTarget(Copyable.class);
            if (copyable != null && copyable.canCopySelection()) {
                copyable.copySelection();
            }
        }
    }
}
