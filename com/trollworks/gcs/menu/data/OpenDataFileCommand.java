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

package com.trollworks.gcs.menu.data;

import com.trollworks.gcs.app.App;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.menu.file.OpenCommand;

import java.awt.event.ActionEvent;
import java.io.File;

import javax.swing.ImageIcon;
import javax.swing.JMenuItem;

/** A command that will open a specific data file. */
public class OpenDataFileCommand extends Command implements Runnable {
	private File	mFile;
	private boolean	mVerify;

	/**
	 * Creates a new {@link OpenDataFileCommand}.
	 * 
	 * @param title The title to use.
	 * @param file The file to open.
	 */
	public OpenDataFileCommand(String title, File file) {
		super(title, new ImageIcon(App.getIconForFile(file)));
		mFile = file;
	}

	/**
	 * Creates a new {@link OpenDataFileCommand} that can only be invoked successfully if
	 * {@link OpenCommand} is enabled.
	 * 
	 * @param file The file to open.
	 */
	public OpenDataFileCommand(File file) {
		super(file.getName());
		mFile = file;
		mVerify = true;
	}

	@Override public void adjustForMenu(JMenuItem item) {
		// Not used. Always enabled.
	}

	@Override public void actionPerformed(ActionEvent event) {
		run();
	}

	public void run() {
		if (mVerify) {
			OpenCommand.INSTANCE.adjustForMenu(null);
			if (!OpenCommand.INSTANCE.isEnabled()) {
				return;
			}
		}
		OpenCommand.INSTANCE.open(mFile);
	}
}
