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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.menu.DynamicMenuItem;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

/** Provides the standard "File" menu. */
public class FileMenuProvider {
    private static List<Command> LIBRARY_EXPORT_TEMPLATE_CMDS;

    public static synchronized List<Command> getLibraryExportTemplateCommands() {
        if (LIBRARY_EXPORT_TEMPLATE_CMDS == null) {
            LIBRARY_EXPORT_TEMPLATE_CMDS = new ArrayList<>();
            Path dir = Library.getMasterRootPath().resolve("Output Templates");
            if (Files.isDirectory(dir)) {
                try {
                    Files.list(dir).
                            sorted((Path p1, Path p2) -> NumericComparator.caselessCompareStrings(PathUtils.getLeafName(p1, true), PathUtils.getLeafName(p2, true))).
                            forEachOrdered((path) -> LIBRARY_EXPORT_TEMPLATE_CMDS.add(new ExportToTextTemplateCommand(path)));
                } catch (IOException exception) {
                    Log.error(exception);
                }
            }
        }
        return LIBRARY_EXPORT_TEMPLATE_CMDS;
    }

    public static List<Command> getModifiableCommands() {
        List<Command> cmds = new ArrayList<>();
        cmds.add(NewCharacterSheetCommand.INSTANCE);
        cmds.add(NewCharacterTemplateCommand.INSTANCE);
        cmds.add(NewAdvantagesLibraryCommand.INSTANCE);
        cmds.add(NewAdvantageModifiersLibraryCommand.INSTANCE);
        cmds.add(NewEquipmentLibraryCommand.INSTANCE);
        cmds.add(NewEquipmentModifiersLibraryCommand.INSTANCE);
        cmds.add(NewNoteLibraryCommand.INSTANCE);
        cmds.add(NewSkillsLibraryCommand.INSTANCE);
        cmds.add(NewSpellsLibraryCommand.INSTANCE);
        cmds.add(OpenCommand.INSTANCE);
        cmds.add(CloseCommand.INSTANCE);
        cmds.add(SaveCommand.INSTANCE);
        cmds.add(SaveAsCommand.INSTANCE);
        cmds.add(ExportToGurpsCalculatorCommand.INSTANCE);
        cmds.add(ExportToPDFCommand.INSTANCE);
        cmds.add(ExportToPNGCommand.INSTANCE);
        cmds.addAll(getLibraryExportTemplateCommands());
        cmds.add(PageSetupCommand.INSTANCE);
        cmds.add(PrintCommand.INSTANCE);
        if (!Platform.isMacintosh()) {
            cmds.add(QuitCommand.INSTANCE);
        }
        return cmds;
    }

    public static JMenu createMenu() {
        JMenu menu    = new JMenu(I18n.Text("File"));
        JMenu newMenu = new JMenu(I18n.Text("New File…"));
        newMenu.add(new DynamicMenuItem(NewCharacterSheetCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewCharacterTemplateCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewAdvantagesLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewAdvantageModifiersLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewEquipmentLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewEquipmentModifiersLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewNoteLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewSkillsLibraryCommand.INSTANCE));
        newMenu.add(new DynamicMenuItem(NewSpellsLibraryCommand.INSTANCE));
        menu.add(newMenu);
        menu.add(new DynamicMenuItem(OpenCommand.INSTANCE));
        menu.add(new RecentFilesMenu());
        menu.add(new DynamicMenuItem(CloseCommand.INSTANCE));
        menu.addSeparator();
        menu.add(new DynamicMenuItem(SaveCommand.INSTANCE));
        menu.add(new DynamicMenuItem(SaveAsCommand.INSTANCE));
        JMenu exportMenu = new JMenu(I18n.Text("Export To…"));
        exportMenu.add(new DynamicMenuItem(ExportToGurpsCalculatorCommand.INSTANCE));
        exportMenu.add(new DynamicMenuItem(ExportToPDFCommand.INSTANCE));
        exportMenu.add(new DynamicMenuItem(ExportToPNGCommand.INSTANCE));
        exportMenu.addSeparator();
        for (Command cmd : getLibraryExportTemplateCommands()) {
            exportMenu.add(new DynamicMenuItem(cmd));
        }
        menu.add(exportMenu);
        menu.addSeparator();
        menu.add(new DynamicMenuItem(PageSetupCommand.INSTANCE));
        menu.add(new DynamicMenuItem(PrintCommand.INSTANCE));
        if (!Platform.isMacintosh()) {
            menu.addSeparator();
            menu.add(new DynamicMenuItem(QuitCommand.INSTANCE));
        }
        DynamicMenuEnabler.add(newMenu);
        DynamicMenuEnabler.add(exportMenu);
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
