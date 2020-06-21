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
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import javax.swing.KeyStroke;
import javax.swing.text.JTextComponent;

/** Provides the "Delete" command. */
public class DeleteCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_DELETE = "Delete";

    /** The singleton {@link DeleteCommand}. */
    public static final DeleteCommand INSTANCE = new DeleteCommand();

    private DeleteCommand() {
        super(I18n.Text("Delete"), CMD_DELETE);
    }

    @Override
    public void adjust() {
        boolean   enable = false;
        Component comp   = getFocusOwner();
        if (comp instanceof JTextComponent && comp.isEnabled()) {
            JTextComponent textComp = (JTextComponent) comp;
            if (textComp.isEditable()) {
                int selectionEnd = textComp.getSelectionEnd();
                enable = textComp.getSelectionStart() != selectionEnd || selectionEnd < textComp.getDocument().getLength();
            }
        } else {
            Deletable deletable = getTarget(Deletable.class);
            if (deletable != null) {
                enable = deletable.canDeleteSelection();
            }
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component comp = getFocusOwner();
        if (comp instanceof JTextComponent && comp.isEnabled()) {
            JTextComponent textComp = (JTextComponent) comp;
            ActionListener listener = textComp.getActionForKeyStroke(KeyStroke.getKeyStroke(KeyEvent.VK_DELETE, 0));
            if (listener != null) {
                listener.actionPerformed(event);
            }
        } else {
            Deletable deletable = getTarget(Deletable.class);
            if (deletable != null && deletable.canDeleteSelection()) {
                deletable.deleteSelection();
            }
        }
    }
}
