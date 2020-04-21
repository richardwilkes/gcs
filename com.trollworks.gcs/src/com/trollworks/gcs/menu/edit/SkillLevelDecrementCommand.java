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

/** Provides "Decrease Skill Level" command */
public class SkillLevelDecrementCommand extends Command {
    public static final String                     CMD_DECREASE_LEVEL = "DecreaseLevel";
    public static final SkillLevelDecrementCommand INSTANCE           = new SkillLevelDecrementCommand();

    private SkillLevelDecrementCommand() {
        super(I18n.Text("Decrease Skill Level"), CMD_DECREASE_LEVEL, KeyEvent.VK_PERIOD);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof SkillLevelIncrementable) {
            SkillLevelIncrementable inc = (SkillLevelIncrementable) focus;
            setEnabled(inc.canDecrementSkillLevel());
            setTitle(inc.getDecrementSkillLevelTitle());
        } else {
            setEnabled(false);
            setTitle(I18n.Text("Decrease Skill Level"));
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ((SkillLevelIncrementable) focus).decrementSkillLevel();
    }

}
