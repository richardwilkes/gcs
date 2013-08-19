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

import java.awt.Window;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;

import javax.swing.JMenuItem;

/** Provides the "Save" command. */
public class SaveCommand extends Command {
	private static String			MSG_SAVE;

	static {
		LocalizedMessages.initialize(SaveCommand.class);
	}

	/** The singleton {@link SaveCommand}. */
	public static final SaveCommand	INSTANCE	= new SaveCommand();

	private SaveCommand() {
		super(MSG_SAVE, KeyEvent.VK_S);
	}

	@Override public void adjustForMenu(JMenuItem item) {
		Window window = getActiveWindow();
		if (window instanceof Saveable) {
			setEnabled(((Saveable) window).isModified());
		} else {
			setEnabled(false);
		}
	}

	@Override public void actionPerformed(ActionEvent event) {
		save((Saveable) getActiveWindow());
	}

	/**
	 * Allows the user to save the file.
	 * 
	 * @param saveable The {@link Saveable} to work on.
	 * @return The file(s) actually written to.
	 */
	public File[] save(Saveable saveable) {
		File file = saveable.getBackingFile();
		if (file != null) {
			File[] files = saveable.saveTo(file);
			for (File one : files) {
				RecentFilesMenu.addRecent(one);
			}
			return files;

		}
		return SaveAsCommand.INSTANCE.saveAs(saveable);
	}
}
