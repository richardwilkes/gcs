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
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

public final class EditMenuProvider {
    private EditMenuProvider() {
    }

    public static List<Command> getModifiableCommands() {
        List<Command> cmds = new ArrayList<>();
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
        cmds.add(SwapDefaultsCommand.INSTANCE);
        cmds.add(ConvertToContainer.INSTANCE);
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.text("Edit"));
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
        JMenu stateMenu = new JMenu(I18n.text("State…"));
        stateMenu.add(new DynamicMenuItem(ToggleStateCommand.INSTANCE));
        stateMenu.addSeparator();
        stateMenu.add(new DynamicMenuItem(IncrementCommand.INSTANCE));
        stateMenu.add(new DynamicMenuItem(DecrementCommand.INSTANCE));
        stateMenu.addSeparator();
        stateMenu.add(new DynamicMenuItem(IncrementUsesCommand.INSTANCE));
        stateMenu.add(new DynamicMenuItem(DecrementUsesCommand.INSTANCE));
        stateMenu.addSeparator();
        stateMenu.add(new DynamicMenuItem(SkillLevelIncrementCommand.INSTANCE));
        stateMenu.add(new DynamicMenuItem(SkillLevelDecrementCommand.INSTANCE));
        stateMenu.addSeparator();
        stateMenu.add(new DynamicMenuItem(TechLevelIncrementCommand.INSTANCE));
        stateMenu.add(new DynamicMenuItem(TechLevelDecrementCommand.INSTANCE));
        stateMenu.addSeparator();
        stateMenu.add(new DynamicMenuItem(SwapDefaultsCommand.INSTANCE));
        menu.add(stateMenu);
        menu.addSeparator();
        menu.add(new DynamicMenuItem(JumpToSearchCommand.INSTANCE));
        DynamicMenuEnabler.add(stateMenu);
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
