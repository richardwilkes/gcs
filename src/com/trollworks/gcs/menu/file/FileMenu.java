/*
 * Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.toolkit.utility.Localization;


import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.menu.DynamicMenuEnabler;
import com.trollworks.toolkit.ui.menu.DynamicMenuItem;
import com.trollworks.toolkit.ui.menu.file.CloseCommand;
import com.trollworks.toolkit.ui.menu.file.ExportToCommand;
import com.trollworks.toolkit.ui.menu.file.OpenCommand;
import com.trollworks.toolkit.ui.menu.file.PageSetupCommand;
import com.trollworks.toolkit.ui.menu.file.PrintCommand;
import com.trollworks.toolkit.ui.menu.file.QuitCommand;
import com.trollworks.toolkit.ui.menu.file.RecentFilesMenu;
import com.trollworks.toolkit.ui.menu.file.SaveAsCommand;
import com.trollworks.toolkit.ui.menu.file.SaveCommand;
import com.trollworks.toolkit.utility.Platform;

import java.util.HashSet;

import javax.swing.JMenu;

/** The standard "File" menu. */
public class FileMenu extends JMenu {
	@Localize("File")
	private static String FILE;

	static {
		Localization.initialize();
	}

	/**
	 * @return The set of {@link Command}s that this menu provides that can have their accelerators
	 *         modified.
	 */
	public static HashSet<Command> getCommands() {
		HashSet<Command> cmds = new HashSet<>();
		cmds.add(NewCharacterSheetCommand.INSTANCE);
		cmds.add(NewCharacterTemplateCommand.INSTANCE);
		cmds.add(NewLibraryCommand.INSTANCE);
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

	/** Creates a new {@link FileMenu}. */
	public FileMenu() {
		super(FILE);
		add(new DynamicMenuItem(NewCharacterSheetCommand.INSTANCE));
		add(new DynamicMenuItem(NewCharacterTemplateCommand.INSTANCE));
		add(new DynamicMenuItem(NewLibraryCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(OpenCommand.INSTANCE));
		add(new RecentFilesMenu());
		add(new DynamicMenuItem(CloseCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(SaveCommand.INSTANCE));
		add(new DynamicMenuItem(SaveAsCommand.INSTANCE));
		addSeparator();
		add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_HTML));
		add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_PDF));
		add(new DynamicMenuItem(ExportToCommand.EXPORT_TO_PNG));
		addSeparator();
		add(new DynamicMenuItem(PageSetupCommand.INSTANCE));
		add(new DynamicMenuItem(PrintCommand.INSTANCE));
		if (!Platform.isMacintosh()) {
			addSeparator();
			add(new DynamicMenuItem(QuitCommand.INSTANCE));
		}
		DynamicMenuEnabler.add(this);
	}
}
