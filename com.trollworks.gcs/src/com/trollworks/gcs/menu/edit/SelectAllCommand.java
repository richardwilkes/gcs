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

/** Provides the "Select All" command. */
public class SelectAllCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_SELECT_ALL = "SelectAll";

    /** The singleton {@link SelectAllCommand}. */
    public static final SelectAllCommand INSTANCE = new SelectAllCommand();

    private SelectAllCommand() {
        super(I18n.Text("Select All"), CMD_SELECT_ALL, KeyEvent.VK_A);
    }

    @Override
    public void adjust() {
        boolean   enable = false;
        Component comp   = getFocusOwner();
        if (comp instanceof JTextComponent && comp.isEnabled()) {
            JTextComponent textComp = (JTextComponent) comp;
            enable = textComp.getSelectionEnd() - textComp.getSelectionStart() != textComp.getDocument().getLength();
        } else {
            SelectAllCapable selectable = getTarget(SelectAllCapable.class);
            if (selectable != null) {
                enable = selectable.canSelectAll();
            }
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component comp = getFocusOwner();
        if (comp instanceof JTextComponent) {
            ((JTextComponent) comp).selectAll();
        } else {
            SelectAllCapable selectable = getTarget(SelectAllCapable.class);
            if (selectable != null && selectable.canSelectAll()) {
                selectable.selectAll();
            }
        }
    }
}
