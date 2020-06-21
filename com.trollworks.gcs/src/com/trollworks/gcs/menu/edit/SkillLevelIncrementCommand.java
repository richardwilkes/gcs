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

/** Provides "Increase Skill Level" command */
public class SkillLevelIncrementCommand extends Command {
    public static final String                     CMD_INCREASE_LEVEL = "IncreaseLevel";
    public static final SkillLevelIncrementCommand INSTANCE           = new SkillLevelIncrementCommand();

    private SkillLevelIncrementCommand() {
        super(I18n.Text("Increase Skill Level"), CMD_INCREASE_LEVEL, KeyEvent.VK_SLASH);
    }

    @Override
    public void adjust() {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        if (focus instanceof SkillLevelIncrementable) {
            SkillLevelIncrementable inc = (SkillLevelIncrementable) focus;
            setEnabled(inc.canIncrementSkillLevel());
            setTitle(inc.getIncrementSkillLevelTitle());
        } else {
            setEnabled(false);
            setTitle(I18n.Text("Increase Skill Level"));
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Component focus = getFocusOwner();
        if (focus instanceof OutlineProxy) {
            focus = ((OutlineProxy) focus).getRealOutline();
        }
        ((SkillLevelIncrementable) focus).incrementSkillLevel();
    }

}
