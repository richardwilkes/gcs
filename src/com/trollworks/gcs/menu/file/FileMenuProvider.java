/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.file;

import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.DynamicMenuItem;
import com.trollworks.toolkit.ui.menu.MenuProvider;
import com.trollworks.toolkit.ui.menu.file.CloseCommand;
import com.trollworks.toolkit.ui.menu.file.ExportToCommand;
import com.trollworks.toolkit.ui.menu.file.OpenCommand;
import com.trollworks.toolkit.ui.menu.file.PageSetupCommand;
import com.trollworks.toolkit.ui.menu.file.PrintCommand;
import com.trollworks.toolkit.ui.menu.file.QuitCommand;
import com.trollworks.toolkit.ui.menu.file.RecentFilesMenu;
import com.trollworks.toolkit.ui.menu.file.SaveAsCommand;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.Platform;

import java.util.HashSet;
import java.util.Set;

import javax.swing.JMenu;

/** Provides the standard "File" menu. */
public class FileMenuProvider implements MenuProvider {
	@Localize("File")
	@Localize(locale = "de", value = "Datei")
	@Localize(locale = "ru", value = "Файл")
	private static String		FILE;

	static {
		Localization.initialize();
	}

	public static final String	NAME	= "File";	//$NON-NLS-1$

	@Override
	public Set<Command> getModifiableCommands() {
		Set<Command> cmds = new HashSet<>();
		cmds.add(NewCharacterSheetCommand.INSTANCE);
		cmds.add(NewCharacterTemplateCommand.INSTANCE);
		cmds.add(NewAdvantagesLibraryCommand.INSTANCE);
		cmds.add(NewEquipmentLibraryCommand.INSTANCE);
		cmds.add(NewSkillsLibraryCommand.INSTANCE);
		cmds.add(NewSpellsLibraryCommand.INSTANCE);
		cmds.add(OpenCommand.INSTANCE);
		cmds.add(CloseCommand.INSTANCE);
		cmds.add(SaveCommand.INSTANCE);
		cmds.add(SaveAsCommand.INSTANCE);
		cmds.add(ExportToCommand.EXPORT_TO_HTML);
		cmds.add(ExportToCommand.EXPORT_TO_PDF);
		cmds.add(ExportToCommand.EXPORT_TO_PNG);
		cmds.add(PageSetupCommand.INSTANCE);
		cmds.add(PrintCommand.INSTANCE);
		if (!Platform.isMacintosh()) {
			cmds.add(QuitCommand.INSTANCE);
		}
		return cmds;
	}

	@Override
	public JMenu createMenu() {
		JMenu menu = new JMenu(FILE);
		menu.setName(NAME);
		menu.add(new DynamicMenuItem(NewCharacterSheetCommand.INSTANCE));
		menu.add(new DynamicMenuItem(NewCharacterTemplateCommand.INSTANCE));
		menu.add(new DynamicMenuItem(NewAdvantagesLibraryCommand.INSTANCE));
		menu.add(new DynamicMenuItem(NewEquipmentLibraryCommand.INSTANCE));
		menu.add(new DynamicMenuItem(NewSkillsLibraryCommand.INSTANCE));
		menu.add(new DynamicMenuItem(NewSpellsLibraryCommand.INSTANCE));
		menu.addSeparator();
		menu.add(new DynamicMenuItem(OpenCommand.INSTANCE));
		menu.add(new RecentFilesMenu());
		menu.add(new DynamicMenuItem(CloseCommand.INSTANCE));
		menu.addSeparator();
		menu.add(new DynamicMenuItem(SaveCommand.INSTANCE));
		menu.add(new DynamicMenuItem(SaveAsCommand.INSTANCE));
		menu.addSeparator();
		menu.add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_HTML));
		menu.add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_PDF));
		menu.add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_PNG));
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
