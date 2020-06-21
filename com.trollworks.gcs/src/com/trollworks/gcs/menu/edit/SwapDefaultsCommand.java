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
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Swaps the default values of two skills that default to each other */
public class SwapDefaultsCommand extends Command {
    /** The action command this command will issue. */
    public static final String              CMD_SWAP_DEFAULTS = "SwapDefaults";
    /** The singleton {@link SwapDefaultsCommand}. */
    public static final SwapDefaultsCommand INSTANCE          = new SwapDefaultsCommand();

    private SwapDefaultsCommand() {
        super(I18n.Text("Swap Defaults"), CMD_SWAP_DEFAULTS, KeyEvent.VK_X, Command.SHIFTED_COMMAND_MODIFIER);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof SkillOutline) {
            SkillOutline skillOutline = (SkillOutline) focus;
            setEnabled(skillOutline.canSwapDefaults());
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
        ((SkillOutline) focus).swapDefaults();
    }

}
