/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Duplicate" command. */
public final class DuplicateCommand extends Command {
    /** The action command this command will issue. */
    public static final String           CMD_DUPLICATE = "Duplicate";
    /** The singleton {@link DuplicateCommand}. */
    public static final DuplicateCommand INSTANCE      = new DuplicateCommand();

    private DuplicateCommand() {
        super(I18n.text("Duplicate"), CMD_DUPLICATE, KeyEvent.VK_U);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        setEnabled(focus instanceof Duplicatable && ((Duplicatable) focus).canDuplicateSelection());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof Duplicatable) {
            ((Duplicatable) focus).duplicateSelection();
        }
    }
}
