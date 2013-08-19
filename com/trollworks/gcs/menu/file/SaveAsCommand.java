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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.widgets.StdFileDialog;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;

import javax.swing.JMenuItem;

/** Provides the "Save As..." command. */
public class SaveAsCommand extends Command {
	private static String				MSG_SAVE_AS;

	static {
		LocalizedMessages.initialize(SaveAsCommand.class);
	}

	/** The singleton {@link SaveAsCommand}. */
	public static final SaveAsCommand	INSTANCE	= new SaveAsCommand();

	private SaveAsCommand() {
		super(MSG_SAVE_AS, KeyEvent.VK_S, SHIFTED_COMMAND_MODIFIER);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		setEnabled(getActiveWindow() instanceof Saveable);
	}

	@Override public void actionPerformed(ActionEvent event) {
		saveAs((Saveable) getActiveWindow());
	}

	/**
	 * Allows the user to save the file under another name.
	 * 
	 * @param saveable The {@link Saveable} to work on.
	 * @return The file(s) actually written to.
	 */
	public File[] saveAs(Saveable saveable) {
		String path = saveable.getPreferredSavePath();
		File result = StdFileDialog.choose(saveable instanceof Component ? (Component) saveable : null, false, MSG_SAVE_AS, Path.getParent(path), Path.getLeafName(path), saveable.getAllowedExtensions());
		File[] files = result != null ? saveable.saveTo(result) : new File[0];
		for (File file : files) {
			RecentFilesMenu.addRecent(file);
		}
		return files;
	}
}
