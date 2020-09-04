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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.menu.MenuHelpers;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

/** Provides the "Item" menu. */
public final class ItemMenuProvider {
    private ItemMenuProvider() {
    }

    public static List<Command> getModifiableCommands() {
        List<Command> cmds = new ArrayList<>();
        cmds.add(OpenEditorCommand.INSTANCE);
        cmds.add(CopyToSheetCommand.INSTANCE);
        cmds.add(CopyToTemplateCommand.INSTANCE);
        cmds.add(ApplyTemplateCommand.INSTANCE);
        cmds.add(OpenPageReferenceCommand.OPEN_ONE_INSTANCE);
        cmds.add(OpenPageReferenceCommand.OPEN_EACH_INSTANCE);
        cmds.add(NewAdvantageCommand.INSTANCE);
        cmds.add(NewAdvantageCommand.CONTAINER_INSTANCE);
        cmds.add(NewSkillCommand.INSTANCE);
        cmds.add(NewSkillCommand.CONTAINER_INSTANCE);
        cmds.add(NewSkillCommand.TECHNIQUE_INSTANCE);
        cmds.add(NewSpellCommand.INSTANCE);
        cmds.add(NewSpellCommand.CONTAINER_INSTANCE);
        cmds.add(NewSpellCommand.RITUAL_MAGIC_INSTANCE);
        cmds.add(NewEquipmentCommand.CARRIED_INSTANCE);
        cmds.add(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE);
        cmds.add(NewEquipmentCommand.NOT_CARRIED_INSTANCE);
        cmds.add(NewEquipmentCommand.NOT_CARRIED_CONTAINER_INSTANCE);
        cmds.add(NewNoteCommand.INSTANCE);
        cmds.add(NewNoteCommand.CONTAINER_INSTANCE);
        cmds.add(AddNaturalAttacksAdvantageCommand.INSTANCE);
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Item"));
        menu.add(MenuHelpers.createSubMenu(I18n.Text("Advantages"), NewAdvantageCommand.INSTANCE, NewAdvantageCommand.CONTAINER_INSTANCE, null, NewAdvantageModifierCommand.INSTANCE, NewAdvantageModifierCommand.CONTAINER_INSTANCE, null, AddNaturalAttacksAdvantageCommand.INSTANCE));
        menu.add(MenuHelpers.createSubMenu(I18n.Text("Skills"), NewSkillCommand.INSTANCE, NewSkillCommand.CONTAINER_INSTANCE, null, NewSkillCommand.TECHNIQUE_INSTANCE));
        menu.add(MenuHelpers.createSubMenu(I18n.Text("Spells"), NewSpellCommand.INSTANCE, NewSpellCommand.CONTAINER_INSTANCE, null, NewSpellCommand.RITUAL_MAGIC_INSTANCE));
        menu.add(MenuHelpers.createSubMenu(I18n.Text("Equipment"), NewEquipmentCommand.CARRIED_INSTANCE, NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE, null, NewEquipmentCommand.NOT_CARRIED_INSTANCE, NewEquipmentCommand.NOT_CARRIED_CONTAINER_INSTANCE, null, NewEquipmentModifierCommand.INSTANCE, NewEquipmentModifierCommand.CONTAINER_INSTANCE));
        menu.add(MenuHelpers.createSubMenu(I18n.Text("Notes"), NewNoteCommand.INSTANCE, NewNoteCommand.CONTAINER_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(OpenEditorCommand.INSTANCE));
        menu.add(new DynamicMenuItem(CopyToSheetCommand.INSTANCE));
        menu.add(new DynamicMenuItem(CopyToTemplateCommand.INSTANCE));
        menu.add(new DynamicMenuItem(ApplyTemplateCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(OpenPageReferenceCommand.OPEN_ONE_INSTANCE));
        menu.add(new DynamicMenuItem(OpenPageReferenceCommand.OPEN_EACH_INSTANCE));
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
