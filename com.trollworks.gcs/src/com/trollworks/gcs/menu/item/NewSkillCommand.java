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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillsDockable;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Skill" command. */
public class NewSkillCommand extends Command {
    /** The action command this command will issue. */
    public static final String          CMD_NEW_SKILL           = "NewSkill";
    /** The action command this command will issue. */
    public static final String          CMD_NEW_SKILL_CONTAINER = "NewSkillContainer";
    /** The action command this command will issue. */
    public static final String          CMD_NEW_TECHNIQUE       = "NewTechnique";
    /** The "New Skill" command. */
    public static final NewSkillCommand INSTANCE                = new NewSkillCommand(false, false, I18n.Text("New Skill"), CMD_NEW_SKILL, KeyEvent.VK_K, COMMAND_MODIFIER);
    /** The "New Skill Container" command. */
    public static final NewSkillCommand CONTAINER_INSTANCE      = new NewSkillCommand(true, false, I18n.Text("New Skill Container"), CMD_NEW_SKILL_CONTAINER, KeyEvent.VK_K, SHIFTED_COMMAND_MODIFIER);
    /** The "New Technique" command. */
    public static final NewSkillCommand TECHNIQUE_INSTANCE      = new NewSkillCommand(false, true, I18n.Text("New Technique"), CMD_NEW_TECHNIQUE, KeyEvent.VK_T, COMMAND_MODIFIER);
    private             boolean         mContainer;
    private             boolean         mTechnique;

    private NewSkillCommand(boolean container, boolean isTechnique, String title, String cmd, int keyCode, int modifiers) {
        super(title, cmd, keyCode, modifiers);
        mContainer = container;
        mTechnique = isTechnique;
    }

    @Override
    public void adjust() {
        SkillsDockable skills = getTarget(SkillsDockable.class);
        if (skills != null) {
            setEnabled(!skills.getOutline().getModel().isLocked());
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                setEnabled(true);
            } else {
                setEnabled(getTarget(TemplateDockable.class) != null);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline    outline;
        DataFile       dataFile;
        SkillsDockable skills = getTarget(SkillsDockable.class);
        if (skills != null) {
            dataFile = skills.getDataFile();
            outline = skills.getOutline();
            if (outline.getModel().isLocked()) {
                return;
            }
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                dataFile = sheet.getDataFile();
                outline = sheet.getSheet().getSkillOutline();
            } else {
                TemplateDockable template = getTarget(TemplateDockable.class);
                if (template != null) {
                    dataFile = template.getDataFile();
                    outline = template.getTemplate().getSkillOutline();
                } else {
                    return;
                }
            }
        }
        Skill skill = mTechnique ? new Technique(dataFile) : new Skill(dataFile, mContainer);
        outline.addRow(skill, getTitle(), false);
        outline.getModel().select(skill, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
