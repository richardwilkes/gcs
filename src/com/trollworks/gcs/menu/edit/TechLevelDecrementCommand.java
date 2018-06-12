/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.outline.OutlineProxy;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Decrement" command. */
public class TechLevelDecrementCommand extends Command {
    @Localize("Decrement Tech Level")
    private static String TITLE;

    static {
        Localization.initialize();
    }

    /** The action command this command will issue. */
    public static final String                    CMD_DECREMENT_TL = "DecrementTL"; //$NON-NLS-1$

    /** The singleton {@link TechLevelDecrementCommand}. */
    public static final TechLevelDecrementCommand INSTANCE         = new TechLevelDecrementCommand();

    private TechLevelDecrementCommand() {
        super(TITLE, CMD_DECREMENT_TL, KeyEvent.VK_OPEN_BRACKET);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof TechLevelIncrementable) {
            TechLevelIncrementable inc = (TechLevelIncrementable) focus;
            setEnabled(inc.canDecrementTechLevel());
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
        ((TechLevelIncrementable) focus).decrementTechLevel();
    }
}
