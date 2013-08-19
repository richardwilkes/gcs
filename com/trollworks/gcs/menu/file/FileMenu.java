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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.menu.DynamicMenuEnabler;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.io.LocalizedMessages;

import javax.swing.JMenu;
import javax.swing.JMenuItem;

/** The standard "File" menu. */
public class FileMenu extends JMenu {
	private static String	MSG_FILE;

	static {
		LocalizedMessages.initialize(FileMenu.class);
	}

	/** Creates a new {@link FileMenu}. */
	public FileMenu() {
		super(MSG_FILE);
		add(new JMenuItem(NewCharacterSheetCommand.INSTANCE));
		add(new JMenuItem(NewCharacterTemplateCommand.INSTANCE));
		add(new JMenuItem(NewListCommand.ADVANTAGES));
		add(new JMenuItem(NewListCommand.SKILLS));
		add(new JMenuItem(NewListCommand.SPELLS));
		add(new JMenuItem(NewListCommand.EQUIPMENT));
		addSeparator();
		add(new JMenuItem(OpenCommand.INSTANCE));
		add(new RecentFilesMenu());
		add(new JMenuItem(CloseCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(SaveCommand.INSTANCE));
		add(new JMenuItem(SaveAsCommand.INSTANCE));
		addSeparator();
		add(new JMenuItem(ExportToCommand.EXPORT_TO_HTML));
		add(new JMenuItem(ExportToCommand.EXPORT_TO_PDF));
		add(new JMenuItem(ExportToCommand.EXPORT_TO_PNG));
		addSeparator();
		add(new JMenuItem(PageSetupCommand.INSTANCE));
		add(new JMenuItem(PrintCommand.INSTANCE));
		if (!Platform.isMacintosh()) {
			addSeparator();
			add(new JMenuItem(QuitCommand.INSTANCE));
		}
		DynamicMenuEnabler.add(this);
	}
}
