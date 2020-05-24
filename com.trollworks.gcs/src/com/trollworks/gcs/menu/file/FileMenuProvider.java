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
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JMenu;

/** Provides the standard "File" menu. */
public class FileMenuProvider {
    private static List<Command> MASTER_LIBRARY_EXPORT_TEMPLATE_CMDS;
    private static List<Command> USER_LIBRARY_EXPORT_TEMPLATE_CMDS;

    public static synchronized List<Command> getMasterLibraryExportTemplateCommands() {
        if (MASTER_LIBRARY_EXPORT_TEMPLATE_CMDS == null) {
            MASTER_LIBRARY_EXPORT_TEMPLATE_CMDS = generateLibraryExportTemplateCommands(true);
        }
        return MASTER_LIBRARY_EXPORT_TEMPLATE_CMDS;
    }

    public static synchronized List<Command> getUserLibraryExportTemplateCommands() {
        if (USER_LIBRARY_EXPORT_TEMPLATE_CMDS == null) {
            USER_LIBRARY_EXPORT_TEMPLATE_CMDS = generateLibraryExportTemplateCommands(false);
        }
        return USER_LIBRARY_EXPORT_TEMPLATE_CMDS;
    }

    private static List<Command> generateLibraryExportTemplateCommands(boolean master) {
        List<Command> cmds = new ArrayList<>();
        Path          dir  = (master ? Library.getMasterRootPath() : Library.getUserRootPath()).resolve("Output Templates");
        if (Files.isDirectory(dir)) {
            // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
            // directory results in leaving state around that prevents future move & delete
            // operations. Only use this style of access for directory listings to avoid that.
            try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                for (Path path : stream) {
                    cmds.add(new ExportToTextTemplateCommand(path, master));
                }
            } catch (IOException exception) {
                Log.error(exception);
            }
            cmds.sort((c1, c2) -> NumericComparator.caselessCompareStrings(PathUtils.getLeafName(c1.getTitle(), true), PathUtils.getLeafName(c2.getTitle(), true)));
        }
        return cmds;
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
        cmds.add(ExportToGURPSCalculatorCommand.INSTANCE);
        cmds.add(ExportToPDFCommand.INSTANCE);
        cmds.add(ExportToPNGCommand.INSTANCE);
        cmds.addAll(getMasterLibraryExportTemplateCommands());
        cmds.addAll(getUserLibraryExportTemplateCommands());
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
        exportMenu.add(new DynamicMenuItem(ExportToGURPSCalculatorCommand.INSTANCE));
        exportMenu.add(new DynamicMenuItem(ExportToPDFCommand.INSTANCE));
        exportMenu.add(new DynamicMenuItem(ExportToPNGCommand.INSTANCE));
        boolean needSep = true;
        for (Command cmd : getMasterLibraryExportTemplateCommands()) {
            if (needSep) {
                exportMenu.addSeparator();
                needSep = false;
            }
            exportMenu.add(new DynamicMenuItem(cmd));
        }
        needSep = true;
        for (Command cmd : getUserLibraryExportTemplateCommands()) {
            if (needSep) {
                exportMenu.addSeparator();
                needSep = false;
            }
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
