/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2008 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.file;

import com.trollworks.ttk.menu.Command;
import com.trollworks.ttk.menu.DynamicMenuEnabler;
import com.trollworks.ttk.menu.DynamicMenuItem;
import com.trollworks.ttk.menu.file.CloseCommand;
import com.trollworks.ttk.menu.file.ExportToCommand;
import com.trollworks.ttk.menu.file.OpenCommand;
import com.trollworks.ttk.menu.file.PageSetupCommand;
import com.trollworks.ttk.menu.file.PrintCommand;
import com.trollworks.ttk.menu.file.QuitCommand;
import com.trollworks.ttk.menu.file.RecentFilesMenu;
import com.trollworks.ttk.menu.file.SaveAsCommand;
import com.trollworks.ttk.menu.file.SaveCommand;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.utility.Platform;

import java.util.HashSet;

import javax.swing.JMenu;

/** The standard "File" menu. */
public class FileMenu extends JMenu {
	private static String	MSG_FILE;

	static {
		LocalizedMessages.initialize(FileMenu.class);
	}

	/**
	 * @return The set of {@link Command}s that this menu provides that can have their accelerators
	 *         modified.
	 */
	public static HashSet<Command> getCommands() {
		HashSet<Command> cmds = new HashSet<Command>();
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
		super(MSG_FILE);
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
