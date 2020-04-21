/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

/** Provides the "Item" menu. */
public class ItemMenuProvider {
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
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Item"));
        menu.add(new DynamicMenuItem(OpenEditorCommand.INSTANCE));
        menu.add(new DynamicMenuItem(CopyToSheetCommand.INSTANCE));
        menu.add(new DynamicMenuItem(CopyToTemplateCommand.INSTANCE));
        menu.add(new DynamicMenuItem(ApplyTemplateCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(OpenPageReferenceCommand.OPEN_ONE_INSTANCE));
        menu.add(new DynamicMenuItem(OpenPageReferenceCommand.OPEN_EACH_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(NewAdvantageCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewAdvantageCommand.CONTAINER_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(NewSkillCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewSkillCommand.CONTAINER_INSTANCE));
        menu.add(new DynamicMenuItem(NewSkillCommand.TECHNIQUE_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(NewSpellCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewSpellCommand.CONTAINER_INSTANCE));
        menu.add(new DynamicMenuItem(NewSpellCommand.RITUAL_MAGIC_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_INSTANCE));
        menu.add(new DynamicMenuItem(NewEquipmentCommand.CARRIED_CONTAINER_INSTANCE));
        menu.add(new DynamicMenuItem(NewEquipmentCommand.NOT_CARRIED_INSTANCE));
        menu.add(new DynamicMenuItem(NewEquipmentCommand.NOT_CARRIED_CONTAINER_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(NewNoteCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewNoteCommand.CONTAINER_INSTANCE));
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
