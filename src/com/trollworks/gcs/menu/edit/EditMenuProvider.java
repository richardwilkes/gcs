/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.edit;

import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicCheckBoxMenuItem;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.DynamicMenuItem;
import com.trollworks.toolkit.ui.menu.MenuProvider;
import com.trollworks.toolkit.ui.menu.edit.CopyCommand;
import com.trollworks.toolkit.ui.menu.edit.CutCommand;
import com.trollworks.toolkit.ui.menu.edit.DeleteCommand;
import com.trollworks.toolkit.ui.menu.edit.JumpToSearchCommand;
import com.trollworks.toolkit.ui.menu.edit.PasteCommand;
import com.trollworks.toolkit.ui.menu.edit.PreferencesCommand;
import com.trollworks.toolkit.ui.menu.edit.RedoCommand;
import com.trollworks.toolkit.ui.menu.edit.SelectAllCommand;
import com.trollworks.toolkit.ui.menu.edit.UndoCommand;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.Platform;

import java.util.HashSet;
import java.util.Set;

import javax.swing.JMenu;

/** Provides the standard "Edit" menu. */
public class EditMenuProvider implements MenuProvider {
    public static final String NAME = "Edit";

    @Override
    public Set<Command> getModifiableCommands() {
        Set<Command> cmds = new HashSet<>();
        cmds.add(UndoCommand.INSTANCE);
        cmds.add(RedoCommand.INSTANCE);
        cmds.add(CutCommand.INSTANCE);
        cmds.add(CopyCommand.INSTANCE);
        cmds.add(PasteCommand.INSTANCE);
        cmds.add(DuplicateCommand.INSTANCE);
        cmds.add(DeleteCommand.INSTANCE);
        cmds.add(SelectAllCommand.INSTANCE);
        cmds.add(IncrementCommand.INSTANCE);
        cmds.add(DecrementCommand.INSTANCE);
        cmds.add(IncrementUsesCommand.INSTANCE);
        cmds.add(DecrementUsesCommand.INSTANCE);
        cmds.add(SkillLevelIncrementCommand.INSTANCE);
        cmds.add(SkillLevelDecrementCommand.INSTANCE);
        cmds.add(TechLevelIncrementCommand.INSTANCE);
        cmds.add(TechLevelDecrementCommand.INSTANCE);
        cmds.add(ToggleStateCommand.INSTANCE);
        cmds.add(JumpToSearchCommand.INSTANCE);
        cmds.add(RandomizeDescriptionCommand.INSTANCE);
        cmds.add(RandomizeNameCommand.FEMALE_INSTANCE);
        cmds.add(RandomizeNameCommand.MALE_INSTANCE);
        cmds.add(AddNaturalPunchCommand.INSTANCE);
        cmds.add(AddNaturalKickCommand.INSTANCE);
        cmds.add(AddNaturalKickWithBootsCommand.INSTANCE);
        cmds.add(SwapDefaultsCommand.INSTANCE);
        cmds.add(ConvertToContainer.INSTANCE);
        if (!Platform.isMacintosh()) {
            cmds.add(PreferencesCommand.INSTANCE);
        }
        return cmds;
    }

    @Override
    public JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("Edit"));
        menu.setName(NAME);
        menu.add(new DynamicMenuItem(UndoCommand.INSTANCE));
        menu.add(new DynamicMenuItem(RedoCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(CutCommand.INSTANCE));
        menu.add(new DynamicMenuItem(CopyCommand.INSTANCE));
        menu.add(new DynamicMenuItem(PasteCommand.INSTANCE));
        menu.add(new DynamicMenuItem(DuplicateCommand.INSTANCE));
        menu.add(new DynamicMenuItem(DeleteCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SelectAllCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(ConvertToContainer.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(IncrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(DecrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(IncrementUsesCommand.INSTANCE));
        menu.add(new DynamicMenuItem(DecrementUsesCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SkillLevelIncrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SkillLevelDecrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(TechLevelIncrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(TechLevelDecrementCommand.INSTANCE));
        menu.add(new DynamicMenuItem(ToggleStateCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SwapDefaultsCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(JumpToSearchCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(RandomizeDescriptionCommand.INSTANCE));
        menu.add(new DynamicMenuItem(RandomizeNameCommand.FEMALE_INSTANCE));
        menu.add(new DynamicMenuItem(RandomizeNameCommand.MALE_INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicCheckBoxMenuItem(AddNaturalPunchCommand.INSTANCE));
        menu.add(new DynamicCheckBoxMenuItem(AddNaturalKickCommand.INSTANCE));
        menu.add(new DynamicCheckBoxMenuItem(AddNaturalKickWithBootsCommand.INSTANCE));
        if (!Platform.isMacintosh()) {
            menu.addSeparator();
            menu.add(new DynamicMenuItem(PreferencesCommand.INSTANCE));
        }
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
