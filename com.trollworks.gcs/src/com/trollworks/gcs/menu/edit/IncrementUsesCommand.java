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

package com.trollworks.gcs.menu.edit;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Increment Uses" command. */
public class IncrementUsesCommand extends Command {
    /** The action command this command will issue. */
    public static final String               CMD_INCREMENT_USES = "IncrementUses";
    /** The singleton {@link IncrementUsesCommand}. */
    public static final IncrementUsesCommand INSTANCE           = new IncrementUsesCommand();

    private IncrementUsesCommand() {
        super(I18n.Text("Increment Uses"), CMD_INCREMENT_USES, KeyEvent.VK_UP);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof UsesIncrementable) {
            setEnabled(((UsesIncrementable) focus).canIncrementUses());
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
        ((UsesIncrementable) focus).incrementUses();
    }
}
