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

/** Provides the "Decrement" command. */
public class DecrementCommand extends Command {
    /** The action command this command will issue. */
    public static final String           CMD_DECREMENT = "Decrement";
    /** The singleton {@link DecrementCommand}. */
    public static final DecrementCommand INSTANCE      = new DecrementCommand();

    private DecrementCommand() {
        super(I18n.Text("Decrement"), CMD_DECREMENT, KeyEvent.VK_MINUS);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof Incrementable) {
            Incrementable inc = (Incrementable) focus;
            setTitle(inc.getDecrementTitle());
            setEnabled(inc.canDecrement());
        } else {
            setTitle(I18n.Text("Decrement"));
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ((Incrementable) focus).decrement();
    }
}
