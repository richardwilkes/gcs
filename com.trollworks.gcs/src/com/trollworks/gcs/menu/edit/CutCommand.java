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

/** Provides the "Cut" command. */
public class CutCommand extends Command {
    /** The action command this command will issue. */
    public static final String CMD_CUT = "Cut";

    /** The singleton {@link CutCommand}. */
    public static final CutCommand INSTANCE = new CutCommand();

    private CutCommand() {
        super(I18n.Text("Cut"), CMD_CUT, KeyEvent.VK_X);
    }

    @Override
    public void adjust() {
        boolean   enable = false;
        Component comp   = getFocusOwner();
        if (comp instanceof JTextComponent && comp.isEnabled()) {
            JTextComponent textComp = (JTextComponent) comp;
            if (textComp.isEditable()) {
                enable = textComp.getSelectionStart() != textComp.getSelectionEnd();
            }
        } else {
            Cutable cutable = getTarget(Cutable.class);
            if (cutable != null) {
                enable = cutable.canCutSelection();
            }
        }
        setEnabled(enable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component comp = getFocusOwner();
        if (comp instanceof JTextComponent) {
            ((JTextComponent) comp).cut();
        } else {
            Cutable cutable = getTarget(Cutable.class);
            if (cutable != null && cutable.canCutSelection()) {
                cutable.cutSelection();
            }
        }
    }
}
