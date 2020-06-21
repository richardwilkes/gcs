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

/** Provides the "Increment" command. */
public class IncrementCommand extends Command {
    /** The action command this command will issue. */
    public static final String           CMD_INCREMENT = "Increment";
    /** The singleton {@link IncrementCommand}. */
    public static final IncrementCommand INSTANCE      = new IncrementCommand();

    private IncrementCommand() {
        super(I18n.Text("Increment"), CMD_INCREMENT, KeyEvent.VK_EQUALS);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof Incrementable) {
            Incrementable inc = (Incrementable) focus;
            setTitle(inc.getIncrementTitle());
            setEnabled(inc.canIncrement());
        } else {
            setTitle(I18n.Text("Increment"));
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ((Incrementable) focus).increment();
    }
}
