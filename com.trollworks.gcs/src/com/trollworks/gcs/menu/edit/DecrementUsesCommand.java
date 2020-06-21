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
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Decrement Uses" command. */
public class DecrementUsesCommand extends Command {
    /** The action command this command will issue. */
    public static final String               CMD_DECREMENT_USES = "DecrementUses";
    /** The singleton {@link DecrementUsesCommand}. */
    public static final DecrementUsesCommand INSTANCE           = new DecrementUsesCommand();

    private DecrementUsesCommand() {
        super(I18n.Text("Decrement Uses"), CMD_DECREMENT_USES, KeyEvent.VK_DOWN);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof UsesIncrementable) {
            setEnabled(((UsesIncrementable) focus).canDecrementUses());
        } else {
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ((UsesIncrementable) focus).decrementUses();
    }
}
