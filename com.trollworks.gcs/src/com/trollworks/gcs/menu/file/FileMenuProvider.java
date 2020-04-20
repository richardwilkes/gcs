/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Platform;

import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

/** Provides the standard "File" menu. */
public class FileMenuProvider {
    public static List<Command> getModifiableCommands() {
        List<Command> cmds = new ArrayList<>();
        cmds.add(NewCharacterSheetCommand.INSTANCE);
        cmds.add(NewCharacterTemplateCommand.INSTANCE);
        cmds.add(NewAdvantagesLibraryCommand.INSTANCE);
        cmds.add(NewEquipmentLibraryCommand.INSTANCE);
        cmds.add(NewNoteLibraryCommand.INSTANCE);
        cmds.add(NewSkillsLibraryCommand.INSTANCE);
        cmds.add(NewSpellsLibraryCommand.INSTANCE);
        cmds.add(OpenCommand.INSTANCE);
        cmds.add(CloseCommand.INSTANCE);
        cmds.add(SaveCommand.INSTANCE);
        cmds.add(SaveAsCommand.INSTANCE);
        cmds.add(ExportToGurpsCalculatorCommand.INSTANCE);
        cmds.add(PageSetupCommand.INSTANCE);
        cmds.add(PrintCommand.INSTANCE);
        if (!Platform.isMacintosh()) {
            cmds.add(QuitCommand.INSTANCE);
        }
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu = new JMenu(I18n.Text("File"));
        menu.add(new DynamicMenuItem(NewCharacterSheetCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewCharacterTemplateCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewAdvantagesLibraryCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewEquipmentLibraryCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewNoteLibraryCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewSkillsLibraryCommand.INSTANCE));
        menu.add(new DynamicMenuItem(NewSpellsLibraryCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(OpenCommand.INSTANCE));
        menu.add(new RecentFilesMenu());
        menu.add(new DynamicMenuItem(CloseCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(SaveCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SaveAsCommand.INSTANCE));
        menu.add(new DynamicMenuItem(ExportToGurpsCalculatorCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(PageSetupCommand.INSTANCE));
        menu.add(new DynamicMenuItem(PrintCommand.INSTANCE));
        if (!Platform.isMacintosh()) {
            menu.addSeparator();
            menu.add(new DynamicMenuItem(QuitCommand.INSTANCE));
        }
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
